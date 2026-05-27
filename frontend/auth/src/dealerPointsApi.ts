const API = ''

export type DealerPoint = {
  id: string
  name: string
  address: string
  created_at: number
  updated_at: number
}

export type DealerPointForm = {
  name: string
  address?: string
}

export type LegalEntity = {
  id: string
  name: string
  inn: string
  address: string
  created_at: number
  updated_at: number
}

export type LegalEntityForm = {
  name: string
  inn?: string
  address?: string
}

export type Warehouse = {
  id: string
  dealer_point_id: string
  legal_entity_id: string
  type: 'cars' | 'parts'
  name: string
  created_at: number
  updated_at: number
}

export type WarehouseForm = {
  dealer_point_id: string
  legal_entity_id: string
  type: 'cars' | 'parts'
  name: string
}

function getAuthHeaders(): HeadersInit {
  const token = sessionStorage.getItem('dealer_access_token')
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  }
}

async function handleRes(res: Response) {
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

// Dealer points
export async function listDealerPoints(params: { limit?: number; offset?: number; search?: string }): Promise<{ dealer_points: DealerPoint[]; total: number }> {
  const sp = new URLSearchParams()
  if (params.limit != null) sp.set('limit', String(params.limit))
  if (params.offset != null) sp.set('offset', String(params.offset))
  if (params.search) sp.set('search', params.search)
  const res = await fetch(`${API}/api/dealer-points?${sp}`, { headers: getAuthHeaders() })
  return handleRes(res)
}

export async function getDealerPoint(id: string): Promise<DealerPoint> {
  const res = await fetch(`${API}/api/dealer-points/${id}`, { headers: getAuthHeaders() })
  return handleRes(res)
}

export async function createDealerPoint(data: DealerPointForm): Promise<DealerPoint> {
  const res = await fetch(`${API}/api/dealer-points`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  return handleRes(res)
}

export async function updateDealerPoint(id: string, data: Partial<DealerPointForm>): Promise<DealerPoint> {
  const res = await fetch(`${API}/api/dealer-points/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  return handleRes(res)
}

export async function deleteDealerPoint(id: string): Promise<void> {
  const res = await fetch(`${API}/api/dealer-points/${id}`, { method: 'DELETE', headers: getAuthHeaders() })
  if (!res.ok && res.status !== 204) await handleRes(res)
}

// Legal entities
export async function listLegalEntities(params: { limit?: number; offset?: number; search?: string }): Promise<{ legal_entities: LegalEntity[]; total: number }> {
  const sp = new URLSearchParams()
  if (params.limit != null) sp.set('limit', String(params.limit))
  if (params.offset != null) sp.set('offset', String(params.offset))
  if (params.search) sp.set('search', params.search)
  const res = await fetch(`${API}/api/legal-entities?${sp}`, { headers: getAuthHeaders() })
  return handleRes(res)
}

export async function getLegalEntity(id: string): Promise<LegalEntity> {
  const res = await fetch(`${API}/api/legal-entities/${id}`, { headers: getAuthHeaders() })
  return handleRes(res)
}

export async function createLegalEntity(data: LegalEntityForm): Promise<LegalEntity> {
  const res = await fetch(`${API}/api/legal-entities`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  return handleRes(res)
}

export async function updateLegalEntity(id: string, data: Partial<LegalEntityForm>): Promise<LegalEntity> {
  const res = await fetch(`${API}/api/legal-entities/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  return handleRes(res)
}

export async function deleteLegalEntity(id: string): Promise<void> {
  const res = await fetch(`${API}/api/legal-entities/${id}`, { method: 'DELETE', headers: getAuthHeaders() })
  if (!res.ok && res.status !== 204) await handleRes(res)
}

export async function listLegalEntitiesByDealerPoint(dealerPointId: string): Promise<LegalEntity[]> {
  const res = await fetch(`${API}/api/dealer-points/${dealerPointId}/legal-entities`, { headers: getAuthHeaders() })
  const data = await handleRes(res)
  return (data as { legal_entities?: LegalEntity[] }).legal_entities ?? []
}

export async function linkLegalEntityToDealerPoint(dealerPointId: string, legalEntityId: string): Promise<void> {
  const res = await fetch(`${API}/api/dealer-points/${dealerPointId}/legal-entities`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ legal_entity_id: legalEntityId }),
  })
  if (!res.ok && res.status !== 204) await handleRes(res)
}

export async function unlinkLegalEntityFromDealerPoint(dealerPointId: string, legalEntityId: string): Promise<void> {
  const res = await fetch(`${API}/api/dealer-points/${dealerPointId}/legal-entities/${legalEntityId}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  })
  if (!res.ok && res.status !== 204) await handleRes(res)
}

// Warehouses
export async function listWarehouses(params: {
  limit?: number
  offset?: number
  dealer_point_id?: string
  legal_entity_id?: string
  type?: 'cars' | 'parts'
}): Promise<{ warehouses: Warehouse[]; total: number }> {
  const sp = new URLSearchParams()
  if (params.limit != null) sp.set('limit', String(params.limit))
  if (params.offset != null) sp.set('offset', String(params.offset))
  if (params.dealer_point_id) sp.set('dealer_point_id', params.dealer_point_id)
  if (params.legal_entity_id) sp.set('legal_entity_id', params.legal_entity_id)
  if (params.type) sp.set('type', params.type)
  const res = await fetch(`${API}/api/warehouses?${sp}`, { headers: getAuthHeaders() })
  return handleRes(res)
}

export async function getWarehouse(id: string): Promise<Warehouse> {
  const res = await fetch(`${API}/api/warehouses/${id}`, { headers: getAuthHeaders() })
  return handleRes(res)
}

export async function createWarehouse(data: WarehouseForm): Promise<Warehouse> {
  const res = await fetch(`${API}/api/warehouses`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  return handleRes(res)
}

export async function updateWarehouse(id: string, data: { name?: string }): Promise<Warehouse> {
  const res = await fetch(`${API}/api/warehouses/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  return handleRes(res)
}

export async function deleteWarehouse(id: string): Promise<void> {
  const res = await fetch(`${API}/api/warehouses/${id}`, { method: 'DELETE', headers: getAuthHeaders() })
  if (!res.ok && res.status !== 204) await handleRes(res)
}
