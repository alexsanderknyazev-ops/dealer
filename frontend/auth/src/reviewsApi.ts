const API = ''

export type EmployeeReview = {
  id: string
  review_id: string
  client_id: string
  user_id: string
  client_email: string
  client_full_name: string
  dealer_point_id: string
  vehicle_id: string
  vehicle_vin: string
  vehicle_make: string
  vehicle_model: string
  vehicle_year: number
  rating: number
  text: string
  status: string
  occurred_at: number
  created_at: number
}

function getAuthHeaders(): HeadersInit {
  const token = sessionStorage.getItem('dealer_access_token')
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  }
}

async function readError(res: Response): Promise<string> {
  const data = await res.json().catch(() => ({}))
  const body = data as { message?: string; error?: string }
  return body.message || body.error || res.statusText || 'Ошибка запроса'
}

export async function listReviews(params: {
  limit?: number
  offset?: number
  client_id?: string
  dealer_point_id?: string
  status?: string
}): Promise<{ reviews: EmployeeReview[]; total: number }> {
  const sp = new URLSearchParams()
  if (params.limit != null) sp.set('limit', String(params.limit))
  if (params.offset != null) sp.set('offset', String(params.offset))
  if (params.client_id) sp.set('client_id', params.client_id)
  if (params.dealer_point_id) sp.set('dealer_point_id', params.dealer_point_id)
  if (params.status) sp.set('status', params.status)
  const res = await fetch(`${API}/api/reviews?${sp}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await readError(res))
  const data = await res.json()
  return {
    reviews: data.reviews ?? [],
    total: Number(data.total ?? 0),
  }
}
