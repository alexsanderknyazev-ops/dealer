const API = ''

export class ApiError extends Error {
  status: number
  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

export type WorkOrderLabor = {
  id?: string
  work_id: string
  description?: string
  quantity: string
  unit_price: string
  amount?: string
  executor_id: string
  executor_name?: string
  sort_order?: number
  work_code?: string
  work_name?: string
  labor_hours?: string
}

export type WorkOrderPart = {
  id?: string
  part_id: string
  warehouse_id: string
  description: string
  quantity: string
  unit_price: string
  amount?: string
  issued?: boolean
  sort_order?: number
  part_sku?: string
  warehouse_name?: string
  part_name?: string
}

export type WorkOrder = {
  id: string
  order_number: string
  customer_id: string
  vehicle_id: string
  dealer_point_id: string
  warehouse_id: string
  repair_type: string
  status: string
  service_advisor_id: string
  service_advisor_name: string
  complaint: string
  diagnosis: string
  mileage_km: number
  labor_cost: string
  parts_cost: string
  total_cost: string
  opened_at: number
  closed_at: number
  parts_issued: boolean
  parts_issued_at: number
  notes: string
  created_at: number
  updated_at: number
  labor: WorkOrderLabor[]
  parts: WorkOrderPart[]
  movement_document_id: string
  movement_document_status: string
  customer_name: string
  vehicle_vin: string
  vehicle_label: string
}

export type WorkOrderForm = {
  customer_id: string
  vehicle_id: string
  dealer_point_id?: string
  warehouse_id?: string
  brand_id?: string
  repair_type?: string
  status?: string
  service_advisor_id?: string
  complaint?: string
  diagnosis?: string
  mileage_km?: number
  opened_at?: number
  closed_at?: number
  notes?: string
  labor?: WorkOrderLabor[]
  parts?: WorkOrderPart[]
}

function getAuthHeaders(): HeadersInit {
  const token = sessionStorage.getItem('dealer_access_token')
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  }
}

function toApiError(status: number, fallbackMessage: string): ApiError {
  if (status === 401) return new ApiError('Сессия истекла. Войдите снова.', 401)
  if (status === 403) return new ApiError('Недостаточно прав для этой операции.', 403)
  return new ApiError(fallbackMessage, status)
}

async function readErrorMessage(res: Response): Promise<string> {
  const data = await res.json().catch(() => ({}))
  const body = data as { message?: string; error?: string }
  return body.message || body.error || res.statusText || 'Ошибка запроса'
}

export async function listWorkOrders(params: {
  limit?: number
  offset?: number
  status?: string
  repair_type?: string
  customer_id?: string
  vehicle_id?: string
}): Promise<{ work_orders: WorkOrder[]; total: number }> {
  const sp = new URLSearchParams()
  if (params.limit != null) sp.set('limit', String(params.limit))
  if (params.offset != null) sp.set('offset', String(params.offset))
  if (params.status) sp.set('status', params.status)
  if (params.repair_type) sp.set('repair_type', params.repair_type)
  if (params.customer_id) sp.set('customer_id', params.customer_id)
  if (params.vehicle_id) sp.set('vehicle_id', params.vehicle_id)
  const res = await fetch(`${API}/api/work-orders?${sp}`, { headers: getAuthHeaders() })
  if (!res.ok) throw toApiError(res.status, await readErrorMessage(res))
  return res.json()
}

export async function getWorkOrder(id: string): Promise<WorkOrder> {
  const res = await fetch(`${API}/api/work-orders/${id}`, { headers: getAuthHeaders() })
  if (!res.ok) throw toApiError(res.status, await readErrorMessage(res))
  return res.json()
}

export async function createWorkOrder(data: WorkOrderForm): Promise<WorkOrder> {
  const res = await fetch(`${API}/api/work-orders`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!res.ok) throw toApiError(res.status, await readErrorMessage(res))
  return res.json()
}

export async function updateWorkOrder(id: string, data: WorkOrderForm): Promise<WorkOrder> {
  const res = await fetch(`${API}/api/work-orders/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!res.ok) throw toApiError(res.status, await readErrorMessage(res))
  return res.json()
}

export async function deleteWorkOrder(id: string): Promise<void> {
  const res = await fetch(`${API}/api/work-orders/${id}`, { method: 'DELETE', headers: getAuthHeaders() })
  if (!res.ok) throw toApiError(res.status, await readErrorMessage(res))
}

export async function movePartsToWork(id: string, issuedBy?: string): Promise<WorkOrder> {
  const res = await fetch(`${API}/api/work-orders/${id}/move-parts-to-work`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ issued_by: issuedBy || '' }),
  })
  if (!res.ok) throw toApiError(res.status, await readErrorMessage(res))
  return res.json()
}
