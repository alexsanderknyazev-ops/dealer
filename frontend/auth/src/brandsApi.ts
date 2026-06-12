import { BRANDS_PATH, brandsResourcePath } from './brandsPaths'
import { HTTP_HEADER_CONTENT_TYPE, HTTP_MIME_JSON } from './httpHeaders'

const API = ''

export type Brand = {
  id: string
  name: string
  created_at: number
  updated_at: number
}

export type BrandForm = {
  name: string
}

export type BrandLaborRate = {
  id: string
  brand_id: string
  dealer_point_id: string
  warranty_hour_price: string
  commercial_hour_price: string
  created_at: number
  updated_at: number
}

export type BrandLaborRateForm = {
  brand_id: string
  dealer_point_id: string
  warranty_hour_price: string
  commercial_hour_price: string
}

export type ResolvedLaborRate = {
  warranty_hour_price: string
  commercial_hour_price: string
  hour_price: string
  found: boolean
}

const LABOR_RATES_PATH = '/api/brand-labor-rates'

async function readErrorMessage(res: Response): Promise<string> {
  const text = await res.text()
  try {
    const body = JSON.parse(text) as { message?: string; error?: string }
    return body.message || body.error || res.statusText || 'Ошибка запроса'
  } catch {
    if (text.trimStart().startsWith('<')) {
      return 'Сервер вернул HTML вместо JSON — проверьте, что auth-service и gateway пересобраны'
    }
    return res.statusText || 'Ошибка запроса'
  }
}

function getAuthHeaders(): HeadersInit {
  const token = sessionStorage.getItem('dealer_access_token')
  return {
    [HTTP_HEADER_CONTENT_TYPE]: HTTP_MIME_JSON,
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  }
}

export async function listBrands(params: { limit?: number; offset?: number; search?: string }): Promise<{ brands: Brand[]; total: number }> {
  const sp = new URLSearchParams()
  if (params.limit != null) sp.set('limit', String(params.limit))
  if (params.offset != null) sp.set('offset', String(params.offset))
  if (params.search) sp.set('search', params.search)
  const res = await fetch(`${API}${BRANDS_PATH}?${sp}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function getBrand(id: string): Promise<Brand> {
  const res = await fetch(`${API}${brandsResourcePath(id)}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function createBrand(data: BrandForm): Promise<Brand> {
  const res = await fetch(`${API}${BRANDS_PATH}`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function updateBrand(id: string, data: Partial<BrandForm>): Promise<Brand> {
  const res = await fetch(`${API}${brandsResourcePath(id)}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
  return res.json()
}

export async function deleteBrand(id: string): Promise<void> {
  const res = await fetch(`${API}${brandsResourcePath(id)}`, { method: 'DELETE', headers: getAuthHeaders() })
  if (!res.ok && res.status !== 204) throw new Error(await res.json().then((b: { error?: string }) => b.error).catch(() => res.statusText))
}

export async function listBrandLaborRates(params: {
  limit?: number
  offset?: number
  brand_id?: string
  dealer_point_id?: string
}): Promise<{ brand_labor_rates: BrandLaborRate[]; total: number }> {
  const sp = new URLSearchParams()
  if (params.limit != null) sp.set('limit', String(params.limit))
  if (params.offset != null) sp.set('offset', String(params.offset))
  if (params.brand_id) sp.set('brand_id', params.brand_id)
  if (params.dealer_point_id) sp.set('dealer_point_id', params.dealer_point_id)
  const res = await fetch(`${API}${LABOR_RATES_PATH}?${sp}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await readErrorMessage(res))
  return res.json()
}

export async function updateBrandLaborRate(data: BrandLaborRateForm): Promise<BrandLaborRate> {
  const res = await fetch(`${API}${LABOR_RATES_PATH}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error(await readErrorMessage(res))
  const body = await res.json()
  return body.brand_labor_rate ?? body
}

export async function deleteBrandLaborRate(id: string): Promise<void> {
  const res = await fetch(`${API}${LABOR_RATES_PATH}/${id}`, { method: 'DELETE', headers: getAuthHeaders() })
  if (!res.ok && res.status !== 204) throw new Error(await readErrorMessage(res))
}

export async function resolveBrandLaborRate(params: {
  brand_id: string
  dealer_point_id: string
  repair_type: string
}): Promise<ResolvedLaborRate> {
  const sp = new URLSearchParams({
    brand_id: params.brand_id,
    dealer_point_id: params.dealer_point_id,
    repair_type: params.repair_type,
  })
  const res = await fetch(`${API}${LABOR_RATES_PATH}/resolve?${sp}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await readErrorMessage(res))
  return res.json()
}
