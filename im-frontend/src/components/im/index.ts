import type { Component } from 'vue'
import { MessageType } from '@/sdk/im'
import TextMessage from './TextMessage.vue'
import ImageMessage from './ImageMessage.vue'
import AudioMessage from './AudioMessage.vue'
import VideoMessage from './VideoMessage.vue'
import CardMessage from './CardMessage.vue'
import MessageStatus from './MessageStatus.vue'
import LinkMessage from './LinkMessage.vue'
import SystemEventMessage from './SystemEventMessage.vue'

export {
  TextMessage,
  ImageMessage,
  AudioMessage,
  VideoMessage,
  CardMessage,
  MessageStatus,
  LinkMessage,
  SystemEventMessage,
}

export const MessageComponents: Partial<Record<MessageType, Component>> = {
  [MessageType.Text]: TextMessage,
  [MessageType.SystemEvent]: SystemEventMessage,
  [MessageType.Image]: ImageMessage,
  [MessageType.Audio]: AudioMessage,
  [MessageType.Video]: VideoMessage,
  [MessageType.Card]: CardMessage,
  [MessageType.Link]: LinkMessage,
}
