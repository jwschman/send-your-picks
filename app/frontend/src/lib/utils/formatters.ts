// Date/time formatters

/** "Jan 5, 2025" */
export function formatShortDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  })
}

/** "Thu, Jan 5" */
export function formatWeekdayDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric'
  })
}

/** "Thu, Jan 5, 3:00 PM" */
export function formatWeekdayDateTime(dateString: string): string {
  return new Date(dateString).toLocaleString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit'
  })
}

/** "Jan 5, 2025, 3:00 PM" */
export function formatDateTime(dateString: string): string {
  return new Date(dateString).toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: 'numeric',
    minute: '2-digit'
  })
}

/** "3:00 PM" */
export function formatTime(dateString: string): string {
  return new Date(dateString).toLocaleTimeString('en-US', {
    hour: 'numeric',
    minute: '2-digit'
  })
}

/** "1/5/2025, 3:00:00 PM" or "Never" */
export function formatLastSignIn(timestamp: string | null): string {
  if (!timestamp) return 'Never'
  return new Date(timestamp).toLocaleString()
}

// Team helpers

/** "Kansas City Chiefs" */
export function getFullTeamName(city: string, name: string): string {
  return `${city} ${name}`
}

// Avatar utilities

export const DEFAULT_AVATAR = '/images/avatars/default-avatar.jpg'

/** "J" or "?" */
export function getAvatarInitial(username: string | null): string {
  if (!username) return '?'
  return username.charAt(0).toUpperCase()
}
