import { useUserStore } from '@/stores/user'

const CHECK_INTERVAL_MS = 60_000

let checkTimer: ReturnType<typeof setInterval> | null = null
let started = false

async function checkAndRefresh() {
  if (document.visibilityState !== 'visible') {
    return
  }

  const userStore = useUserStore()
  await userStore.ensureValidToken({ logoutOnAuthError: true })
}

function onVisibilityChange() {
  if (document.visibilityState === 'visible') {
    void checkAndRefresh()
  }
}

export function startTokenRefresh() {
  if (started) {
    return
  }

  started = true
  document.addEventListener('visibilitychange', onVisibilityChange)
  checkTimer = setInterval(() => {
    void checkAndRefresh()
  }, CHECK_INTERVAL_MS)
  void checkAndRefresh()
}

export function stopTokenRefresh() {
  if (!started) {
    return
  }

  started = false
  document.removeEventListener('visibilitychange', onVisibilityChange)

  if (checkTimer) {
    clearInterval(checkTimer)
    checkTimer = null
  }
}
