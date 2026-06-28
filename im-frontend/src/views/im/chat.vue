<template>
  <div class="chat-page">
    <div class="chat-layout">
      <aside class="chat-sidebar">
        <div class="sidebar-header">
          <span class="sidebar-title">消息</span>
          <button class="sidebar-icon-btn" type="button" aria-label="搜索会话" @click="openConversationSearch">
            <i class="ri-search-line"></i>
          </button>
        </div>
        <ConversationList
          embedded
          navigate-mode="replace"
          :active-conversation-id="conversationId"
        />
        <div ref="sidebarMenuRef" class="sidebar-footer">
          <button
            class="sidebar-footer-trigger"
            type="button"
            :class="{ 'is-active': showSidebarMenu }"
            aria-label="更多"
            @click="showSidebarMenu = !showSidebarMenu"
          >
            <div class="sidebar-footer-user" :title="userStore.userInfo?.nickname || '当前用户'">
              <img
                v-if="userStore.userInfo?.avatar"
                class="sidebar-footer-avatar"
                :src="userStore.userInfo.avatar"
                alt=""
              >
              <div v-else class="sidebar-footer-avatar sidebar-footer-avatar-fallback">
                <i class="ri-user-3-line"></i>
              </div>
              <span class="sidebar-footer-name">
                {{ userStore.userInfo?.nickname || '当前用户' }}
              </span>
            </div>
            <span class="sidebar-menu-btn" aria-hidden="true">
              <i class="ri-menu-line"></i>
            </span>
          </button>
          <div v-if="showSidebarMenu" class="sidebar-menu">
            <div class="sidebar-user">
              <img
                class="sidebar-user-avatar"
                :src="userStore.userInfo?.avatar || ''"
                alt=""
              >
              <span class="sidebar-user-name">
                {{ userStore.userInfo?.nickname || '当前用户' }}
              </span>
            </div>
            <button type="button" class="sidebar-menu-item" @click="handleLogout">
              <i class="ri-logout-box-r-line"></i>
              <span>退出登录</span>
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
          <div class="user-info">
            <h1 class="title">{{ conversationTitle }}</h1>
          </div>
          <i v-if="!isConnected"  class="ri-loader-4-line nav-reconnect-icon" aria-label="连接中"></i>
        </div>
        <button class="nav-side-btn" type="button" aria-label="会话信息" @click="openConversationInfo">
          <i class="ri-more-line"></i>
        </button>
      </div>
    </div>

    <div class="message-container" @click="showSidebarMenu = false">
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
        <template v-for="item in timelineItems" :key="item.id">
          <div v-if="item.type === 'time'" class="message-time-divider">
            {{ item.text }}
          </div>
          <SystemEventMessage
            v-else-if="isSystemEventMessage(item.message)"
            :message="item.message"
          />
          <div
            v-else
            class="message-item"
            :class="{ 'message-mine': item.message.from_id === currentUserId }"
          >
            <div v-if="getMessageAvatar(item.message)" class="message-avatar">
              <img
                :src="getMessageAvatar(item.message)"
                alt=""
              >
            </div>
            <div class="message-wrapper">
              <component
                :is="MessageComponents[item.message.type || MessageType.Text] ?? MessageComponents[MessageType.Text]"
                :message="item.message"
                :isMine="item.message.from_id === currentUserId"
                @retry="retryMessage(item.message)"
              />
            </div>
          </div>
        </template>
      </div>
    </div>

    <div v-if="peerUserType === 'system'" class="system-notice-bar">
      <i class="ri-information-line"></i>
      <span>系统消息，暂不支持回复</span>
    </div>
    <div v-else class="message-input-container">
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
            <MessageMoreOptions
              @select-file="handleSelectFile"
              @upload-success="handleUploadSuccess"
              @upload-error="handleUploadError"
            />
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
    </div>
      </div>
      <div v-else class="chat-main chat-empty-main">
        <div class="chat-empty-state">
          <i class="ri-chat-3-line"></i>
        </div>
      </div>
    </div>
    <ConversationSearchModal
      v-model="showConversationSearch"
      navigate-mode="replace"
      :active-conversation-id="conversationId"
    />
    <ConversationInfoDrawer
      v-model="showConversationInfoDrawer"
      :participants="conversationInfoParticipants"
      :is-group="isGroupConversation"
      :group-id="conversation?.group_id"
      @invite="handleInviteMembers"
      @search="handleSearchInConversation"
      @clear="handleClearConversationHistory"
    />
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
import type { Payload } from '@/sdk/im'
import type { ChatMessage, Conversation, UploadState } from '@/types/im'
import type { UserInfo } from '@/types/user'
import { MessageComponents } from '@/components/im'
import MessageMoreOptions from '@/components/im/MessageMoreOptions.vue'
import MultilineInput from '@/components/im/MultilineInput.vue'
import ConversationList from '@/components/im/ConversationList.vue'
import ConversationSearchModal from '@/components/im/ConversationSearchModal.vue'
import ConversationInfoDrawer from '@/components/im/ConversationInfoDrawer.vue'
import SystemEventMessage from '@/components/im/SystemEventMessage.vue'
import { usePageTitleNotification } from '@/composables/usePageTitleNotification'
import { buildMessageTimeline } from '@/utils/im/timeline'
import {
  getConversationDisplayName,
  getPeerUserId,
  isGroupConversation as isGroupConversationModel,
} from '@/utils/im/conversation'
import { collectSystemEventUserIds } from '@/utils/im/systemEvent'
import { readImageDimensions, getFileFormat } from '@/utils/file'
import { uploadIMFile } from '@/utils/upload'

