const API = ''

export type Deal = {
  id: string
  customer_id: string
  vehicle_id: string
  amount: string
  stage: string
  assigned_to: string
  notes: string
  created_at: number
  updated_at: number
}

export type DealForm = {
  customer_id: string
  vehicle_id: string
  amount?: string
  stage?: string
  assigned_to?: string
  notes?: string
}

function getAuthHeaders(): HeadersInit {
  const token = sessionStorage.getItem('dealer_access_token')
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  }
}

export async function listDeals(params: {
  limit?: number
  offset?: number
  stage?: string
  customer_id?: string
}): Promise<{ deals: Deal[]; total: number }> {
  const sp = new URLSearchParams()
  if (params.limit != null) sp.set('limit', String(params.limit))
  if (params.offset != null) sp.set('offset', String(params.offset))
  if (params.stage) sp.set('stage', params.stage)
  if (params.customer_id) sp.set('customer_id', params.customer_id)
  const res = await fetch(`${API}/api/deals?${sp}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function getDeal(id: string): Promise<Deal> {
  const res = await fetch(`${API}/api/deals/${id}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function createDeal(data: DealForm): Promise<Deal> {
  const res = await fetch(`${API}/api/deals`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function updateDeal(id: string, data: Partial<DealForm>): Promise<Deal> {
  const res = await fetch(`${API}/api/deals/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function deleteDeal(id: string): Promise<void> {
  const res = await fetch(`${API}/api/deals/${id}`, { method: 'DELETE', headers: getAuthHeaders() })
  if (!res.ok && res.status !== 204) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
}
