const API = ''

export type Brand = {
  id: string
  name: string
  created_at: number
  updated_at: number
}

export type BrandForm = {
  name: string
}

function getAuthHeaders(): HeadersInit {
  const token = sessionStorage.getItem('dealer_access_token')
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  }
}

export async function listBrands(params: { limit?: number; offset?: number; search?: string }): Promise<{ brands: Brand[]; total: number }> {
  const sp = new URLSearchParams()
  if (params.limit != null) sp.set('limit', String(params.limit))
  if (params.offset != null) sp.set('offset', String(params.offset))
  if (params.search) sp.set('search', params.search)
  const res = await fetch(`${API}/api/brands?${sp}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function getBrand(id: string): Promise<Brand> {
  const res = await fetch(`${API}/api/brands/${id}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function createBrand(data: BrandForm): Promise<Brand> {
  const res = await fetch(`${API}/api/brands`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function updateBrand(id: string, data: Partial<BrandForm>): Promise<Brand> {
  const res = await fetch(`${API}/api/brands/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function deleteBrand(id: string): Promise<void> {
  const res = await fetch(`${API}/api/brands/${id}`, { method: 'DELETE', headers: getAuthHeaders() })
  if (!res.ok && res.status !== 204) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
}
