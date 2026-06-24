import { ref, computed } from 'vue'
import { useIMStore } from '@/stores/im'
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
const error = ref<string | null>(null)
let loadPromise: Promise<void> | null = null

export function useConversationList() {
  const imStore = useIMStore()
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

  async function loadConversations() {
    if (loadPromise) {
      return loadPromise
    }

    loading.value = true
    error.value = null

    loadPromise = (async () => {
      try {
        const sdk = ensureImSDK()
        if (!sdk) {
          error.value = '未登录'
          return
        }

        const list = await sdk.getConversations()
        conversations.value = sortConversationsByActivity(list ?? [])
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

  function handleIncomingMessage(message: Message, activePeerId?: string) {
    const userId = currentUserId.value
    if (!userId) return
    if (message.from_id !== userId && message.to_id !== userId) return

    conversations.value = applyIncomingMessage(
      conversations.value,
      message,
      userId,
      activePeerId,
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
  }

  return {
    conversations,
    loading,
    error,
    currentUserId,
    loadConversations,
    handleIncomingMessage,
    clearUnreadForPeer,
    getPeerUserIds,
    resetConversations,
  }
}
