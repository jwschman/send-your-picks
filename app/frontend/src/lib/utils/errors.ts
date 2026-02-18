import { toasts } from '$lib/stores/toast'
import { logApi } from '$lib/utils/logger'

export function handleApiError(err: unknown, fallbackMessage = 'An error occurred'): string {
  const message = err instanceof Error ? err.message : fallbackMessage
  toasts.error(message)

  logApi('RESPONSE_ERROR', {
    error: message,
    status: (err as { status?: number })?.status
  })

  return message
}
