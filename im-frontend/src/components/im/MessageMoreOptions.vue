<!-- 消息更多选项组件 -->
<template>
  <div class="message-more-options" :class="{ 'is-active': modelValue }">
    <div class="options-grid">
      <div class="option-item" @click="openImagePicker">
        <div class="option-icon">
          <i class="ri-image-line"></i>
        </div>
        <div class="option-label">图片</div>
        <input
          ref="imageInputRef"
          type="file"
          accept=".jpg,.jpeg,.png,.gif,.webp,image/jpeg,image/png,image/gif,image/webp"
          class="hidden-input"
          @change="handleImageChange"
        >
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { showToast } from '@/plugins/toast'
import { MessageType, type MediaInfo } from '@/sdk/im'
import { takeInputFile, readImageDimensions, getFileFormat } from '@/utils/file'
import { uploadIMFile } from '@/utils/upload'

const IMAGE_MAX_SIZE = 10 * 1024 * 1024
const IMAGE_ALLOWED_EXTENSIONS = new Set(['jpg', 'jpeg', 'png', 'gif', 'webp'])

type PendingMediaInfo = MediaInfo & { uploading?: boolean }

defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'select-file': [file: File, type: MessageType, fileInfo: PendingMediaInfo]
  'upload-success': [file: File, type: MessageType, fileInfo: MediaInfo]
  'upload-error': [file: File, type: MessageType]
}>()

const imageInputRef = ref<HTMLInputElement | null>(null)

function openImagePicker() {
  imageInputRef.value?.click()
}

function closePanel() {
  emit('update:modelValue', false)
}

function assertFileSize(file: File, maxBytes: number, message: string): boolean {
  if (file.size <= maxBytes) {
    return true
  }
  showToast(message)
  return false
}

function getFileExtension(file: File): string {
  const name = file.name.trim().toLowerCase()
  const ext = name.includes('.') ? name.split('.').pop() : ''
  return ext ?? ''
}

function assertImageExtension(file: File): boolean {
  if (IMAGE_ALLOWED_EXTENSIONS.has(getFileExtension(file))) {
    return true
  }
  showToast('仅支持 JPG、PNG、GIF、WEBP 图片')
  return false
}

async function uploadImage(file: File) {
  const localUrl = URL.createObjectURL(file)
  const format = getFileFormat(file)

  emit('select-file', file, MessageType.Image, {
    url: localUrl,
    size: file.size,
    format,
    width: 0,
    height: 0,
    uploading: true,
  })
  closePanel()

  try {
    const uploaded = await uploadIMFile(file)
    const dimensions = await readImageDimensions(file)

    const mediaInfo: MediaInfo = {
      url: uploaded.url,
      size: uploaded.size,
      width: uploaded.width ?? dimensions.width,
      height: uploaded.height ?? dimensions.height,
      format: uploaded.format ?? format,
    }

    emit('upload-success', file, MessageType.Image, mediaInfo)
  } catch (error) {
    console.error('图片上传失败:', error)
    showToast('上传失败')
    emit('upload-error', file, MessageType.Image)
  }
}

async function handleImageChange(event: Event) {
  const file = takeInputFile(event)
  if (!file) return
  if (!assertImageExtension(file)) return
  if (!assertFileSize(file, IMAGE_MAX_SIZE, '图片大小不能超过10MB')) return

  await uploadImage(file)
}
</script>

<style scoped>
.message-more-options {
  overflow: hidden;
  max-height: 0;
  opacity: 0;
  border-top: 1px solid transparent;
  transition: max-height 0.25s ease, opacity 0.2s ease, border-color 0.2s ease;
}

.message-more-options.is-active {
  max-height: 140px;
  opacity: 1;
  border-top-color: var(--border-color-light);
}

.options-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--spacing-base);
  padding: var(--spacing-base);
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

.hidden-input {
  display: none;
}
</style>
