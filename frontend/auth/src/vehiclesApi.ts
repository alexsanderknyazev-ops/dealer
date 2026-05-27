const API = ''

export type Vehicle = {
  id: string
  vin: string
  make: string
  model: string
  year: number
  mileage_km: number
  price: string
  status: string
  color: string
  notes: string
  brand_id?: string | null
  dealer_point_id?: string | null
  legal_entity_id?: string | null
  warehouse_id?: string | null
  created_at: number
  updated_at: number
}

export type VehicleForm = {
  vin: string
  make?: string
  model?: string
  year?: number
  mileage_km?: number
  price?: string
  status?: string
  color?: string
  notes?: string
  brand_id?: string | null
  dealer_point_id?: string | null
  legal_entity_id?: string | null
  warehouse_id?: string | null
}

function getAuthHeaders(): HeadersInit {
  const token = sessionStorage.getItem('dealer_access_token')
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  }
}

export async function listVehicles(params: {
  limit?: number
  offset?: number
  search?: string
  status?: string
  brand_id?: string
  dealer_point_id?: string
  legal_entity_id?: string
  warehouse_id?: string
}): Promise<{ vehicles: Vehicle[]; total: number }> {
  const sp = new URLSearchParams()
  if (params.limit != null) sp.set('limit', String(params.limit))
  if (params.offset != null) sp.set('offset', String(params.offset))
  if (params.search) sp.set('search', params.search)
  if (params.status) sp.set('status', params.status)
  if (params.brand_id) sp.set('brand_id', params.brand_id)
  if (params.dealer_point_id) sp.set('dealer_point_id', params.dealer_point_id)
  if (params.legal_entity_id) sp.set('legal_entity_id', params.legal_entity_id)
  if (params.warehouse_id) sp.set('warehouse_id', params.warehouse_id)
  const res = await fetch(`${API}/api/vehicles?${sp}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function getVehicle(id: string): Promise<Vehicle> {
  const res = await fetch(`${API}/api/vehicles/${id}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function createVehicle(data: VehicleForm): Promise<Vehicle> {
  const res = await fetch(`${API}/api/vehicles`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function updateVehicle(id: string, data: Partial<VehicleForm>): Promise<Vehicle> {
  const res = await fetch(`${API}/api/vehicles/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function deleteVehicle(id: string): Promise<void> {
  const res = await fetch(`${API}/api/vehicles/${id}`, { method: 'DELETE', headers: getAuthHeaders() })
  if (!res.ok && res.status !== 204) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
}
