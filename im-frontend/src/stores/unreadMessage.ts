import { defineStore } from 'pinia'
import { ref } from 'vue'
import { request } from '@/utils/request'
import { normalizeUnreadCount } from '@/utils/im/format'
import { useUserStore } from './user'

export const useUnreadMessageStore = defineStore('unreadMessage', () => {
    const unreadCount = ref(0)
    const heartbeatTimer = ref<ReturnType<typeof setInterval> | null>(null)
    const fetchPromise = ref<Promise<void> | null>(null)
    const lastFetchAt = ref(0)
    const FETCH_DEDUP_MS = 2000

    const userStore = useUserStore()

    // 获取未读消息数
    const fetchUnreadCount = async (options: { force?: boolean } = {}) => {
        if (fetchPromise.value) {
            return fetchPromise.value
        }

        if (!options.force && Date.now() - lastFetchAt.value < FETCH_DEDUP_MS) {
            return
        }

        fetchPromise.value = (async () => {
            try {
                const response = await request('/im/api/messages/unread/count') as { unread_count: number }
                if (response) {
                    unreadCount.value = normalizeUnreadCount(response.unread_count)
                    lastFetchAt.value = Date.now()
                }
            } catch (error) {
                console.error('获取未读消息数失败:', error)
            } finally {
                fetchPromise.value = null
            }
        })()

        return fetchPromise.value
    }

    // 设置未读消息数
    const setUnreadCount = (count: number) => {
        unreadCount.value = normalizeUnreadCount(count)
    }

    // 增加未读消息数
    const increment = () => {
        unreadCount.value = normalizeUnreadCount(unreadCount.value + 1)
    }

    // 重置未读消息数
    const reset = () => {
        unreadCount.value = 0
        lastFetchAt.value = 0
    }

    // 启动心跳
    const startHeartbeat = (interval = 60000) => {
        stopHeartbeat()
        if (!userStore.token) return

        fetchUnreadCount({ force: true })
        heartbeatTimer.value = setInterval(() => {
            fetchUnreadCount({ force: true })
        }, interval)
    }

    // 停止心跳
    const stopHeartbeat = () => {
        if (heartbeatTimer.value) {
            clearInterval(heartbeatTimer.value)
            heartbeatTimer.value = null
        }
    }

    return {
        unreadCount,
        fetchUnreadCount,
        setUnreadCount,
        increment,
        reset,
        startHeartbeat,
        stopHeartbeat
    }
}) 
