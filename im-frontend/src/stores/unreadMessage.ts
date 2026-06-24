import { defineStore } from 'pinia'
import { ref } from 'vue'
import { normalizeUnreadCount } from '@/utils/im/format'

export const useUnreadMessageStore = defineStore('unreadMessage', () => {
    const unreadCount = ref(0)

    // 获取未读消息数
    const fetchUnreadCount = async () => {}

    // 设置未读消息数
    const setUnreadCount = (count: number) => {
        unreadCount.value = normalizeUnreadCount(count)
    }

    // 增加未读消息数
    const increment = () => {
        unreadCount.value = normalizeUnreadCount(unreadCount.value + 1)
    }

    // 减少未读消息数
    const decrement = (count = 1) => {
        unreadCount.value = normalizeUnreadCount(unreadCount.value - count)
    }

    // 重置未读消息数
    const reset = () => {
        unreadCount.value = 0
    }

    // 启动心跳
    const startHeartbeat = () => {}

    // 停止心跳
    const stopHeartbeat = () => {}

    return {
        unreadCount,
        fetchUnreadCount,
        setUnreadCount,
        increment,
        decrement,
        reset,
        startHeartbeat,
        stopHeartbeat
    }
}) 
