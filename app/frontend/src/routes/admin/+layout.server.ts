import { redirect, error } from '@sveltejs/kit'
import type { LayoutServerLoad } from './$types'

function getRoleFromAccessToken(token: string): string | null {
  try {
    const payload = token.split('.')[1]
    const decoded = JSON.parse(Buffer.from(payload, 'base64').toString())
    return decoded.user_role || decoded.role || null
  } catch {
    return null
  }
}

export const load: LayoutServerLoad = async ({ locals }) => {
  const { session } = await locals.safeGetSession()

  if (!session) {
    throw redirect(303, '/login')
  }

  const role = getRoleFromAccessToken(session.access_token)

  if (role !== 'admin') {
    throw error(403, 'Insufficient permissions')
  }

  return {
    userRole: role
  }
}
