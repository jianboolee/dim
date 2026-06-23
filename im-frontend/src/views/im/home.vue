<template>
  <div class="page">
    <div class="nav-bar">
      <div class="nav-bar-left">
        <div class="nav-bar-title">消息</div>
      </div>
      <div class="nav-bar-right">
        <div class="nav-bar-right-item">
          <button class="btn btn-primary" type="button">
            <i class="bi bi-three-dots"></i>
          </button>
        </div>
      </div>
    </div>

    <div v-if="loading" class="state-block">
      <div class="spinner"></div>
    </div>

    <div v-else-if="error" class="state-block state-error">
      <p>{{ error }}</p>
      <button type="button" class="btn btn-default" @click="refresh">重试</button>
    </div>

    <div v-else class="conversation-list">
      <div
        v-for="item in conversationItems"
        :key="item.id"
        class="conversation-item"
        @click="goToChat(item.peerId)"
      >
        <div class="avatar">
          <img v-if="item.avatar" :src="item.avatar" alt="">
          <PlaceholderImage
            v-else
            bgColor="#EFF1F8"
            text=""
            aspect="1:1"
          />
          <span v-if="item.unreadCount" class="unread-badge">
            {{ item.unreadBadge }}
          </span>
        </div>

        <div class="conversation-info">
          <div class="conversation-header">
            <span class="nickname">{{ item.displayName }}</span>
          </div>
          <div class="last-message">{{ item.lastMessagePreview }}</div>
          <div class="time">{{ item.time }}</div>
        </div>

        <div v-if="item.previewImage" class="conversation-image">
          <img :src="item.previewImage" alt="">
        </div>
      </div>

      <div v-if="conversationItems.length === 0" class="empty-state">
        <i class="bi bi-chat"></i>
        <p>暂无消息</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useIMStore } from '@/stores/im'
import { useConversationList } from '@/composables/useConversationList'
import { useUserProfiles } from '@/composables/useUserProfiles'
import {
  formatConversationTime,
  formatLastMessagePreview,
  formatUnreadBadge,
} from '@/utils/im/format'
import { getPeerUserId, getUnreadCount } from '@/utils/im/conversation'
import PlaceholderImage from '@/components/common/PlaceholderImage.vue'

const router = useRouter()
const userStore = useUserStore()
const imStore = useIMStore()

const {
  conversations,
  loading,
  error,
  currentUserId,
  loadConversations,
  handleIncomingMessage,
  getPeerUserIds,
} = useConversationList()

const { userMap, fetchUsers } = useUserProfiles()

const conversationItems = computed(() => {
  const uid = currentUserId.value

  return conversations.value.map((conversation) => {
    const peerId = getPeerUserId(conversation, uid)
    const profile = userMap.value[peerId]
    const unreadCount = getUnreadCount(conversation, uid)

    return {
      id: conversation.id,
      peerId,
      avatar: profile?.avatar,
      displayName: profile?.nickname ?? (peerId ? `用户${peerId.slice(-4)}` : '未知用户'),
      lastMessagePreview: formatLastMessagePreview(conversation.last_message),
      time: formatConversationTime(
        conversation.last_message?.created_at ?? conversation.updated_at,
      ),
      unreadCount,
      unreadBadge: formatUnreadBadge(unreadCount),
      previewImage: conversation.image_url,
    }
  })
})

const goToChat = (peerId: string) => {
  if (!peerId) return
  router.push({ name: 'im-chat', params: { userId: peerId } })
}

const onIncomingMessage = async (message: Parameters<typeof handleIncomingMessage>[0]) => {
  const peerIds = [message.from_id, message.to_id].filter(
    (id): id is string => Boolean(id) && id !== currentUserId.value,
  )
  await fetchUsers(peerIds)
  handleIncomingMessage(message)
}

const refresh = async () => {
  await loadConversations()
  await fetchUsers(getPeerUserIds())
}

onMounted(async () => {
  if (!userStore.token) {
    router.replace({ name: 'im-login', query: { redirect: '/im/home' } })
    return
  }

  imStore.initSDK()
  imStore.addMessageHandler(onIncomingMessage)

  await refresh()
})

onUnmounted(() => {
  imStore.removeMessageHandler(onIncomingMessage)
})
</script>

<style scoped>
.page {
  min-height: 100dvh;
  background: white;
}

.nav-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-base);
  background: white;
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

.state-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-base);
  padding: 48px var(--spacing-base);
  color: var(--text-color-light);
}

.state-error p {
  margin: 0;
  color: var(--error-color);
}

.spinner {
  width: 24px;
  height: 24px;
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

.conversation-list {
  padding: 0;
}

.conversation-item {
  display: flex;
  align-items: flex-start;
  padding: var(--spacing-base);
  background: white;
  cursor: pointer;
}

.avatar {
  position: relative;
  width: 48px;
  height: 48px;
  margin-right: var(--spacing-base);
  background: var(--bg-color);
  border-radius: 50%;
  flex-shrink: 0;
}

.avatar img {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
}

.unread-badge {
  position: absolute;
  top: -4px;
  right: -4px;
  min-width: 18px;
  height: 18px;
  padding: 0 4px;
  background: var(--error-color);
  color: white;
  border-radius: 9px;
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.conversation-info {
  flex: 1;
  min-width: 0;
}

.conversation-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.nickname {
  font-weight: 500;
  font-size: var(--font-size-base);
}

.time {
  margin-top: 4px;
  font-size: 12px;
  color: var(--text-color-light);
}

.conversation-image {
  margin-left: var(--spacing-small);
  background-color: var(--bg-color);
  overflow: hidden;
  width: 60px;
  max-height: 60px;
  border-radius: 12px;
  aspect-ratio: 1 / 1;
  flex-shrink: 0;
}

.conversation-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.last-message {
  color: var(--text-color-secondary);
  font-size: 13px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px var(--spacing-base);
  color: var(--text-color-light);
}

.empty-state i {
  font-size: 48px;
  margin-bottom: var(--spacing-base);
}

.empty-state p {
  font-size: var(--font-size-base);
  margin: 0;
}
</style>
