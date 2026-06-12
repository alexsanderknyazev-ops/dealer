import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Plus } from 'lucide-react'
import type { Employee } from './employeesApi'
import * as api from './employeesApi'
import { PageHeader } from '@/components/common/PageHeader'
import { ErrorAlert } from '@/components/common/ErrorAlert'
import { LoadingState } from '@/components/common/LoadingState'
import { EmptyState } from '@/components/common/EmptyState'
import { Pagination } from '@/components/common/Pagination'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { NativeSelect } from '@/components/ui/native-select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

export function Employees() {
  const [list, setList] = useState<Employee[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [positionFilter, setPositionFilter] = useState('')
  const [activeOnly, setActiveOnly] = useState(true)
  const [page, setPage] = useState(0)
  const [retry, setRetry] = useState(0)
  const limit = 20

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    api
      .listEmployees({
        limit,
        offset: page * limit,
        search: search || undefined,
        position: positionFilter || undefined,
        active_only: activeOnly,
      })
      .then((r) => {
        if (!cancelled) {
          setList(r.employees)
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
  }, [page, search, positionFilter, activeOnly, retry])

  return (
    <div className="mx-auto w-full max-w-5xl">
      <PageHeader
        title="Сотрудники"
        subtitle="Справочник персонала СТО и связь с учётными записями"
        action={
          <Button asChild>
            <Link to="/employees/new">
              <Plus className="mr-2 h-4 w-4" />
              Добавить
            </Link>
          </Button>
        }
      />

      <div className="mb-4 flex flex-wrap gap-3">
        <Input
          className="max-w-sm"
          type="search"
          placeholder="Поиск по ФИО, телефону..."
          value={search}
          onChange={(e) => {
            setSearch(e.target.value)
            setPage(0)
          }}
        />
        <NativeSelect
          className="max-w-[220px]"
          value={positionFilter}
          onChange={(e) => {
            setPositionFilter(e.target.value)
            setPage(0)
          }}
        >
          <option value="">Все должности</option>
          {Object.entries(api.POSITION_LABEL).map(([value, label]) => (
            <option key={value} value={value}>{label}</option>
          ))}
        </NativeSelect>
        <NativeSelect
          className="max-w-[180px]"
          value={activeOnly ? 'active' : 'all'}
          onChange={(e) => {
            setActiveOnly(e.target.value === 'active')
            setPage(0)
          }}
        >
          <option value="active">Только активные</option>
          <option value="all">Все</option>
        </NativeSelect>
      </div>

      {error && <ErrorAlert message={error} onRetry={() => setRetry((r) => r + 1)} />}

      {loading ? (
        <LoadingState />
      ) : list.length === 0 && !error ? (
        <EmptyState>Нет сотрудников. Нажмите «Добавить».</EmptyState>
      ) : (
        <>
          <Card>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>ФИО</TableHead>
                    <TableHead>Должность</TableHead>
                    <TableHead>Отдел</TableHead>
                    <TableHead>Телефон</TableHead>
                    <TableHead>Статус</TableHead>
                    <TableHead />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {list.map((e) => (
                    <TableRow key={e.id}>
                      <TableCell className="font-medium">{e.full_name}</TableCell>
                      <TableCell>{api.positionLabel(e.position)}</TableCell>
                      <TableCell>{e.department || '—'}</TableCell>
                      <TableCell>{e.phone || '—'}</TableCell>
                      <TableCell>
                        <Badge variant={e.active ? 'default' : 'secondary'}>
                          {e.active ? 'Активен' : 'Неактивен'}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Button variant="link" className="h-auto p-0" asChild>
                          <Link to={`/employees/${e.id}`}>Открыть</Link>
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
