function decodeBase64(str: string): string {
  // Works in both browser (atob) and Node.js (Buffer)
  if (typeof atob === 'function') {
    return atob(str)
  }
  return Buffer.from(str, 'base64').toString('utf-8')
}

export function getRoleFromToken(token: string | undefined): string | null {
  if (!token) return null
  try {
    const payload = token.split('.')[1]
    const decoded = JSON.parse(decodeBase64(payload))
    return decoded.user_role || decoded.role || null
  } catch {
    return null
  }
}

export function isCommissioner(role: string | null): boolean {
  return role === 'commissioner' || role === 'admin'
}

export function isAdmin(role: string | null): boolean {
  return role === 'admin'
}
