<template>
  <div class="page">
    <div class="nav-bar">
      <div class="nav-bar-left">
        <div class="nav-bar-title">消息</div>
      </div>
      <div class="nav-bar-right">
        <div class="nav-bar-right-item">
          <button class="btn btn-primary" type="button">
            <i class="ri-more-line"></i>
          </button>
        </div>
      </div>
    </div>

    <ConversationList navigate-mode="push" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useIMStore } from '@/stores/im'
import { MessageType, type Message } from '@/sdk/im'
import { usePageTitleNotification } from '@/composables/usePageTitleNotification'
import ConversationList from '@/components/im/ConversationList.vue'

const router = useRouter()
const userStore = useUserStore()
const imStore = useIMStore()
const { notifyNewMessage } = usePageTitleNotification('消息')

const onNewMessageForTitle = (message: Message) => {
  const myId = userStore.userInfo?.id
  if (!myId) return
  if (message.type === MessageType.Ping || message.type === MessageType.Pong) return
  if (message.to_id === myId && message.from_id && message.from_id !== myId) {
    notifyNewMessage('消息')
  }
}

onMounted(() => {
  if (!userStore.token) {
    router.replace({ name: 'im-login', query: { redirect: '/im/home' } })
    return
  }

  imStore.initSDK()
  imStore.addMessageHandler(onNewMessageForTitle)
})

onUnmounted(() => {
  imStore.removeMessageHandler(onNewMessageForTitle)
})
</script>

<style scoped>
.page {
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
  background: white;
}

.nav-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-base);
  background: white;
  flex-shrink: 0;
}

.nav-bar-left {
  flex: 1;
}

.nav-bar-right {
  flex: 0;
}

.nav-bar-right-item {
  display: flex;
  gap: var(--spacing-mini);
}

.nav-bar-title {
  font-weight: 600;
}
</style>
