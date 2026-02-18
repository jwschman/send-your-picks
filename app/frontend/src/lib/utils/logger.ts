// Structured logging utilities for debugging production issues
// Logs to console (appears in container logs) with structured format

type LogLevel = 'info' | 'warn' | 'error'

interface LogEntry {
  timestamp: string
  category: string
  event: string
  level: LogLevel
  details?: Record<string, unknown>
}

function formatLog(entry: LogEntry): string {
  const parts = [
    `[${entry.category}]`,
    `[${entry.timestamp}]`,
    `[${entry.event}]`
  ]

  if (entry.details && Object.keys(entry.details).length > 0) {
    // Format details inline for readability
    const detailParts = Object.entries(entry.details).map(([k, v]) => {
      if (typeof v === 'string') return `${k}="${v}"`
      if (typeof v === 'object') return `${k}=${JSON.stringify(v)}`
      return `${k}=${v}`
    })
    parts.push(detailParts.join(' '))
  }

  return parts.join(' ')
}

function log(category: string, event: string, level: LogLevel, details?: Record<string, unknown>): void {
  const entry: LogEntry = {
    timestamp: new Date().toISOString(),
    category,
    event,
    level,
    details
  }

  const message = formatLog(entry)

  switch (level) {
    case 'error':
      console.error(message)
      break
    case 'warn':
      console.warn(message)
      break
    default:
      console.log(message)
  }
}

// API logging
export type ApiEventType =
  | 'REQUEST'
  | 'RESPONSE_SUCCESS'
  | 'RESPONSE_ERROR'
  | 'NETWORK_ERROR'

export function logApi(
  event: ApiEventType,
  details: {
    method?: string
    endpoint?: string
    status?: number
    error?: string
    durationMs?: number
  }
): void {
  const level: LogLevel = event.includes('ERROR') ? 'error' : 'info'
  log('API', event, level, details)
}

// Picks logging
export type PicksEventType =
  | 'PAGE_LOAD'
  | 'PICK_SELECTED'
  | 'PICK_DESELECTED'
  | 'SUBMIT_ATTEMPT'
  | 'SUBMIT_SUCCESS'
  | 'SUBMIT_FAILURE'
  | 'GAME_LOCKED'
  | 'VALIDATION_ERROR'

export function logPicks(
  event: PicksEventType,
  details?: {
    weekId?: string
    seasonId?: string
    gameId?: string
    team?: string
    pickCount?: number
    lockedCount?: number
    error?: string
  }
): void {
  const level: LogLevel = event.includes('FAILURE') || event.includes('ERROR') ? 'error' : 'info'
  log('PICKS', event, level, details)
}

// Commissioner/Week state logging
export type WeekEventType =
  | 'PAGE_LOAD'
  | 'ACTIVATION_ATTEMPT'
  | 'ACTIVATION_SUCCESS'
  | 'ACTIVATION_FAILURE'
  | 'SPREADS_IMPORT_ATTEMPT'
  | 'SPREADS_IMPORT_SUCCESS'
  | 'SPREADS_IMPORT_FAILURE'
  | 'SPREADS_SAVE_ATTEMPT'
  | 'SPREADS_SAVE_SUCCESS'
  | 'SPREADS_SAVE_FAILURE'
  | 'SPREADS_VALIDATION_ERROR'
  | 'STATE_TRANSITION'

export function logWeek(
  event: WeekEventType,
  details?: {
    weekId?: string
    seasonId?: string
    status?: string
    newStatus?: string
    gameCount?: number
    spreadCount?: number
    error?: string
    source?: string
  }
): void {
  const level: LogLevel = event.includes('FAILURE') || event.includes('ERROR') ? 'error' : 'info'
  log('WEEK', event, level, details)
}

// User/profile logging
export type UserEventType =
  | 'INFO_FETCH_SUCCESS'
  | 'INFO_FETCH_FAILURE'
  | 'PROFILE_FETCH_SUCCESS'
  | 'PROFILE_FETCH_FAILURE'

export function logUser(
  event: UserEventType,
  details?: {
    error?: string
    hasSession?: boolean
  }
): void {
  const level: LogLevel = event.includes('FAILURE') ? 'error' : 'info'
  log('USER', event, level, details)
}

// Standings/data logging
export type DataEventType =
  | 'STANDINGS_LOAD_SUCCESS'
  | 'STANDINGS_LOAD_FAILURE'
  | 'STANDINGS_UNAVAILABLE'
  | 'CHART_RENDER_SUCCESS'
  | 'CHART_RENDER_FAILURE'

export function logData(
  event: DataEventType,
  details?: {
    seasonId?: string
    weekId?: string
    reason?: string
    error?: string
  }
): void {
  const level: LogLevel = event.includes('FAILURE') ? 'error' : 'info'
  log('DATA', event, level, details)
}
