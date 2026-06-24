import { defineStore } from 'pinia'
import { ref } from 'vue'

const DEFAULT_TITLE = '消息'
const NOTIFY_PREFIX = '【有新消息】'

export const usePageNotificationStore = defineStore('pageNotification', () => {
  const baseTitle = ref(DEFAULT_TITLE)
  const hasUnreadNotice = ref(false)
  const mounted = ref(false)

  function applyTitle() {
    if (typeof document === 'undefined') return
    document.title = hasUnreadNotice.value
      ? `${NOTIFY_PREFIX}${baseTitle.value}`
      : baseTitle.value
  }

  function setBaseTitle(title: string) {
    baseTitle.value = title.trim() || DEFAULT_TITLE
    applyTitle()
  }

  function notifyNewMessage() {
    if (typeof document === 'undefined') return
    if (document.visibilityState === 'visible') return

    hasUnreadNotice.value = true
    applyTitle()
  }

  function clearNotice() {
    hasUnreadNotice.value = false
    applyTitle()
  }

  function handleVisibilityChange() {
    if (document.visibilityState === 'visible') {
      clearNotice()
    }
  }

  function mount() {
    if (mounted.value || typeof document === 'undefined') return

    mounted.value = true
    document.addEventListener('visibilitychange', handleVisibilityChange)
    window.addEventListener('focus', clearNotice)
    applyTitle()
  }

  function unmount() {
    if (!mounted.value || typeof document === 'undefined') return

    mounted.value = false
    document.removeEventListener('visibilitychange', handleVisibilityChange)
    window.removeEventListener('focus', clearNotice)
  }

  return {
    baseTitle,
    hasUnreadNotice,
    setBaseTitle,
    notifyNewMessage,
    clearNotice,
    mount,
    unmount,
  }
})

