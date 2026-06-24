const REFRESH_THRESHOLD_MS = 5 * 60 * 1000

interface JwtPayload {
  exp?: number
  iat?: number
  sub?: string
}

function decodeBase64Url(value: string): string {
  const normalized = value.replace(/-/g, '+').replace(/_/g, '/')
  const padding = normalized.length % 4 === 0 ? '' : '='.repeat(4 - (normalized.length % 4))
  return atob(normalized + padding)
}

export function parseJwtPayload(token: string): JwtPayload | null {
  const parts = token.split('.')
  if (parts.length !== 3) {
    return null
  }

  try {
    const payload = JSON.parse(decodeBase64Url(parts[1]!)) as JwtPayload
    return payload
  } catch {
    return null
  }
}

export function getTokenExpiryMs(token: string): number | null {
  const payload = parseJwtPayload(token)
  if (typeof payload?.exp !== 'number') {
    return null
  }

  return payload.exp * 1000
}

export function isTokenExpiringSoon(
  token: string,
  thresholdMs = REFRESH_THRESHOLD_MS,
): boolean {
  const expiryMs = getTokenExpiryMs(token)
  if (expiryMs == null) {
    return true
  }

  return expiryMs - Date.now() <= thresholdMs
}

export { REFRESH_THRESHOLD_MS }
