import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import { request } from '@/utils/request'
import { RefreshAccessTokenError, refreshAccessToken } from '@/utils/authRefresh'
import { startTokenRefresh, stopTokenRefresh } from '@/composables/useTokenRefresh'
import { useConversationList } from '@/composables/useConversationList'
import { useIMStore } from '@/stores/im'
import { useUnreadMessageStore } from '@/stores/unreadMessage'
import { getTokenExpiryMs, isTokenExpiringSoon } from '@/utils/token'
import type { UserInfo } from '@/types/user'
import type { ApiResponse } from '@/types/api'

const TOKEN_KEY = 'im-token'

let refreshPromise: Promise<string | null> | null = null

interface EnsureValidTokenOptions {
  force?: boolean
  logoutOnAuthError?: boolean
}

interface SetTokenOptions {
  startRefresh?: boolean
}

export const useUserStore = defineStore('user', () => {
  const token = ref<string | null>(null)
  const userInfo = ref<UserInfo | null>(null)
  const unreadMessageStore = useUnreadMessageStore()

  watch(token, (newToken, oldToken) => {
    const imStore = useIMStore()

    if (newToken && newToken !== oldToken && oldToken) {
      void imStore.reconnectWithLatestToken()
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

    try {
      setToken(stored, { startRefresh: false })
      const validToken = await ensureValidToken({ logoutOnAuthError: true })
      if (!validToken) {
        logout()
        return
      }
      await fetchUser()
      startTokenRefresh()
    } catch {
      if (!token.value) {
        logout()
        return
      }
      startTokenRefresh()
    }
  }

  const setToken = (newToken: string, options: SetTokenOptions = {}) => {
    token.value = newToken
    localStorage.setItem(TOKEN_KEY, newToken)
    if (options.startRefresh !== false) {
      startTokenRefresh()
    }
  }

  const setUserInfo = (info: UserInfo) => {
    userInfo.value = info
  }

  const logout = () => {
    stopTokenRefresh()
    token.value = null
    userInfo.value = null
    localStorage.removeItem(TOKEN_KEY)
    const { resetConversations } = useConversationList()
    resetConversations()
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

  const hasUsableToken = (value: string) => {
    const expiryMs = getTokenExpiryMs(value)
    return expiryMs != null && expiryMs > Date.now()
  }

  const ensureValidToken = async (
    options: EnsureValidTokenOptions = {},
  ): Promise<string | null> => {
    const currentToken = token.value
    if (!currentToken) {
      return null
    }

    if (!options.force && !isTokenExpiringSoon(currentToken)) {
      return currentToken
    }

    try {
      return await refreshToken()
    } catch (error) {
      if (
        error instanceof RefreshAccessTokenError &&
        error.reason === 'auth' &&
        options.logoutOnAuthError !== false
      ) {
        logout()
        return null
      }

      if (!options.force && hasUsableToken(currentToken)) {
        return currentToken
      }

      return null
    }
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
    ensureValidToken,
    fetchUser,
    initialize,
  }
})
