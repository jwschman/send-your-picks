// src/hooks.server.ts
import { env } from '$env/dynamic/public'
import { createServerClient } from '@supabase/ssr'
import type { Handle } from '@sveltejs/kit'
import { logAuth, getJwtDebugInfo } from '$lib/utils/authLogger'

export const handle: Handle = async ({ event, resolve }) => {
  event.locals.supabase = createServerClient(env.PUBLIC_SUPABASE_URL, env.PUBLIC_SUPABASE_PUBLISHABLE_KEY, {
    cookies: {
      getAll: () => event.cookies.getAll(),
      /**
       * Note: You have to add the `path` variable to the
       * set and remove method due to sveltekit's cookie API
       * requiring this to be set, setting the path to `/`
       * will replicate previous/standard behaviour (https://kit.svelte.dev/docs/types#public-types-cookies)
       */
      setAll: (cookiesToSet) => {
        cookiesToSet.forEach(({ name, value, options }) => {
          event.cookies.set(name, value, { ...options, path: '/' })
        })
      },
    },
  })

  /**
   * Unlike `supabase.auth.getSession`, which is unsafe on the server because it
   * doesn't validate the JWT, this function validates the JWT by first calling
   * `getUser` and aborts early if the JWT signature is invalid.
   */
  event.locals.safeGetSession = async () => {
    // Get the raw session first to extract token for debugging
    const {
      data: { session: rawSession },
    } = await event.locals.supabase.auth.getSession()

    const accessToken = rawSession?.access_token

    const {
      data: { user },
      error,
    } = await event.locals.supabase.auth.getUser()

    if (error) {
      // Only log if there was actually a token that failed validation
      // Don't log "Auth session missing!" for unauthenticated visitors - that's expected
      if (accessToken) {
        const jwtInfo = getJwtDebugInfo(accessToken)

        // Determine the type of error
        let eventType: 'JWT_ERROR' | 'SESSION_EXPIRED' | 'SESSION_INVALID' = 'JWT_ERROR'
        if (error.message?.toLowerCase().includes('expired')) {
          eventType = 'SESSION_EXPIRED'
        } else if (error.message?.toLowerCase().includes('invalid')) {
          eventType = 'SESSION_INVALID'
        }

        logAuth(eventType, {
          email: rawSession?.user?.email,
          error: error,
          details: {
            path: event.url.pathname,
            errorCode: error.code,
            errorStatus: error.status,
            jwt: jwtInfo,
            serverTime: new Date().toISOString()
          }
        })
      }

      return { session: null, user: null }
    }

    // Log successful session validations only for authenticated routes (reduce noise)
    // Uncomment if you want verbose logging:
    // if (user) {
    //   logAuth('SESSION_VALID', {
    //     email: user.email,
    //     details: { path: event.url.pathname }
    //   })
    // }

    const {
      data: { session },
    } = await event.locals.supabase.auth.getSession()
    return { session, user }
  }

  return resolve(event, {
    filterSerializedResponseHeaders(name: string) {
      return name === 'content-range' || name === 'x-supabase-api-version'
    },
  })
}