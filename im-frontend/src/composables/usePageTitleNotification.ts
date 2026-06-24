import { onMounted, onUnmounted, ref } from 'vue'

const NOTIFY_PREFIX = '【有新消息】'

export function usePageTitleNotification(defaultTitle = '消息') {
  const baseTitle = ref(defaultTitle)
  let hasNotification = false

  const restoreTitle = () => {
    document.title = baseTitle.value
    hasNotification = false
  }

  const setBaseTitle = (title: string) => {
    const next = title.trim() || defaultTitle
    baseTitle.value = next
    if (!hasNotification) {
      document.title = next
    }
  }

  const notifyNewMessage = (suffix = '') => {
    if (document.visibilityState === 'visible') {
      return
    }

    document.title = suffix ? `${NOTIFY_PREFIX}${suffix}` : NOTIFY_PREFIX
    hasNotification = true
  }

  const onVisibilityChange = () => {
    if (document.visibilityState === 'visible') {
      restoreTitle()
    }
  }

  onMounted(() => {
    setBaseTitle(defaultTitle)
    document.addEventListener('visibilitychange', onVisibilityChange)
    window.addEventListener('focus', restoreTitle)
  })

  onUnmounted(() => {
    document.removeEventListener('visibilitychange', onVisibilityChange)
    window.removeEventListener('focus', restoreTitle)
    restoreTitle()
  })

  return {
    notifyNewMessage,
    restoreTitle,
    setBaseTitle,
  }
}
