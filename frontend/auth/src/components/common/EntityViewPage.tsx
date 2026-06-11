import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { LoadingState } from '@/components/common/LoadingState'
import { ErrorAlert } from '@/components/common/ErrorAlert'

export type EntityField = { label: string; value: ReactNode }

export function EntityViewPage({
  title,
  backTo,
  backLabel,
  editTo,
  onDelete,
  fields,
  loading,
  error,
}: {
  title: string
  backTo: string
  backLabel: string
  editTo?: string
  onDelete?: () => void
  fields: EntityField[]
  loading?: boolean
  error?: string | null
}) {
  if (loading) return <LoadingState />
  if (error) return <ErrorAlert message={error} />

  return (
    <div className="mx-auto w-full max-w-2xl space-y-4">
      <Card>
        <CardHeader className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <CardTitle className="text-xl">{title}</CardTitle>
          <div className="flex flex-wrap gap-2">
            {editTo && (
              <Button variant="outline" size="sm" asChild>
                <Link to={editTo}>Редактировать</Link>
              </Button>
            )}
            {onDelete && (
              <Button variant="destructive" size="sm" onClick={onDelete}>
                Удалить
              </Button>
            )}
          </div>
        </CardHeader>
        <CardContent>
          <dl className="grid gap-4 sm:grid-cols-[minmax(120px,auto)_1fr]">
            {fields.map(({ label, value }) => (
              <div key={label} className="contents">
                <dt className="text-sm font-medium text-muted-foreground">{label}</dt>
                <dd className="text-sm">{value}</dd>
              </div>
            ))}
          </dl>
        </CardContent>
      </Card>
      <Button variant="link" className="h-auto p-0" asChild>
        <Link to={backTo}>← {backLabel}</Link>
      </Button>
    </div>
  )
}
