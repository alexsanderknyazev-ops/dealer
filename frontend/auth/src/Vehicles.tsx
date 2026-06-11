import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Plus } from 'lucide-react'
import type { Vehicle } from './vehiclesApi'
import * as api from './vehiclesApi'
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
import { Badge } from '@/components/ui/badge'

const STATUS_LABEL: Record<string, string> = {
  available: 'В наличии',
  sold: 'Продан',
  reserved: 'Зарезервирован',
}

export function Vehicles() {
  const [list, setList] = useState<Vehicle[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [page, setPage] = useState(0)
  const limit = 20

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    api
      .listVehicles({ limit, offset: page * limit, search: search || undefined, status: statusFilter || undefined })
      .then((r) => {
        if (!cancelled) {
          setList(r.vehicles)
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
  }, [page, search, statusFilter])

  return (
    <div className="mx-auto w-full max-w-6xl">
      <PageHeader
        title="Автомобили"
        action={
          <Button asChild>
            <Link to="/vehicles/new">
              <Plus className="mr-2 h-4 w-4" />
              Добавить
            </Link>
          </Button>
        }
      />
      <div className="mb-4 flex flex-col gap-3 sm:flex-row">
        <Input
          type="search"
          placeholder="Поиск по VIN, марке, модели..."
          value={search}
          onChange={(e) => {
            setSearch(e.target.value)
            setPage(0)
          }}
        />
        <NativeSelect
          className="sm:max-w-[200px]"
          value={statusFilter}
          onChange={(e) => {
            setStatusFilter(e.target.value)
            setPage(0)
          }}
        >
          <option value="">Все статусы</option>
          <option value="available">В наличии</option>
          <option value="reserved">Зарезервирован</option>
          <option value="sold">Продан</option>
        </NativeSelect>
      </div>
      {error && <ErrorAlert message={`${error}. Запустите vehicles-service: make run-vehicles`} />}
      {loading ? (
        <LoadingState />
      ) : list.length === 0 && !error ? (
        <EmptyState>Нет автомобилей</EmptyState>
      ) : (
        <>
          <Card>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>VIN</TableHead>
                    <TableHead>Марка / Модель</TableHead>
                    <TableHead>Год</TableHead>
                    <TableHead>Пробег</TableHead>
                    <TableHead>Цена</TableHead>
                    <TableHead>Статус</TableHead>
                    <TableHead />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {list.map((v) => (
                    <TableRow key={v.id}>
                      <TableCell className="font-mono text-xs">{v.vin}</TableCell>
                      <TableCell>
                        {v.make} {v.model}
                      </TableCell>
                      <TableCell>{v.year}</TableCell>
                      <TableCell>{v.mileage_km.toLocaleString('ru')} км</TableCell>
                      <TableCell>{v.price ? Number(v.price).toLocaleString('ru') : '—'}</TableCell>
                      <TableCell>
                        <Badge variant="secondary">{STATUS_LABEL[v.status] || v.status}</Badge>
                      </TableCell>
                      <TableCell>
                        <Button variant="link" className="h-auto p-0" asChild>
                          <Link to={`/vehicles/${v.id}`}>Открыть</Link>
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
