import { config } from '@/config'
import { SUCCESS_CODE, type ApiResponse } from '@/types/api'
import axios from '@/plugins/axios'
import { isAxiosError } from 'axios'
import { getDeviceMeta } from '@/utils/device'

export interface AccessTokenData {
  token: string
  expires_in: number
  refresh_token: string
  session_id: string
}

export type AuthActionErrorReason = 'auth' | 'network' | 'server'

export class AuthActionError extends Error {
  reason: AuthActionErrorReason
  status?: number

  constructor(reason: AuthActionErrorReason, message: string, status?: number) {
    super(message)
    this.name = 'AuthActionError'
    this.reason = reason
    this.status = status
  }
}

function normalizeAuthError(error: unknown, fallbackMessage: string): never {
  const status = isAxiosError(error) ? error.response?.status : undefined
  if (status === 401 || status === 403) {
    throw new AuthActionError('auth', '登录态已失效', status)
  }
  if (status != null) {
    throw new AuthActionError('server', fallbackMessage, status)
  }
  throw new AuthActionError('network', fallbackMessage)
}

export async function exchangeAccessToken(
  accessToken: string,
): Promise<AccessTokenData> {
  try {
    const response = await axios.post<ApiResponse<AccessTokenData>>(
      '/im/api/auth/exchange',
      { device: getDeviceMeta() },
      {
        baseURL: config.baseURL,
        headers: {
          Authorization: `Bearer ${accessToken}`,
        },
      },
    )

    if (response.data.code === SUCCESS_CODE && response.data.data?.token) {
      return response.data.data
    }

    throw new AuthActionError('server', response.data.message || '建立登录会话失败')
  } catch (error) {
    if (error instanceof AuthActionError) {
      throw error
    }
    normalizeAuthError(error, '建立登录会话失败')
  }
}

export async function refreshAccessToken(refreshToken: string): Promise<AccessTokenData> {
  try {
    const response = await axios.post<ApiResponse<AccessTokenData>>(
      '/im/api/auth/refresh',
      {
        refresh_token: refreshToken,
        device: getDeviceMeta(),
      },
      {
        baseURL: config.baseURL,
        headers: {
          'X-Refresh-Token': refreshToken,
        },
      },
    )

    if (response.data.code === SUCCESS_CODE && response.data.data?.token) {
      return response.data.data
    }

    throw new AuthActionError('server', response.data.message || '刷新登录态失败')
  } catch (error) {
    if (error instanceof AuthActionError) {
      throw error
    }
    normalizeAuthError(error, '刷新登录态失败')
  }
}

export async function logoutSession(refreshToken?: string | null): Promise<void> {
  try {
    await axios.post(
      '/im/api/auth/logout',
      refreshToken ? { refresh_token: refreshToken } : {},
      {
        baseURL: config.baseURL,
        headers: refreshToken
          ? {
              'X-Refresh-Token': refreshToken,
            }
          : undefined,
      },
    )
  } catch (error) {
    if (isAxiosError(error) && error.response?.status && error.response.status < 500) {
      return
    }
    throw error
  }
}
