import { config } from '@/config'
import type { ApiResponse } from '@/types/api'
import { request } from './request'

export interface UploadedFile {
  url: string
  filename: string
  size: number
  width?: number
  height?: number
  format?: string
}

export const UPLOAD_TIMEOUT_MS = 60000

function normalizeUploadPayload(data: unknown): UploadedFile[] {
  if (Array.isArray(data)) {
    return data as UploadedFile[]
  }

  if (data && typeof data === 'object' && 'code' in data) {
    const wrapped = data as ApiResponse<UploadedFile[] | UploadedFile>
    if (wrapped.code !== 200 || wrapped.data == null) {
      throw new Error(wrapped.message || '上传失败')
    }
    return Array.isArray(wrapped.data) ? wrapped.data : [wrapped.data]
  }

  throw new Error('上传响应格式无效')
}

/**
 * 上传文件到 IM 服务
 */
export async function uploadIMFiles(files: File | File[]): Promise<UploadedFile[]> {
  const formData = new FormData()
  const list = Array.isArray(files) ? files : [files]
  list.forEach((file) => formData.append('files', file))

  const response = await request<unknown>(config.api.uploads, {
    method: 'POST',
    body: formData,
    headers: {},
    timeout: UPLOAD_TIMEOUT_MS,
  })

  const uploaded = normalizeUploadPayload(response)
  if (!uploaded.length) {
    throw new Error('上传失败')
  }

  return uploaded
}

/** 上传单个文件，无有效结果时抛错 */
export async function uploadIMFile(file: File): Promise<UploadedFile> {
  const [uploaded] = await uploadIMFiles(file)
  if (!uploaded) {
    throw new Error('上传失败')
  }
  return uploaded
}
