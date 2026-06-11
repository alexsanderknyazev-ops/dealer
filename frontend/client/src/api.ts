const API_BASE = import.meta.env.VITE_API_URL || ''

export type AuthResponse = {
  user_id: string
  email: string
  access_token: string
  refresh_token: string
  expires_at: number
  client_id?: string
}

export type RefreshResponse = {
  access_token: string
  refresh_token: string
  expires_at: number
}

export type MeResponse = {
  user_id: string
  email: string
  valid: boolean
}

export type ClientVehicle = {
  id: string
  vehicle_id: string
  vin: string
  make: string
  model: string
  year: number
  added_at: number
}

export type ClientProfile = {
  id: string
  user_id: string
  email: string
  full_name: string
  phone: string
  vehicles: ClientVehicle[]
  created_at: number
}

export type Review = {
  id: string
  client_id: string
  dealer_point_id: string
  vehicle_id: string
  rating: number
  text: string
  status: string
  created_at: number
  updated_at: number
}

export class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

async function readError(res: Response): Promise<string> {
  const data = await res.json().catch(() => ({}))
  const body = data as { message?: string; error?: string }
  return body.message || body.error || res.statusText || 'Ошибка запроса'
}

async function request<T>(path: string, opts: RequestInit = {}): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...opts,
    headers: {
      'Content-Type': 'application/json',
      ...opts.headers,
    },
  })
  if (!res.ok) throw new ApiError(await readError(res), res.status)
  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}

function authHeaders(accessToken: string): HeadersInit {
  return { Authorization: `Bearer ${accessToken}` }
}

export async function registerClient(payload: {
  email: string
  password: string
  full_name: string
  phone?: string
  vin?: string
}): Promise<AuthResponse> {
  return request<AuthResponse>('/api/client/register', { method: 'POST', body: JSON.stringify(payload) })
}

export async function login(email: string, password: string): Promise<AuthResponse> {
  return request<AuthResponse>('/api/login', { method: 'POST', body: JSON.stringify({ email, password }) })
}

export async function refresh(refreshToken: string): Promise<RefreshResponse> {
  return request<RefreshResponse>('/api/refresh', {
    method: 'POST',
    body: JSON.stringify({ refresh_token: refreshToken }),
  })
}

export async function logout(refreshToken: string): Promise<void> {
  await request('/api/logout', { method: 'POST', body: JSON.stringify({ refresh_token: refreshToken }) })
}

export async function me(accessToken: string): Promise<MeResponse> {
  return request<MeResponse>('/api/me', { headers: authHeaders(accessToken) })
}

export async function getProfile(accessToken: string): Promise<ClientProfile> {
  return request<ClientProfile>('/api/client/profile', { headers: authHeaders(accessToken) })
}

export async function listVehicles(accessToken: string): Promise<{ vehicles: ClientVehicle[] }> {
  return request<{ vehicles: ClientVehicle[] }>('/api/client/vehicles', { headers: authHeaders(accessToken) })
}

export async function addVehicle(accessToken: string, vin: string): Promise<ClientVehicle> {
  return request<ClientVehicle>('/api/client/vehicles', {
    method: 'POST',
    headers: authHeaders(accessToken),
    body: JSON.stringify({ vin }),
  })
}

export async function listReviews(accessToken: string): Promise<{ reviews: Review[] }> {
  return request<{ reviews: Review[] }>('/api/client/reviews', { headers: authHeaders(accessToken) })
}

export async function createReview(
  accessToken: string,
  payload: { vehicle_id: string; rating: number; text: string },
): Promise<Review> {
  return request<Review>('/api/client/reviews', {
    method: 'POST',
    headers: authHeaders(accessToken),
    body: JSON.stringify(payload),
  })
}
