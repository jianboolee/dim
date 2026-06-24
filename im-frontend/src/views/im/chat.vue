<template>
  <div class="chat-page">
    <div class="chat-layout">
      <aside class="chat-sidebar">
        <div class="sidebar-header">
          <span class="sidebar-title">消息</span>
          <button class="sidebar-icon-btn" type="button" aria-label="搜索会话">
            <i class="ri-search-line"></i>
          </button>
        </div>
        <ConversationList embedded navigate-mode="replace" :active-conversation-id="conversationId" />
        <div ref="sidebarMenuRef" class="sidebar-footer">
          <button
            class="sidebar-menu-btn"
            type="button"
            :class="{ 'is-active': showSidebarMenu }"
            aria-label="更多"
            @click="showSidebarMenu = !showSidebarMenu"
          >
            <i class="ri-menu-line"></i>
          </button>
          <div v-if="showSidebarMenu" class="sidebar-menu">
            <button type="button" class="sidebar-menu-item" @click="handleLogout">
              退出登录
            </button>
          </div>
        </div>
      </aside>

      <div v-if="imTabStore.isSuspended" class="chat-main chat-suspended-main">
        <div class="chat-suspended-state">
          <i class="ri-computer-line"></i>
          <p>已在其他标签页打开</p>
          <button type="button" @click="handleTakeoverTab">在此标签页使用</button>
        </div>
      </div>
      <div v-else-if="hasSelectedConversation" class="chat-main">
    <div class="nav-bar">
      <div class="nav-bar-content">
        <!-- <button class="nav-side-btn back-btn" type="button" @click="handleBack">
          <i class="ri-arrow-left-s-line"></i>
        </button> -->
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

    <div class="message-container" @click="showMoreOptions = false; showSidebarMenu = false">
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
            :minRows="inputMinRows"
            :maxRows="inputMaxRows"
            @enter="sendMessage"
            @focus="handleMessageInputFocus"
          />
          <div class="message-input-actions">
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
              :disabled="!messageText.trim()"
              @click="sendMessage"
            >
              <i class="ri-arrow-up-line"></i>
            </button>
          </div>
        </div>
      </div>

      <MessageMoreOptions
        v-model="showMoreOptions"
        @select-file="handleSelectFile"
        @upload-success="handleUploadSuccess"
        @upload-error="handleUploadError"
      />
    </div>
      </div>
      <div v-else class="chat-main chat-empty-main">
        <div class="chat-empty-state">
          <i class="ri-chat-3-line"></i>
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
import { useIMTabStore } from '@/stores/imTab'
import { useUnreadMessageStore } from '@/stores/unreadMessage'
import { useConversationList } from '@/composables/useConversationList'
import { useUserProfiles } from '@/composables/useUserProfiles'
import { MessageType } from '@/sdk/im'
import type { MediaInfo } from '@/sdk/im'
import type { ChatMessage, Conversation } from '@/types/im'
import type { UserInfo } from '@/types/user'
import { MessageComponents } from '@/components/im'
import MessageMoreOptions from '@/components/im/MessageMoreOptions.vue'
import MultilineInput from '@/components/im/MultilineInput.vue'
import ConversationList from '@/components/im/ConversationList.vue'
import { usePageTitleNotification } from '@/composables/usePageTitleNotification'

const props = defineProps<{
  conversationId: string
}>()

const router = useRouter()
const userStore = useUserStore()
const imStore = useIMStore()
const imTabStore = useIMTabStore()
const unreadMessageStore = useUnreadMessageStore()
const { conversations, clearUnreadForPeer } = useConversationList()
const { userMap, fetchUser, mergeUsers } = useUserProfiles()

const messageText = ref('')
const messages = ref<ChatMessage[]>([])
const conversation = ref<Conversation | null>(null)
const targetUser = ref<UserInfo | null>(null)
const messageListRef = ref<HTMLElement | null>(null)
const sidebarMenuRef = ref<HTMLElement | null>(null)
const showMoreOptions = ref(false)
const showSidebarMenu = ref(false)
const isMobileViewport = ref(false)
let cleanupViewportListener: (() => void) | null = null

