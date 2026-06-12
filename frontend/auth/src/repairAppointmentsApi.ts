const API = ''

export class ApiError extends Error {
  status: number
  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

export type RepairAppointmentSlot = {
  start_at: number
  end_at: number
  available: boolean
  label: string
}

export type RepairAppointmentLabor = {
  id: string
  work_id: string
  description: string
  quantity: string
  unit_price: string
  sort_order: number
  work_code: string
  work_name: string
}

export type RepairAppointmentPart = {
  id: string
  part_id: string
  part_name: string
  part_sku: string
  warehouse_id: string
  warehouse_name: string
  quantity: string
  unit_price: string
  notes: string
  sort_order: number
}

export type RepairAppointment = {
  id: string
  appointment_number: string
  customer_id: string
  customer_name: string
  vehicle_id: string
  vehicle_vin: string
  vehicle_label: string
  dealer_point_id: string
  warehouse_id: string
  scheduled_start: number
  scheduled_end: number
  status: string
  work_order_id: string
  work_order_number: string
  complaint: string
  notes: string
  labor: RepairAppointmentLabor[]
  parts: RepairAppointmentPart[]
  created_at: number
  updated_at: number
}

export type LaborLineInput = {
  work_id?: string
  description?: string
  quantity?: string
  unit_price?: string
  sort_order?: number
}

export type PartLineInput = {
  part_id: string
  warehouse_id: string
  quantity?: string
  unit_price?: string
  notes?: string
  sort_order?: number
}

export type RepairAppointmentForm = {
  customer_id: string
  vehicle_id: string
  warehouse_id?: string
  scheduled_start: number
  scheduled_end: number
  complaint?: string
  notes?: string
  labor: LaborLineInput[]
  parts: PartLineInput[]
}

function getAuthHeaders(): HeadersInit {
  const token = sessionStorage.getItem('dealer_access_token')
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  }
}

async function readErrorMessage(res: Response): Promise<string> {
  const data = await res.json().catch(() => ({}))
  const body = data as { message?: string; error?: string }
  return body.message || body.error || res.statusText || 'Ошибка запроса'
}

export async function listSlots(date: string): Promise<{ slots: RepairAppointmentSlot[] }> {
  const sp = new URLSearchParams({ date })
  const res = await fetch(`${API}/api/repair-appointment-slots?${sp}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new ApiError(await readErrorMessage(res), res.status)
  return res.json()
}

export async function listRepairAppointments(params: {
  limit?: number
  offset?: number
  status?: string
  date_from?: number
  date_to?: number
}): Promise<{ appointments: RepairAppointment[]; total: number }> {
  const sp = new URLSearchParams()
  if (params.limit != null) sp.set('limit', String(params.limit))
  if (params.offset != null) sp.set('offset', String(params.offset))
  if (params.status) sp.set('status', params.status)
  if (params.date_from != null) sp.set('date_from', String(params.date_from))
  if (params.date_to != null) sp.set('date_to', String(params.date_to))
  const res = await fetch(`${API}/api/repair-appointments?${sp}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new ApiError(await readErrorMessage(res), res.status)
  return res.json()
}

export async function getRepairAppointment(id: string): Promise<RepairAppointment> {
  const res = await fetch(`${API}/api/repair-appointments/${id}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new ApiError(await readErrorMessage(res), res.status)
  return res.json()
}

export async function createRepairAppointment(payload: RepairAppointmentForm, createdBy?: string): Promise<RepairAppointment> {
  const res = await fetch(`${API}/api/repair-appointments`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ ...payload, created_by: createdBy || '' }),
  })
  if (!res.ok) throw new ApiError(await readErrorMessage(res), res.status)
  return res.json()
}

export async function updateRepairAppointment(id: string, payload: RepairAppointmentForm): Promise<RepairAppointment> {
  const res = await fetch(`${API}/api/repair-appointments/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify({ ...payload, replace_lines: true }),
  })
  if (!res.ok) throw new ApiError(await readErrorMessage(res), res.status)
  return res.json()
}

export async function cancelRepairAppointment(id: string): Promise<RepairAppointment> {
  const res = await fetch(`${API}/api/repair-appointments/${id}/cancel`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({}),
  })
  if (!res.ok) throw new ApiError(await readErrorMessage(res), res.status)
  return res.json()
}

export async function createWorkOrderFromAppointment(id: string) {
  const res = await fetch(`${API}/api/repair-appointments/${id}/create-work-order`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({}),
  })
  if (!res.ok) throw new ApiError(await readErrorMessage(res), res.status)
  return res.json() as Promise<{ work_order_id: string; work_order_number: string; appointment: RepairAppointment }>
}
