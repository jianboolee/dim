type ToastOptions = {
  duration?: number
}

let toastElement: HTMLDivElement | null = null
let toastTimer: ReturnType<typeof window.setTimeout> | null = null

function ensureToastElement() {
  if (toastElement) return toastElement

  toastElement = document.createElement('div')
  toastElement.className = 'app-toast'
  toastElement.setAttribute('role', 'status')
  toastElement.setAttribute('aria-live', 'polite')
  document.body.appendChild(toastElement)

  if (!document.getElementById('app-toast-style')) {
    const style = document.createElement('style')
    style.id = 'app-toast-style'
    style.textContent = `
      .app-toast {
        position: fixed;
        left: 50%;
        bottom: max(28px, env(safe-area-inset-bottom));
        z-index: 9999;
        max-width: min(320px, calc(100vw - 48px));
        padding: 9px 14px;
        border-radius: 8px;
        background: rgba(22, 24, 29, 0.9);
        color: #fff;
        font-size: 14px;
        line-height: 1.45;
        text-align: center;
        pointer-events: none;
        opacity: 0;
        transform: translate(-50%, 8px);
        transition: opacity 0.18s ease, transform 0.18s ease;
        word-break: break-word;
      }

      .app-toast.is-visible {
        opacity: 1;
        transform: translate(-50%, 0);
      }
    `
    document.head.appendChild(style)
  }

  return toastElement
}

export function showToast(message: string, options: ToastOptions = {}) {
  if (typeof document === 'undefined') return

  const element = ensureToastElement()
  element.textContent = message

  if (toastTimer) {
    window.clearTimeout(toastTimer)
    toastTimer = null
  }

  window.requestAnimationFrame(() => {
    element.classList.add('is-visible')
  })

  toastTimer = window.setTimeout(() => {
    element.classList.remove('is-visible')
  }, options.duration ?? 2000)
}

