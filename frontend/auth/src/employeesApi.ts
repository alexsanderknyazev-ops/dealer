import { EMPLOYEES_PATH, employeesResourcePath } from './employeesPaths'
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

export type EmployeeForm = {
  user_id?: string
  full_name: string
  position: string
  department?: string
  phone?: string
  active?: boolean
}

export const POSITION_LABEL: Record<string, string> = {
  admin: 'Администратор',
  manager: 'Менеджер',
  sales: 'Продажи',
  master: 'Мастер',
  service_advisor: 'Мастер-консультант',
  parts_manager: 'Менеджер запчастей',
  storekeeper: 'Кладовщик',
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

export async function getEmployee(id: string): Promise<Employee> {
  const res = await fetch(`${API}${employeesResourcePath(id)}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function createEmployee(data: EmployeeForm): Promise<Employee> {
  const res = await fetch(`${API}${EMPLOYEES_PATH}`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function updateEmployee(id: string, data: Partial<EmployeeForm>): Promise<Employee> {
  const res = await fetch(`${API}${employeesResourcePath(id)}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function deleteEmployee(id: string): Promise<void> {
  const res = await fetch(`${API}${employeesResourcePath(id)}`, { method: 'DELETE', headers: getAuthHeaders() })
  if (!res.ok && res.status !== 204) {
    throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  }
}

export function employeeRefId(emp: Employee): string {
  return emp.user_id || emp.id
}

export function positionLabel(position: string): string {
  return POSITION_LABEL[position] || position || '—'
}
