import {
  AUTH_PATH_LOGIN,
  AUTH_PATH_LOGOUT,
  AUTH_PATH_ME,
  AUTH_PATH_REFRESH,
  AUTH_PATH_REGISTER,
} from './apiPaths'
import { HTTP_HEADER_CONTENT_TYPE, HTTP_MIME_JSON } from './httpHeaders'

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
      [HTTP_HEADER_CONTENT_TYPE]: HTTP_MIME_JSON,
      ...opts.headers,
    },
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error((data as { error?: string }).error || res.statusText)
  return data as T
}

export async function register(p: RegisterPayload): Promise<AuthResponse> {
  return request<AuthResponse>(AUTH_PATH_REGISTER, { method: 'POST', body: JSON.stringify(p) })
}

export async function login(p: LoginPayload): Promise<AuthResponse> {
  return request<AuthResponse>(AUTH_PATH_LOGIN, { method: 'POST', body: JSON.stringify(p) })
}

export async function refresh(refreshToken: string): Promise<RefreshResponse> {
  return request<RefreshResponse>(AUTH_PATH_REFRESH, {
    method: 'POST',
    body: JSON.stringify({ refresh_token: refreshToken }),
  })
}

export async function logout(refreshToken: string): Promise<void> {
  await request(AUTH_PATH_LOGOUT, {
    method: 'POST',
    body: JSON.stringify({ refresh_token: refreshToken }),
  })
}

export async function me(accessToken: string): Promise<MeResponse> {
  return request<MeResponse>(AUTH_PATH_ME, {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}
