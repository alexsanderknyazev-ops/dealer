const API = ''

export type Supplier = {
  id: string
  name: string
  inn: string
  phone: string
  email: string
  notes: string
  created_at: number
  updated_at: number
}

function getAuthHeaders(): HeadersInit {
  const token = sessionStorage.getItem('dealer_access_token')
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  }
}

export async function listSuppliers(params: {
  limit?: number
  offset?: number
  search?: string
}): Promise<{ suppliers: Supplier[]; total: number }> {
  const sp = new URLSearchParams()
  if (params.limit != null) sp.set('limit', String(params.limit))
  if (params.offset != null) sp.set('offset', String(params.offset))
  if (params.search) sp.set('search', params.search)
  const res = await fetch(`${API}/api/suppliers?${sp}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}
