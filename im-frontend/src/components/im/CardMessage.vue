<!-- 卡片消息组件 -->
<template>
  <div class="message-content message-card" :class="{ 'message-failed': message.status === 'failed' }">
    <div class="message-arrow"></div>
    <div @click="handleCardClick(message)" class="card-link">
      <div class="card-thumb">
        <template v-if="!imageError">
          <img 
            v-if="message.card_info?.image_url"
            :src="message.card_info?.image_url" 
            alt=""
            @error="handleImageError(message)"
          >
          <PlaceholderImage 
            v-else
            text=""
            bgColor="#EFF1F8"
            aspect="1:1"
          />
        </template>
        <PlaceholderImage 
          v-else
          text=""
          bgColor="#EFF1F8"
          aspect="1:1"
        />
      </div>
      <div class="card-info">
        <div class="card-title">{{ message.card_info?.title || '' }}</div>
        <div class="card-desc">{{ message.card_info?.description }}</div>
        <div class="card-price" v-if="message.card_info?.price">
          <span class="currency">{{ message.card_info?.currency === 'CNY' ? '¥' : message.card_info?.currency }}</span>
          <span class="amount">{{ formatPrice(message.card_info?.price) }}</span>
        </div>
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
import { ref } from 'vue'
import type { Message } from '@/types/im'
import MessageStatus from './MessageStatus.vue'
import PlaceholderImage from '../common/PlaceholderImage.vue'
import { useRouter } from 'vue-router'

const router = useRouter()

defineProps<{
  message: Message
  isMine: boolean
}>()

defineEmits<{
  (e: 'retry'): void
}>()

const imageError = ref(false)

// 图片加载失败处理
const handleImageError = (message: Message) => {
  imageError.value = true
  if (message.card_info?.image_url) {
    message.card_info.image_url = ''  // 清空图片，触发显示占位图
  }
}

// 格式化价格
const formatPrice = (price: number) => {
  return price.toLocaleString('zh-CN', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  })
}

const handleCardClick = (message: Message) => {
  if (message.card_info?.path) {
    router.push(message.card_info?.path)
  }
}
</script>

<style scoped>
.message-content {
  position: relative;
  border-radius: 12px;
  background: var(--bg-color);
  max-width: 180px;
  width: 80%;
  padding: 0;
}

.message-content.message-card {
  max-width: 180px;
  width: 80%;
  padding: 0;
  overflow: hidden;
  background: var(--bg-color);
}

.card-link {
  display: flex;
  flex-direction: column;
  text-decoration: none;
  border: 1px solid var(--border-color-light);
  border-radius: 12px;
  overflow: hidden;
  background: var(--bg-color);
}

.card-thumb {
  width: 100%;
  aspect-ratio: 1/1;
  background: #EFF1F8;
  overflow: hidden;
}

.card-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.card-info {
  padding: 8px 12px;
  flex: 1;
  min-width: 0;
}

.card-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-color-dark);
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  line-height: 1.4;
}

.card-desc {
  font-size: 12px;
  color: var(--text-color-light);
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
}

.card-price {
  font-size: 14px;
  color: var(--price-color);
  font-weight: 600;
  display: flex;
  align-items: baseline;
  gap: 2px;
}

.card-price .currency {
  font-size: 12px;
  font-weight: normal;
}

.card-price .amount {
  font-family: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
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

.message-mine .message-content {
  background: var(--bg-color);
}

.message-mine .message-arrow {
  left: auto;
  right: -6px;
  border-width: 6px 0 6px 6px;
  border-color: transparent transparent transparent var(--primary-color);
}

.message-failed {
  opacity: 0.7;
}

.message-mine .message-failed {
  background: var(--error-color-light);
}


</style> 