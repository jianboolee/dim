import { MessageType, type Conversation, type Message } from '@/sdk/im'
import { normalizeUnreadCount } from '@/utils/im/format'

export function getPeerUserId(conversation: Conversation, currentUserId: string): string {
  return conversation.participants.find((id) => id !== currentUserId) ?? ''
}

export function getUnreadCount(conversation: Conversation, currentUserId: string): number {
  if (!currentUserId) return 0
  return normalizeUnreadCount(conversation.user_states?.[currentUserId]?.unread_count ?? 0)
}

export function sortConversationsByActivity(conversations: Conversation[]): Conversation[] {
  return [...conversations].sort((a, b) => {
    const timeA = a.last_activity || a.last_message?.created_at || a.updated_at
    const timeB = b.last_activity || b.last_message?.created_at || b.updated_at
    return new Date(timeB).getTime() - new Date(timeA).getTime()
  })
}

function previewImageFromMessage(message: Message): string {
  if (message.type === MessageType.Card) return message.card_info?.image_url ?? ''
  if (message.type === MessageType.Image) return message.media_info?.url ?? ''
  return ''
}

function matchesMessage(conversation: Conversation, message: Message): boolean {
  if (message.conversation_id && conversation.id === message.conversation_id) return true

  const fromId = message.from_id
  if (!fromId) return false
  return conversation.participants.includes(fromId) && conversation.participants.includes(message.to_id)
}

function maxTime(...values: Array<string | undefined>): string {
  const timestamps = values
    .map((value) => (value ? new Date(value).getTime() : 0))
    .filter((value) => Number.isFinite(value) && value > 0)
  const max = Math.max(...timestamps)
  return max > 0 ? new Date(max).toISOString() : new Date().toISOString()
}

export function buildConversationFromMessage(message: Message, currentUserId: string): Conversation {
  const fromId = message.from_id ?? ''
  const timestamp = message.created_at ?? new Date().toISOString()

  return {
    id: message.conversation_id ?? [fromId, message.to_id].sort().join(':'),
    type: 'private',
    participants: [fromId, message.to_id],
    last_message: message,
    image_url: previewImageFromMessage(message),
    user_states: message.to_id === currentUserId
      ? { [currentUserId]: { unread_count: 1 } }
      : {},
    created_at: timestamp,
    updated_at: timestamp,
    last_activity: timestamp,
  }
}

/** 收到实时消息后，不可变地更新会话列表 */
export function applyIncomingMessage(
  conversations: Conversation[],
  message: Message,
  currentUserId: string,
  activeConversationId?: string,
): Conversation[] {
  const index = conversations.findIndex((conversation) => matchesMessage(conversation, message))

  if (index === -1) {
    const created = buildConversationFromMessage(message, currentUserId)
    if (activeConversationId && created.id === activeConversationId && message.to_id === currentUserId) {
      created.user_states = { [currentUserId]: { unread_count: 0 } }
    }
    return sortConversationsByActivity([created, ...conversations])
  }

  const existing = conversations[index]!
  const previewImage = previewImageFromMessage(message)
  const shouldIncrementUnread =
    message.to_id === currentUserId && existing.id !== activeConversationId

  const updated: Conversation = {
    ...existing,
    last_message: message,
    updated_at: message.created_at ?? existing.updated_at,
    last_activity: maxTime(message.created_at, existing.last_activity, existing.updated_at),
    image_url: previewImage || existing.image_url,
    user_states: shouldIncrementUnread
      ? {
          ...existing.user_states,
          [currentUserId]: {
            ...existing.user_states?.[currentUserId],
            unread_count: getUnreadCount(existing, currentUserId) + 1,
          },
        }
      : existing.user_states,
  }

  const next = [...conversations]
  next[index] = updated
  return sortConversationsByActivity(next)
}

/** 清除指定会话的未读数（进入聊天或标已读后同步侧栏） */
export function withClearedUnreadForPeer(
  conversations: Conversation[],
  peerId: string,
  currentUserId: string,
): Conversation[] {
  if (!peerId || !currentUserId) {
    return conversations
  }

  return conversations.map((conversation) => {
    if (getPeerUserId(conversation, currentUserId) !== peerId) {
      return conversation
    }

    return {
      ...conversation,
      user_states: {
        ...conversation.user_states,
        [currentUserId]: {
          ...conversation.user_states?.[currentUserId],
          unread_count: 0,
        },
      },
    }
  })
}

export function collectPeerUserIds(conversations: Conversation[], currentUserId: string): string[] {
  const ids = new Set<string>()

  for (const conversation of conversations) {
    const peerId = getPeerUserId(conversation, currentUserId)
    if (peerId) ids.add(peerId)
  }

  return [...ids]
}
