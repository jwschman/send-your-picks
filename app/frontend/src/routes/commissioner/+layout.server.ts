import { redirect, error } from '@sveltejs/kit'
import type { LayoutServerLoad } from './$types'
import { getRoleFromToken, isCommissioner } from '$lib/utils/auth'

export const load: LayoutServerLoad = async ({ locals }) => {
  const { session } = await locals.safeGetSession()

  if (!session) {
    throw redirect(303, '/login')
  }

  const role = getRoleFromToken(session.access_token)

  if (!isCommissioner(role)) {
    throw error(403, 'Insufficient permissions')
  }

  return {
    userRole: role
  }
}
