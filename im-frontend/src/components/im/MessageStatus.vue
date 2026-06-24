<!-- 消息状态组件 -->
<template>
  <div class="message-status">
    <i v-if="status === 'sending'" class="bi bi-clock"></i>
    <i v-else-if="status === 'failed'" class="bi bi-exclamation-circle" @click="$emit('retry')"></i>
    <i v-else-if="status === 'sent'" class="bi bi-check"></i>
    <i v-else-if="status === 'read'" class="bi bi-check-all"></i>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  status?: 'sending' | 'failed' | 'sent' | 'delivered' | 'read'
}>()

defineEmits<{
  (e: 'retry'): void
}>()
</script>

<style scoped>
.message-status {
  position: absolute;
  right: 6px;
  bottom: 6px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.7);
}

.message-status i {
  cursor: pointer;
}

.message-status i.bi-exclamation-circle {
  color: var(--error-color);
}
</style> 