const props = defineProps<{
  conversationId: string
}>()

const router = useRouter()
const userStore = useUserStore()
const imStore = useIMStore()
const imTabStore = useIMTabStore()
const unreadMessageStore = useUnreadMessageStore()
const {
  conversations,
  clearUnreadForPeer,
  clearUnreadForConversation,
  handleIncomingMessage: updateConversationByMessage,
  ensureConversationInList,
  requestScrollToConversation,
} = useConversationList()
const { userMap, fetchUser, fetchUsers, mergeUsers } = useUserProfiles()

const messageText = ref('')
const messages = ref<ChatMessage[]>([])
const conversation = ref<Conversation | null>(null)
const targetUser = ref<UserInfo | null>(null)
const messageListRef = ref<HTMLElement | null>(null)
const sidebarMenuRef = ref<HTMLElement | null>(null)
const showSidebarMenu = ref(false)
const showConversationSearch = ref(false)
const showConversationInfoDrawer = ref(false)
const isMobileViewport = ref(false)
let cleanupViewportListener: (() => void) | null = null
const pendingUploadMessageIds = new WeakMap<File, string>()
const messageDrafts = new Map<string, string>()

const isConnected = computed(() => imStore.isConnected)
const currentUserId = computed(() => userStore.userInfo?.id)
const conversationId = computed(() => props.conversationId)
const hasSelectedConversation = computed(() => Boolean(conversationId.value))
const inputMinRows = computed(() => (isMobileViewport.value ? 1 : 2))
const inputMaxRows = computed(() => (isMobileViewport.value ? 10 : 15))
const timelineItems = computed(() => buildMessageTimeline(messages.value))
const isGroupConversation = computed(() => isGroupConversationModel(conversation.value))
const peerUserId = computed(() => {
  if (!conversation.value || isGroupConversation.value) return ''
  return conversation.value.peer_user_info?.id
    || getPeerUserId(conversation.value, currentUserId.value || '')
})
const peerUserType = computed(
  () => isGroupConversation.value
    ? ''
    : conversation.value?.peer_user_info?.type || userMap.value[peerUserId.value]?.type || '',
)
const conversationTitle = computed(() => (
  conversation.value
    ? getConversationDisplayName(conversation.value, currentUserId.value || '')
    : '-'
))
const conversationInfoParticipants = computed<UserInfo[]>(() => {
  const currentId = currentUserId.value
  if (isGroupConversation.value) {
    return (conversation.value?.group_info?.members ?? [])
      .map((member) => member.user_info ?? userMap.value[member.user_id] ?? { id: member.user_id })
      .filter((user) => user.id)
  }

  const participantIds = conversation.value?.participants.filter((id) => id && id !== currentId) ?? []
  const usersById = new Map<string, UserInfo>()

  const peerInfo = conversation.value?.peer_user_info
  if (peerInfo?.id && peerInfo.id !== currentId) {
    usersById.set(peerInfo.id, peerInfo)
  }

  if (targetUser.value?.id && targetUser.value.id !== currentId) {
    usersById.set(targetUser.value.id, targetUser.value)
  }

  participantIds.forEach((id) => {
    usersById.set(id, userMap.value[id] ?? usersById.get(id) ?? { id })
  })

  return [...usersById.values()]
})

