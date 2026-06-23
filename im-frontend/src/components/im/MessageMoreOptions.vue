<!-- 消息更多选项组件 -->
<template>
  <div class="message-more-options" :class="{ 'is-active': modelValue }">
    <div class="options-grid">
      <div class="option-item" @click="handleImageClick">
        <div class="option-icon">
          <i class="bi bi-image"></i>
        </div>
        <div class="option-label">图片</div>
        <input 
          type="file" 
          ref="imageInputRef"
          accept="image/*"
          style="display: none"
          @change="handleImageChange"
        >
      </div>
      
      <div class="option-item" @click="handleVideoClick" v-if="false">
        <div class="option-icon">
          <i class="bi bi-camera-video"></i>
        </div>
        <div class="option-label">视频</div>
        <input 
          type="file" 
          ref="videoInputRef"
          accept="video/*"
          style="display: none"
          @change="handleVideoChange"
        >
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { MessageType } from '@/sdk/im'
import { showToast, showLoadingToast, closeToast } from 'vant'
import { request } from '@/utils/request'


interface UploadImage {
    url: string
    filename: string
    size: number
    width: number
    height: number
    format: string
}


/**
 * 上传文件
 * @param files 单个文件或文件数组
 * @returns 上传结果，包含文件URL等信息
 */
const uploadFiles = async (files: File | File[]) => {
  const formData = new FormData()
  
  if (Array.isArray(files)) {
    files.forEach(file => {
      formData.append('files', file)
    })
  } else {
    formData.append('files', files)
  }
  
  try {
    const response = await request('/api/im/uploads', {
      method: 'POST',
      body: formData,
      headers: {
        // 不设置 Content-Type，让浏览器自动设置包含 boundary 的值
      }
    })
    if (response) {
      return response
    }
    throw new Error(response.message || '上传失败')
  } catch (error) {
    console.error('文件上传失败:', error)
    throw error
  }
} 

// 获取图片尺寸
const getImageDimensions = (file: File): Promise<{ width: number; height: number }> => {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.onload = () => {
      resolve({
        width: img.width,
        height: img.height
      })
    }
    img.onerror = () => {
      reject(new Error('获取图片尺寸失败'))
    }
    img.src = URL.createObjectURL(file)
  })
}

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'select-file', file: File, type: string, fileInfo?: any): void
  (e: 'upload-success', file: File, type: string, fileInfo: any): void
  (e: 'upload-error', file: File, type: string): void
}>()

const imageInputRef = ref<HTMLInputElement | null>(null)
const videoInputRef = ref<HTMLInputElement | null>(null)

const handleImageClick = () => {
  imageInputRef.value?.click()
}

const handleVideoClick = () => {
  videoInputRef.value?.click()
}

const handleImageChange = async (event: Event) => {
  const input = event.target as HTMLInputElement
  if (input.files?.length) {
    const file = input.files[0]
    
    // 检查文件大小（限制为 10MB）
    if (file.size > 10 * 1024 * 1024) {
      showToast('图片大小不能超过10MB')
      input.value = ''
      return
    }
    
    try {
      // 获取图片尺寸
    //   const dimensions = await getImageDimensions(file)
      
      // 先创建本地预览URL
      const localUrl = URL.createObjectURL(file)
      
      // 发送临时消息
      emit('select-file', file, MessageType.Image, {
        url: localUrl,
        size: file.size,
        uploading: true // 标记为上传中
      })
      
      emit('update:modelValue', false)
      
      // 上传文件
    //   showLoadingToast({
    //     message: '正在上传...',
    //     forbidClick: true,
    //   })
      
      const result = await uploadFiles(file)
      
      // 发送上传成功的消息
      emit('upload-success', file, MessageType.Image, {
        ...result[0],
      })
      
    } catch (error) {
      console.error('图片上传失败:', error)
      showToast('上传失败')
      emit('upload-error', file, MessageType.Image)
    } finally {
      closeToast()
      input.value = '' // 清空选择，允许选择相同文件
    }
  }
}

const handleVideoChange = async (event: Event) => {
  const input = event.target as HTMLInputElement
  if (input.files?.length) {
    const file = input.files[0]
    
    // 检查文件大小（限制为 50MB）
    if (file.size > 50 * 1024 * 1024) {
      showToast('视频大小不能超过50MB')
      input.value = ''
      return
    }
    
    try {
      // 创建本地预览URL
      const localUrl = URL.createObjectURL(file)
      
      // 发送临时消息
      emit('select-file', file, MessageType.Video, {
        url: localUrl,
        size: file.size,
        format: file.type.split('/')[1],
        uploading: true // 标记为上传中
      })
      
      emit('update:modelValue', false)
      
      // 上传文件
      showLoadingToast({
        message: '正在上传...',
        forbidClick: true,
      })
      
      const result = await uploadFiles(file)
      
      // 发送上传成功的消息
      emit('upload-success', file, MessageType.Video, result[0])
      
    } catch (error) {
      console.error('上传视频失败:', error)
      showToast('上传失败')
      emit('upload-error', file, MessageType.Video)
    } finally {
      closeToast()
      input.value = '' // 清空选择，允许选择相同文件
    }
  }
}
</script>

<style scoped>
.message-more-options {
  position: relative;
  background: white;
  padding: 0 var(--spacing-base);
  transform: translateY(100%);
  opacity: 0;
  height: 0;
  visibility: hidden;
  transition: all 0.3s ease;
}

.message-more-options.is-active {
  transform: translateY(0);
  opacity: 1;
  visibility: visible;
  height: auto;
}

.options-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--spacing-base);
  padding: var(--spacing-base) 0;
}

.option-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-mini);
  cursor: pointer;
}

.option-icon {
  width: 50px;
  height: 50px;
  border-radius: 12px;
  background: var(--bg-color);
  display: flex;
  align-items: center;
  justify-content: center;
}

.option-icon i {
  font-size: 24px;
  color: var(--text-color-dark);
}

.option-label {
  font-size: 12px;
  color: var(--text-color-dark);
}
</style> 