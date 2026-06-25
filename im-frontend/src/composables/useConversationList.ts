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
const searchResults = ref<Conversation[]>([])
const loading = ref(false)
const loadingMore = ref(false)
const searching = ref(false)
const searchingMore = ref(false)
const error = ref<string | null>(null)
const searchError = ref<string | null>(null)
const hasMore = ref(true)
const nextCursor = ref<string | undefined>()
const searchHasMore = ref(false)
const searchNextCursor = ref<string | undefined>()
const activeSearchKeyword = ref('')
const pendingScrollRequest = ref<{ conversationId: string; nonce: number } | null>(null)
let loadPromise: Promise<void> | null = null
let loadMorePromise: Promise<void> | null = null
let searchPromise: Promise<void> | null = null
let searchMorePromise: Promise<void> | null = null
let searchRequestId = 0
let scrollRequestNonce = 0

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

  async function searchConversations(keyword: string) {
    const normalizedKeyword = keyword.trim()
    activeSearchKeyword.value = normalizedKeyword
    searchNextCursor.value = undefined
    searchHasMore.value = false
    searchError.value = null
    searchingMore.value = false
    searchMorePromise = null

    const requestId = ++searchRequestId
    if (!normalizedKeyword) {
      searchResults.value = []
      searching.value = false
      searchPromise = null
      return
    }

    searching.value = true

    searchPromise = (async () => {
      try {
        const sdk = ensureImSDK()
        if (!sdk) {
          if (!imTabStore.isSuspended && requestId === searchRequestId) {
            searchError.value = '未登录'
          }
          return
        }

        const page = await sdk.getConversationPage({
          limit: PAGE_SIZE,
          q: normalizedKeyword,
        })
        if (requestId !== searchRequestId) return

        searchResults.value = sortConversationsByActivity(page.items ?? [])
        searchNextCursor.value = page.next_cursor
        searchHasMore.value = page.has_more
      } catch (err) {
        if (requestId !== searchRequestId) return
        console.error('搜索会话失败:', err)
        searchError.value = '搜索会话失败'
      } finally {
        if (requestId === searchRequestId) {
          searching.value = false
          searchPromise = null
        }
      }
    })()

    return searchPromise
  }

  async function loadMoreSearchConversations() {
    if (searchPromise) {
      await searchPromise
    }
    if (searchMorePromise) {
      return searchMorePromise
    }
    if (!activeSearchKeyword.value || !searchHasMore.value || !searchNextCursor.value || searching.value) {
      return
    }

    const requestId = searchRequestId
    searchingMore.value = true
    searchError.value = null

    searchMorePromise = (async () => {
      try {
        const sdk = ensureImSDK()
        if (!sdk) {
          if (!imTabStore.isSuspended && requestId === searchRequestId) {
            searchError.value = '未登录'
          }
          return
        }

        const page = await sdk.getConversationPage({
          limit: PAGE_SIZE,
          cursor: searchNextCursor.value,
          q: activeSearchKeyword.value,
        })
        if (requestId !== searchRequestId) return

        const byId = new Map(searchResults.value.map((conversation) => [conversation.id, conversation]))
        for (const conversation of page.items ?? []) {
          byId.set(conversation.id, conversation)
        }
        searchResults.value = sortConversationsByActivity([...byId.values()])
        searchNextCursor.value = page.next_cursor
        searchHasMore.value = page.has_more
      } catch (err) {
        if (requestId !== searchRequestId) return
        console.error('加载更多搜索会话失败:', err)
        searchError.value = '加载更多搜索会话失败'
      } finally {
        if (requestId === searchRequestId) {
          searchingMore.value = false
          searchMorePromise = null
        }
      }
    })()

    return searchMorePromise
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

  function upsertConversation(conversation: Conversation) {
    if (!conversation.id) return

    const byId = new Map(conversations.value.map((item) => [item.id, item]))
    byId.set(conversation.id, conversation)
    conversations.value = sortConversationsByActivity([...byId.values()])
  }

  async function ensureConversationInList(conversationId: string) {
    if (!conversationId) return null

    const existing = conversations.value.find((conversation) => conversation.id === conversationId)
    if (existing) return existing

    const sdk = ensureImSDK()
    if (!sdk) return null

    const conversation = await sdk.getConversation(conversationId)
    upsertConversation(conversation)
    return conversation
  }

  async function openConversationInList(conversationId: string) {
    if (!conversationId) return null

    const sdk = ensureImSDK()
    if (!sdk) return null

    const conversation = await sdk.openConversation(conversationId)
    upsertConversation(conversation)
    return conversation
  }

  function requestScrollToConversation(conversationId: string) {
    if (!conversationId) return
    pendingScrollRequest.value = {
      conversationId,
      nonce: ++scrollRequestNonce,
    }
  }

  function getPeerUserIds() {
    return collectPeerUserIds(conversations.value, currentUserId.value)
  }

  function resetConversations() {
    conversations.value = []
    searchResults.value = []
    error.value = null
    searchError.value = null
    loading.value = false
    loadingMore.value = false
    searching.value = false
    searchingMore.value = false
    hasMore.value = true
    nextCursor.value = undefined
    searchHasMore.value = false
    searchNextCursor.value = undefined
    activeSearchKeyword.value = ''
    pendingScrollRequest.value = null
    loadPromise = null
    loadMorePromise = null
    searchPromise = null
    searchMorePromise = null
    searchRequestId += 1
  }

  return {
    conversations,
    searchResults,
    loading,
    loadingMore,
    searching,
    searchingMore,
    error,
    searchError,
    hasMore,
    searchHasMore,
    pendingScrollRequest,
    currentUserId,
    loadConversations,
    loadMoreConversations,
    searchConversations,
    loadMoreSearchConversations,
    handleIncomingMessage,
    clearUnreadForPeer,
    upsertConversation,
    ensureConversationInList,
    openConversationInList,
    requestScrollToConversation,
    getPeerUserIds,
    resetConversations,
  }
}
