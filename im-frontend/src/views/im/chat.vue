<template>
  <div class="chat-page">
    <div class="nav-bar">
      <div class="nav-bar-content">
        <button class="back-btn" type="button" @click="router.back()">
          <i class="bi bi-chevron-left"></i>
        </button>
        <div class="user-info">
          <div class="user-avatar">
            <img
              v-if="targetUser?.avatar"
              :src="targetUser.avatar"
              alt=""
              @error="handleAvatarError"
            >
            <PlaceholderImage
              v-else
              bgColor="#EFF1F8"
              text=""
              aspect="1:1"
            />
          </div>
          <h1 class="title">{{ targetUser?.nickname || '未知用户' }}</h1>
        </div>
        <button class="btn" type="button">
          <i class="bi bi-three-dots"></i>
        </button>
      </div>
    </div>

    <div v-if="!isConnected" class="connection-status-bar">
      连接中，消息发送暂不可用…
    </div>

    <div class="message-container" @click="showMoreOptions = false">
      <div ref="messageListRef" class="message-list" @scroll="handleScroll">
        <div v-if="loading" class="loading-spinner">
          <div class="spinner"></div>
        </div>
        <div
          v-if="!hasMore && !firstLoad && messages.length > pageSize"
          class="no-more-messages"
        >
          没有更多消息了
        </div>
        <div
          v-for="msg in messages"
          :key="msg.id ?? msg.created_at"
          class="message-item"
          :class="{ 'message-mine': msg.from_id === currentUserId }"
        >
          <div class="message-avatar">
            <img
              :src="msg.from_id === currentUserId ? userStore.userInfo?.avatar || '' : targetUser?.avatar || ''"
              alt=""
            >
          </div>
          <div class="message-wrapper">
            <component
              :is="MessageComponents[msg.type || MessageType.Text] ?? MessageComponents[MessageType.Text]"
              :message="msg"
              :isMine="msg.from_id === currentUserId"
              @retry="retryMessage(msg)"
            />
          </div>
        </div>
      </div>
    </div>

    <div class="message-input-container">
      <div class="message-input">
        <div class="message-input-content">
          <MultilineInput
            v-model="messageText"
            placeholder="发消息..."
            :maxRows="4"
            @enter="sendMessage"
            @focus="handleMessageInputFocus"
          />
          <button
            type="button"
            class="message-input-button plus-btn"
            :disabled="messageText.trim().length > 0"
            @click="handlePlusClick"
          >
            <i v-if="!showMoreOptions" class="bi bi-plus-circle"></i>
            <i v-else class="bi bi-x-circle"></i>
          </button>
        </div>
        <button
          type="button"
          class="send-btn"
          :disabled="!messageText.trim() || !isConnected"
          @click="sendMessage"
        >
          <i class="bi bi-send-fill"></i>
        </button>
      </div>

      <MessageMoreOptions
        v-model="showMoreOptions"
        @select-file="handleSelectFile"
        @upload-success="handleUploadSuccess"
        @upload-error="handleUploadError"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, watch, provide, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { useUserStore } from '@/stores/user'
import { useIMStore } from '@/stores/im'
import { request } from '@/utils/request'
import { MessageType } from '@/sdk/im'
import type { MediaInfo } from '@/sdk/im'
import type { ChatMessage } from '@/types/im'
import type { UserInfo } from '@/types/user'
import { MessageComponents } from '@/components/im'
import MessageMoreOptions from '@/components/im/MessageMoreOptions.vue'
import MultilineInput from '@/components/im/MultilineInput.vue'
import PlaceholderImage from '@/components/common/PlaceholderImage.vue'

interface ApiResponse<T> {
  code: number
  data: T
}

const props = defineProps<{
  userId: string
}>()

const router = useRouter()
const userStore = useUserStore()
const imStore = useIMStore()

const messageText = ref('')
const messages = ref<ChatMessage[]>([])
const targetUser = ref<UserInfo | null>(null)
const messageListRef = ref<HTMLElement | null>(null)
const showMoreOptions = ref(false)

const isConnected = computed(() => imStore.isConnected)
const currentUserId = computed(() => userStore.userInfo?.id)
const peerUserId = computed(() => props.userId)

const pageSize = 20
const loading = ref(false)
const hasMore = ref(true)
const firstLoad = ref(true)
const initialized = ref(false)

const chatImages = computed(() =>
  messages.value
    .filter((msg) => msg.type === MessageType.Image && msg.media_info?.url)
    .map((msg) => msg.media_info!.url),
)

provide('chatImages', chatImages)

