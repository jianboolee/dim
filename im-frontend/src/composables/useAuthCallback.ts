import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useIMStore } from '@/stores/im'
import { request } from '@/utils/request'
import type { ApiResponse } from '@/types/api'

interface ConversationDetail {
  id: string
  participants: string[]
}

export function useAuthCallback() {
  const route = useRoute()
  const router = useRouter()
  const userStore = useUserStore()
  const imStore = useIMStore()

  const loading = ref(true)
  const error = ref<string | null>(null)

  async function completeEnter() {
    const token = typeof route.query.token === 'string' ? route.query.token : ''
    const conversationId =
      typeof route.query.conversation_id === 'string' ? route.query.conversation_id : ''

    if (!token || !conversationId) {
      error.value = '缺少 token 或 conversation_id 参数'
      loading.value = false
      return
    }

    try {
      userStore.setToken(token)
      await userStore.fetchUser()

      const response = await request<ApiResponse<ConversationDetail>>(
        `/im/api/conversations/${conversationId}`,
      )

      if (response.code !== 200 || !response.data) {
        throw new Error('无法加载会话')
      }

      const myId = userStore.userInfo?.id
      if (!myId) {
        throw new Error('无法获取当前用户')
      }

      const peerId = response.data.participants.find((id) => id !== myId)
      if (!peerId) {
        throw new Error('无法解析会话对方')
      }

      await router.replace({ name: 'im-chat', params: { userId: peerId } })
      imStore.initSDK()
    } catch (err) {
      console.error('SSO 登录失败:', err)
      error.value = '登录失败，请从业务系统重新进入'
      userStore.logout()
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    error,
    completeEnter,
  }
}
