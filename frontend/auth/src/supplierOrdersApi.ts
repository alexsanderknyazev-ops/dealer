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

export type SupplierOrder = {
  id: string
  order_number: string
  status: string
  supplier_id: string
  supplier_name: string
  receipt_warehouse_id: string
  receipt_warehouse_name: string
  fulfillment_movement_document_id: string
  fulfillment_movement_document_number: string
  fulfillment_work_order_id: string
  fulfillment_work_order_number: string
  customer_order_id: string
  customer_order_number: string
  notes: string
  created_at: number
  updated_at: number
  lines: PartOrderLine[]
}

export type SupplierOrderForm = {
  supplier_id: string
  receipt_warehouse_id: string
  customer_order_id?: string
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

export async function listSupplierOrders(params: {
  limit?: number
  offset?: number
  status?: string
}): Promise<{ orders: SupplierOrder[]; total: number }> {
  const sp = new URLSearchParams()
  if (params.limit != null) sp.set('limit', String(params.limit))
  if (params.offset != null) sp.set('offset', String(params.offset))
  if (params.status) sp.set('status', params.status)
  const res = await fetch(`${API}/api/supplier-orders?${sp}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new ApiError(await readErrorMessage(res), res.status)
  return res.json()
}

export async function getSupplierOrder(id: string): Promise<SupplierOrder> {
  const res = await fetch(`${API}/api/supplier-orders/${id}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new ApiError(await readErrorMessage(res), res.status)
  return res.json()
}

export async function createSupplierOrder(payload: SupplierOrderForm, createdBy?: string): Promise<SupplierOrder> {
  const res = await fetch(`${API}/api/supplier-orders`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ ...payload, created_by: createdBy || '' }),
  })
  if (!res.ok) throw new ApiError(await readErrorMessage(res), res.status)
  return res.json()
}

export async function updateSupplierOrder(id: string, payload: SupplierOrderForm): Promise<SupplierOrder> {
  const res = await fetch(`${API}/api/supplier-orders/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify({ ...payload, replace_lines: true }),
  })
  if (!res.ok) throw new ApiError(await readErrorMessage(res), res.status)
  return res.json()
}

export async function cancelSupplierOrder(id: string): Promise<SupplierOrder> {
  const res = await fetch(`${API}/api/supplier-orders/${id}/cancel`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({}),
  })
  if (!res.ok) throw new ApiError(await readErrorMessage(res), res.status)
  return res.json()
}

export async function createReceiptFromSupplierOrder(id: string, createdBy?: string) {
  const res = await fetch(`${API}/api/supplier-orders/${id}/create-receipt`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ created_by: createdBy || '' }),
  })
  if (!res.ok) throw new ApiError(await readErrorMessage(res), res.status)
  return res.json()
}

export async function createWorkOrderFromSupplierOrder(
  id: string,
  payload: { customer_id: string; vehicle_id?: string; vehicle_vin?: string; notes?: string },
) {
  const res = await fetch(`${API}/api/supplier-orders/${id}/create-work-order`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(payload),
  })
  if (!res.ok) throw new ApiError(await readErrorMessage(res), res.status)
  return res.json() as Promise<{ work_order_id: string; work_order_number: string }>
}
