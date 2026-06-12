const API = ''

export class ApiError extends Error {
  status: number
  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

export type PartOrderLine = {
  id: string
  part_id: string
  part_name: string
  part_sku: string
  quantity: number
  unit_price: string
  notes: string
  sort_order: number
}

export type PartOrderLineInput = {
  part_id: string
  quantity: number
  unit_price: string
  notes?: string
  sort_order?: number
}

export type CustomerOrder = {
  id: string
  order_number: string
  status: string
  customer_id: string
  customer_name: string
  vehicle_id: string
  vehicle_vin: string
  vehicle_label: string
  issue_warehouse_id: string
  issue_warehouse_name: string
  fulfillment_movement_document_id: string
  fulfillment_movement_document_number: string
  fulfillment_work_order_id: string
  fulfillment_work_order_number: string
  notes: string
  created_at: number
  updated_at: number
  lines: PartOrderLine[]
}

export type CustomerOrderForm = {
  customer_id: string
  vehicle_id?: string
  vehicle_vin?: string
  issue_warehouse_id: string
  notes?: string
  lines: PartOrderLineInput[]
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

export async function listCustomerOrders(params: {
  limit?: number
  offset?: number
  status?: string
}): Promise<{ orders: CustomerOrder[]; total: number }> {
  const sp = new URLSearchParams()
  if (params.limit != null) sp.set('limit', String(params.limit))
  if (params.offset != null) sp.set('offset', String(params.offset))
  if (params.status) sp.set('status', params.status)
  const res = await fetch(`${API}/api/customer-orders?${sp}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new ApiError(await readErrorMessage(res), res.status)
  return res.json()
}

export async function getCustomerOrder(id: string): Promise<CustomerOrder> {
  const res = await fetch(`${API}/api/customer-orders/${id}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new ApiError(await readErrorMessage(res), res.status)
  return res.json()
}

export async function createCustomerOrder(payload: CustomerOrderForm, createdBy?: string): Promise<CustomerOrder> {
  const res = await fetch(`${API}/api/customer-orders`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ ...payload, created_by: createdBy || '' }),
  })
  if (!res.ok) throw new ApiError(await readErrorMessage(res), res.status)
  return res.json()
}

export async function updateCustomerOrder(id: string, payload: CustomerOrderForm): Promise<CustomerOrder> {
  const res = await fetch(`${API}/api/customer-orders/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify({
      ...payload,
      clear_vehicle: !payload.vehicle_id && !payload.vehicle_vin,
      replace_lines: true,
    }),
  })
  if (!res.ok) throw new ApiError(await readErrorMessage(res), res.status)
  return res.json()
}

export async function cancelCustomerOrder(id: string): Promise<CustomerOrder> {
  const res = await fetch(`${API}/api/customer-orders/${id}/cancel`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({}),
  })
  if (!res.ok) throw new ApiError(await readErrorMessage(res), res.status)
  return res.json()
}

export async function createSaleFromCustomerOrder(id: string, createdBy?: string) {
  const res = await fetch(`${API}/api/customer-orders/${id}/create-sale`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ created_by: createdBy || '' }),
  })
  if (!res.ok) throw new ApiError(await readErrorMessage(res), res.status)
  return res.json()
}

export async function createWorkOrderFromCustomerOrder(
  id: string,
  payload?: { vehicle_id?: string; vehicle_vin?: string; notes?: string },
) {
  const res = await fetch(`${API}/api/customer-orders/${id}/create-work-order`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(payload ?? {}),
  })
  if (!res.ok) throw new ApiError(await readErrorMessage(res), res.status)
  return res.json() as Promise<{ work_order_id: string; work_order_number: string }>
}
