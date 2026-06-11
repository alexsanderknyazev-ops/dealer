import { useCallback, useEffect, useState } from 'react'
import { Plus } from 'lucide-react'
import { useAuth } from '@/auth'
import * as api from '@/api'
import { PageHeader } from '@/components/common/PageHeader'
import { LoadingState } from '@/components/common/LoadingState'
import { ErrorAlert } from '@/components/common/ErrorAlert'
import { EmptyState } from '@/components/common/EmptyState'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

export function Vehicles() {
  const { getAccessToken } = useAuth()
  const [list, setList] = useState<api.ClientVehicle[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [vin, setVin] = useState('')
  const [adding, setAdding] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const token = await getAccessToken()
      if (!token) throw new Error('Сессия истекла')
      const r = await api.listVehicles(token)
      setList(r.vehicles ?? [])
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка загрузки')
    } finally {
      setLoading(false)
    }
  }, [getAccessToken])

  useEffect(() => {
    load()
  }, [load])

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault()
    const value = vin.trim()
    if (!value) return
    setAdding(true)
    setError(null)
    try {
      const token = await getAccessToken()
      if (!token) throw new Error('Сессия истекла')
      await api.addVehicle(token, value)
      setVin('')
      await load()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось добавить автомобиль')
    } finally {
      setAdding(false)
    }
  }

  return (
    <div className="mx-auto w-full max-w-4xl space-y-4">
      <PageHeader title="Мои автомобили" />

      <form onSubmit={handleAdd} className="flex flex-col gap-2 sm:flex-row">
        <Input
          placeholder="VIN нового автомобиля"
          value={vin}
          onChange={(e) => setVin(e.target.value.toUpperCase())}
          disabled={adding}
        />
        <Button type="submit" disabled={adding || !vin.trim()}>
          <Plus className="mr-2 h-4 w-4" />
          Добавить
        </Button>
      </form>

      {error && <ErrorAlert message={error} onRetry={load} />}
      {loading ? (
        <LoadingState />
      ) : list.length === 0 && !error ? (
        <EmptyState>Нет привязанных автомобилей. Добавьте VIN выше.</EmptyState>
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>VIN</TableHead>
                  <TableHead>Марка / модель</TableHead>
                  <TableHead>Год</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {list.map((v) => (
                  <TableRow key={v.id}>
                    <TableCell className="font-mono text-xs">{v.vin}</TableCell>
                    <TableCell>
                      {v.make} {v.model}
                    </TableCell>
                    <TableCell>{v.year || '—'}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