const isCurrentChatMessage = (message: ChatMessage) => {
  const myId = userStore.userInfo?.id
  if (!myId) return false

  return (
    (message.from_id === peerUserId.value && message.to_id === myId) ||
    (message.from_id === myId && message.to_id === peerUserId.value)
  )
}

const isNearBottom = () => {
  if (!messageListRef.value) return false
  const { scrollTop, scrollHeight, clientHeight } = messageListRef.value
  return scrollHeight - scrollTop - clientHeight < 120
}

const scrollToBottom = (smooth = true, force = false) => {
  if (!force && !isNearBottom()) return

  nextTick(() => {
    if (!messageListRef.value) return
    const maxScrollTop = messageListRef.value.scrollHeight - messageListRef.value.clientHeight
    messageListRef.value.scrollTo({
      top: maxScrollTop,
      behavior: smooth ? 'smooth' : 'auto',
    })
  })
}

const handleMessageInputFocus = () => {
  scrollToBottom(true, true)
  showMoreOptions.value = false
}

const handleNewMessage = (message: ChatMessage) => {
  if (!isCurrentChatMessage(message)) return

  const existingIndex = messages.value.findIndex((msg) => msg.id && msg.id === message.id)

  if (existingIndex === -1) {
    messages.value.push({
      ...message,
      status: message.from_id === currentUserId.value ? 'sent' : message.status,
    })
  } else {
    messages.value[existingIndex] = {
      ...message,
      status: message.from_id === currentUserId.value ? 'sent' : message.status,
    }
  }

  messages.value = [...messages.value]
  scrollToBottom(true, message.from_id === currentUserId.value)

  if (message.from_id === peerUserId.value && message.id) {
    markMessageAsRead(message.id)
  }
}

const sendMessage = async () => {
  if (!messageText.value.trim() || !currentUserId.value || !isConnected.value) return

  const content = messageText.value.trim()
  messageText.value = ''

  const tempMessage: ChatMessage = {
    id: `temp-${Date.now()}`,
    from_id: currentUserId.value,
    to_id: peerUserId.value,
    type: MessageType.Text,
    content,
    status: 'sending',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  }

  messages.value = [...messages.value, tempMessage]
  scrollToBottom(true, true)

  try {
    const response = await imStore.imSDK?.sendMessage(peerUserId.value, MessageType.Text, content)
    const messageIndex = messages.value.findIndex((msg) => msg.id === tempMessage.id)
    if (response && messageIndex !== -1) {
      messages.value[messageIndex] = { ...response, status: 'sent' }
      messages.value = [...messages.value]
    }
  } catch (error) {
    console.error('发送消息失败:', error)
    const messageIndex = messages.value.findIndex((msg) => msg.id === tempMessage.id)
    if (messageIndex !== -1 && messages.value[messageIndex]) {
      messages.value[messageIndex].status = 'failed'
      messages.value = [...messages.value]
    }
    showToast('发送失败')
  }
}

const handleScroll = (e: Event) => {
  const target = e.target as HTMLElement
  if (!target || !initialized.value || firstLoad.value || loading.value || !hasMore.value) return
  if (target.scrollTop <= 50) {
    fetchHistoryMessages(true)
  }
}

const fetchHistoryMessages = async (loadMore = false) => {
  if (loading.value) return
  if (!loadMore && !firstLoad.value && !hasMore.value) return

  try {
    loading.value = true
    if (!imStore.imSDK) return

    const oldScrollTop = messageListRef.value?.scrollTop || 0
    const oldScrollHeight = messageListRef.value?.scrollHeight || 0

    const response = await imStore.imSDK.getMessages({
      to_id: peerUserId.value,
      limit: pageSize,
      skip: loadMore ? messages.value.length : 0,
    })

    const newMessages = response.sort(
      (a, b) => new Date(a.created_at ?? 0).getTime() - new Date(b.created_at ?? 0).getTime(),
    )

    if (loadMore) {
      messages.value = [...newMessages, ...messages.value]
      nextTick(() => {
        if (!messageListRef.value) return
        const scrollDiff = messageListRef.value.scrollHeight - oldScrollHeight
        messageListRef.value.scrollTop = oldScrollTop + scrollDiff
      })
    } else {
      messages.value = newMessages
      scrollToBottom(false)
    }

    hasMore.value = newMessages.length === pageSize

    const unreadMessages = messages.value.filter(
      (msg) => msg.status !== 'read' && msg.from_id === peerUserId.value && msg.id,
    )
    for (const msg of unreadMessages) {
      await markMessageAsRead(msg.id!)
    }
  } catch (error) {
    console.error('获取历史消息失败:', error)
    showToast('加载消息失败')
  } finally {
    loading.value = false
    if (!loadMore) {
      firstLoad.value = false
      initialized.value = true
    }
  }
}

