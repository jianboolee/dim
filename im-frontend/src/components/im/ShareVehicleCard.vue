<template>
  <Transition
    enter-active-class="slide-enter-active"
    leave-active-class="slide-leave-active"
    @after-leave="$emit('update:modelValue', false)"
  >
    <div v-if="show && vehicleInfo" class="share-vehicle-card">
      <div class="vehicle-info">
        <img :src="vehicleInfo?.images[0]" class="vehicle-image" />
        <div class="vehicle-details">
          <div class="vehicle-title">
            <div class="vehicle-title-text">
              {{ vehicleInfo?.description.slice(0, 50) }}
            </div>
            <button class="close-btn" @click="show = false">
              <i class="bi bi-x-lg"></i>
            </button>
          </div>
          <div class="vehicle-footer">
            <div class="vehicle-price">
              ¥ {{ vehicleInfo?.price }}
            </div>
            <van-button 
              type="primary" 
              class="btn send-btn"
              block 
              @click="handleSend"
              :loading="sending"
            >
              发送到聊天
            </van-button>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRoute } from 'vue-router'
import { request } from '@/utils/request'
import { showToast } from 'vant'
import { Button as VanButton } from 'vant'
import type { CardInfo } from '@/sdk/im'
import { config } from '@/config'
const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'send', cardInfo: CardInfo): void
}>()

const route = useRoute()
const vehicleInfo = ref<any>(null)
const sending = ref(false)

// 计算属性用于控制弹窗显示
const show = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

// 获取车辆详情
const fetchVehicleInfo = async (id: string) => {
  try {
    const response = await request(`/api/used/rvs/${id}`)
    if (response.code === 200) {
      vehicleInfo.value = response.data
    }
  } catch (error) {
    console.error('获取车辆详情失败:', error)
    showToast('获取车辆信息失败')
  }
}

// 发送卡片消息
const handleSend = async () => {
  if (!vehicleInfo.value) return
  
  sending.value = true
  try {
    const cardInfo: CardInfo = {
      title: vehicleInfo.value.title || vehicleInfo.value.description.slice(0, 50),
      description: vehicleInfo.value.brand + ' ' + vehicleInfo.value.year + '年 ' + vehicleInfo.value.mileage + '公里',
      path: `/detail/${vehicleInfo.value.id}`,
      image_url: vehicleInfo.value.images[0] || '',
      price: vehicleInfo.value.price,
      currency: 'CNY'
    }
    
    emit('send', cardInfo)
    show.value = false
  } catch (error) {
    console.error('发送失败:', error)
    showToast('发送失败')
  } finally {
    sending.value = false
  }
}

// 监听路由参数变化
watch(
  () => route.query,
  (query) => {
    if (query.from === 'detail' && query.id) {
      show.value = true
      fetchVehicleInfo(query.id as string)
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.share-vehicle-card {
  position: fixed;
  top: 56px; /* nav-bar 高度 + 1px border */
  left: 16px;
  right: 16px;
  background: white;
  z-index: 99;
  border-bottom: 1px solid var(--border-color-light);
  padding: 0 0;
  border-radius: 12px;
  box-shadow: 0 4px 8px 0 rgba(0, 0, 0, 0.2);
  transform-origin: top center;
  will-change: transform, opacity;
}

.close-btn {
  border: none;
  background: none;
  padding: var(--spacing-mini);
  cursor: pointer;
  color: var(--text-color-light);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: -2px;
}

.close-btn:hover {
  color: var(--text-color-dark);
}

.vehicle-info {
  display: flex;
  padding: var(--spacing-small) var(--spacing-base);
}

.vehicle-image {
  width: 60px;
  height: 60px;
  object-fit: cover;
  border-radius: 4px;
  margin-right: var(--spacing-base);
}

.vehicle-details {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.vehicle-title {
  font-size: 12px;
  font-weight: 500;
  line-height: 1.4;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--spacing-small);
  margin-bottom: var(--spacing-small);
}

.vehicle-title-text {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  word-break: break-all;
  flex: 1;
  min-width: 0; /* 确保文本可以正确换行 */
}

.vehicle-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-small);
}

.vehicle-price {
  font-size: 14px;
  color: var(--price-color);
  font-weight: bold;
}

.send-btn {
  font-size: 12px;
  font-weight: 400;
  height: 28px;
  padding: 4px var(--spacing-base);
  background: var(--primary-color);
  border:none;
  border-radius: 12px;
  color: white;
  display: inline-block;
  width: auto;
}

/* 添加动画相关样式 */
.slide-enter-active {
  animation: slideIn 0.3s ease-out;
}

.slide-leave-active {
  animation: fadeOut 0.2s ease-in;
}

@keyframes slideIn {
  from {
    transform: translateY(-100%);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

@keyframes fadeOut {
  from {
    opacity: 1;
  }
  to {
    opacity: 0;
  }
}
</style>
