import { request } from './request'

interface UploadResponse {
  code: number
  message: string
  data: {
    url: string
    filename: string
    size: number
  }[]
}

/**
 * 上传文件
 * @param files 单个文件或文件数组
 * @returns 上传结果，包含文件URL等信息
 */
export const uploadFiles = async (files: File | File[]) => {
  const formData = new FormData()
  
  if (Array.isArray(files)) {
    files.forEach(file => {
      formData.append('files', file)
    })
  } else {
    formData.append('files', files)
  }
  
  try {
    const response = await request('/api/used/upload', {
      method: 'POST',
      body: formData,
      headers: {
        // 不设置 Content-Type，让浏览器自动设置包含 boundary 的值
      }
    }) as UploadResponse
    
    if (response.code === 200) {
      return response.data
    }
    throw new Error(response.message || '上传失败')
  } catch (error) {
    console.error('文件上传失败:', error)
    throw error
  }
} 