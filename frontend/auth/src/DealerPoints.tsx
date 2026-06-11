import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Plus } from 'lucide-react'
import type { DealerPoint } from './dealerPointsApi'
import * as api from './dealerPointsApi'
import { PageHeader } from '@/components/common/PageHeader'
import { ErrorAlert } from '@/components/common/ErrorAlert'
import { LoadingState } from '@/components/common/LoadingState'
import { EmptyState } from '@/components/common/EmptyState'
import { Pagination } from '@/components/common/Pagination'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

export function DealerPoints() {
  const [list, setList] = useState<DealerPoint[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(0)
  const limit = 20

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    api
      .listDealerPoints({ limit, offset: page * limit, search: search || undefined })
      .then((r) => {
        if (!cancelled) {
          setList(r.dealer_points)
          setTotal(r.total)
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setList([])
          setError(err instanceof Error ? err.message : 'Ошибка загрузки')
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [page, search])

  return (
    <div className="mx-auto w-full max-w-5xl">
      <PageHeader
        title="Дилерские точки"
        action={
          <Button asChild>
            <Link to="/dealer-points/new">
              <Plus className="mr-2 h-4 w-4" />
              Добавить
            </Link>
          </Button>
        }
      />
      <div className="mb-4">
        <Input type="search" placeholder="Поиск по названию, адресу..." value={search} onChange={(e) => { setSearch(e.target.value); setPage(0) }} />
      </div>
      {error && <ErrorAlert message={error} />}
      {loading ? (
        <LoadingState />
      ) : list.length === 0 && !error ? (
        <EmptyState>Нет дилерских точек. Нажмите «Добавить».</EmptyState>
      ) : (
        <>
          <Card>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Название</TableHead>
                    <TableHead>Адрес</TableHead>
                    <TableHead />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {list.map((d) => (
                    <TableRow key={d.id}>
                      <TableCell>{d.name}</TableCell>
                      <TableCell>{d.address || '—'}</TableCell>
                      <TableCell className="space-x-2 whitespace-nowrap">
                        <Button variant="link" className="h-auto p-0" asChild>
                          <Link to={`/dealer-points/${d.id}/edit`}>Изменить</Link>
                        </Button>
                        <Button variant="link" className="h-auto p-0" asChild>
                          <Link to={`/legal-entities?dealer_point_id=${d.id}`}>Юр. лица</Link>
                        </Button>
                        <Button variant="link" className="h-auto p-0" asChild>
                          <Link to={`/warehouses?dealer_point_id=${d.id}`}>Склады</Link>
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
          <Pagination page={page} total={total} limit={limit} onPageChange={setPage} />
        </>
      )}
    </div>
  )
}
