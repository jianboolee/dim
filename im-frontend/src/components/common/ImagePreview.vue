<!-- 图片预览组件 -->
<template>
  <teleport to="body">
    <div class="image-preview-container">
      <!-- 预览内容由 vant 注入 -->
    </div>
  </teleport>
</template>

<script setup lang="ts">
import { showImagePreview } from 'vant'
import { ref } from 'vue'

interface PreviewOptions {
  images: string[]  // 图片数组
  startPosition?: number  // 起始位置
  showIndex?: boolean  // 是否显示索引
  closeable?: boolean  // 是否显示关闭按钮
  closeIconPosition?: string  // 关闭按钮位置
  swipeDuration?: number  // 动画时长
  onClose?: () => void  // 关闭回调
  onChange?: (index: number) => void  // 切换回调
  maxZoom?: number
  minZoom?: number
  closeOnPopstate?: boolean
  className?: string
}

// 当前预览实例
const instance = ref<any>(null)

// 预览方法
const preview = (options: PreviewOptions) => {
  const defaultOptions = {
    showIndex: true,
    closeable: true,
    closeIconPosition: 'top-right',
    swipeDuration: 300,
    maxZoom: 3,
    minZoom: 1/3,
    closeOnPopstate: true,
    startPosition: 0,
    className: 'custom-image-preview',
  }

  // 合并配置
  const finalOptions = {
    ...defaultOptions,
    ...options,
    onClose: () => {
      options.onClose?.()
      instance.value = null
    }
  }

  // 打开预览
  instance.value = showImagePreview({
    ...finalOptions,
    onScale: (index: number, scale: number) => {
      console.log('scale:', scale)
    },
    onSwipe: (swipe: string) => {
      if (swipe === 'down') {
        close()
      }
    }
  })

  return instance.value
}

// 关闭预览
const close = () => {
  if (instance.value) {
    instance.value.close()
    instance.value = null
  }
}

// 切换到指定索引
const swipeTo = (index: number, immediate?: boolean) => {
  if (instance.value) {
    instance.value.swipeTo(index, immediate)
  }
}

// 导出方法
defineExpose({
  preview,
  close,
  swipeTo
})
</script>

<style>
:root {
  --van-image-preview-index-text-color: white;
  --van-image-preview-index-font-size: 14px;
  --van-image-preview-index-line-height: 1.2;
  --van-image-preview-index-text-shadow: 0 1px 1px rgba(0, 0, 0, 0.3);
}

.custom-image-preview {
  --van-image-preview-close-icon-size: 22px;
  --van-image-preview-close-icon-color: white;
  --van-image-preview-close-icon-margin: 16px;
  --van-image-preview-close-icon-z-index: 2;
}

.custom-image-preview .van-image-preview__index {
  padding: 4px 10px;
  border-radius: 12px;
  background: rgba(0, 0, 0, 0.3);
  font-weight: 500;
}

.custom-image-preview .van-image-preview__image {
  will-change: transform;
  transition: transform 0.3s;
}

.custom-image-preview .van-image-preview__overlay {
  background: rgba(0, 0, 0, 0.9);
}

.custom-image-preview .van-image-preview__close-icon {
  opacity: 0.8;
  transition: opacity 0.2s;
}

.custom-image-preview .van-image-preview__close-icon:active {
  opacity: 1;
}

.custom-image-preview .van-swipe {
  height: 100vh !important;
}
</style> 