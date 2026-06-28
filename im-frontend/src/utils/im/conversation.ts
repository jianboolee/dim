import type { Conversation, Message, LastMessageSnapshot } from '@/sdk/im'
import { normalizeUnreadCount } from '@/utils/im/format'

function toSnapshot(message: Message): LastMessageSnapshot {
  return {
    content: message.content,
    type: message.type,
    created_at: message.created_at ?? new Date().toISOString(),
  }
}

export function getPeerUserId(conversation: Conversation, currentUserId: string): string {
  return conversation.participants.find((id) => id !== currentUserId) ?? ''
}

export function isGroupConversation(conversation?: Conversation | null): boolean {
  return conversation?.type === 'group'
}

export function getConversationDisplayName(conversation: Conversation, currentUserId: string): string {
  if (conversation.display_name) return conversation.display_name
  if (isGroupConversation(conversation)) return conversation.group_info?.name || '群聊'

  const peerId = getPeerUserId(conversation, currentUserId)
  return conversation.peer_user_info?.nickname
    || conversation.to_user_info?.nickname
    || (peerId ? `用户${peerId.slice(-4)}` : '未知用户')
}

export function getConversationDisplayAvatar(conversation: Conversation): string {
  return conversation.display_avatar
    || conversation.group_info?.avatar_url
    || conversation.peer_user_info?.avatar
    || conversation.to_user_info?.avatar
    || ''
}

export function getUnreadCount(conversation: Conversation, currentUserId: string): number {
  if (!currentUserId) return 0
  return normalizeUnreadCount(
    conversation.member_state?.unread_count
      ?? conversation.user_states?.[currentUserId]?.unread_count
      ?? 0,
  )
}

export function sortConversationsByActivity(conversations: Conversation[]): Conversation[] {
  return [...conversations].sort((a, b) => {
    const timeA = a.last_activity || a.last_message?.created_at || a.updated_at
    const timeB = b.last_activity || b.last_message?.created_at || b.updated_at
    return new Date(timeB).getTime() - new Date(timeA).getTime()
  })
}

function matchesMessage(conversation: Conversation, message: Message): boolean {
  if (message.conversation_id && conversation.id === message.conversation_id) return true
  if (isGroupConversation(conversation)) return false

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
    type: message.to_id ? 'private' : 'group',
    participants: message.to_id ? [fromId, message.to_id] : [fromId, currentUserId],
    display_name: message.to_id ? undefined : '群聊',
    last_message: toSnapshot(message),
    image_url: '',
    user_states: message.to_id === currentUserId
      ? { [currentUserId]: { unread_count: 1 } }
      : {},
    member_state: {
      status: 'active',
      last_read_seq: 0,
      unread_count: message.to_id === currentUserId ? 1 : 0,
    },
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
  const shouldIncrementUnread =
    !isGroupConversation(existing) &&
    message.to_id === currentUserId && existing.id !== activeConversationId

  const updated: Conversation = {
    ...existing,
    last_message: toSnapshot(message),
    updated_at: message.created_at ?? existing.updated_at,
    last_activity: maxTime(message.created_at, existing.last_activity, existing.updated_at),
    user_states: shouldIncrementUnread
      ? {
          ...existing.user_states,
          [currentUserId]: {
            ...existing.user_states?.[currentUserId],
            unread_count: getUnreadCount(existing, currentUserId) + 1,
          },
        }
      : existing.user_states,
    member_state: shouldIncrementUnread
      ? {
          ...existing.member_state,
          status: existing.member_state?.status ?? 'active',
          last_read_seq: existing.member_state?.last_read_seq ?? 0,
          unread_count: getUnreadCount(existing, currentUserId) + 1,
        }
      : existing.member_state,
  }

  const next = [...conversations]
  next[index] = updated
  return sortConversationsByActivity(next)
}

/** 清除指定会话的本地未读数（进入聊天或当前会话收到消息后同步侧栏） */
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
      member_state: {
        ...conversation.member_state,
        status: conversation.member_state?.status ?? 'active',
        last_read_seq: conversation.member_state?.last_read_seq ?? 0,
        unread_count: 0,
      },
    }
  })
}

export function withClearedUnreadForConversation(
  conversations: Conversation[],
  conversationId: string,
): Conversation[] {
  if (!conversationId) return conversations

  return conversations.map((conversation) => {
    if (conversation.id !== conversationId) {
      return conversation
    }

    return {
      ...conversation,
      member_state: {
        ...conversation.member_state,
        status: conversation.member_state?.status ?? 'active',
        last_read_seq: conversation.member_state?.last_read_seq ?? 0,
        unread_count: 0,
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
