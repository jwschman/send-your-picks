import { redirect } from '@sveltejs/kit'
import type { PageServerLoad } from './$types'

export const load: PageServerLoad = async ({ locals: { safeGetSession } }) => {
  const { session } = await safeGetSession()
  
  // If the user is already logged in, redirect them to dashboard
  if (session) {
    redirect(303, '/dashboard')
  }

  return {}
}