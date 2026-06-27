import { defineStore } from 'pinia'
import { ref } from 'vue'
import { request } from '@/utils/request'
import { normalizeUnreadCount } from '@/utils/im/format'
import { SUCCESS_CODE, type ApiResponse } from '@/types/api'

const SYNC_INTERVAL_MS = 60_000

type UnreadCountResponse = {
  unread_count: number
}

type UnreadMessage =
  | { type: 'count-updated'; count: number }
  | { type: 'refresh-requested' }

let syncTimer: ReturnType<typeof setInterval> | null = null
let channel: BroadcastChannel | null = null
let channelBound = false
let visibilityBound = false
let fetchPromise: Promise<number> | null = null

function getChannel() {
  if (typeof window === 'undefined' || channel) {
    return channel
  }

  channel = new BroadcastChannel('d-im-unread-count')
  return channel
}

function clearSyncTimer() {
  if (syncTimer) {
    clearInterval(syncTimer)
    syncTimer = null
  }
}

export const useUnreadMessageStore = defineStore('unreadMessage', () => {
  const unreadCount = ref(0)

  const setUnreadCount = (count: number, options: { broadcast?: boolean } = {}) => {
    unreadCount.value = normalizeUnreadCount(count)

    if (options.broadcast !== false) {
      getChannel()?.postMessage({
        type: 'count-updated',
        count: unreadCount.value,
      } satisfies UnreadMessage)
    }
  }

  const fetchUnreadCount = async (options: { broadcast?: boolean } = {}): Promise<number> => {
    if (fetchPromise) {
      return fetchPromise
    }

    fetchPromise = (async () => {
      try {
        const response = await request<ApiResponse<UnreadCountResponse>>('/im/api/messages/unread/count')
        if (response.code !== SUCCESS_CODE) {
          throw new Error(response.message || '获取未读消息数失败')
        }

        const nextCount = normalizeUnreadCount(response.data?.unread_count ?? 0)
        setUnreadCount(nextCount, { broadcast: options.broadcast !== false })
        return nextCount
      } finally {
        fetchPromise = null
      }
    })()

    return fetchPromise
  }

  const requestRefresh = () => {
    getChannel()?.postMessage({ type: 'refresh-requested' } satisfies UnreadMessage)
  }

  const increment = (count = 1) => {
    setUnreadCount(unreadCount.value + count)
  }

  const decrement = (count = 1) => {
    setUnreadCount(unreadCount.value - count)
  }

  const reset = () => {
    setUnreadCount(0)
  }

  const startHeartbeat = () => {
    console.log('startHeartbeat')
    if (typeof document === 'undefined') {
      return
    }

    const nextChannel = getChannel()
    if (nextChannel && !channelBound) {
      nextChannel.addEventListener('message', (event: MessageEvent<UnreadMessage>) => {
        const message = event.data
        if (!message) {
          return
        }

        if (message.type === 'count-updated') {
          setUnreadCount(message.count, { broadcast: false })
          return
        }

        if (message.type === 'refresh-requested') {
          void fetchUnreadCount({ broadcast: true }).catch((error) => {
            console.error('刷新未读消息数失败:', error)
          })
        }
      })
      channelBound = true
    }

    if (!visibilityBound) {
      document.addEventListener('visibilitychange', () => {
        if (document.visibilityState === 'visible') {
          void fetchUnreadCount().catch((error) => {
            console.error('同步未读消息数失败:', error)
          })
        }
      })
      visibilityBound = true
    }

    if (!syncTimer) {
      syncTimer = setInterval(() => {
        if (document.visibilityState !== 'visible') {
          return
        }
        void fetchUnreadCount().catch((error) => {
          console.error('轮询未读消息数失败:', error)
        })
      }, SYNC_INTERVAL_MS)
    }

    void fetchUnreadCount().catch((error) => {
      console.error('初始化未读消息数失败:', error)
    })
  }

  const stopHeartbeat = () => {
    clearSyncTimer()
  }

  return {
    unreadCount,
    fetchUnreadCount,
    setUnreadCount,
    increment,
    decrement,
    reset,
    startHeartbeat,
    stopHeartbeat,
    requestRefresh,
  }
})
