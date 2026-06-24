<template>
  <div class="conversation-list-root" :class="{ 'is-embedded': embedded }">
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
        :class="{ 'is-active': activePeerId === item.peerId }"
        @click="selectConversation(item.peerId)"
      >
        <div class="avatar">
          <img v-if="item.avatar" :src="item.avatar" alt="">
          <PlaceholderImage
            v-else
            bgColor="#EFF1F8"
            text=""
            aspect="1:1"
          />
          <span v-if="item.unreadCount > 0" class="unread-badge">
            {{ item.unreadBadge }}
          </span>
        </div>

        <div class="conversation-info">
          <div class="conversation-header">
            <span class="nickname">{{ item.displayName }}</span>
            <span class="time">{{ item.time }}</span>
          </div>
          <div class="last-message">{{ item.lastMessagePreview }}</div>
        </div>

        <div v-if="item.previewImage" class="conversation-image">
          <img :src="item.previewImage" alt="">
        </div>
      </div>

      <div v-if="conversationItems.length === 0" class="empty-state">
        <i class="ri-chat-3-line"></i>
        <p>暂无消息</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
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

const props = withDefaults(
  defineProps<{
    /** 当前聊天对象，用于高亮 */
    activePeerId?: string
    /** 嵌入聊天页侧栏时使用（仅影响样式） */
    embedded?: boolean
    /** 点击会话后的路由行为：home 用 push，侧栏用 replace */
    navigateMode?: 'push' | 'replace' | 'none'
  }>(),
  {
    activePeerId: '',
    embedded: false,
    navigateMode: 'push',
  },
)

const emit = defineEmits<{
  select: [peerId: string]
}>()

const router = useRouter()
const imStore = useIMStore()

const {
  conversations,
  loading,
  error,
  currentUserId,
  loadConversations,
  handleIncomingMessage,
  clearUnreadForPeer,
} = useConversationList()

const { userMap, fetchUsers, mergeUsers } = useUserProfiles()

const conversationItems = computed(() => {
  const uid = currentUserId.value

  return conversations.value.map((conversation) => {
    const peerId = getPeerUserId(conversation, uid)
    const profile = conversation.to_user_info ?? userMap.value[peerId]
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

const selectConversation = (peerId: string) => {
  if (!peerId || peerId === props.activePeerId) return

  clearUnreadForPeer(peerId)
  emit('select', peerId)

  if (props.navigateMode === 'none') return

  const location = { name: 'im-chat' as const, params: { userId: peerId } }
  if (props.navigateMode === 'replace') {
    router.replace(location)
    return
  }

  router.push(location)
}

const onIncomingMessage = async (message: Parameters<typeof handleIncomingMessage>[0]) => {
  const peerIds = [message.from_id, message.to_id].filter(
    (id): id is string => Boolean(id) && id !== currentUserId.value,
  )
  await fetchUsers(peerIds)
  handleIncomingMessage(message, props.activePeerId || undefined)
}

const refresh = async () => {
  await loadConversations()
  mergeUsers(conversations.value.map((conversation) => conversation.to_user_info))
}

onMounted(async () => {
  imStore.initSDK()
  imStore.addMessageHandler(onIncomingMessage)
  await refresh()
})

onUnmounted(() => {
  imStore.removeMessageHandler(onIncomingMessage)
})

watch(
  () => imStore.imSDK,
  (sdk, prev) => {
    if (sdk && !prev && conversations.value.length === 0 && !loading.value) {
      void refresh()
    }
  },
)

defineExpose({ refresh })
</script>

<style scoped>
.conversation-list-root {
  display: flex;
  flex-direction: column;
  min-height: 0;
  flex: 1;
}

.conversation-list-root.is-embedded {
  background: var(--bg-color);
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
  flex: 1;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
}

.conversation-item {
  display: flex;
  align-items: flex-start;
  padding: 10px var(--spacing-base);
  background: transparent;
  cursor: pointer;
  transition: background-color 0.15s ease;
}

.is-embedded .conversation-item:hover {
  background: var(--bg-color-gray);
}

.is-embedded .conversation-item.is-active {
  background: var(--bg-color-gray);
}

.conversation-list-root:not(.is-embedded) .conversation-item {
  background: white;
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
  padding-bottom: 2px;
}

.conversation-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.nickname {
  font-weight: 500;
  font-size: var(--font-size-base);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.time {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--text-color-light);
}

.conversation-image {
  margin-left: var(--spacing-small);
  background-color: var(--bg-color);
  overflow: hidden;
  width: 48px;
  max-height: 48px;
  border-radius: 8px;
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
  font-size: 14px;
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
