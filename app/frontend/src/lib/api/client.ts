import { env } from '$env/dynamic/public'
import { logApi } from '$lib/utils/logger'

export class ApiError extends Error {
  status: number
  constructor(message: string, status: number) {
    super(message)
    this.status = status
    this.name = 'ApiError'
  }
}

async function request<T>(
  endpoint: string,
  token: string | undefined,
  options: RequestInit = {}
): Promise<T> {
  const method = options.method || 'GET'
  const startTime = Date.now()

  try {
    const response = await fetch(`${env.PUBLIC_API_BASE_URL}${endpoint}`, {
      ...options,
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
        ...options.headers
      }
    })

    const durationMs = Date.now() - startTime

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      const errorMessage = errorData.error || 'Request failed'

      logApi('RESPONSE_ERROR', {
        method,
        endpoint,
        status: response.status,
        error: errorMessage,
        durationMs
      })

      throw new ApiError(errorMessage, response.status)
    }

    logApi('RESPONSE_SUCCESS', {
      method,
      endpoint,
      status: response.status,
      durationMs
    })

    return response.json()
  } catch (err) {
    // If it's already an ApiError, rethrow (already logged)
    if (err instanceof ApiError) {
      throw err
    }

    // Network error or other fetch failure
    const durationMs = Date.now() - startTime
    logApi('NETWORK_ERROR', {
      method,
      endpoint,
      error: err instanceof Error ? err.message : 'Unknown error',
      durationMs
    })

    throw err
  }
}

export const api = {
  get: <T>(endpoint: string, token?: string) =>
    request<T>(endpoint, token),

  post: <T>(endpoint: string, body: unknown, token?: string) =>
    request<T>(endpoint, token, { method: 'POST', body: JSON.stringify(body) }),

  put: <T>(endpoint: string, body: unknown, token?: string) =>
    request<T>(endpoint, token, { method: 'PUT', body: JSON.stringify(body) }),

  patch: <T>(endpoint: string, body: unknown, token?: string) =>
    request<T>(endpoint, token, { method: 'PATCH', body: JSON.stringify(body) }),

  delete: <T>(endpoint: string, token?: string) =>
    request<T>(endpoint, token, { method: 'DELETE' })
}
