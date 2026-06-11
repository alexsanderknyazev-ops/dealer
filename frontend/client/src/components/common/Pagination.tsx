import { Button } from '@/components/ui/button'

type PaginationProps = {
  page: number
  total: number
  limit: number
  onPageChange: (page: number) => void
  showTotal?: boolean
}

export function Pagination({ page, total, limit, onPageChange, showTotal }: PaginationProps) {
  if (total <= limit) return null
  const totalPages = Math.ceil(total / limit) || 1

  return (
    <div className="mt-4 flex flex-wrap items-center gap-3 text-sm text-muted-foreground">
      {showTotal && <span>Всего: {total}</span>}
      <Button type="button" variant="outline" size="sm" disabled={page === 0} onClick={() => onPageChange(page - 1)}>
        Назад
      </Button>
      <span>
        Стр. {page + 1} из {totalPages}
      </span>
      <Button
        type="button"
        variant="outline"
        size="sm"
        disabled={(page + 1) * limit >= total}
        onClick={() => onPageChange(page + 1)}
      >
        Вперёд
      </Button>
    </div>
  )
}
