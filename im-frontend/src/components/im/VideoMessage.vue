<!-- 视频消息组件 -->
<template>
  <div class="message-content message-video" :class="{ 'message-failed': message.status === 'failed' }">
    <div class="message-arrow"></div>
    <div class="video-container" @click="playVideo">
      <video 
        ref="videoRef"
        class="video-js vjs-default-skin"
        preload="none"
        :poster="message.media_info?.thumbnail"
        playsinline
        webkit-playsinline
      >
        <source 
          v-for="(type, index) in getVideoSources(message.media_info)" 
          :key="index"
          :src="type.src"
          :type="type.type"
        />
        <p class="vjs-no-js">
          您的设备不支持视频播放
        </p>
      </video>
      <div class="video-overlay" v-if="!isPlaying">
        <i class="bi bi-play-circle-fill"></i>
        <span class="duration" v-if="message.media_info?.duration">
          {{ formatDuration(message.media_info.duration) }}
        </span>
      </div>
      <div v-if="loading" class="loading-overlay">
        <div class="spinner"></div>
      </div>
      <div v-if="error" class="error-overlay">
        <i class="bi bi-exclamation-circle"></i>
        <span>视频加载失败</span>
      </div>
    </div>
    <MessageStatus 
      v-if="isMine" 
      :status="message.status" 
      @retry="$emit('retry')" 
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import type { Message } from '@/types/im'
import MessageStatus from './MessageStatus.vue'
import videojs from 'video.js'
import 'video.js/dist/video-js.css'

const props = defineProps<{
  message: Message
  isMine: boolean
}>()

defineEmits<{
  (e: 'retry'): void
}>()

const videoRef = ref<HTMLVideoElement | null>(null)
const player = ref<any>(null)
const isPlaying = ref(false)
const loading = ref(true)
const error = ref(false)

// 触摸相关变量
let touchStartY = 0
let touchEndY = 0

// 处理触摸开始
const handleTouchStart = (e: TouchEvent) => {
  const touch = e.touches[0]
  if (!touch) return
  touchStartY = touch.clientY
}

// 处理触摸结束
const handleTouchEnd = (e: TouchEvent) => {
  const touch = e.changedTouches[0]
  if (!touch) return
  touchEndY = touch.clientY
  const deltaY = touchEndY - touchStartY
  
  // 如果是向下滑动超过50px
  if (deltaY > 50) {
    // 退出全屏
    if (document.fullscreenElement) {
      document.exitFullscreen()
    }
    // 暂停播放
    player.value?.pause()
    isPlaying.value = false
  }
}

// 播放视频
const playVideo = () => {
  if (!player.value || error.value) return
  
  // 如果已经在全屏状态
  if (document.fullscreenElement) {
    // 切换播放/暂停状态
    if (player.value.paused()) {
      player.value.play()
    } else {
      player.value.pause()
    }
    return
  }
  
  // 如果不在全屏状态，则进入全屏并播放
  const videoElement = player.value.el()
  if (videoElement.requestFullscreen) {
    videoElement.requestFullscreen().then(() => {
      player.value.play()
    })
  }
  
  // 添加触摸事件监听
  videoElement.addEventListener('touchstart', handleTouchStart)
  videoElement.addEventListener('touchend', handleTouchEnd)
}

// 监听全屏变化
const handleFullscreenChange = () => {
  if (!document.fullscreenElement && player.value) {
    // 退出全屏时暂停播放
    player.value?.pause()
    isPlaying.value = false
    
    // 移除触摸事件监听
    const videoElement = player.value.el()
    videoElement.removeEventListener('touchstart', handleTouchStart)
    videoElement.removeEventListener('touchend', handleTouchEnd)
  }
}

