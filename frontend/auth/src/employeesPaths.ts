export const EMPLOYEES_PATH = '/api/employees'

export function employeesResourcePath(id: string): string {
  return `${EMPLOYEES_PATH}/${id}`
}
