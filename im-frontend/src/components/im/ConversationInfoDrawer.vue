<template>
  <Teleport to="body">
    <Transition name="drawer-fade">
      <div
        v-if="modelValue"
        class="conversation-info-mask"
        @click.self="close"
      >
        <Transition name="drawer-slide" appear>
          <aside
            class="conversation-info-drawer"
            role="dialog"
            aria-modal="true"
            aria-label="会话信息"
            tabindex="-1"
            @keydown.esc="close"
          >
            <!-- <header class="drawer-header">
              <h2>会话信息</h2>
              <button type="button" class="drawer-close-btn" aria-label="关闭" @click="close">
                <i class="ri-close-line"></i>
              </button>
            </header> -->

            <section class="drawer-section participants-section">
              <div v-if="isGroup && groupId" class="group-meta-row">
                <span>群 ID</span>
                <strong>{{ groupId }}</strong>
              </div>
              <div class="participant-list">
                <div
                  v-for="user in displayParticipants"
                  :key="user.id"
                  class="participant-item"
                >
                  <img class="participant-avatar" :src="user.avatar || ''" alt="">
                  <span class="participant-name">{{ user.nickname || user.id }}</span>
                </div>
              </div>
            </section>

            <section class="drawer-section action-section">
              <button v-if="isGroup" type="button" class="drawer-row action-row" @click="handleInvite">
                <span>邀请成员</span>
                <i class="ri-add-line"></i>
              </button>
              <button type="button" class="drawer-row action-row" @click="handleSearch">
                <span>查找聊天内容</span>
                <i class="ri-arrow-right-s-line"></i>
              </button>
            </section>

            <section class="drawer-section setting-section">
              <div class="drawer-row switch-row">
                <span>置顶聊天</span>
                <button
                  type="button"
                  class="switch-control"
                  :class="{ 'is-on': pinned }"
                  :aria-pressed="pinned"
                  @click="pinned = !pinned"
                >
                  <span></span>
                </button>
              </div>
              <div class="drawer-row switch-row">
                <span>消息免打扰</span>
                <button
                  type="button"
                  class="switch-control"
                  :class="{ 'is-on': muted }"
                  :aria-pressed="muted"
                  @click="muted = !muted"
                >
                  <span></span>
                </button>
              </div>
            </section>

            <section class="drawer-section danger-section">
              <button type="button" class="clear-history-btn" @click="handleClear">
                清空聊天记录
              </button>
            </section>
          </aside>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import type { UserInfo } from '@/types/user'

const props = defineProps<{
  modelValue: boolean
  participants: UserInfo[]
  isGroup?: boolean
  groupId?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  search: []
  clear: []
  invite: []
}>()

const pinned = ref(false)
const muted = ref(false)

const displayParticipants = computed(() => props.participants.filter((user) => user.id))

const close = () => {
  emit('update:modelValue', false)
}

const handleSearch = () => {
  emit('search')
}

const handleInvite = () => {
  emit('invite')
}

const handleClear = () => {
  emit('clear')
}

watch(
  () => props.modelValue,
  (visible) => {
    document.body.style.overflow = visible ? 'hidden' : ''
  },
)

onUnmounted(() => {
  document.body.style.overflow = ''
})
</script>

<style scoped>
.conversation-info-mask {
  position: fixed;
  inset: 0;
  z-index: 1100;
  display: flex;
  justify-content: flex-end;
  background: rgba(15, 23, 42, 0);
}

.conversation-info-drawer {
  width: min(360px, 100vw);
  height: 100dvh;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  background: #f7f8fb;
  outline: none;
  border-left: 1px solid var(--border-color-light);
}

.drawer-header {
  min-height: 52px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  border-bottom: 1px solid var(--border-color-light);
  background: #fff;
}

.drawer-header h2 {
  margin: 0;
  color: var(--text-color-dark);
  font-size: 15px;
  font-weight: 600;
  line-height: 1.3;
}

.drawer-close-btn {
  width: 34px;
  height: 34px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--text-color-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.drawer-close-btn:hover {
  background: #f0f3f8;
  color: var(--text-color-dark);
}

.drawer-close-btn i {
  font-size: 20px;
  line-height: 1;
}

.drawer-section {
  background: #fff;
  border-bottom: 1px solid var(--border-color-light);
}

.participants-section {
  padding: 18px 16px 16px;
}

.group-meta-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
  color: var(--text-color-secondary);
  font-size: 12px;
}

.group-meta-row strong {
  min-width: 0;
  color: var(--text-color-dark);
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.participant-list {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
}

.participant-item {
  width: 64px;
  min-width: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 7px;
}

.participant-avatar {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  object-fit: cover;
  background: #edf1f7;
}

.participant-name {
  width: 100%;
  color: var(--text-color-secondary);
  font-size: 12px;
  line-height: 1.25;
  text-align: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.action-section,
.setting-section,
.danger-section {
  border-bottom: 1px solid var(--border-color-light);
}

.drawer-row {
  width: 100%;
  min-height: 54px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 0 16px;
  border: none;
  background: #fff;
  color: var(--text-color-dark);
  font-size: 13px;
  line-height: 1.35;
  text-align: left;
}

.action-row {
  cursor: pointer;
}

.action-row i {
  flex-shrink: 0;
  color: #a2a9b6;
  font-size: 22px;
  line-height: 1;
}

.switch-row {
  cursor: default;
}

.switch-row + .switch-row {
  border-top: 1px solid var(--border-color-light);
}

.switch-control {
  width: 36px;
  height: 22px;
  flex-shrink: 0;
  padding: 2px;
  border: none;
  border-radius: 999px;
  background: #e8ebf0;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.switch-control span {
  width: 18px;
  height: 18px;
  display: block;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 1px 4px rgba(15, 23, 42, 0.18);
  transition: transform 0.2s ease;
}

.switch-control.is-on {
  background: #07c160;
}

.switch-control.is-on span {
  transform: translateX(14px);
}

.danger-section {
  padding: 14px 16px;
}

.clear-history-btn {
  width: 100%;
  min-height: 42px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--danger-color);
  font-size: 13px;
  line-height: 1.35;
  cursor: pointer;
}

.clear-history-btn:hover {
  background: rgba(255, 77, 79, 0.08);
}

.drawer-fade-enter-active,
.drawer-fade-leave-active {
  transition: opacity 0.18s ease;
}

.drawer-fade-enter-from,
.drawer-fade-leave-to {
  opacity: 0;
}

.drawer-slide-enter-active,
.drawer-slide-leave-active {
  transition: transform 0.22s ease;
}

.drawer-slide-enter-from,
.drawer-slide-leave-to {
  transform: translateX(100%);
}

@media (max-width: 767px) {
  .conversation-info-mask {
    background: rgba(15, 23, 42, 0.18);
  }

  .conversation-info-drawer {
    width: min(88vw, 360px);
  }
}
</style>
