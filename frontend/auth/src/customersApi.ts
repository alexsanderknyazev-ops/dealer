const API = '' // same origin via proxy

export type Customer = {
  id: string
  name: string
  email: string
  phone: string
  customer_type: string
  inn: string
  address: string
  notes: string
  created_at: number
  updated_at: number
}

export type CustomerForm = {
  name: string
  email?: string
  phone?: string
  customer_type?: string
  inn?: string
  address?: string
  notes?: string
}

function getAuthHeaders(): HeadersInit {
  const token = sessionStorage.getItem('dealer_access_token')
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  }
}

export async function listCustomers(params: { limit?: number; offset?: number; search?: string }): Promise<{ customers: Customer[]; total: number }> {
  const sp = new URLSearchParams()
  if (params.limit != null) sp.set('limit', String(params.limit))
  if (params.offset != null) sp.set('offset', String(params.offset))
  if (params.search) sp.set('search', params.search)
  const res = await fetch(`${API}/api/customers?${sp}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function getCustomer(id: string): Promise<Customer> {
  const res = await fetch(`${API}/api/customers/${id}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function createCustomer(data: CustomerForm): Promise<Customer> {
  const res = await fetch(`${API}/api/customers`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function updateCustomer(id: string, data: Partial<CustomerForm>): Promise<Customer> {
  const res = await fetch(`${API}/api/customers/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function deleteCustomer(id: string): Promise<void> {
  const res = await fetch(`${API}/api/customers/${id}`, { method: 'DELETE', headers: getAuthHeaders() })
  if (!res.ok && res.status !== 204) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
}
