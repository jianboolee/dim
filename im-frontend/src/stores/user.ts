import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import { request } from '@/utils/request'
import { useIMStore } from '@/stores/im'
import { useUnreadMessageStore } from '@/stores/unreadMessage'
import type { UserInfo } from '@/types/user'
import type { ApiResponse } from '@/types/api'

const TOKEN_KEY = 'im-token'

export const useUserStore = defineStore('user', () => {
  const token = ref<string | null>(null)
  const userInfo = ref<UserInfo | null>(null)
  const unreadMessageStore = useUnreadMessageStore()

  watch(token, (newToken, oldToken) => {
    const imStore = useIMStore()

    if (newToken && newToken !== oldToken && imStore.isConnected) {
      imStore.closeConnection()
      imStore.initSDK()
    } else if (!newToken) {
      imStore.closeConnection()
    }

    if (newToken) {
      unreadMessageStore.reset()
      unreadMessageStore.startHeartbeat()
    } else {
      unreadMessageStore.stopHeartbeat()
    }
  })

  const initialize = async () => {
    const stored = localStorage.getItem(TOKEN_KEY)
    if (stored) {
      token.value = stored
      try {
        await fetchUser()
      } catch {
        logout()
      }
    }
  }

  const setToken = (newToken: string) => {
    token.value = newToken
    localStorage.setItem(TOKEN_KEY, newToken)
  }

  const setUserInfo = (info: UserInfo) => {
    userInfo.value = info
  }

  const logout = () => {
    token.value = null
    userInfo.value = null
    localStorage.removeItem(TOKEN_KEY)
  }

  const fetchUser = async (): Promise<UserInfo> => {
    const response = await request<ApiResponse<UserInfo>>('/im/api/users/me')
    if (response.code === 200 && response.data?.id) {
      setUserInfo(response.data)
      return response.data
    }
    throw new Error('无法获取用户信息')
  }

  return {
    token,
    userInfo,
    setToken,
    setUserInfo,
    logout,
    fetchUser,
    initialize,
  }
})
