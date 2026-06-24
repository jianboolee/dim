import { defineStore } from 'pinia'
import { ref } from 'vue'
import { request } from '@/utils/request'
import { normalizeUnreadCount } from '@/utils/im/format'
import { useIMStore } from './im'
import { useUserStore } from './user'

export const useUnreadMessageStore = defineStore('unreadMessage', () => {
    const unreadCount = ref(0)
    const heartbeatTimer = ref<ReturnType<typeof setInterval> | null>(null)

    const imStore = useIMStore()
    const userStore = useUserStore()

    // 获取未读消息数
    const fetchUnreadCount = async () => {
        try {
            const response = await request('/im/api/messages/unread/count') as { unread_count: number }
            if (response) {
                unreadCount.value = normalizeUnreadCount(response.unread_count)
            }
        } catch (error) {
            console.error('获取未读消息数失败:', error)
        }
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
    }

    // 启动心跳
    const startHeartbeat = (interval = 60000) => {
        // 清除可能存在的旧定时器
        stopHeartbeat()
        if (!userStore.token) return

        if (imStore.isConnected) {
            imStore.closeConnection()
        }

        // 立即执行一次
        fetchUnreadCount()
        // 定期执行
        heartbeatTimer.value = setInterval(() => {
            fetchUnreadCount()
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