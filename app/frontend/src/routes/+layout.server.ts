// src/routes/+layout.server.ts
import { redirect } from '@sveltejs/kit'
import type { LayoutServerLoad } from './$types'

const PUBLIC_ROUTES = ['/login', '/auth']

function isPublicRoute(pathname: string): boolean {
  return PUBLIC_ROUTES.some(route => pathname === route || pathname.startsWith(route + '/'))
}

export const load: LayoutServerLoad = async ({ locals: { safeGetSession }, cookies, url }) => {
  const { session, user } = await safeGetSession()

  if (!session && !isPublicRoute(url.pathname) && url.pathname !== '/') {
    throw redirect(303, '/login')
  }

  return {
    session,
    user,
    cookies: cookies.getAll(),
  }
}