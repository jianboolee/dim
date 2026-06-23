<!-- 文本消息组件 -->
<template>
  <div class="message-content message-text" :class="{ 'message-failed': message.status === 'failed' }">
    <div class="message-arrow"></div>
    <span v-html="formattedContent"></span>
    <MessageStatus 
      v-if="isMine" 
      :status="message.status" 
      @retry="$emit('retry')" 
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Message } from '@/types/im'
import MessageStatus from './MessageStatus.vue'

const props = defineProps<{
  message: Message
  isMine: boolean
}>()

defineEmits<{
  (e: 'retry'): void
}>()

// 链接正则表达式
const urlRegex = /(https?:\/\/[^\s<]+[^<.,:;"')\]\s])/g

// 格式化内容，将链接转换为可点击的形式
const formattedContent = computed(() => {
  if (!props.message.content) return ''
  
  // 转义HTML特殊字符
  const escapedContent = props.message.content
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
  
  // 替换链接为可点击的形式
  return escapedContent.replace(urlRegex, (url) => {
    return `<a href="${url}" target="_blank" rel="noopener noreferrer">${url}</a>`
  })
})
</script>

<style scoped>
.message-content.message-text {
  max-width: 80%;
  white-space: pre-line;
  position: relative;
  padding: 8px 12px;
  padding-right: 24px;
  border-radius: 12px;
  background: var(--bg-color);
  word-break: break-all;
  font-size: 15px;
  line-height: 1.75;
  font-weight: 400;
  color: var(--text-color-dark);
}

.message-arrow {
  position: absolute;
  top: 12px;
  left: -6px;
  width: 0;
  height: 0;
  border-style: solid;
  border-width: 6px 6px 6px 0;
  border-color: transparent var(--bg-color) transparent transparent;
}

/* 链接基础样式 */
.message-content :deep(a) {
  color: var(--primary-color);
  text-decoration: underline;
  word-break: break-all;
  transition: opacity 0.2s;
}

.message-content :deep(a:hover) {
  opacity: 0.8;
}

/* 我发送的消息样式 */
.message-mine .message-content {
  background: #4b86f8;
  background: linear-gradient(45deg, #4b86f8 0%, #427be8 100%);
  color: #fff;
}

.message-mine .message-arrow {
  left: auto;
  right: -6px;
  border-width: 6px 0 6px 6px;
  border-color: transparent transparent transparent #427be8;
}

/* 我发送的消息中的链接样式 */
.message-mine .message-content :deep(a) {
  color: #fff;
}

/* 发送失败的消息样式 */
.message-failed {
  background: var(--bg-color-light) !important;
  color: var(--text-color-light) !important;
}

:deep(.message-mine) .message-failed {
  background: var(--error-color-light) !important;
}
</style> 
