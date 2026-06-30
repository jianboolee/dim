import type { ChatMessage } from '@/types/im'

export function formatSystemEventMessage(message: ChatMessage): string {
  return message.preview_text || message.content || '群聊状态已更新'
}