const pageTitle = computed(() => conversationTitle.value || '消息')
const { setBaseTitle } = usePageTitleNotification('消息')

const mergeConversationUsers = (item: Conversation) => {
  mergeUsers([
    item.peer_user_info,
    ...(item.group_info?.members ?? []).map((member) => member.user_info),
  ])
}

const getMessageAvatar = (message: ChatMessage) => {
  if (message.from_id === currentUserId.value) {
    return userStore.userInfo?.avatar || ''
  }
  if (isGroupConversation.value) {
    return userMap.value[message.from_id || '']?.avatar || ''
  }
  return targetUser.value?.avatar || ''
}

const isSystemEventMessage = (message: ChatMessage) => message.type === MessageType.SystemEvent

const collectMessageUserIds = (items: ChatMessage[]) => {
  const ids = new Set<string>()
  for (const message of items) {
    if (message.from_id) ids.add(message.from_id)
    if (isSystemEventMessage(message)) {
      for (const userId of collectSystemEventUserIds(message)) {
        ids.add(userId)
      }
    }
  }
  return [...ids]
}

watch(pageTitle, (title) => {
  setBaseTitle(title)
}, { immediate: true })

watch(messageText, (value) => {
  const id = conversationId.value
  if (!id) return

  if (value) {
    messageDrafts.set(id, value)
  } else {
    messageDrafts.delete(id)
  }
})

