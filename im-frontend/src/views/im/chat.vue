<template>
  <div class="chat-page">
    <div class="chat-layout">
      <aside class="chat-sidebar">
        <div class="sidebar-header">
          <span class="sidebar-title">消息</span>
        </div>
        <ConversationList embedded navigate-mode="replace" :active-peer-id="peerUserId" />
      </aside>

      <div class="chat-main">
    <div class="nav-bar">
      <div class="nav-bar-content">
        <button class="nav-side-btn back-btn" type="button" @click="handleBack">
          <i class="ri-arrow-left-s-line"></i>
        </button>
        <div class="nav-bar-center">
          <div v-if="isConnected" class="user-info">
            <h1 class="title">{{ targetUser?.nickname || '未知用户' }}</h1>
          </div>
          <i v-else class="ri-loader-4-line nav-reconnect-icon" aria-label="连接中"></i>
        </div>
        <button class="nav-side-btn" type="button">
          <i class="ri-more-line"></i>
        </button>
      </div>
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
        </div>
        <button
          type="button"
          class="addon-btn"
          :class="{ 'is-active': showMoreOptions }"
          aria-label="更多"
          @click="handlePlusClick"
        >
          <i :class="showMoreOptions ? 'ri-close-line' : 'ri-add-line'"></i>
        </button>
        <button
          type="button"
          class="send-btn"
          :disabled="!messageText.trim() || !isConnected"
          @click="sendMessage"
        >
          发送
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
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, watch, provide, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from '@/plugins/toast'
import { useUserStore } from '@/stores/user'
import { useIMStore } from '@/stores/im'
import { useUnreadMessageStore } from '@/stores/unreadMessage'
import { useConversationList } from '@/composables/useConversationList'
import { useUserProfiles } from '@/composables/useUserProfiles'
import { MessageType } from '@/sdk/im'
import type { MediaInfo } from '@/sdk/im'
import type { ChatMessage } from '@/types/im'
import type { UserInfo } from '@/types/user'
import { MessageComponents } from '@/components/im'
import MessageMoreOptions from '@/components/im/MessageMoreOptions.vue'
import MultilineInput from '@/components/im/MultilineInput.vue'
import ConversationList from '@/components/im/ConversationList.vue'
import { usePageTitleNotification } from '@/composables/usePageTitleNotification'

const props = defineProps<{
  userId: string
}>()

const router = useRouter()
const userStore = useUserStore()
const imStore = useIMStore()
const unreadMessageStore = useUnreadMessageStore()
const { conversations, clearUnreadForPeer } = useConversationList()
const { userMap, fetchUser, mergeUsers } = useUserProfiles()

const messageText = ref('')
const messages = ref<ChatMessage[]>([])
const targetUser = ref<UserInfo | null>(null)
const messageListRef = ref<HTMLElement | null>(null)
const showMoreOptions = ref(false)

const isConnected = computed(() => imStore.isConnected)
const currentUserId = computed(() => userStore.userInfo?.id)
const peerUserId = computed(() => props.userId)

const pageTitle = computed(() => targetUser.value?.nickname || '消息')
const { setBaseTitle } = usePageTitleNotification('消息')

watch(pageTitle, (title) => {
  setBaseTitle(title)
}, { immediate: true })

watch(
  () => conversations.value.find(
    (conversation) => conversation.to_user_info?.id === peerUserId.value,
  )?.to_user_info,
  (user) => {
    if (!user) return
    mergeUsers([user])
    targetUser.value = user
  },
)

const handleBack = () => {
  const back = window.history.state?.back
  const backToAnotherChat =
    typeof back === 'string' && /\/im\/chat\/[^/]+/.test(back)

  if (typeof back === 'string' && !backToAnotherChat) {
    router.back()
    return
  }

  router.replace({ name: 'im-home' })
}

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

