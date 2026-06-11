import { WORKS_PATH, worksResourcePath } from './worksPaths'
import { HTTP_HEADER_CONTENT_TYPE, HTTP_MIME_JSON } from './httpHeaders'

const API = ''

export type Work = {
  id: string
  code: string
  name: string
  category: string
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

export async function listWorks(params: {
  limit?: number
  offset?: number
  search?: string
  category?: string
}): Promise<{ works: Work[]; total: number }> {
  const sp = new URLSearchParams()
  if (params.limit != null) sp.set('limit', String(params.limit))
  if (params.offset != null) sp.set('offset', String(params.offset))
  if (params.search) sp.set('search', params.search)
  if (params.category) sp.set('category', params.category)
  const res = await fetch(`${API}${WORKS_PATH}?${sp}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function getWork(id: string): Promise<Work> {
  const res = await fetch(`${API}${worksResourcePath(id)}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}
