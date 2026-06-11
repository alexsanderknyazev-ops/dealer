import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { Plus } from 'lucide-react'
import type { LegalEntity } from './dealerPointsApi'
import * as api from './dealerPointsApi'
import { PageHeader } from '@/components/common/PageHeader'
import { ErrorAlert } from '@/components/common/ErrorAlert'
import { LoadingState } from '@/components/common/LoadingState'
import { EmptyState } from '@/components/common/EmptyState'
import { Pagination } from '@/components/common/Pagination'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { NativeSelect } from '@/components/ui/native-select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

export function LegalEntities() {
  const [searchParams, setSearchParams] = useSearchParams()
  const dealerPointId = searchParams.get('dealer_point_id') || ''
  const [list, setList] = useState<LegalEntity[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(0)
  const [points, setPoints] = useState<{ id: string; name: string }[]>([])
  const limit = 20

  useEffect(() => {
    api.listDealerPoints({ limit: 200 }).then((r) => setPoints(r.dealer_points)).catch(() => setPoints([]))
  }, [])

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    if (dealerPointId) {
      api
        .listLegalEntitiesByDealerPoint(dealerPointId)
        .then((linked) => {
          if (!cancelled) {
            const filtered = search
              ? linked.filter(
                  (e) =>
                    e.name.toLowerCase().includes(search.toLowerCase()) || (e.inn && e.inn.includes(search)),
                )
              : linked
            setList(filtered.slice(page * limit, (page + 1) * limit))
            setTotal(filtered.length)
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
    } else {
      api
        .listLegalEntities({ limit, offset: page * limit, search: search || undefined })
        .then((r) => {
          if (!cancelled) {
            setList(r.legal_entities)
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
    }
    return () => {
      cancelled = true
    }
  }, [page, search, dealerPointId])

  const pointName = points.find((p) => p.id === dealerPointId)?.name

  return (
    <div className="mx-auto w-full max-w-5xl">
      <PageHeader
        title={
          <>
            Юридические лица
            {pointName && <span className="ml-2 text-base font-normal text-muted-foreground">— {pointName}</span>}
          </>
        }
        action={
          <Button asChild>
            <Link to="/legal-entities/new">
              <Plus className="mr-2 h-4 w-4" />
              Добавить
            </Link>
          </Button>
        }
      />
      <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center">
        {dealerPointId && (
          <Button variant="link" className="h-auto justify-start p-0" asChild>
            <Link to="/legal-entities">Все юр. лица</Link>
          </Button>
        )}
        {points.length > 0 && (
          <NativeSelect
            className="sm:max-w-xs"
            value={dealerPointId}
            onChange={(e) => {
              const v = e.target.value
              setPage(0)
              const next = new URLSearchParams(searchParams)
              if (v) next.set('dealer_point_id', v)
              else next.delete('dealer_point_id')
              setSearchParams(next)
            }}
          >
            <option value="">Все дилерские точки</option>
            {points.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </NativeSelect>
        )}
        <Input
          type="search"
          placeholder="Поиск по названию, ИНН..."
          value={search}
          onChange={(e) => {
            setSearch(e.target.value)
            setPage(0)
          }}
        />
      </div>
      {error && <ErrorAlert message={error} />}
      {loading ? (
        <LoadingState />
      ) : list.length === 0 && !error ? (
        <EmptyState>Нет юридических лиц.</EmptyState>
      ) : (
        <>
          <Card>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Название</TableHead>
                    <TableHead>ИНН</TableHead>
                    <TableHead>Адрес</TableHead>
                    <TableHead />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {list.map((e) => (
                    <TableRow key={e.id}>
                      <TableCell>{e.name}</TableCell>
                      <TableCell>{e.inn || '—'}</TableCell>
                      <TableCell>{e.address || '—'}</TableCell>
                      <TableCell>
                        <Button variant="link" className="h-auto p-0" asChild>
                          <Link to={`/legal-entities/${e.id}/edit`}>Изменить</Link>
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
