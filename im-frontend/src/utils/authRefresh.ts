import { config } from '@/config'
import type { ApiResponse } from '@/types/api'
import axios from '@/plugins/axios'

export interface RefreshTokenData {
  token: string
  expires_in: number
}

export async function refreshAccessToken(currentToken: string): Promise<RefreshTokenData | null> {
  try {
    const response = await axios.post<ApiResponse<RefreshTokenData>>(
      '/im/api/auth/refresh',
      {},
      {
        baseURL: config.baseURL,
        headers: {
          Authorization: `Bearer ${currentToken}`,
        },
      },
    )

    if (response.data.code === 200 && response.data.data?.token) {
      return response.data.data
    }

    return null
  } catch {
    return null
  }
}
