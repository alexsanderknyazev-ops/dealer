export const AUTH_ERROR_EVENT = 'dealer:auth-error'

export class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

function notifyAuthError(status: number) {
  if (typeof window === 'undefined') return
  window.dispatchEvent(new CustomEvent(AUTH_ERROR_EVENT, { detail: { status } }))
}

export function createApiError(status: number, fallbackMessage: string): ApiError {
  if (status === 401) {
    notifyAuthError(status)
    return new ApiError('Сессия истекла. Войдите снова.', status)
  }
  if (status === 403) {
    notifyAuthError(status)
    return new ApiError('Недостаточно прав для этой операции.', status)
  }
  return new ApiError(fallbackMessage || 'Ошибка запроса', status)
}

export async function readApiErrorMessage(res: Response): Promise<string> {
  return res
    .json()
    .then((b: { error?: string; message?: string }) => b.error || b.message || res.statusText || 'Ошибка запроса')
    .catch(() => res.statusText || 'Ошибка запроса')
}
