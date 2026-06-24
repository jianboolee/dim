import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import { request } from '@/utils/request'
import { refreshAccessToken } from '@/utils/authRefresh'
import { startTokenRefresh, stopTokenRefresh } from '@/composables/useTokenRefresh'
import { useIMStore } from '@/stores/im'
import { useUnreadMessageStore } from '@/stores/unreadMessage'
import type { UserInfo } from '@/types/user'
import type { ApiResponse } from '@/types/api'

const TOKEN_KEY = 'im-token'

let refreshPromise: Promise<string | null> | null = null

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
    if (!stored) {
      return
    }

    setToken(stored)

    try {
      await fetchUser()
    } catch {
      logout()
    }
  }

  const setToken = (newToken: string) => {
    token.value = newToken
    localStorage.setItem(TOKEN_KEY, newToken)
    startTokenRefresh()
  }

  const setUserInfo = (info: UserInfo) => {
    userInfo.value = info
  }

  const logout = () => {
    stopTokenRefresh()
    token.value = null
    userInfo.value = null
    localStorage.removeItem(TOKEN_KEY)
  }

  const refreshToken = async (): Promise<string | null> => {
    if (!token.value) {
      return null
    }

    if (refreshPromise) {
      return refreshPromise
    }

    const currentToken = token.value
    refreshPromise = (async () => {
      try {
        const result = await refreshAccessToken(currentToken)
        if (result?.token) {
          setToken(result.token)
          return result.token
        }
        return null
      } finally {
        refreshPromise = null
      }
    })()

    return refreshPromise
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
    refreshToken,
    fetchUser,
    initialize,
  }
})