const markMessageAsRead = async (messageId: string) => {
  try {
    await imStore.imSDK?.markMessageAsRead(messageId)
  } catch (error) {
    console.error('标记消息已读失败:', error)
  }
}

const fetchTargetUser = async () => {
  try {
    const response = await request<ApiResponse<UserInfo>>(`/api/im/users/${peerUserId.value}`)
    if (response.code === 200) {
      targetUser.value = response.data
    }
  } catch (error) {
    console.error('获取用户信息失败:', error)
  }
}

const retryMessage = async (message: ChatMessage) => {
  const messageIndex = messages.value.findIndex((msg) => msg.id === message.id)
  if (messageIndex === -1) return

  const current = messages.value[messageIndex]
  if (!current) return

  current.status = 'sending'
  messages.value = [...messages.value]

  try {
    const response = await imStore.imSDK?.sendMessage(
      peerUserId.value,
      message.type ?? MessageType.Text,
      message.content,
      message.media_info,
      message.card_info,
      message.link_info,
    )
    if (response) {
      messages.value[messageIndex] = { ...response, status: 'sent' }
      messages.value = [...messages.value]
    }
  } catch (error) {
    console.error('重新发送消息失败:', error)
    if (messages.value[messageIndex]) {
      messages.value[messageIndex].status = 'failed'
      messages.value = [...messages.value]
    }
    showToast('发送失败')
  }
}

const handleAvatarError = () => {
  if (targetUser.value) {
    targetUser.value.avatar = undefined
  }
}

const handlePlusClick = () => {
  showMoreOptions.value = !showMoreOptions.value
}

const handleSelectFile = (_file: File, type: string, fileInfo: MediaInfo & { uploading?: boolean }) => {
  if (!currentUserId.value) return

  const messageType = type as MessageType
  const tempMessage: ChatMessage = {
    id: `temp-${Date.now()}`,
    from_id: currentUserId.value,
    to_id: peerUserId.value,
    type: messageType,
    content: messageType === MessageType.Image ? '[图片]' : '[视频]',
    status: 'sending',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    media_info: { ...fileInfo, uploading: true },
  }

  messages.value = [...messages.value, tempMessage]
  scrollToBottom(true, true)
}

const handleUploadSuccess = async (_file: File, type: string, fileInfo: MediaInfo) => {
  const messageType = type as MessageType
  const messageIndex = messages.value.findIndex(
    (msg) => msg.status === 'sending' && msg.type === messageType && msg.media_info?.uploading,
  )
  if (messageIndex === -1) return

  const current = messages.value[messageIndex]
  if (!current) return

  const previewUrl = current.media_info?.url

  try {
    const response = await imStore.imSDK?.sendMessage(peerUserId.value, messageType, '', fileInfo)
    if (response) {
      messages.value[messageIndex] = { ...response, status: 'sent' }
      messages.value = [...messages.value]
      if (previewUrl?.startsWith('blob:')) {
        URL.revokeObjectURL(previewUrl)
      }
    }
  } catch (error) {
    console.error('发送媒体消息失败:', error)
    if (messages.value[messageIndex]) {
      messages.value[messageIndex].status = 'failed'
      messages.value = [...messages.value]
    }
    showToast('发送失败')
  }
}

const handleUploadError = (_file: File, type: string) => {
  const messageType = type as MessageType
  const messageIndex = messages.value.findIndex(
    (msg) => msg.status === 'sending' && msg.type === messageType && msg.media_info?.uploading,
  )
  if (messageIndex === -1) return

  const current = messages.value[messageIndex]
  if (!current) return

  const previewUrl = current.media_info?.url
  current.status = 'failed'
  messages.value = [...messages.value]

  if (previewUrl?.startsWith('blob:')) {
    URL.revokeObjectURL(previewUrl)
  }
}

const waitForConnection = (timeoutMs = 10000) =>
  new Promise<void>((resolve, reject) => {
    if (imStore.isConnected) {
      resolve()
      return
    }

    const timer = window.setTimeout(() => {
      stop()
      reject(new Error('WebSocket connection timeout'))
    }, timeoutMs)

    const stop = watch(
      () => imStore.isConnected,
      (connected) => {
        if (connected) {
          window.clearTimeout(timer)
          stop()
          resolve()
        }
      },
      { immediate: true },
    )
  })

