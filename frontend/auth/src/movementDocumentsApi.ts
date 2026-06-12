const API = ''

export class ApiError extends Error {
  status: number
  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

export type MovementDocumentLine = {
  id: string
  part_id: string
  warehouse_id: string
  destination_warehouse_id: string
  destination_warehouse_name: string
  source_stock_quantity: number
  quantity: number
  reference_line_id: string
  notes: string
  sort_order: number
  part_name: string
  part_sku: string
  warehouse_name: string
}

export type MovementDocumentLineInput = {
  part_id: string
  warehouse_id: string
  destination_warehouse_id?: string
  quantity: number
  notes?: string
  sort_order?: number
}

export type MovementDocument = {
  id: string
  document_number: string
  status: string
  movement_type: string
  reference_type: string
  reference_id: string
  reference_label: string
  customer_name: string
  vehicle_vin: string
  vehicle_label: string
  parent_document_id: string
  parent_document_number: string
  notes: string
  created_by: string
  confirmed_by: string
  created_by_name: string
  confirmed_by_name: string
  created_at: number
  confirmed_at: number
  updated_at: number
  lines: MovementDocumentLine[]
}

export type MovementDocumentForm = {
  movement_type: string
  notes?: string
  lines: MovementDocumentLineInput[]
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

function toApiError(status: number, msg: string): ApiError {
  return new ApiError(msg, status)
}

export async function getMovementDocument(id: string): Promise<MovementDocument> {
  const res = await fetch(`${API}/api/movement-documents/${id}`, { headers: getAuthHeaders() })
  if (!res.ok) throw toApiError(res.status, await readErrorMessage(res))
  return res.json()
}

export async function listMovementDocuments(params: {
  limit?: number
  offset?: number
  status?: string
  reference_type?: string
  reference_id?: string
}): Promise<{ documents: MovementDocument[]; total: number }> {
  const sp = new URLSearchParams()
  if (params.limit != null) sp.set('limit', String(params.limit))
  if (params.offset != null) sp.set('offset', String(params.offset))
  if (params.status) sp.set('status', params.status)
  if (params.reference_type) sp.set('reference_type', params.reference_type)
  if (params.reference_id) sp.set('reference_id', params.reference_id)
  const res = await fetch(`${API}/api/movement-documents?${sp}`, { headers: getAuthHeaders() })
  if (!res.ok) throw toApiError(res.status, await readErrorMessage(res))
  return res.json()
}

export async function startMovementDocument(id: string): Promise<MovementDocument> {
  const res = await fetch(`${API}/api/movement-documents/${id}/start`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({}),
  })
  if (!res.ok) throw toApiError(res.status, await readErrorMessage(res))
  return res.json()
}

export async function closeMovementDocument(id: string, closedBy?: string): Promise<MovementDocument> {
  const res = await fetch(`${API}/api/movement-documents/${id}/close`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ closed_by: closedBy || '' }),
  })
  if (!res.ok) throw toApiError(res.status, await readErrorMessage(res))
  return res.json()
}

/** @deprecated используйте closeMovementDocument */
export async function confirmMovementDocument(id: string, confirmedBy?: string): Promise<MovementDocument> {
  return closeMovementDocument(id, confirmedBy)
}

export async function createMovementDocument(
  payload: MovementDocumentForm,
  createdBy?: string,
): Promise<MovementDocument> {
  const res = await fetch(`${API}/api/movement-documents`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({
      movement_type: payload.movement_type,
      notes: payload.notes || '',
      lines: payload.lines.map((l, i) => ({
        part_id: l.part_id,
        warehouse_id: l.warehouse_id,
        destination_warehouse_id: l.destination_warehouse_id || '',
        quantity: l.quantity,
        notes: l.notes || '',
        sort_order: l.sort_order ?? i,
      })),
      created_by: createdBy || '',
    }),
  })
  if (!res.ok) throw toApiError(res.status, await readErrorMessage(res))
  return res.json()
}

export async function updateMovementDocument(
  id: string,
  payload: MovementDocumentForm,
): Promise<MovementDocument> {
  const res = await fetch(`${API}/api/movement-documents/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify({
      movement_type: payload.movement_type,
      notes: payload.notes || '',
      replace_lines: true,
      lines: payload.lines.map((l, i) => ({
        part_id: l.part_id,
        warehouse_id: l.warehouse_id,
        destination_warehouse_id: l.destination_warehouse_id || '',
        quantity: l.quantity,
        notes: l.notes || '',
        sort_order: l.sort_order ?? i,
      })),
    }),
  })
  if (!res.ok) throw toApiError(res.status, await readErrorMessage(res))
  return res.json()
}

export async function cancelMovementDocument(id: string): Promise<MovementDocument> {
  const res = await fetch(`${API}/api/movement-documents/${id}/cancel`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({}),
  })
  if (!res.ok) throw toApiError(res.status, await readErrorMessage(res))
  return res.json()
}

export async function createProductionExtraction(
  parentId: string,
  createdBy?: string,
): Promise<MovementDocument> {
  const res = await fetch(`${API}/api/movement-documents/${parentId}/create-production-extraction`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ created_by: createdBy || '' }),
  })
  if (!res.ok) throw toApiError(res.status, await readErrorMessage(res))
  return res.json()
}
