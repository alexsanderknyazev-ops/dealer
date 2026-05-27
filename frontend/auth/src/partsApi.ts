const API = ''

export type PartFolder = {
  id: string
  name: string
  parent_id: string
  created_at: number
  updated_at: number
}

export type PartStockRow = {
  warehouse_id: string
  quantity: number
}

export type Part = {
  id: string
  sku: string
  name: string
  category: string
  folder_id: string
  brand_id?: string
  dealer_point_id?: string
  legal_entity_id?: string
  warehouse_id?: string
  quantity: number
  unit: string
  price: string
  location: string
  notes: string
  created_at: number
  updated_at: number
  /** Остатки по складам (одна запчасть может быть на нескольких складах) */
  stock?: PartStockRow[]
}

export type PartForm = {
  sku: string
  name?: string
  category?: string
  folder_id?: string
  brand_id?: string
  dealer_point_id?: string
  legal_entity_id?: string
  warehouse_id?: string
  quantity?: number
  unit?: string
  price?: string
  location?: string
  notes?: string
  /** Остатки по складам: [{ warehouse_id, quantity }] */
  stock?: PartStockRow[]
}

function getAuthHeaders(): HeadersInit {
  const token = sessionStorage.getItem('dealer_access_token')
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  }
}

// Folders
export async function listFolders(parentId?: string): Promise<{ folders: PartFolder[] }> {
  const sp = new URLSearchParams()
  if (parentId) sp.set('parent_id', parentId)
  const res = await fetch(`${API}/api/parts/folders?${sp}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function getFolder(id: string): Promise<PartFolder> {
  const res = await fetch(`${API}/api/parts/folders/${id}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function createFolder(data: { name: string; parent_id?: string }): Promise<PartFolder> {
  const res = await fetch(`${API}/api/parts/folders`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function updateFolder(id: string, data: { name?: string; parent_id?: string }): Promise<PartFolder> {
  const res = await fetch(`${API}/api/parts/folders/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function deleteFolder(id: string): Promise<void> {
  const res = await fetch(`${API}/api/parts/folders/${id}`, { method: 'DELETE', headers: getAuthHeaders() })
  if (!res.ok && res.status !== 204) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
}

// Load all folders recursively for dropdown (flat list with indent)
export async function loadAllFoldersFlat(): Promise<{ id: string; name: string; level: number }[]> {
  const out: { id: string; name: string; level: number }[] = []
  async function addChildren(parentId: string | undefined, level: number) {
    const { folders } = await listFolders(parentId)
    for (const f of folders) {
      out.push({ id: f.id, name: f.name, level })
      await addChildren(f.id, level + 1)
    }
  }
  await addChildren(undefined, 0)
  return out
}

export async function listParts(params: {
  limit?: number
  offset?: number
  search?: string
  category?: string
  folder_id?: string
  brand_id?: string
  dealer_point_id?: string
  legal_entity_id?: string
  warehouse_id?: string
}): Promise<{ parts: Part[]; total: number }> {
  const sp = new URLSearchParams()
  if (params.limit != null) sp.set('limit', String(params.limit))
  if (params.offset != null) sp.set('offset', String(params.offset))
  if (params.search) sp.set('search', params.search)
  if (params.category) sp.set('category', params.category)
  if (params.folder_id) sp.set('folder_id', params.folder_id)
  if (params.brand_id) sp.set('brand_id', params.brand_id)
  if (params.dealer_point_id) sp.set('dealer_point_id', params.dealer_point_id)
  if (params.legal_entity_id) sp.set('legal_entity_id', params.legal_entity_id)
  if (params.warehouse_id) sp.set('warehouse_id', params.warehouse_id)
  const res = await fetch(`${API}/api/parts?${sp}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function getPart(id: string): Promise<Part> {
  const res = await fetch(`${API}/api/parts/${id}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function createPart(data: PartForm): Promise<Part> {
  const res = await fetch(`${API}/api/parts`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function updatePart(id: string, data: Partial<PartForm>): Promise<Part> {
  const res = await fetch(`${API}/api/parts/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function deletePart(id: string): Promise<void> {
  const res = await fetch(`${API}/api/parts/${id}`, { method: 'DELETE', headers: getAuthHeaders() })
  if (!res.ok && res.status !== 204) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
}
