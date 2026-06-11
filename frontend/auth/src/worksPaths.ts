/** Works API URL segments (single definition for fetch URLs). */
export const WORKS_PATH = '/api/works'

export function worksResourcePath(id: string): string {
  return `${WORKS_PATH}/${id}`
}
