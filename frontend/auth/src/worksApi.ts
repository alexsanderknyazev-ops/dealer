import { WORKS_PATH, worksResourcePath } from './worksPaths'
import { HTTP_HEADER_CONTENT_TYPE, HTTP_MIME_JSON } from './httpHeaders'

const API = ''

export type WorkFolder = {
  id: string
  name: string
  parent_id: string
  created_at: number
  updated_at: number
}

export type Work = {
  id: string
  code: string
  name: string
  category: string
  folder_id: string
  labor_hours: string
  unit_price: string
  notes: string
  created_at: number
  updated_at: number
}

export type WorkForm = {
  code: string
  name: string
  category?: string
  folder_id?: string
  labor_hours?: string
  unit_price?: string
  notes?: string
}

function getAuthHeaders(): HeadersInit {
  const token = sessionStorage.getItem('dealer_access_token')
  return {
    [HTTP_HEADER_CONTENT_TYPE]: HTTP_MIME_JSON,
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  }
}

export async function listFolders(parentId?: string): Promise<{ folders: WorkFolder[] }> {
  const sp = new URLSearchParams()
  if (parentId) sp.set('parent_id', parentId)
  const res = await fetch(`${API}${WORKS_PATH}/folders?${sp}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function getFolder(id: string): Promise<WorkFolder> {
  const res = await fetch(`${API}${WORKS_PATH}/folders/${id}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function createFolder(data: { name: string; parent_id?: string }): Promise<WorkFolder> {
  const res = await fetch(`${API}${WORKS_PATH}/folders`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function deleteFolder(id: string): Promise<void> {
  const res = await fetch(`${API}${WORKS_PATH}/folders/${id}`, { method: 'DELETE', headers: getAuthHeaders() })
  if (!res.ok && res.status !== 204) {
    throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  }
}

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

export async function listWorks(params: {
  limit?: number
  offset?: number
  search?: string
  category?: string
  folder_id?: string
}): Promise<{ works: Work[]; total: number }> {
  const sp = new URLSearchParams()
  if (params.limit != null) sp.set('limit', String(params.limit))
  if (params.offset != null) sp.set('offset', String(params.offset))
  if (params.search) sp.set('search', params.search)
  if (params.category) sp.set('category', params.category)
  if (params.folder_id) sp.set('folder_id', params.folder_id)
  const res = await fetch(`${API}${WORKS_PATH}?${sp}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function getWork(id: string): Promise<Work> {
  const res = await fetch(`${API}${worksResourcePath(id)}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function createWork(data: WorkForm): Promise<Work> {
  const res = await fetch(`${API}${WORKS_PATH}`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function updateWork(id: string, data: Partial<WorkForm>): Promise<Work> {
  const res = await fetch(`${API}${worksResourcePath(id)}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function deleteWork(id: string): Promise<void> {
  const res = await fetch(`${API}${worksResourcePath(id)}`, { method: 'DELETE', headers: getAuthHeaders() })
  if (!res.ok && res.status !== 204) {
    throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  }
}
