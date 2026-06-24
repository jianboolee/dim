import { ref, computed } from 'vue'
import { useIMStore } from '@/stores/im'
import { useIMTabStore } from '@/stores/imTab'
import { useUserStore } from '@/stores/user'
import type { Conversation, Message } from '@/sdk/im'
import {
  applyIncomingMessage,
  collectPeerUserIds,
  sortConversationsByActivity,
  withClearedUnreadForPeer,
} from '@/utils/im/conversation'

const conversations = ref<Conversation[]>([])
const loading = ref(false)
const loadingMore = ref(false)
const error = ref<string | null>(null)
const hasMore = ref(true)
const nextCursor = ref<string | undefined>()
let loadPromise: Promise<void> | null = null
let loadMorePromise: Promise<void> | null = null

const PAGE_SIZE = 20

export function useConversationList() {
  const imStore = useIMStore()
  const imTabStore = useIMTabStore()
  const userStore = useUserStore()

  const currentUserId = computed(() => userStore.userInfo?.id ?? '')

  function ensureImSDK() {
    if (imStore.imSDK) {
      return imStore.imSDK
    }
    if (!userStore.token) {
      return null
    }
    return imStore.initSDK()
  }

  async function loadConversations(options: { reset?: boolean } = {}) {
    if (loadPromise) {
      return loadPromise
    }

    if (options.reset !== false) {
      nextCursor.value = undefined
      hasMore.value = true
    }

    loading.value = true
    error.value = null

    loadPromise = (async () => {
      try {
        const sdk = ensureImSDK()
        if (!sdk) {
          if (!imTabStore.isSuspended) {
            error.value = '未登录'
          }
          return
        }

        const page = await sdk.getConversationPage({ limit: PAGE_SIZE })
        conversations.value = sortConversationsByActivity(page.items ?? [])
        nextCursor.value = page.next_cursor
        hasMore.value = page.has_more
      } catch (err) {
        console.error('获取会话列表失败:', err)
        error.value = '获取会话列表失败'
      } finally {
        loading.value = false
        loadPromise = null
      }
    })()

    return loadPromise
  }

  async function loadMoreConversations() {
    if (loadPromise) {
      await loadPromise
    }
    if (loadMorePromise) {
      return loadMorePromise
    }
    if (!hasMore.value || !nextCursor.value || loading.value) {
      return
    }

    loadingMore.value = true
    error.value = null

    loadMorePromise = (async () => {
      try {
        const sdk = ensureImSDK()
        if (!sdk) {
          if (!imTabStore.isSuspended) {
            error.value = '未登录'
          }
          return
        }

        const page = await sdk.getConversationPage({
          limit: PAGE_SIZE,
          cursor: nextCursor.value,
        })
        const byId = new Map(conversations.value.map((conversation) => [conversation.id, conversation]))
        for (const conversation of page.items ?? []) {
          byId.set(conversation.id, conversation)
        }
        conversations.value = sortConversationsByActivity([...byId.values()])
        nextCursor.value = page.next_cursor
        hasMore.value = page.has_more
      } catch (err) {
        console.error('加载更多会话失败:', err)
        error.value = '加载更多会话失败'
      } finally {
        loadingMore.value = false
        loadMorePromise = null
      }
    })()

    return loadMorePromise
  }

  function handleIncomingMessage(message: Message, activeConversationId?: string) {
    const userId = currentUserId.value
    if (!userId) return
    if (message.from_id !== userId && message.to_id !== userId) return

    conversations.value = applyIncomingMessage(
      conversations.value,
      message,
      userId,
      activeConversationId,
    )
  }

  function clearUnreadForPeer(peerId: string) {
    const userId = currentUserId.value
    if (!userId || !peerId) return

    conversations.value = withClearedUnreadForPeer(conversations.value, peerId, userId)
  }

  function getPeerUserIds() {
    return collectPeerUserIds(conversations.value, currentUserId.value)
  }

  function resetConversations() {
    conversations.value = []
    error.value = null
    loading.value = false
    loadingMore.value = false
    hasMore.value = true
    nextCursor.value = undefined
    loadPromise = null
    loadMorePromise = null
  }

  return {
    conversations,
    loading,
    loadingMore,
    error,
    hasMore,
    currentUserId,
    loadConversations,
    loadMoreConversations,
    handleIncomingMessage,
    clearUnreadForPeer,
    getPeerUserIds,
    resetConversations,
  }
}
