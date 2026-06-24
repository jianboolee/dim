import type { Message, MediaInfo } from '@/sdk/im'

export type {
  Message,
  MessageType,
  MessageStatus,
  MediaInfo,
  CardInfo,
  LinkInfo,
  Conversation,
  ConversationPage,
} from '@/sdk/im'

/** 聊天 UI 层扩展的消息状态 */
export type ChatMessageStatus = 'sending' | 'failed' | 'sent' | 'delivered' | 'read'

export type ChatMessage = Omit<Message, 'status' | 'media_info'> & {
  status?: ChatMessageStatus
  media_info?: MediaInfo & { uploading?: boolean }
}
