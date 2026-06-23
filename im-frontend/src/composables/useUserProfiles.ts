import { ref, computed } from 'vue'
import { request } from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { UserInfo } from '@/types/user'

export function useUserProfiles() {
  const cache = ref<Record<string, UserInfo>>({})

  async function fetchUser(userId: string): Promise<UserInfo | null> {
    if (!userId) return null

    const cached = cache.value[userId]
    if (cached) return cached

    try {
      const response = await request<ApiResponse<UserInfo>>(`/api/im/users/${userId}`)
      if (response.code === 200) {
        cache.value = { ...cache.value, [userId]: response.data }
        return response.data
      }
    } catch (error) {
      console.error('获取用户信息失败:', error)
    }

    return null
  }

  async function fetchUsers(userIds: Iterable<string>) {
    const uniqueIds = [...new Set(userIds)].filter(Boolean)
    await Promise.all(uniqueIds.map((userId) => fetchUser(userId)))
  }

  const userMap = computed(() => cache.value)

  return {
    userMap,
    fetchUser,
    fetchUsers,
  }
}