// 格式化时长
const formatDuration = (seconds: number) => {
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = Math.floor(seconds % 60)
  return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`
}

// 获取视频源
const getVideoSources = (mediaInfo: any) => {
  if (!mediaInfo?.url) return []
  
  const url = mediaInfo.url
  const format = mediaInfo.format?.toLowerCase()
  const sources = []
  
  // 添加原始格式
  if (format) {
    sources.push({
      src: url,
      type: `video/${format === 'mov' ? 'quicktime' : format}`
    })
  }
  
  // 如果是 MOV 格式，添加 MP4 作为备选
  if (format === 'mov') {
    const mp4Url = url.replace(/\.[^.]+$/, '.mp4')
    sources.push({
      src: mp4Url,
      type: 'video/mp4'
    })
  }
  
  // 如果没有识别出格式，使用默认的 MP4
  if (!format) {
    sources.push({
      src: url,
      type: 'video/mp4'
    })
  }
  
  return sources
}

onMounted(() => {
  if (!videoRef.value) return

  // 初始化 Video.js 播放器
  player.value = videojs(videoRef.value, {
    controls: true,
    autoplay: false,
    preload: 'metadata',
    fluid: true,
    playsinline: false,
    controlBar: {
      playToggle: true,
      currentTimeDisplay: false,
      timeDivider: false,
      durationDisplay: false,
      progressControl: true,
      volumePanel: false,
      fullscreenToggle: true,
      pictureInPictureToggle: false
    },
    techOrder: ['html5'],
    sources: getVideoSources(props.message.media_info)
  })

  // 监听播放器事件
  player.value.on('loadedmetadata', () => {
    loading.value = false
    error.value = false
  })

  // 监听播放状态变化
  player.value.on('play', () => {
    isPlaying.value = true
  })

  player.value.on('pause', () => {
    isPlaying.value = false
  })

  // 监听全屏变化
  player.value.on('fullscreenchange', () => {
    const isFullscreen = player.value.isFullscreen()
    player.value.controls(isFullscreen)
  })

  // 添加全屏变化监听
  document.addEventListener('fullscreenchange', handleFullscreenChange)
})

onUnmounted(() => {
  if (player.value) {
    const videoElement = player.value.el()
    videoElement.removeEventListener('touchstart', handleTouchStart)
    videoElement.removeEventListener('touchend', handleTouchEnd)
    document.removeEventListener('fullscreenchange', handleFullscreenChange)
    player.value.dispose()
  }
})

// 计算视频宽高比
const getAspectRatio = computed(() => {
  const { width, height } = props.message.media_info || {}
  if (width && height) {
    // 如果是横屏视频
    if (width > height) {
      return '16/9'
    }
    // 如果是竖屏视频
    return '9/16'
  }
  // 默认竖屏比例
  return '9/16'
})
</script>

<style scoped>
.message-content {
  position: relative;
  padding: 4px;
  border-radius: 12px;
  background: var(--bg-color);
  max-width: 120px;
}

.message-content.message-video {
  max-width: 120px;
  width: 80%;
  padding: 0;
}

.video-container {
  position: relative;
  width: 100%;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  background: #000;
  aspect-ratio: v-bind('getAspectRatio');
}

:deep(.video-js) {
  width: 100%;
  height: 100%;
  border-radius: 8px;
}

:deep(.vjs-poster) {
  background-size: cover;
}

:deep(.vjs-big-play-button) {
  display: none;
}

:deep(.vjs-control-bar) {
  background-color: rgba(0, 0, 0, 0.5);
  height: 2.5em;
}

.video-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1;
}

.video-overlay i {
  font-size: 48px;
  color: white;
  opacity: 0.8;
}

.duration {
  position: absolute;
  bottom: 8px;
  right: 8px;
  font-size: 12px;
  color: white;
  background: rgba(0, 0, 0, 0.5);
  padding: 2px 4px;
  border-radius: 4px;
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2;
}

.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid var(--border-color);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
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
  background: var(--bg-color);
}

:deep(.message-mine) .message-arrow {
  left: auto;
  right: -6px;
  border-width: 6px 0 6px 6px;
  border-color: transparent transparent transparent var(--bg-color);
}

.message-mine .message-arrow {
  left: auto;
  right: -6px;
  border-width: 6px 0 6px 6px;
  border-color: transparent transparent transparent var(--bg-color);
}

.message-failed {
  background: var(--bg-color-light);
}

:deep(.message-mine) .message-failed {
  background: var(--error-color-light);
}

.error-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  z-index: 3;
  color: white;
}

.error-overlay i {
  font-size: 32px;
  margin-bottom: 8px;
}

.error-overlay span {
  font-size: 12px;
}

/* 非全屏时隐藏控制条 */
:deep(.video-js:not(.vjs-fullscreen) .vjs-control-bar) {
  display: none !important;
}

/* 全屏时显示控制条 */
:deep(.video-js.vjs-fullscreen .vjs-control-bar) {
  display: flex !important;
  opacity: 0;
  transition: opacity 0.3s;
}

:deep(.video-js.vjs-fullscreen:hover .vjs-control-bar) {
  opacity: 1;
}

/* 确保视频在全屏时填满屏幕 */
:deep(.video-js.vjs-fullscreen video) {
  object-fit: contain;
}
</style> 
