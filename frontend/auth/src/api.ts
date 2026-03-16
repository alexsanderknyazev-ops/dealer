const API_BASE = import.meta.env.VITE_API_URL || ''

export type RegisterPayload = { email: string; password: string; name?: string; phone?: string }
export type LoginPayload = { email: string; password: string }
export type AuthResponse = {
  user_id: string
  email: string
  access_token: string
  refresh_token: string
  expires_at: number
}
export type RefreshResponse = { access_token: string; refresh_token: string; expires_at: number }
export type MeResponse = { user_id: string; email: string; valid: boolean }

async function request<T>(path: string, opts: RequestInit = {}): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...opts,
    headers: {
      'Content-Type': 'application/json',
      ...opts.headers,
    },
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error((data as { error?: string }).error || res.statusText)
  return data as T
}

export async function register(p: RegisterPayload): Promise<AuthResponse> {
  return request<AuthResponse>('/api/register', { method: 'POST', body: JSON.stringify(p) })
}

export async function login(p: LoginPayload): Promise<AuthResponse> {
  return request<AuthResponse>('/api/login', { method: 'POST', body: JSON.stringify(p) })
}

export async function refresh(refreshToken: string): Promise<RefreshResponse> {
  return request<RefreshResponse>('/api/refresh', {
    method: 'POST',
    body: JSON.stringify({ refresh_token: refreshToken }),
  })
}

export async function logout(refreshToken: string): Promise<void> {
  await request('/api/logout', {
    method: 'POST',
    body: JSON.stringify({ refresh_token: refreshToken }),
  })
}

export async function me(accessToken: string): Promise<MeResponse> {
  return request<MeResponse>('/api/me', {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}