const sortMessages = (items: ChatMessage[]) =>
  items.sort((a, b) => {
    const timeDiff =
      new Date(a.created_at ?? 0).getTime() - new Date(b.created_at ?? 0).getTime()
    if (timeDiff !== 0) return timeDiff
    return String(a.id ?? '').localeCompare(String(b.id ?? ''))
  })

const mergeMessages = (incoming: ChatMessage[]) => {
  if (incoming.length === 0) return

  const byId = new Map<string, ChatMessage>()
  const withoutId: ChatMessage[] = []

  for (const msg of messages.value) {
    if (msg.id) {
      byId.set(msg.id, msg)
    } else {
      withoutId.push(msg)
    }
  }

  for (const msg of incoming) {
    if (!msg.id) {
      withoutId.push(msg)
      continue
    }

    const existing = byId.get(msg.id)
    byId.set(msg.id, {
      ...existing,
      ...msg,
      status: msg.from_id === currentUserId.value ? 'sent' : msg.status,
    })
  }

  messages.value = sortMessages([...withoutId, ...byId.values()])
}

const handleNewMessage = async (message: ChatMessage) => {
  if (!isCurrentChatMessage(message)) return

  mergeMessages([message])
  scrollToBottom(true, message.from_id === currentUserId.value)

  if (message.from_id === peerUserId.value && message.id) {
    await markMessageAsRead(message.id)
    await syncUnreadState()
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

    const oldestMessage = loadMore && messages.value.length > 0 ? messages.value[0] : undefined
    const beforeId = oldestMessage?.id && !oldestMessage.id.startsWith('temp-')
      ? oldestMessage.id
      : undefined

    const response = await imStore.imSDK.getMessages({
      receiver_id: peerUserId.value,
      limit: pageSize,
      before_id: beforeId,
    })

    const newMessages = response.sort(
      (a, b) => new Date(a.created_at ?? 0).getTime() - new Date(b.created_at ?? 0).getTime(),
    )

    if (loadMore) {
      mergeMessages(newMessages)
      nextTick(() => {
        if (!messageListRef.value) return
        const scrollDiff = messageListRef.value.scrollHeight - oldScrollHeight
        messageListRef.value.scrollTop = oldScrollTop + scrollDiff
      })
    } else {
      messages.value = sortMessages(newMessages)
      scrollToBottom(false)
    }

    hasMore.value = newMessages.length === pageSize

    const unreadMessages = messages.value.filter(
      (msg) => msg.status !== 'read' && msg.from_id === peerUserId.value && msg.id,
    )
    const markedIds = new Set<string>()
    for (const msg of unreadMessages) {
      if (!msg.id || markedIds.has(msg.id)) continue
      markedIds.add(msg.id)
      await markMessageAsRead(msg.id)
      msg.status = 'read'
    }

    if (!loadMore && markedIds.size > 0) {
      await syncUnreadState()
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

const syncLatestMessages = async () => {
  if (!initialized.value || !imStore.imSDK || loading.value) return

  let latestMessage = [...messages.value]
    .reverse()
    .find((msg) => msg.id && !msg.id.startsWith('temp-'))
  const wasNearBottom = isNearBottom()

  try {
    const incoming: ChatMessage[] = []

    while (true) {
      const response = await imStore.imSDK.getMessages({
        receiver_id: peerUserId.value,
        limit: pageSize,
        after_id: latestMessage?.id,
      })
      const currentPage = response.filter(isCurrentChatMessage)
      incoming.push(...currentPage)

      latestMessage = [...currentPage]
        .reverse()
        .find((msg) => msg.id && !msg.id.startsWith('temp-'))

      if (response.length < pageSize || !latestMessage?.id) {
        break
      }
    }

    if (incoming.length === 0) {
      await syncUnreadState()
      return
    }

    mergeMessages(incoming)

    const unreadMessages = incoming.filter(
      (msg) => msg.status !== 'read' && msg.from_id === peerUserId.value && msg.id,
    )
    for (const msg of unreadMessages) {
      if (!msg.id) continue
      await markMessageAsRead(msg.id)
    }

    await syncUnreadState()
    scrollToBottom(true, wasNearBottom)
  } catch (error) {
    console.error('同步最新消息失败:', error)
  }
}

const syncUnreadState = async () => {
  clearUnreadForPeer(peerUserId.value)
  await unreadMessageStore.fetchUnreadCount({ force: true })
}

const markMessageAsRead = async (messageId: string) => {
  try {
    await imStore.imSDK?.markMessageAsRead(messageId)
  } catch (error) {
    console.error('标记消息已读失败:', error)
  }
}

const fetchTargetUser = async () => {
  const fromConversation = conversations.value.find(
    (conversation) => conversation.to_user_info?.id === peerUserId.value,
  )?.to_user_info
  if (fromConversation) {
    mergeUsers([fromConversation])
    targetUser.value = fromConversation
    return
  }

  const cached = userMap.value[peerUserId.value]
  if (cached) {
    targetUser.value = cached
    return
  }

  targetUser.value = await fetchUser(peerUserId.value)
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

const handlePlusClick = () => {
  showMoreOptions.value = !showMoreOptions.value
  if (showMoreOptions.value) {
    nextTick(() => scrollToBottom(true, true))
  }
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
  clearUnreadForPeer(peerUserId.value)
  imStore.initSDK()
  imStore.addMessageHandler(handleNewMessage)

  try {
    await Promise.all([fetchTargetUser(), waitForConnection()])
    await fetchHistoryMessages()
    await unreadMessageStore.fetchUnreadCount()
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

watch(
  () => imStore.isConnected,
  (connected, wasConnected) => {
    if (connected && wasConnected === false) {
      syncLatestMessages()
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
  background: white;
}

.chat-layout {
  display: flex;
  height: 100%;
  min-height: 0;
}

.chat-sidebar {
  width: 300px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: var(--bg-color);
  border-right: 1px solid var(--border-color-light);
  min-height: 0;
}

.sidebar-header {
  flex-shrink: 0;
  padding: 14px var(--spacing-base);
  border-bottom: 1px solid var(--border-color-light);
}

.sidebar-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-color-dark);
}

.chat-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  background: white;
}

@media (max-width: 767px) {
  .chat-sidebar {
    display: none;
  }
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
  margin: 0 auto;
  width: 100%;
  display: flex;
  align-items: center;
  padding: 0 var(--spacing-base);
}

.nav-side-btn {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  border: none;
  background: none;
  padding: 0;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-color-dark);
}

.nav-side-btn i {
  font-size: 22px;
  line-height: 1;
}

.back-btn {
  justify-content: flex-start;
  width: auto;
  padding-right: 4px;
}

.nav-bar-center {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 0;
}

.nav-reconnect-icon {
  font-size: 20px;
  color: var(--text-color-secondary);
  animation: nav-reconnect-spin 0.8s linear infinite;
}

@keyframes nav-reconnect-spin {
  to {
    transform: rotate(360deg);
  }
}

.user-info {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 0;
  max-width: 100%;
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
  flex: 1;
  min-width: 0;
  border: 1px solid var(--border-color-light);
  background: #f6f8fc;
  border-radius: 20px;
  min-height: 40px;
  max-height: calc(24px * 4 + 16px);
  overflow: hidden;
}

.addon-btn {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  padding: 0;
  border: none;
  border-radius: 50%;
  background: #f6f8fc;
  color: #666;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.2s ease, background 0.2s ease;
}

.addon-btn i {
  font-size: 22px;
  line-height: 1;
}

.addon-btn.is-active {
  color: #1989fa;
  background: #e8f3ff;
}

.addon-btn:active {
  opacity: 0.75;
}

.message-input .send-btn {
  flex-shrink: 0;
  padding: 8px 16px;
  border: none;
  border-radius: 20px;
  background: #1989fa;
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: opacity 0.2s ease;
  font-size: 15px;
  line-height: 1.4;
  font-weight: 500;
  min-height: 40px;
}

.message-input .send-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.message-input .send-btn:not(:disabled):active {
  opacity: 0.85;
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
