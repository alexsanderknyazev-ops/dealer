export const REVIEW_STATUS_LABEL: Record<string, string> = {
  pending: 'На модерации',
  published: 'Опубликован',
  rejected: 'Отклонён',
}

export function formatReviewDateTime(ts: number) {
  return ts ? new Date(ts * 1000).toLocaleString('ru-RU') : '—'
}

export function ratingStars(rating: number) {
  const n = Math.max(0, Math.min(5, Math.round(rating)))
  return '★'.repeat(n) + '☆'.repeat(5 - n)
}

export function truncateReviewText(text: string, max = 60) {
  const t = text.trim()
  if (t.length <= max) return t
  return `${t.slice(0, max)}…`
}