const isConnected = computed(() => imStore.isConnected)
const currentUserId = computed(() => userStore.userInfo?.id)
const conversationId = computed(() => props.conversationId)
const hasSelectedConversation = computed(() => Boolean(conversationId.value))
const inputMinRows = computed(() => (isMobileViewport.value ? 1 : 2))
const inputMaxRows = computed(() => (isMobileViewport.value ? 4 : 6))
const peerUserId = computed(
  () =>
    conversation.value?.to_user_info?.id ||
    conversation.value?.participants.find((id) => id !== currentUserId.value) ||
    '',
)

const pageTitle = computed(() => targetUser.value?.nickname || '消息')
const { setBaseTitle } = usePageTitleNotification('消息')

watch(pageTitle, (title) => {
  setBaseTitle(title)
}, { immediate: true })

watch(
  () => conversations.value.find(
    (item) => item.id === conversationId.value,
  ),
  (user) => {
    if (!user) return
    conversation.value = user
    if (user.to_user_info) {
      mergeUsers([user.to_user_info])
      targetUser.value = user.to_user_info
    }
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

const handleLogout = () => {
  showSidebarMenu.value = false
  imTabStore.reset()
  userStore.logout()
  router.replace({ name: 'im-login' })
}

const handleTakeoverTab = () => {
  imTabStore.claimActive()
  initChat()
}

const handleDocumentPointerDown = (event: PointerEvent) => {
  if (!showSidebarMenu.value) return
  const target = event.target
  if (target instanceof Node && sidebarMenuRef.value?.contains(target)) {
    return
  }
  showSidebarMenu.value = false
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

  if (message.conversation_id && message.conversation_id === conversationId.value) {
    return true
  }

  if (!peerUserId.value) return false

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

const createClientMessageId = () => {
  const random =
    typeof crypto !== 'undefined' && 'randomUUID' in crypto
      ? crypto.randomUUID()
      : `${Date.now()}-${Math.random().toString(36).slice(2)}`
  return `cmid_${currentUserId.value ?? 'unknown'}_${random}`
}

const mergeMessages = (incoming: ChatMessage[]) => {
  if (incoming.length === 0) return

  const next = [...messages.value]

  for (const msg of incoming) {
    const existingIndex = next.findIndex((item) => {
      if (msg.client_message_id && item.client_message_id === msg.client_message_id) return true
      if (msg.id && item.id === msg.id) return true
      return false
    })

    const existing = existingIndex === -1 ? undefined : next[existingIndex]
    const merged = {
      ...existing,
      ...msg,
      status: msg.from_id === currentUserId.value ? 'sent' : msg.status,
    }

    if (existingIndex === -1) {
      next.push(merged)
    } else {
      next[existingIndex] = merged
    }
  }

  const seen = new Set<string>()
  const deduped = next.filter((msg) => {
    const key = msg.client_message_id
      ? `client:${msg.client_message_id}`
      : msg.id
        ? `id:${msg.id}`
        : ''
    if (!key) return true
    if (seen.has(key)) return false
    seen.add(key)
    return true
  })

  messages.value = sortMessages(deduped)
}

const confirmPendingMessage = (pendingId: string, confirmed: ChatMessage) => {
  if (!confirmed.client_message_id && !confirmed.id) {
    messages.value = messages.value.filter((msg) => msg.id !== pendingId)
  }
  mergeMessages([{ ...confirmed, status: 'sent' }])
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
  if (!messageText.value.trim() || !currentUserId.value) return
  if (!peerUserId.value) {
    showToast('无效的会话')
    return
  }

  const content = messageText.value.trim()
  const clientMessageId = createClientMessageId()
  messageText.value = ''

  const tempMessage: ChatMessage = {
    id: `temp-${clientMessageId}`,
    client_message_id: clientMessageId,
    conversation_id: conversationId.value,
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
    const response = await imStore.imSDK?.sendMessage(
      conversationId.value,
      MessageType.Text,
      content,
      undefined,
      undefined,
      undefined,
      clientMessageId,
    )
    if (response) {
      confirmPendingMessage(tempMessage.id!, response)
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

    const response = await imStore.imSDK.getConversationMessages(conversationId.value, {
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
      const response = await imStore.imSDK.getConversationMessages(conversationId.value, {
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
  if (peerUserId.value) {
    clearUnreadForPeer(peerUserId.value)
  }
  unreadMessageStore.decrement()
}

const markMessageAsRead = async (messageId: string) => {
  try {
    await imStore.imSDK?.markMessageAsRead(messageId)
  } catch (error) {
    console.error('标记消息已读失败:', error)
  }
}

const fetchTargetUser = async () => {
  if (conversation.value?.to_user_info) {
    mergeUsers([conversation.value.to_user_info])
    targetUser.value = conversation.value.to_user_info
    return
  }

  if (!peerUserId.value) {
    targetUser.value = null
    return
  }

  const fromConversation = conversations.value.find(
    (item) => item.id === conversationId.value,
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

const fetchConversation = async () => {
  const existing = conversations.value.find((item) => item.id === conversationId.value)
  if (existing) {
    conversation.value = existing
    if (existing.to_user_info) {
      mergeUsers([existing.to_user_info])
      targetUser.value = existing.to_user_info
    }
    return
  }

  if (!imStore.imSDK) return

  conversation.value = await imStore.imSDK.getConversation(conversationId.value)
  if (conversation.value.to_user_info) {
    mergeUsers([conversation.value.to_user_info])
    targetUser.value = conversation.value.to_user_info
  }
}

const retryMessage = async (message: ChatMessage) => {
  const messageIndex = messages.value.findIndex((msg) => msg.id === message.id)
  if (messageIndex === -1) return

  const current = messages.value[messageIndex]
  if (!current) return

  const clientMessageId = current.client_message_id ?? createClientMessageId()
  current.status = 'sending'
  current.client_message_id = clientMessageId
  messages.value = [...messages.value]

  try {
    const response = await imStore.imSDK?.sendMessage(
      conversationId.value,
      message.type ?? MessageType.Text,
      message.content,
      message.media_info,
      message.card_info,
      message.link_info,
      clientMessageId,
    )
    if (response) {
      confirmPendingMessage(message.id!, response)
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
  if (!currentUserId.value || !peerUserId.value) return

  const messageType = type as MessageType
  const clientMessageId = createClientMessageId()
  const tempMessage: ChatMessage = {
    id: `temp-${clientMessageId}`,
    client_message_id: clientMessageId,
    conversation_id: conversationId.value,
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
    const response = await imStore.imSDK?.sendMessage(
      conversationId.value,
      messageType,
      '',
      fileInfo,
      undefined,
      undefined,
      current.client_message_id,
    )
    if (response) {
      confirmPendingMessage(current.id!, response)
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
  conversation.value = null
}

const initChat = async () => {
  if (!userStore.token) {
    router.replace({ name: 'im-login', query: { redirect: router.currentRoute.value.fullPath } })
    return
  }

  if (imTabStore.isSuspended) {
    resetChatState()
    imStore.closeConnection()
    return
  }

  if (!conversationId.value) {
    resetChatState()
    imStore.initSDK()
    imStore.addMessageHandler(handleNewMessage)
    return
  }

  resetChatState()
  imStore.initSDK()
  imStore.addMessageHandler(handleNewMessage)

  try {
    await fetchConversation()
    if (!peerUserId.value) {
      throw new Error('invalid conversation participants')
    }
    clearUnreadForPeer(peerUserId.value)
    await Promise.all([fetchTargetUser(), waitForConnection()])
    await fetchHistoryMessages()
  } catch (error) {
    console.error('初始化聊天失败:', error)
    showToast('连接失败，请稍后重试')
  }
}

onMounted(() => {
  const mediaQuery = window.matchMedia('(max-width: 767px)')
  const handleViewportChange = (event: MediaQueryListEvent) => {
    isMobileViewport.value = event.matches
  }
  isMobileViewport.value = mediaQuery.matches
  mediaQuery.addEventListener('change', handleViewportChange)
  document.addEventListener('pointerdown', handleDocumentPointerDown)
  cleanupViewportListener = () => {
    mediaQuery.removeEventListener('change', handleViewportChange)
    document.removeEventListener('pointerdown', handleDocumentPointerDown)
  }

  initChat()
})

watch(
  () => imTabStore.isPrimaryTab,
  (isPrimary, wasPrimary) => {
    if (!imTabStore.initialized || isPrimary === wasPrimary) return

    if (!isPrimary) {
      resetChatState()
      imStore.closeConnection()
      return
    }

    initChat()
  },
)

watch(
  () => props.conversationId,
  (nextConversationId, prevConversationId) => {
    if (nextConversationId !== prevConversationId) {
      imStore.removeMessageHandler(handleNewMessage)
      initChat()
    }
  },
)

watch(
  () => imStore.isConnected,
  (connected, wasConnected) => {
    if (hasSelectedConversation.value && connected && wasConnected === false) {
      syncLatestMessages()
    }
  },
)

onUnmounted(() => {
  cleanupViewportListener?.()
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
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--spacing-small);
}

.sidebar-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-color-dark);
}

.sidebar-icon-btn,
.sidebar-menu-btn {
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--text-color-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.15s ease, color 0.15s ease;
}

.sidebar-icon-btn i,
.sidebar-menu-btn i {
  font-size: 20px;
  line-height: 1;
}

.sidebar-icon-btn:hover,
.sidebar-menu-btn:hover,
.sidebar-menu-btn.is-active {
  background: var(--bg-color-gray);
  color: var(--text-color-dark);
}

.sidebar-footer {
  position: relative;
  flex-shrink: 0;
  padding: 10px var(--spacing-base);
  border-top: 1px solid var(--border-color-light);
}

.sidebar-menu {
  position: absolute;
  left: var(--spacing-base);
  bottom: calc(100% + 6px);
  min-width: 132px;
  padding: 6px;
  border: 1px solid var(--border-color-light);
  border-radius: 8px;
  background: white;
  box-shadow: 0 10px 28px rgba(15, 23, 42, 0.12);
  z-index: 20;
}

.sidebar-menu-item {
  width: 100%;
  border: none;
  background: transparent;
  border-radius: 6px;
  padding: 8px 10px;
  color: var(--text-color-dark);
  font-size: 14px;
  line-height: 1.4;
  text-align: left;
  cursor: pointer;
}

.sidebar-menu-item:hover {
  background: var(--bg-color-gray);
}

.chat-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  background: white;
}

.chat-empty-main {
  align-items: center;
  justify-content: center;
  background: #fafbfe;
}

.chat-empty-state {
  width: 88px;
  height: 88px;
  border-radius: 50%;
  color: #8a96aa;
  display: flex;
  align-items: center;
  justify-content: center;
}

.chat-empty-state i {
  font-size: 38px;
  line-height: 1;
}

.chat-suspended-main {
  align-items: center;
  justify-content: center;
  background: #fafbfe;
}

.chat-suspended-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
  color: #687386;
}

.chat-suspended-state i {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  background: #eef3fb;
  color: #8a96aa;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  line-height: 1;
}

.chat-suspended-state p {
  margin: 0;
  font-size: 14px;
}

.chat-suspended-state button {
  border: none;
  border-radius: 8px;
  background: #4b86f8;
  color: white;
  padding: 8px 14px;
  font-size: 14px;
  cursor: pointer;
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
  justify-content: flex-start;
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
  padding-bottom: env(safe-area-inset-bottom);
  z-index: 10;
}

.message-input {
  padding: 10px var(--spacing-base);
  display: flex;
}

.message-input-content {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  flex: 1;
  min-width: 0;
  border: 1px solid var(--border-color-light);
  background: #f6f8fc;
  border-radius: 18px;
  min-height: 64px;
  overflow: visible;
  padding: 0 8px 0 16px;
}

@media (max-width: 767px) {
  .message-input {
    padding: 8px 10px;
  }

  .message-input-content {
    min-height: 42px;
    border-radius: 21px;
    padding-left: 14px;
  }
}

.message-input-actions {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 0;
}

.addon-btn {
  flex-shrink: 0;
  width: 24px;
  height: 24px;
  padding: 0;
  border: 2px solid #333;
  border-radius: 50%;
  background: transparent;
  color: #333;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.2s ease, background 0.2s ease;
}

.addon-btn i {
  font-size: 20px;
  line-height: 1;
}

.addon-btn.is-active {
  background: #ededed;

}

.addon-btn:active {
  opacity: 0.75;
}

.message-input .send-btn {
  flex-shrink: 0;
  width: 28px;
  height: 28px;
  padding: 0;
  border: none;
  border-radius: 50%;
  background: #4b86f8;
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: opacity 0.2s ease;
  font-size: 18px;
  line-height: 1.4;
  font-weight: 500;
}

.message-input .send-btn:disabled {
  opacity: 0.38;
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
