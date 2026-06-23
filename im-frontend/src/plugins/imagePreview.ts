import { showImagePreview, type ImagePreviewOptions } from 'vant'

type PreviewOptions = Omit<ImagePreviewOptions, 'images'> & {
  images: string[]
}

let instance: any = null

export const imagePreview = {
  preview(options: PreviewOptions) {
    const defaultOptions = {
      showIndex: true,
      closeable: true,
      closeIconPosition: 'top-right' as const,
      swipeDuration: 300,
      maxZoom: 3,
      minZoom: 1/3,
      closeOnPopstate: true,
      startPosition: 0
    }

    // 合并配置
    const finalOptions = {
      ...defaultOptions,
      ...options,
      onClose: () => {
        options.onClose?.()
        instance = null
      }
    }

    // 关闭已存在的实例
    if (instance) {
      instance.close()
    }

    // 打开预览
    instance = showImagePreview(finalOptions)
    return instance
  },

  close() {
    if (instance) {
      instance.close()
      instance = null
    }
  },

  swipeTo(index: number, immediate?: boolean) {
    if (instance) {
      instance.swipeTo(index, immediate)
    }
  }
}

export default {
  install(app: any) {
    app.config.globalProperties.$imagePreview = imagePreview
  }
} 