const resetChatState = () => {
  initialized.value = false
  firstLoad.value = true
  hasMore.value = true
  messages.value = []
  messageText.value = ''
  showMoreOptions.value = false
  targetUser.value = null
}

const initChat = async () => {
  if (!userStore.token) {
    router.replace({ name: 'im-login', query: { redirect: router.currentRoute.value.fullPath } })
    return
  }

  if (!peerUserId.value) {
    showToast('无效的会话')
    router.replace({ name: 'im-home' })
    return
  }

  resetChatState()
  imStore.initSDK()
  imStore.addMessageHandler(handleNewMessage)

  try {
    await Promise.all([fetchTargetUser(), waitForConnection()])
    await fetchHistoryMessages()
  } catch (error) {
    console.error('初始化聊天失败:', error)
    showToast('连接失败，请稍后重试')
  }
}

onMounted(() => {
  initChat()
})

watch(
  () => props.userId,
  (userId, prevUserId) => {
    if (userId && userId !== prevUserId) {
      imStore.removeMessageHandler(handleNewMessage)
      initChat()
    }
  },
)

onUnmounted(() => {
  imStore.removeMessageHandler(handleNewMessage)
})
</script>

<style scoped>
.chat-page {
  height: 100dvh;
  display: flex;
  flex-direction: column;
  background: white;
}

.nav-bar {
  position: sticky;
  top: 0;
  background: white;
  z-index: 100;
  border-bottom: 1px solid var(--border-color-light);
}

.nav-bar-content {
  height: 50px;
  max-width: 768px;
  margin: 0 auto;
  width: 100%;
  display: flex;
  align-items: center;
  padding: 0 var(--spacing-base);
}

.back-btn {
  border: none;
  background: none;
  padding: 8px 16px 8px 0;
  cursor: pointer;
}

.user-info {
  flex: 1;
  display: flex;
  align-items: center;
  min-width: 0;
}

.user-avatar {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
  margin-right: var(--spacing-small);
}

.user-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.title {
  font-size: var(--font-size-small);
  font-weight: 600;
  margin: 0;
  line-height: 1.2;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.connection-status-bar {
  padding: 8px var(--spacing-base);
  text-align: center;
  font-size: 13px;
  background-color: var(--bg-color-light);
  color: var(--warning-color);
}

.message-container {
  flex: 1;
  overflow: hidden;
  position: relative;
}

.message-list {
  height: 100%;
  overflow-y: auto;
  padding: var(--spacing-small);
  -webkit-overflow-scrolling: touch;
}

.message-item {
  display: flex;
  align-items: flex-start;
  margin-bottom: var(--spacing-large);
}

.message-mine {
  flex-direction: row-reverse;
}

.message-avatar {
  width: 32px;
  height: 32px;
  margin: 0 var(--spacing-small);
  flex-shrink: 0;
}

.message-avatar img {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
}

.message-wrapper {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  flex-grow: 1;
  min-width: 0;
}

.message-mine .message-wrapper {
  align-items: flex-end;
}

.message-input-container {
  position: sticky;
  bottom: 0;
  left: 0;
  right: 0;
  background: white;
  border-top: 1px solid var(--border-color-light);
  padding-bottom: calc(var(--spacing-base) + env(safe-area-inset-bottom));
  z-index: 10;
}

.message-input {
  padding: var(--spacing-base);
  display: flex;
  gap: var(--spacing-small);
  align-items: flex-end;
}

.message-input-content {
  display: flex;
  align-items: flex-end;
  gap: var(--spacing-small);
  flex: 1;
  border: 1px solid var(--border-color-light);
  background: #f6f8fc;
  border-radius: 20px;
  min-height: 40px;
  max-height: calc(24px * 4 + 16px);
  overflow: hidden;
}

.message-input-button {
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: 20px;
  margin-right: var(--spacing-small);
  padding: 8px;
  line-height: 1;
  display: flex;
  align-items: center;
}

.message-input-button:disabled {
  display: none;
}

.message-input .send-btn {
  padding: 6px 18px;
  border: none;
  border-radius: 28px;
  background: var(--primary-color);
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  font-size: 14px;
  line-height: 1.6;
  font-weight: 600;
}

.message-input .send-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  width: 0;
  padding: 0;
  overflow: hidden;
}

.message-input .send-btn i {
  font-size: 16px;
}

.loading-spinner {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: var(--spacing-small);
  height: 40px;
}

.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid var(--border-color);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.no-more-messages {
  text-align: center;
  color: var(--text-color-light);
  font-size: 12px;
  padding: var(--spacing-small);
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
