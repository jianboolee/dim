import dayjs from 'dayjs'
import { MessageType, type Message } from '@/sdk/im'

const EMPTY_TIME = '0001-01-01T00:00:00Z'

const LAST_MESSAGE_LABELS: Partial<Record<MessageType, string>> = {
  [MessageType.Image]: '[图片]',
  [MessageType.Video]: '[视频]',
  [MessageType.Audio]: '[语音]',
  [MessageType.Link]: '[链接]',
  [MessageType.Card]: '[卡片]',
}

/** 会话列表时间展示：今天显示时分，昨天显示「昨天」，同年显示月日 */
export function formatConversationTime(time?: string): string {
  if (!time || time === EMPTY_TIME) return ''

  const date = dayjs(time)
  const now = dayjs()

  if (date.isSame(now, 'day')) return date.format('HH:mm')
  if (date.isSame(now.subtract(1, 'day'), 'day')) return '昨天'
  if (date.isSame(now, 'year')) return date.format('MM-DD')
  return date.format('YYYY-MM-DD')
}

/** 会话列表最后一条消息预览文案 */
export function formatLastMessagePreview(message?: Message): string {
  if (!message) return ''

  if (message.type && message.type !== MessageType.Text) {
    return LAST_MESSAGE_LABELS[message.type] ?? `[${message.type}]`
  }

  return message.content || ''
}

export function formatUnreadBadge(count: number): string | number {
  return count > 99 ? '99+' : count
}
