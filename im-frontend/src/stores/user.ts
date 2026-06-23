import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import { request } from '@/utils/request'
import { useIMStore } from '@/stores/im'
import { useUnreadMessageStore } from '@/stores/unreadMessage'
import type { UserInfo } from '@/types/user'

interface StoredAccount {
  token: string
  userInfo: UserInfo
  lastUsed: number
}

export const useUserStore = defineStore('user', () => {
  const token = ref<string | null>(null)
  const userInfo = ref<UserInfo | null>(null)
  const accounts = ref<StoredAccount[]>([])
  const unreadMessageStore = useUnreadMessageStore()

  // 监听 token 的变化
  watch(token, (newToken, oldToken) => {
    const imStore = useIMStore()

    // 切换账户
    if (newToken && newToken !== oldToken && imStore.isConnected) {
      // 先关闭旧的连接
      imStore.closeConnection()
      // 初始化新的 SDK 实例并建立连接
      imStore.initSDK()
    } else if (!newToken) {
      // token 被清除时关闭连接
      imStore.closeConnection()
    }

    // 重置未读消息数
    if (newToken) {
      unreadMessageStore.reset()
      unreadMessageStore.startHeartbeat()
    } else {
      unreadMessageStore.stopHeartbeat()
    }
  })

  const initialize = async () => {
    const storedAccounts = localStorage.getItem('user-accounts')
    if (storedAccounts) {
      try {
        const parsedAccounts = JSON.parse(storedAccounts)
        if (Array.isArray(parsedAccounts)) {
          accounts.value = parsedAccounts
          // 获取最后使用的账户
          const lastUsedAccount = parsedAccounts.sort((a, b) => b.lastUsed - a.lastUsed)[0]
          if (lastUsedAccount) {
            token.value = lastUsedAccount.token
            userInfo.value = lastUsedAccount.userInfo
          }
        }
      } catch (e) {
        console.error('Failed to parse accounts:', e)
        localStorage.removeItem('user-accounts')
      }
    }
  }

  const setToken = (newToken: string) => {
    token.value = newToken
    saveAccounts()
  }

  const setUserInfo = (info: UserInfo) => {
    userInfo.value = info
    updateCurrentAccount()
  }

  const updateCurrentAccount = () => {
    if (!token.value || !userInfo.value) return

    const currentTime = Date.now()
    const accountIndex = accounts.value.findIndex(acc => acc.userInfo.id === userInfo.value?.id)

    if (accountIndex >= 0) {
      // 更新现有账户
      accounts.value[accountIndex] = {
        token: token.value,
        userInfo: userInfo.value,
        lastUsed: currentTime
      }
    } else {
      // 添加新账户
      accounts.value.push({
        token: token.value,
        userInfo: userInfo.value,
        lastUsed: currentTime
      })
    }

    saveAccounts()
  }

  const saveAccounts = () => {
    localStorage.setItem('user-accounts', JSON.stringify(accounts.value))
  }

  const switchAccount = (accountId: string) => {
    const account = accounts.value.find(acc => acc.userInfo.id === accountId)
    if (account) {
      token.value = account.token
      userInfo.value = account.userInfo
      account.lastUsed = Date.now()
      saveAccounts()
    }
  }

  const removeAccount = (accountId: string) => {
    const index = accounts.value.findIndex(acc => acc.userInfo.id === accountId)
    if (index >= 0) {
      accounts.value.splice(index, 1)
      saveAccounts()
    }
  }

  const logout = () => {
    token.value = null
    userInfo.value = null
    accounts.value = []
    localStorage.removeItem('user-accounts')
  }

  const fetchUser = async () => {
    try {
      const response = await request('/api/used/users/me') as { code: number, data: UserInfo }
      if (response.code === 200) {
        setUserInfo(response.data)
      }
    } catch (error) {
      console.error('Failed to fetch user info:', error)
    }
  }

  return {
    token,
    userInfo,
    accounts,
    setToken,
    setUserInfo,
    logout,
    fetchUser,
    initialize,
    switchAccount,
    removeAccount
  }
})
