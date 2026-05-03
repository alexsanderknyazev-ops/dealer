/** Brands API URL segments (single definition for fetch URLs). */
export const BRANDS_PATH = '/api/brands'

export function brandsResourcePath(id: string): string {
  return `${BRANDS_PATH}/${id}`
}
