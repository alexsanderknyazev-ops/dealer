import { EMPLOYEES_PATH } from './employeesPaths'
import { HTTP_HEADER_CONTENT_TYPE, HTTP_MIME_JSON } from './httpHeaders'

const API = ''

export type Employee = {
  id: string
  user_id: string
  full_name: string
  position: string
  department: string
  phone: string
  active: boolean
  created_at: number
  updated_at: number
}

function getAuthHeaders(): HeadersInit {
  const token = sessionStorage.getItem('dealer_access_token')
  return {
    [HTTP_HEADER_CONTENT_TYPE]: HTTP_MIME_JSON,
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  }
}

export async function listEmployees(params: {
  limit?: number
  offset?: number
  search?: string
  position?: string
  active_only?: boolean
}): Promise<{ employees: Employee[]; total: number }> {
  const sp = new URLSearchParams()
  if (params.limit != null) sp.set('limit', String(params.limit))
  if (params.offset != null) sp.set('offset', String(params.offset))
  if (params.search) sp.set('search', params.search)
  if (params.position) sp.set('position', params.position)
  if (params.active_only) sp.set('active_only', 'true')
  const res = await fetch(`${API}${EMPLOYEES_PATH}?${sp}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export function employeeRefId(emp: Employee): string {
  return emp.user_id || emp.id
}
