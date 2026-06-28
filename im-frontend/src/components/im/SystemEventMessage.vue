<template>
  <div class="system-event-message">
    {{ text }}
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ChatMessage } from '@/types/im'
import { useUserProfiles } from '@/composables/useUserProfiles'
import { formatSystemEventMessage } from '@/utils/im/systemEvent'

const props = defineProps<{
  message: ChatMessage
}>()

const { userMap } = useUserProfiles()

const text = computed(() => formatSystemEventMessage(props.message, userMap.value))
</script>

<style scoped>
.system-event-message {
  width: fit-content;
  max-width: min(76%, 520px);
  margin: 8px auto 18px;
  padding: 3px 8px;
  border-radius: 10px;
  color: #8a93a3;
  font-size: 12px;
  line-height: 1.45;
  text-align: center;
  word-break: break-word;
}
</style>
