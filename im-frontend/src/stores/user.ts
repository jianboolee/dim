import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import { request } from '@/utils/request'
import {
  AuthActionError,
  exchangeAccessToken,
  logoutSession,
  refreshAccessToken,
} from '@/utils/authRefresh'
import { useConversationList } from '@/composables/useConversationList'
import { useIMStore } from '@/stores/im'
import { useUnreadMessageStore } from '@/stores/unreadMessage'
import { getTokenExpiryMs, isTokenExpiringSoon, REFRESH_THRESHOLD_MS } from '@/utils/token'
import type { UserInfo } from '@/types/user'
import { SUCCESS_CODE, type ApiResponse } from '@/types/api'

let refreshPromise: Promise<string | null> | null = null
let authChannel: BroadcastChannel | null = null
let refreshTimer: ReturnType<typeof setTimeout> | null = null
let visibilityBound = false
let authChannelBound = false

type AuthMessage =
  | { type: 'token-updated'; token: string }
  | { type: 'logout' }

interface EnsureValidTokenOptions {
  force?: boolean
  logoutOnAuthError?: boolean
}

interface LogoutOptions {
  broadcast?: boolean
  revokeSession?: boolean
}

interface SetTokenOptions {
  broadcast?: boolean
}

function createAuthChannel() {
  if (typeof window === 'undefined' || authChannel) {
    return authChannel
  }

  authChannel = new BroadcastChannel('d-im-auth')
  return authChannel
}

function clearRefreshTimer() {
  if (refreshTimer) {
    clearTimeout(refreshTimer)
    refreshTimer = null
  }
}

export const useUserStore = defineStore('user', () => {
  const token = ref<string | null>(null)
  const userInfo = ref<UserInfo | null>(null)
  const unreadMessageStore = useUnreadMessageStore()

  const scheduleRefresh = (nextToken: string | null) => {
    clearRefreshTimer()
    if (!nextToken) {
      return
    }

    const expiryMs = getTokenExpiryMs(nextToken)
    if (expiryMs == null) {
      return
    }

    const delay = Math.max(expiryMs - Date.now() - REFRESH_THRESHOLD_MS, 0)
    refreshTimer = setTimeout(() => {
      void ensureValidToken({ force: true, logoutOnAuthError: true })
    }, delay)
  }

  const bindVisibilityRefresh = () => {
    if (typeof document === 'undefined' || visibilityBound) {
      return
    }

    document.addEventListener('visibilitychange', () => {
      if (document.visibilityState === 'visible') {
        void ensureValidToken({ logoutOnAuthError: true })
      }
    })
    visibilityBound = true
  }

  const setToken = (newToken: string, options: SetTokenOptions = {}) => {
    token.value = newToken
    scheduleRefresh(newToken)

    const imStore = useIMStore()
    imStore.syncAccessToken(newToken)

    if (options.broadcast !== false) {
      createAuthChannel()?.postMessage({ type: 'token-updated', token: newToken } satisfies AuthMessage)
    }
  }

  const clearLocalAuthState = (broadcast = false) => {
    clearRefreshTimer()
    token.value = null
    userInfo.value = null

    const imStore = useIMStore()
    imStore.closeConnection()

    if (broadcast) {
      createAuthChannel()?.postMessage({ type: 'logout' } satisfies AuthMessage)
    }
  }

  const applyLogoutSideEffects = () => {
    const { resetConversations } = useConversationList()
    resetConversations()
  }

  const logout = async (options: LogoutOptions = {}) => {
    const { broadcast = true, revokeSession = true } = options

    if (revokeSession) {
      try {
        await logoutSession()
      } catch (error) {
        console.error('登出会话失败:', error)
      }
    }

    clearLocalAuthState(broadcast)
    applyLogoutSideEffects()
  }

  const hasUsableToken = (value: string | null) => {
    if (!value) {
      return false
    }
    const expiryMs = getTokenExpiryMs(value)
    return expiryMs != null && expiryMs > Date.now()
  }

  const refreshToken = async (
    options: Pick<EnsureValidTokenOptions, 'logoutOnAuthError'> & { broadcast?: boolean } = {},
  ): Promise<string | null> => {
    if (refreshPromise) {
      return refreshPromise
    }

    refreshPromise = (async () => {
      try {
        const result = await refreshAccessToken()
        if (result?.token) {
          setToken(result.token, { broadcast: options.broadcast !== false })
          return result.token
        }
        return null
      } finally {
        refreshPromise = null
      }
    })()

    try {
      return await refreshPromise
    } catch (error) {
      if (hasUsableToken(token.value)) {
        return token.value
      }
      if (
        error instanceof AuthActionError &&
        error.reason === 'auth' &&
        options.logoutOnAuthError !== false
      ) {
        await logout({ revokeSession: false, broadcast: false })
        return null
      }
      throw error
    }
  }

  const ensureValidToken = async (
    options: EnsureValidTokenOptions = {},
  ): Promise<string | null> => {
    const currentToken = token.value
    if (!options.force && currentToken && !isTokenExpiringSoon(currentToken)) {
      return currentToken
    }

    try {
      return await refreshToken({ broadcast: true, logoutOnAuthError: options.logoutOnAuthError })
    } catch (error) {
      if (!options.force && hasUsableToken(currentToken)) {
        return currentToken
      }
      if (
        error instanceof AuthActionError &&
        error.reason === 'auth' &&
        options.logoutOnAuthError !== false
      ) {
        await logout({ revokeSession: false, broadcast: false })
        return null
      }
      return null
    }
  }

  const initialize = async () => {
    bindVisibilityRefresh()
    const channel = createAuthChannel()
    if (channel && !authChannelBound) {
      channel.addEventListener('message', (event: MessageEvent<AuthMessage>) => {
        const message = event.data
        if (!message) {
          return
        }
        if (message.type === 'token-updated') {
          setToken(message.token, { broadcast: false })
          if (!userInfo.value) {
            void fetchUser().catch((error) => {
              console.error('同步用户信息失败:', error)
            })
          }
          return
        }
        if (message.type === 'logout') {
          clearLocalAuthState(false)
          applyLogoutSideEffects()
        }
      })
      authChannelBound = true
    }

    const restoredToken = await ensureValidToken({ force: true, logoutOnAuthError: false })
    if (!restoredToken) {
      clearLocalAuthState(false)
      return
    }

    try {
      await fetchUser()
    } catch (error) {
      console.error('初始化用户信息失败:', error)
      await logout({ revokeSession: false })
    }
  }

  const establishSession = async (entryToken: string): Promise<string> => {
    const result = await exchangeAccessToken(entryToken)
    setToken(result.token)
    return result.token
  }

  const setUserInfo = (info: UserInfo) => {
    userInfo.value = info
  }

  const fetchUser = async (): Promise<UserInfo> => {
    const response = await request<ApiResponse<UserInfo>>('/im/api/users/me')
    if (response.code === SUCCESS_CODE && response.data?.id) {
      setUserInfo(response.data)
      return response.data
    }
    throw new Error('无法获取用户信息')
  }

  watch(token, (newToken, oldToken) => {
    if (newToken && !oldToken) {
      unreadMessageStore.startHeartbeat()
    } else if (!newToken && oldToken) {
      unreadMessageStore.reset()
      unreadMessageStore.stopHeartbeat()
    }
  })

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
    establishSession,
  }
})
