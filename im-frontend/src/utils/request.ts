import { useUserStore } from '@/stores/user'
import { config } from '@/config'
import axios from '@/plugins/axios'
import router from '@/router'

interface ApiError {
  response?: {
    status: number
  }
}

export const request = async <T>(url: string, options: any = {}): Promise<T> => {
  const userStore = useUserStore()
  
  const headers = {
    ...(!(options.body instanceof FormData) && { 'Content-Type': 'application/json' }),
    ...options.headers
  }

  if (userStore.token) {
    headers.Authorization = `Bearer ${userStore.token}`
  }

  try {
    const response = await axios.request<T>({
      url,
      baseURL: config.baseURL,
      method: options.method || 'GET',
      headers,
      data: options.body,
      params: options.params
    })
    return response.data
  } catch (error) {
    if ((error as ApiError).response?.status === 401) {
      userStore.logout()
      // router.push('/login')
    }
    throw error
  }
}
