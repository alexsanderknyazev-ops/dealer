const API = ''

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

async function readJSON<T>(res: Response): Promise<T> {
  const text = await res.text()
  try {
    return JSON.parse(text) as T
  } catch {
    throw new Error('API вернул не JSON — проверьте, что gateway и сервисы статистики запущены')
  }
}

function toNum(v: string | number | undefined | null): number {
  if (v == null) return 0
  if (typeof v === 'number') return v
  const n = Number(v)
  return Number.isFinite(n) ? n : 0
}

export type DealStageCount = {
  stage: string
  count: number
}

export type EmployeeOverview = {
  customers_count: number
  vehicles_count: number
  deals_count: number
  deals_by_stage: DealStageCount[]
  total_revenue: number
  parts_count: number
  dealer_points_count: number
}

export type ReviewStatusCount = {
  status: string
  count: number
}

export type ClientOverview = {
  clients_count: number
  client_vehicles_count: number
  registered_users_count: number
  reviews_count: number
  average_rating: number
  reviews_by_status: ReviewStatusCount[]
}

type EmployeeOverviewRaw = {
  customers_count?: string | number
  vehicles_count?: string | number
  deals_count?: string | number
  deals_by_stage?: { stage?: string; count?: string | number }[]
  total_revenue?: number
  parts_count?: string | number
  dealer_points_count?: string | number
}

type ClientOverviewRaw = {
  clients_count?: string | number
  client_vehicles_count?: string | number
  registered_users_count?: string | number
  reviews_count?: string | number
  average_rating?: number
  reviews_by_status?: { status?: string; count?: string | number }[]
}

function normalizeEmployee(raw: EmployeeOverviewRaw): EmployeeOverview {
  return {
    customers_count: toNum(raw.customers_count),
    vehicles_count: toNum(raw.vehicles_count),
    deals_count: toNum(raw.deals_count),
    deals_by_stage: (raw.deals_by_stage ?? []).map((item) => ({
      stage: item.stage ?? '',
      count: toNum(item.count),
    })),
    total_revenue: raw.total_revenue ?? 0,
    parts_count: toNum(raw.parts_count),
    dealer_points_count: toNum(raw.dealer_points_count),
  }
}

function normalizeClient(raw: ClientOverviewRaw): ClientOverview {
  return {
    clients_count: toNum(raw.clients_count),
    client_vehicles_count: toNum(raw.client_vehicles_count),
    registered_users_count: toNum(raw.registered_users_count),
    reviews_count: toNum(raw.reviews_count),
    average_rating: raw.average_rating ?? 0,
    reviews_by_status: (raw.reviews_by_status ?? []).map((item) => ({
      status: item.status ?? '',
      count: toNum(item.count),
    })),
  }
}

export async function getEmployeeOverview(): Promise<EmployeeOverview> {
  const res = await fetch(`${API}/api/stats/employee/overview`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await readError(res))
  const data = await readJSON<EmployeeOverviewRaw>(res)
  return normalizeEmployee(data)
}

export async function getClientOverview(): Promise<ClientOverview> {
  const res = await fetch(`${API}/api/stats/client/overview`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await readError(res))
  const data = await readJSON<ClientOverviewRaw>(res)
  return normalizeClient(data)
}
