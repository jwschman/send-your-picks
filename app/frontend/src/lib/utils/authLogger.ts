// Auth logging utility for debugging authentication issues
// Logs to console (appears in container logs) with structured format

type AuthEventType =
  | 'LOGIN_ATTEMPT'
  | 'LOGIN_SUCCESS'
  | 'LOGIN_FAILURE'
  | 'SIGNUP_ATTEMPT'
  | 'SIGNUP_SUCCESS'
  | 'SIGNUP_FAILURE'
  | 'SESSION_VALID'
  | 'SESSION_INVALID'
  | 'SESSION_EXPIRED'
  | 'JWT_ERROR'
  | 'TOKEN_REFRESH'
  | 'LOGOUT'

interface AuthLogEntry {
  timestamp: string
  event: AuthEventType
  email?: string
  error?: string
  details?: Record<string, unknown>
}

function formatLog(entry: AuthLogEntry): string {
  const parts = [
    `[AUTH]`,
    `[${entry.timestamp}]`,
    `[${entry.event}]`
  ]

  if (entry.email) {
    parts.push(`email=${entry.email}`)
  }

  if (entry.error) {
    parts.push(`error="${entry.error}"`)
  }

  if (entry.details && Object.keys(entry.details).length > 0) {
    parts.push(`details=${JSON.stringify(entry.details)}`)
  }

  return parts.join(' ')
}

export function logAuth(
  event: AuthEventType,
  options: {
    email?: string
    error?: string | Error
    details?: Record<string, unknown>
  } = {}
): void {
  const entry: AuthLogEntry = {
    timestamp: new Date().toISOString(),
    event,
    email: options.email,
    error: options.error instanceof Error ? options.error.message : options.error,
    details: options.details
  }

  // Use console.error for failures so they stand out in logs
  const isError = event.includes('FAILURE') || event.includes('ERROR') || event.includes('INVALID') || event.includes('EXPIRED')

  if (isError) {
    console.error(formatLog(entry))
  } else {
    console.log(formatLog(entry))
  }
}

// Helper to extract useful JWT info for logging (without exposing sensitive data)
export function getJwtDebugInfo(token: string | undefined): Record<string, unknown> | null {
  if (!token) return null

  try {
    const parts = token.split('.')
    if (parts.length !== 3) {
      return { error: 'Invalid JWT format', parts: parts.length }
    }

    const payload = JSON.parse(
      typeof atob === 'function'
        ? atob(parts[1])
        : Buffer.from(parts[1], 'base64').toString('utf-8')
    )

    const now = Math.floor(Date.now() / 1000)
    const iat = payload.iat
    const exp = payload.exp

    return {
      issued_at: iat ? new Date(iat * 1000).toISOString() : null,
      expires_at: exp ? new Date(exp * 1000).toISOString() : null,
      current_time: new Date(now * 1000).toISOString(),
      seconds_until_expiry: exp ? exp - now : null,
      seconds_since_issued: iat ? now - iat : null,
      is_expired: exp ? now > exp : null,
      issued_in_future: iat ? iat > now : null,  // This catches time sync issues!
      user_role: payload.user_role || payload.role || null,
      sub: payload.sub ? `${payload.sub.substring(0, 8)}...` : null  // Truncated user ID
    }
  } catch (e) {
    return { error: 'Failed to parse JWT', message: e instanceof Error ? e.message : 'unknown' }
  }
}
