import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { Plus } from 'lucide-react'
import type { Warehouse } from './dealerPointsApi'
import * as api from './dealerPointsApi'
import { PageHeader } from '@/components/common/PageHeader'
import { ErrorAlert } from '@/components/common/ErrorAlert'
import { LoadingState } from '@/components/common/LoadingState'
import { EmptyState } from '@/components/common/EmptyState'
import { Pagination } from '@/components/common/Pagination'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { NativeSelect } from '@/components/ui/native-select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { cn } from '@/lib/utils'

const TYPE_LABEL: Record<string, string> = {
  cars: 'Автомобили',
  parts: 'Запчасти',
}

export function Warehouses() {
  const [searchParams, setSearchParams] = useSearchParams()
  const typeFilter = (searchParams.get('type') as 'cars' | 'parts' | '') || ''
  const dealerPointId = searchParams.get('dealer_point_id') || ''
  const [list, setList] = useState<Warehouse[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
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
    api
      .listWarehouses({
        limit,
        offset: page * limit,
        dealer_point_id: dealerPointId || undefined,
        type: typeFilter === 'cars' || typeFilter === 'parts' ? typeFilter : undefined,
      })
      .then((r) => {
        if (!cancelled) {
          setList(r.warehouses)
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
  }, [page, dealerPointId, typeFilter])

  function typeLink(type: '' | 'cars' | 'parts') {
    const next = new URLSearchParams()
    if (dealerPointId) next.set('dealer_point_id', dealerPointId)
    if (type) next.set('type', type)
    const qs = next.toString()
    return qs ? `/warehouses?${qs}` : '/warehouses'
  }

  return (
    <div className="mx-auto w-full max-w-5xl">
      <PageHeader
        title="Склады"
        action={
          <Button asChild>
            <Link to="/warehouses/new">
              <Plus className="mr-2 h-4 w-4" />
              Добавить
            </Link>
          </Button>
        }
      />
      <div className="mb-4 flex flex-wrap items-center gap-2">
        <span className="text-sm text-muted-foreground">Тип:</span>
        <Button variant={!typeFilter ? 'secondary' : 'ghost'} size="sm" asChild>
          <Link to={typeLink('')}>Все</Link>
        </Button>
        <Button variant={typeFilter === 'cars' ? 'secondary' : 'ghost'} size="sm" asChild>
          <Link to={typeLink('cars')}>Склады автомобилей</Link>
        </Button>
        <Button variant={typeFilter === 'parts' ? 'secondary' : 'ghost'} size="sm" asChild>
          <Link to={typeLink('parts')}>Склады запчастей</Link>
        </Button>
        {points.length > 0 && (
          <NativeSelect
            className={cn('ml-auto w-full sm:max-w-xs')}
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
      </div>
      {error && <ErrorAlert message={error} />}
      {loading ? (
        <LoadingState />
      ) : list.length === 0 && !error ? (
        <EmptyState>Нет складов. Нажмите «Добавить» или смените фильтры.</EmptyState>
      ) : (
        <>
          <Card>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Название</TableHead>
                    <TableHead>Тип</TableHead>
                    <TableHead>Дилерская точка</TableHead>
                    <TableHead />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {list.map((w) => (
                    <TableRow key={w.id}>
                      <TableCell>{w.name}</TableCell>
                      <TableCell>{TYPE_LABEL[w.type] || w.type}</TableCell>
                      <TableCell>{points.find((p) => p.id === w.dealer_point_id)?.name ?? w.dealer_point_id}</TableCell>
                      <TableCell>
                        <Button variant="link" className="h-auto p-0" asChild>
                          <Link to={`/warehouses/${w.id}/edit`}>Изменить</Link>
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
