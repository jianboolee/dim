import { ref, computed } from 'vue'
import { useIMStore } from '@/stores/im'
import { useUserStore } from '@/stores/user'
import type { Conversation, Message } from '@/sdk/im'
import {
  applyIncomingMessage,
  collectPeerUserIds,
  sortConversationsByActivity,
} from '@/utils/im/conversation'

export function useConversationList() {
  const imStore = useIMStore()
  const userStore = useUserStore()

  const conversations = ref<Conversation[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const currentUserId = computed(() => userStore.userInfo?.id ?? '')

  async function loadConversations() {
    if (!imStore.imSDK) return

    loading.value = true
    error.value = null

    try {
      const list = await imStore.imSDK.getConversations()
      conversations.value = sortConversationsByActivity(list)
    } catch (err) {
      console.error('获取会话列表失败:', err)
      error.value = '获取会话列表失败'
    } finally {
      loading.value = false
    }
  }

  function handleIncomingMessage(message: Message) {
    const userId = currentUserId.value
    if (!userId) return
    if (message.from_id !== userId && message.to_id !== userId) return

    conversations.value = applyIncomingMessage(conversations.value, message, userId)
  }

  function getPeerUserIds() {
    return collectPeerUserIds(conversations.value, currentUserId.value)
  }

  return {
    conversations,
    loading,
    error,
    currentUserId,
    loadConversations,
    handleIncomingMessage,
    getPeerUserIds,
  }
}
