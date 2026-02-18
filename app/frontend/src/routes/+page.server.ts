import { redirect } from '@sveltejs/kit'
import type { PageServerLoad } from './$types'

export const load: PageServerLoad = async ({ url, locals: { supabase, safeGetSession } }) => {
  // Handle PKCE code exchange (fallback for old emails that redirect to root)
  const code = url.searchParams.get('code')
  if (code) {
    const { error } = await supabase.auth.exchangeCodeForSession(code)
    if (!error) {
      redirect(303, '/dashboard')
    }
    // If code exchange failed, redirect to error page
    redirect(303, '/auth/error')
  }

  const { session } = await safeGetSession()

  if (session) {
    redirect(303, '/dashboard')
  } else {
    redirect(303, '/login')
  }
}