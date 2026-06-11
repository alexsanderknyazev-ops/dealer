import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Plus } from 'lucide-react'
import type { Brand } from './brandsApi'
import * as api from './brandsApi'
import { PageHeader } from '@/components/common/PageHeader'
import { ErrorAlert } from '@/components/common/ErrorAlert'
import { LoadingState } from '@/components/common/LoadingState'
import { EmptyState } from '@/components/common/EmptyState'
import { Pagination } from '@/components/common/Pagination'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

export function Brands() {
  const [list, setList] = useState<Brand[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(0)
  const [retry, setRetry] = useState(0)
  const limit = 50

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    api
      .listBrands({ limit, offset: page * limit, search: search || undefined })
      .then((r) => {
        if (!cancelled) {
          setList(r.brands)
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
  }, [page, search, retry])

  function handleDelete(id: string, name: string) {
    if (!window.confirm(`Удалить бренд «${name}»?`)) return
    api
      .deleteBrand(id)
      .then(() => setRetry((r) => r + 1))
      .catch((e) => setError(e instanceof Error ? e.message : 'Ошибка удаления'))
  }

  return (
    <div className="mx-auto w-full max-w-4xl">
      <PageHeader
        title="Бренды"
        action={
          <Button asChild>
            <Link to="/brands/new">
              <Plus className="mr-2 h-4 w-4" />
              Добавить
            </Link>
          </Button>
        }
      />
      <div className="mb-4">
        <Input type="search" placeholder="Поиск по названию..." value={search} onChange={(e) => { setSearch(e.target.value); setPage(0) }} />
      </div>
      {error && <ErrorAlert message={error} onRetry={() => setRetry((r) => r + 1)} />}
      {loading ? (
        <LoadingState />
      ) : list.length === 0 && !error ? (
        <EmptyState>Нет брендов. Нажмите «Добавить».</EmptyState>
      ) : (
        <>
          <Card>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Название</TableHead>
                    <TableHead />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {list.map((b) => (
                    <TableRow key={b.id}>
                      <TableCell>{b.name}</TableCell>
                      <TableCell className="space-x-2 whitespace-nowrap">
                        <Button variant="link" className="h-auto p-0" asChild>
                          <Link to={`/brands/${b.id}/edit`}>Изменить</Link>
                        </Button>
                        <Button variant="link" className="h-auto p-0 text-destructive" onClick={() => handleDelete(b.id, b.name)}>
                          Удалить
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
          <Pagination page={page} total={total} limit={limit} onPageChange={setPage} showTotal />
        </>
      )}
    </div>
  )
}
