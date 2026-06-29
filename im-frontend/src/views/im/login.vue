<template>
  <div class="login-page">
    <div class="content">
      <i class="ri-chat-3-line icon"></i>
      <h1>即时消息</h1>
      <p>请从业务系统进入聊天，IM 不提供独立登录。</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const redirect = computed(() =>
  typeof route.query.redirect === 'string' ? route.query.redirect : '',
)
</script>

<style scoped>
.login-page {
  min-height: 100dvh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: #f5f6fa;
}

.content {
  max-width: 360px;
  text-align: center;
  color: #333;
}

.icon {
  font-size: 48px;
  color: #4b86f8;
  margin-bottom: 16px;
}

h1 {
  font-size: 22px;
  margin: 0 0 12px;
}

p {
  margin: 0;
  color: #666;
  line-height: 1.6;
}

.hint {
  margin-top: 16px;
  font-size: 13px;
  color: #999;
  word-break: break-all;
}
</style>
