import { config } from '@/config'
import type { ApiResponse } from '@/types/api'
import axios from '@/plugins/axios'
import { isAxiosError } from 'axios'

export interface RefreshTokenData {
  token: string
  expires_in: number
}

export type RefreshAccessTokenErrorReason = 'auth' | 'network' | 'server'

export class RefreshAccessTokenError extends Error {
  reason: RefreshAccessTokenErrorReason
  status?: number

  constructor(reason: RefreshAccessTokenErrorReason, message: string, status?: number) {
    super(message)
    this.name = 'RefreshAccessTokenError'
    this.reason = reason
    this.status = status
  }
}

export async function refreshAccessToken(currentToken: string): Promise<RefreshTokenData> {
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

    throw new RefreshAccessTokenError('server', response.data.message || '刷新登录态失败')
  } catch (error) {
    if (error instanceof RefreshAccessTokenError) {
      throw error
    }

    const status = isAxiosError(error) ? error.response?.status : undefined
    if (status === 401 || status === 403) {
      throw new RefreshAccessTokenError('auth', '登录态已失效', status)
    }
    if (status != null) {
      throw new RefreshAccessTokenError('server', '刷新登录态失败', status)
    }
    throw new RefreshAccessTokenError('network', '网络异常，刷新登录态失败')
  }
}
