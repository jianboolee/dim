<!-- 图片消息组件 -->
<template>
  <div class="message-content message-image">
    <div class="message-arrow"></div>
    <div 
      class="image-container" 
      :style="containerStyle"
    >
      <div class="image-wrapper">
        <ImageView
          :src="message.media_info?.url || ''"
          :alt="message.content"
          placeholderText="图片"
          :width="containerStyle.width"
          @click="preview"
          @load="handleImageLoad"
        />
        
        <!-- 上传中状态 -->
        <div v-if="message.media_info?.uploading" class="upload-overlay">
          <div class="upload-spinner"></div>
        </div>
      </div>
    </div>
    
    <MessageStatus 
      :status="message.status" 
      :is-mine="isMine"
      @retry="$emit('retry')"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, inject, ref, type Ref } from 'vue'
import type { Message } from '@/sdk/im'
import MessageStatus from './MessageStatus.vue'
import ImageView from '@/components/common/ImageView.vue'
import { imagePreview } from '@/plugins/imagePreview'

const props = defineProps<{
  message: Message
  isMine: boolean
}>()

defineEmits<{
  (e: 'retry'): void
}>()

// 注入所有图片URL列表，确保是 ref 类型
const chatImages = inject<Ref<string[]>>('chatImages', ref([]))

// 实际图片尺寸
const actualWidth = ref(0)
const actualHeight = ref(0)

// 处理图片加载完成
const handleImageLoad = (event: Event) => {
  const img = event.target as HTMLImageElement
  actualWidth.value = img.naturalWidth
  actualHeight.value = img.naturalHeight
}

// 获取有效的宽高数据
const getValidDimensions = () => {
  const mediaInfo = props.message.media_info
  
  // 优先使用加载后的实际尺寸
  if (actualWidth.value && actualHeight.value) {
    return {
      width: actualWidth.value,
      height: actualHeight.value
    }
  }
  
  // 其次使用 media_info 中的尺寸
  if (mediaInfo?.width && mediaInfo?.height) {
    return {
      width: mediaInfo.width,
      height: mediaInfo.height
    }
  }
  
  // 最后使用默认尺寸 (4:3)
  return {
    width: 400,
    height: 300
  }
}

// 计算容器样式
const containerStyle = computed(() => {
  const { width, height } = getValidDimensions()
  const aspectRatio = (height / width) * 100
  const maxWidth = width > height ? '180px' : '120px'
  
  return {
    'padding-bottom': `${aspectRatio}%`,
    'width': `${width}px`,
    'max-width': maxWidth,
    'min-width': '100px',
    'min-height': 'auto'
  }
})

// 预览图片
const preview = () => {
  if (!props.message.media_info?.url) return
  // 找到当前图片在列表中的索引
  const index = chatImages.value.findIndex((url: string) => url === props.message.media_info?.url)
  if (index === -1) return
  
  imagePreview.preview({
      images: chatImages.value,
      startPosition: index >= 0 ? index : 0,
      showIndex: chatImages.value.length > 1
    })
}
</script>

<style scoped>
.message-content.message-image {
  padding: 0;
  overflow: hidden;
  background: transparent;
}

.image-container {
  position: relative;
  width: 100%;
  min-width: 120px;
  min-height: auto;
}

.image-wrapper {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 上传中的遮罩层 */
.upload-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
}

.upload-spinner {
  width: 24px;
  height: 24px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.message-arrow {
  position: absolute;
  top: 12px;
  left: -6px;
  width: 0;
  height: 0;
  border-style: solid;
  border-width: 6px 6px 6px 0;
  border-color: transparent var(--bg-color) transparent transparent;
}

:deep(.message-mine) .message-content {
  background: var(--primary-color);
}

:deep(.message-mine) .message-arrow {
  left: auto;
  right: -6px;
  border-width: 6px 0 6px 6px;
  border-color: transparent transparent transparent var(--primary-color);
}

.message-mine .message-arrow {
  left: auto;
  right: -6px;
  border-width: 6px 0 6px 6px;
  border-color: transparent transparent transparent var(--bg-color);
}

</style> 
