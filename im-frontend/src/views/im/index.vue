<template>
  <RouterView />
</template>

<script setup lang="ts">
import { computed, onUnmounted, watch } from 'vue'
import { RouterView } from 'vue-router'
import { useIMStore } from '@/stores/im'
import { useIMTabStore } from '@/stores/imTab'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const imStore = useIMStore()
const imTabStore = useIMTabStore()
const currentUserId = computed(() => userStore.userInfo?.id ?? '')

watch(
  currentUserId,
  (userId) => {
    if (userId) {
      imTabStore.init(userId)
    }
  },
  { immediate: true },
)

watch(
  () => imTabStore.isPrimaryTab,
  (isPrimary, wasPrimary) => {
    if (!imTabStore.initialized || isPrimary === wasPrimary) return
    if (!isPrimary) {
      imStore.closeConnection()
      return
    }
    imStore.initSDK()
  },
)

onUnmounted(() => {
  imTabStore.reset()
})
</script>