watch(
  () => conversations.value.find(
    (item) => item.id === conversationId.value,
  ),
  (user) => {
    if (!user) return
    conversation.value = user
    mergeConversationUsers(user)
    const peerInfo = user.peer_user_info
    if (!isGroupConversationModel(user) && peerInfo) {
      mergeUsers([peerInfo])
      targetUser.value = peerInfo
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

const openConversationSearch = () => {
  showConversationSearch.value = true
  showSidebarMenu.value = false
}

const openConversationInfo = () => {
  showConversationInfoDrawer.value = true
  showSidebarMenu.value = false
}

const handleSearchInConversation = () => {
  showToast('查找聊天内容稍后开放')
}

const handleClearConversationHistory = () => {
  showToast('清空聊天记录稍后开放')
}

const handleInviteMembers = () => {
  showToast('邀请成员稍后开放')
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
    .filter((msg) => msg.type === MessageType.Image && msg.payload?.url)
    .map((msg) => msg.payload!.url!),
)

provide('chatImages', chatImages)

const isCurrentChatMessage = (message: ChatMessage) => {
  return Boolean(message.conversation_id && message.conversation_id === conversationId.value)
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
    return
  }
  mergeMessages([{ ...confirmed, status: 'sent', uploadState: undefined }])
}

const syncConversationByMessage = (message: ChatMessage, shouldScroll = false) => {
  updateConversationByMessage(
    message as Parameters<typeof updateConversationByMessage>[0],
    conversationId.value,
  )
  if (shouldScroll) {
    requestScrollToConversation(conversationId.value)
  }
}

const handleNewMessage = async (message: ChatMessage) => {
  if (!isCurrentChatMessage(message)) return

  if (isGroupConversation.value) {
    await fetchUsers(collectMessageUserIds([message]))
  }
  mergeMessages([message])
  syncConversationByMessage(message, message.from_id === currentUserId.value)
  scrollToBottom(true, message.from_id === currentUserId.value)

  if (!isGroupConversation.value && message.from_id === peerUserId.value) {
    await syncUnreadState()
  }
}

const sendMessage = async () => {
  if (!messageText.value.trim() || !currentUserId.value) return
  if (!isGroupConversation.value && !peerUserId.value) {
    showToast('无效的会话')
    return
  }

  const content = messageText.value.trim()
  const clientMessageId = createClientMessageId()
  messageText.value = ''
  messageDrafts.delete(conversationId.value)

  const tempMessage: ChatMessage = {
    id: `temp-${clientMessageId}`,
    client_message_id: clientMessageId,
    conversation_id: conversationId.value,
    from_id: currentUserId.value,
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
      clientMessageId,
    )
    if (response) {
      confirmPendingMessage(tempMessage.id!, response)
      syncConversationByMessage(response, true)
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
    if (isGroupConversation.value) {
      await fetchUsers(collectMessageUserIds(newMessages))
    }

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

    if (!loadMore) {
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
    if (isGroupConversation.value) {
      await fetchUsers(collectMessageUserIds(incoming))
    }

    await syncUnreadState()
    scrollToBottom(true, wasNearBottom)
  } catch (error) {
    console.error('同步最新消息失败:', error)
  }
}

const syncUnreadState = async () => {
  if (conversationId.value && imStore.imSDK) {
    try {
      await imStore.imSDK.markConversationRead(conversationId.value)
    } catch (error) {
      console.error('标记会话已读失败:', error)
    }
  }
  if (!isGroupConversation.value && peerUserId.value) {
    clearUnreadForPeer(peerUserId.value)
  }
  clearUnreadForConversation(conversationId.value)
  unreadMessageStore.requestRefresh()
  await unreadMessageStore.fetchUnreadCount()
}

const fetchTargetUser = async () => {
  if (conversation.value && isGroupConversation.value) {
    mergeConversationUsers(conversation.value)
    targetUser.value = null
    return
  }

  const peerInfo = conversation.value?.peer_user_info
  if (peerInfo) {
    mergeUsers([peerInfo])
    targetUser.value = peerInfo
    return
  }

  if (!peerUserId.value) {
    targetUser.value = null
    return
  }

  const fromConversation = conversations.value.find(
    (item) => item.id === conversationId.value,
  )
  const fromConversationPeer = fromConversation?.peer_user_info
  if (fromConversationPeer) {
    mergeUsers([fromConversationPeer])
    targetUser.value = fromConversationPeer
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
    mergeConversationUsers(existing)
    const peerInfo = existing.peer_user_info
    if (!isGroupConversationModel(existing) && peerInfo) {
      mergeUsers([peerInfo])
      targetUser.value = peerInfo
    }
  }

  const ensuredConversation = await ensureConversationInList(conversationId.value, {
    activateIfMissing: true,
  })
  if (ensuredConversation) {
    conversation.value = ensuredConversation
  }
  requestScrollToConversation(conversationId.value)
  if (!conversation.value) return

  mergeConversationUsers(conversation.value)
  const peerInfo = conversation.value.peer_user_info
  if (!isGroupConversation.value && peerInfo) {
    mergeUsers([peerInfo])
    targetUser.value = peerInfo
  }
}

const retryMessage = async (message: ChatMessage) => {
  const messageIndex = messages.value.findIndex((msg) => msg.id === message.id)
  if (messageIndex === -1) return

  const current = messages.value[messageIndex]
  if (!current) return

  if (current.type === MessageType.Image && current.uploadState?.localFile) {
    await retryUploadImageMessage(current, messageIndex)
    return
  }

  const clientMessageId = current.client_message_id ?? createClientMessageId()
  current.status = 'sending'
  current.client_message_id = clientMessageId
  messages.value = [...messages.value]

  try {
    const response = await imStore.imSDK?.sendMessage(
      conversationId.value,
      message.type ?? MessageType.Text,
      message.content,
      message.payload,
      clientMessageId,
    )
    if (response) {
      confirmPendingMessage(message.id!, response)
      syncConversationByMessage(response, true)
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

const retryUploadImageMessage = async (message: ChatMessage, messageIndex: number) => {
  const file = message.uploadState?.localFile
  if (!file) {
    showToast('图片文件已失效，请重新选择')
    return
  }

  const current = messages.value[messageIndex]
  if (!current) return

  const clientMessageId = current.client_message_id ?? createClientMessageId()
  current.status = 'sending'
  current.client_message_id = clientMessageId
  current.uploadState = { localFile: file, uploading: true }
  messages.value = [...messages.value]

  try {
    const uploaded = await uploadIMFile(file)
    const dimensions = await readImageDimensions(file)
    const format = uploaded.format ?? getFileFormat(file)
    const w = uploaded.width ?? dimensions.width
    const h = uploaded.height ?? dimensions.height

    const response = await imStore.imSDK?.sendMessage(
      conversationId.value,
      MessageType.Image,
      '',
      {
        url: uploaded.url,
        meta: {
          width: String(w),
          height: String(h),
          size: String(uploaded.size),
          format,
        },
      } as Payload,
      clientMessageId,
    )
    if (response) {
      const previewUrl = current.payload?.url
      confirmPendingMessage(current.id!, response)
      syncConversationByMessage(response, true)
      if (previewUrl?.startsWith('blob:')) {
        URL.revokeObjectURL(previewUrl)
      }
    }
  } catch (error) {
    console.error('重新上传图片失败:', error)
    const latest = messages.value.find((msg) => msg.id === current.id)
    if (latest) {
      latest.status = 'failed'
      latest.uploadState = { localFile: file, uploading: false, uploadFailed: true }
      messages.value = [...messages.value]
    }
    showToast('上传失败，点击重试')
  }
}

// 文件上传预览信息（纯前端本地结构）
interface FilePreview {
  url: string        // 预览 blob URL
  size: number
  width?: number
  height?: number
  format?: string
  uploading?: boolean
}

const handleSelectFile = (file: File, type: string, preview: FilePreview) => {
  if (!currentUserId.value || (!isGroupConversation.value && !peerUserId.value)) return

  const messageType = type as MessageType
  const clientMessageId = createClientMessageId()
  const tempMessageId = `temp-${clientMessageId}`
  const tempMessage: ChatMessage = {
    id: tempMessageId,
    client_message_id: clientMessageId,
    conversation_id: conversationId.value,
    from_id: currentUserId.value,
    type: messageType,
    content: messageType === MessageType.Image ? '[图片]' : '[视频]',
    status: 'sending',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    payload: { url: preview.url },
    uploadState: { uploading: true, localFile: file },
  }

  pendingUploadMessageIds.set(file, tempMessageId)
  messages.value = [...messages.value, tempMessage]
  scrollToBottom(true, true)
}

const handleUploadSuccess = async (file: File, type: string, uploaded: { url: string; size: number; width?: number; height?: number; format?: string }) => {
  const messageType = type as MessageType
  const pendingMessageId = pendingUploadMessageIds.get(file)
  const messageIndex = messages.value.findIndex(
    (msg) =>
      msg.status === 'sending' &&
      msg.type === messageType &&
      msg.uploadState?.uploading &&
      (!pendingMessageId || msg.id === pendingMessageId),
  )
  if (messageIndex === -1) return

  const current = messages.value[messageIndex]
  if (!current) return

  const previewUrl = current.payload?.url

  try {
    const response = await imStore.imSDK?.sendMessage(
      conversationId.value,
      messageType,
      '',
      {
        url: uploaded.url,
        meta: {
          width: uploaded.width != null ? String(uploaded.width) : undefined,
          height: uploaded.height != null ? String(uploaded.height) : undefined,
          size: String(uploaded.size),
          format: uploaded.format,
        },
      } as Payload,
      current.client_message_id,
    )
    if (response) {
      confirmPendingMessage(current.id!, response)
      syncConversationByMessage(response, true)
      pendingUploadMessageIds.delete(file)
      if (previewUrl?.startsWith('blob:')) {
        URL.revokeObjectURL(previewUrl)
      }
    }
  } catch (error) {
    console.error('发送媒体消息失败:', error)
    if (messages.value[messageIndex]) {
      messages.value[messageIndex].status = 'failed'
      messages.value[messageIndex].uploadState = { localFile: file, uploading: false, uploadFailed: true }
      messages.value = [...messages.value]
    }
    pendingUploadMessageIds.delete(file)
    showToast('发送失败')
  }
}

const handleUploadError = (file: File, type: string) => {
  const messageType = type as MessageType
  const pendingMessageId = pendingUploadMessageIds.get(file)
  const messageIndex = messages.value.findIndex(
    (msg) =>
      msg.status === 'sending' &&
      msg.type === messageType &&
      msg.uploadState?.uploading &&
      (!pendingMessageId || msg.id === pendingMessageId),
  )
  if (messageIndex === -1) return

  const current = messages.value[messageIndex]
  if (!current) return

  current.status = 'failed'
  current.uploadState = { localFile: file, uploading: false, uploadFailed: true }
  messages.value = [...messages.value]
  pendingUploadMessageIds.delete(file)
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
  targetUser.value = null
  conversation.value = null
}

const restoreMessageDraft = (id: string) => {
  messageText.value = id ? messageDrafts.get(id) ?? '' : ''
}

const initChat = async () => {
  if (!userStore.token) {
    router.replace({ name: 'im-login', query: { redirect: router.currentRoute.value.fullPath } })
    return
  }

  if (imTabStore.isSuspended) {
    showConversationInfoDrawer.value = false
    resetChatState()
    imStore.closeConnection()
    return
  }

  if (!conversationId.value) {
    showConversationInfoDrawer.value = false
    resetChatState()
    restoreMessageDraft('')
    imStore.initSDK()
    imStore.addMessageHandler(handleNewMessage)
    return
  }

  resetChatState()
  restoreMessageDraft(conversationId.value)
  imStore.initSDK()
  imStore.addMessageHandler(handleNewMessage)

  try {
    await fetchConversation()
    if (!isGroupConversation.value && !peerUserId.value) {
      throw new Error('invalid conversation participants')
    }
    if (!isGroupConversation.value) {
      clearUnreadForPeer(peerUserId.value)
    }
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
      if (prevConversationId && messageText.value) {
        messageDrafts.set(prevConversationId, messageText.value)
      }
      showConversationInfoDrawer.value = false
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

.sidebar-footer-trigger {
  width: 100%;
  border: none;
  padding: 0;
  background: transparent;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  cursor: pointer;
  border-radius: 10px;
  transition: background 0.15s ease;
}

.sidebar-footer-trigger:hover,
.sidebar-footer-trigger.is-active {
  background: var(--bg-color-gray);
}

.sidebar-footer-user {
  min-width: 0;
  flex: 1;
  display: flex;
  align-items: center;
  gap: 10px;
}

.sidebar-footer-avatar {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  border-radius: 50%;
  object-fit: cover;
  background: #eef1f6;
}

.sidebar-footer-avatar-fallback {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-color-secondary);
}

.sidebar-footer-avatar-fallback i {
  font-size: 16px;
  line-height: 1;
}

.sidebar-footer-name {
  min-width: 0;
  color: var(--text-color-dark);
  font-size: 14px;
  font-weight: 500;
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sidebar-menu {
  position: absolute;
  left: var(--spacing-base);
  bottom: calc(100% + 6px);
  width: 270px;
  padding: 6px;
  border: 1px solid var(--border-color-light);
  border-radius: 8px;
  background: white;
  box-shadow: 0 10px 28px rgba(15, 23, 42, 0.12);
  z-index: 20;
}

.sidebar-user {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  padding: 7px 8px 9px;
  margin-bottom: 4px;
  border-bottom: 1px solid var(--border-color-light);
}

.sidebar-user-avatar {
  width: 28px;
  height: 28px;
  flex-shrink: 0;
  border-radius: 50%;
  background: #eef1f6;
  object-fit: cover;
}

.sidebar-user-name {
  min-width: 0;
  color: var(--text-color-dark);
  font-size: 13px;
  font-weight: 500;
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  text-align: left;
}

.sidebar-menu-item i {
  flex-shrink: 0;
  color: var(--text-color-secondary);
  font-size: 16px;
  line-height: 1;
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
  font-size: 16px;
  color: var(--text-color-dark);
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

.message-time-divider {
  width: fit-content;
  max-width: calc(100% - 48px);
  margin: 12px auto 16px;
  padding: 3px 8px;
  border-radius: 10px;
  background: transparent;
  color: #8a93a3;
  font-size: 12px;
  line-height: 1.4;
  text-align: center;
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
  /* opacity: 0.38; */
  cursor: not-allowed;
}

/* .message-input .send-btn:not(:disabled):active {
  opacity: 0.85;
} */

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

.system-notice-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px var(--spacing-base);
  background: #f0f4ff;
  border-top: 1px solid #dce3f5;
  color: #5b6c94;
  font-size: 13px;
}

.system-notice-bar i {
  font-size: 18px;
  line-height: 1;
  color: #8ba0cb;
}
</style>
