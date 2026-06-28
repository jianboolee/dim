import type { UserInfo } from '@/sdk/im'
import type { ChatMessage } from '@/types/im'

type UserMap = Record<string, UserInfo | undefined>

function displayUserName(userId: string, users: UserMap): string {
  if (!userId) return ''
  return users[userId]?.nickname || userId
}

function displayUserNames(userIds: string[] | undefined, users: UserMap): string {
  const names = (userIds ?? [])
    .filter(Boolean)
    .map((userId) => displayUserName(userId, users))

  return names.join('、')
}

export function collectSystemEventUserIds(message: ChatMessage): string[] {
  const ids = new Set<string>()
  if (message.from_id) ids.add(message.from_id)
  if (message.payload?.operator_id) ids.add(message.payload.operator_id)
  for (const userId of message.payload?.target_user_ids ?? []) {
    if (userId) ids.add(userId)
  }
  return [...ids]
}

export function formatSystemEventMessage(message: ChatMessage, users: UserMap): string {
  const payload = message.payload
  const operatorName = displayUserName(payload?.operator_id || message.from_id || '', users)
  const targetNames = displayUserNames(payload?.target_user_ids, users)
  const actor = operatorName || targetNames

  switch (payload?.event_type) {
    case 'group_created':
      return actor ? `${actor}创建了群聊` : message.content
    case 'member_joined':
      return targetNames ? `${targetNames}加入了群聊` : message.content
    case 'member_kicked':
      return targetNames ? `${targetNames}被移出群聊` : message.content
    case 'member_left':
      return actor ? `${actor}退出了群聊` : message.content
    case 'group_dissolved':
      return '群聊已解散'
    case 'group_name_updated':
      return payload.after_value ? `群名修改为 ${payload.after_value}` : message.content
    case 'group_avatar_updated':
      return actor ? `${actor}修改了群头像` : message.content
    case 'admin_added':
      return targetNames ? `${targetNames}被设置为管理员` : message.content
    case 'admin_removed':
      return targetNames ? `${targetNames}被取消管理员` : message.content
    default:
      return message.content || '群聊状态已更新'
  }
}
