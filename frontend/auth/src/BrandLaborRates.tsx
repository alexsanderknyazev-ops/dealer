import { useEffect, useState } from 'react'
import { Plus } from 'lucide-react'
import type { BrandLaborRate, BrandLaborRateForm } from './brandsApi'
import * as brandsApi from './brandsApi'
import * as dealerPointsApi from './dealerPointsApi'
import { PageHeader } from '@/components/common/PageHeader'
import { ErrorAlert } from '@/components/common/ErrorAlert'
import { LoadingState } from '@/components/common/LoadingState'
import { EmptyState } from '@/components/common/EmptyState'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { NativeSelect } from '@/components/ui/native-select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { FormField } from '@/components/common/FormField'

const emptyForm = (): BrandLaborRateForm => ({
  brand_id: '',
  dealer_point_id: '',
  warranty_hour_price: '0',
  commercial_hour_price: '0',
})

export function BrandLaborRates() {
  const [list, setList] = useState<BrandLaborRate[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [retry, setRetry] = useState(0)
  const [brands, setBrands] = useState<brandsApi.Brand[]>([])
  const [points, setPoints] = useState<dealerPointsApi.DealerPoint[]>([])
  const [form, setForm] = useState<BrandLaborRateForm>(emptyForm)
  const [saving, setSaving] = useState(false)
  const [filterBrand, setFilterBrand] = useState('')
  const [filterPoint, setFilterPoint] = useState('')

  const brandName = (id: string) => brands.find((b) => b.id === id)?.name || id
  const pointName = (id: string) => points.find((p) => p.id === id)?.name || id

  function load() {
    setLoading(true)
    setError(null)
    brandsApi
      .listBrandLaborRates({
        limit: 500,
        brand_id: filterBrand || undefined,
        dealer_point_id: filterPoint || undefined,
      })
      .then((r) => setList(r.brand_labor_rates))
      .catch((e) => setError(e instanceof Error ? e.message : 'Ошибка загрузки'))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    brandsApi.listBrands({ limit: 500 }).then((r) => setBrands(r.brands)).catch(() => {})
    dealerPointsApi.listDealerPoints({ limit: 200 }).then((r) => setPoints(r.dealer_points)).catch(() => {})
  }, [])

  useEffect(() => {
    load()
  }, [retry, filterBrand, filterPoint])

  async function handleSave(e: React.FormEvent) {
    e.preventDefault()
    if (!form.brand_id || !form.dealer_point_id) {
      setError('Выберите бренд и дилерскую точку')
      return
    }
    setSaving(true)
    setError(null)
    try {
      await brandsApi.updateBrandLaborRate(form)
      setForm(emptyForm())
      setRetry((r) => r + 1)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка сохранения')
    } finally {
      setSaving(false)
    }
  }

  function handleDelete(id: string) {
    if (!window.confirm('Удалить тариф нормо-часа?')) return
    brandsApi
      .deleteBrandLaborRate(id)
      .then(() => setRetry((r) => r + 1))
      .catch((e) => setError(e instanceof Error ? e.message : 'Ошибка удаления'))
  }

  function startEdit(row: BrandLaborRate) {
    setForm({
      brand_id: row.brand_id,
      dealer_point_id: row.dealer_point_id,
      warranty_hour_price: row.warranty_hour_price,
      commercial_hour_price: row.commercial_hour_price,
    })
  }

  return (
    <div className="mx-auto w-full max-w-5xl space-y-6">
      <PageHeader title="Стоимость нормо-часа" subtitle="Тарифы по бренду и дилерской точке: гарантийный и коммерческий" />

      {error && <ErrorAlert message={error} onRetry={() => setRetry((r) => r + 1)} />}

      <Card>
        <CardContent className="pt-6">
          <form onSubmit={handleSave} className="grid gap-4 md:grid-cols-2 lg:grid-cols-5">
            <FormField label="Бренд" required>
              <NativeSelect
                value={form.brand_id}
                onChange={(e) => setForm((f) => ({ ...f, brand_id: e.target.value }))}
                required
              >
                <option value="">Выберите бренд</option>
                {brands.map((b) => (
                  <option key={b.id} value={b.id}>{b.name}</option>
                ))}
              </NativeSelect>
            </FormField>
            <FormField label="Дилерская точка" required>
              <NativeSelect
                value={form.dealer_point_id}
                onChange={(e) => setForm((f) => ({ ...f, dealer_point_id: e.target.value }))}
                required
              >
                <option value="">Выберите точку</option>
                {points.map((p) => (
                  <option key={p.id} value={p.id}>{p.name}</option>
                ))}
              </NativeSelect>
            </FormField>
            <FormField label="Гарантийный н/ч, ₽" required>
              <Input
                type="number"
                min={0}
                step="0.01"
                value={form.warranty_hour_price}
                onChange={(e) => setForm((f) => ({ ...f, warranty_hour_price: e.target.value }))}
                required
              />
            </FormField>
            <FormField label="Коммерческий н/ч, ₽" required>
              <Input
                type="number"
                min={0}
                step="0.01"
                value={form.commercial_hour_price}
                onChange={(e) => setForm((f) => ({ ...f, commercial_hour_price: e.target.value }))}
                required
              />
            </FormField>
            <div className="flex items-end gap-2">
              <Button type="submit" disabled={saving} className="w-full">
                <Plus className="mr-2 h-4 w-4" />
                {saving ? 'Сохранение…' : 'Сохранить'}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>

      <div className="grid gap-4 md:grid-cols-2">
        <FormField label="Фильтр: бренд">
          <NativeSelect value={filterBrand} onChange={(e) => setFilterBrand(e.target.value)}>
            <option value="">Все бренды</option>
            {brands.map((b) => (
              <option key={b.id} value={b.id}>{b.name}</option>
            ))}
          </NativeSelect>
        </FormField>
        <FormField label="Фильтр: дилерская точка">
          <NativeSelect value={filterPoint} onChange={(e) => setFilterPoint(e.target.value)}>
            <option value="">Все точки</option>
            {points.map((p) => (
              <option key={p.id} value={p.id}>{p.name}</option>
            ))}
          </NativeSelect>
        </FormField>
      </div>

      {loading ? (
        <LoadingState />
      ) : list.length === 0 ? (
        <EmptyState>Тарифы не заданы. Добавьте первую запись выше.</EmptyState>
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Бренд</TableHead>
                  <TableHead>Дилерская точка</TableHead>
                  <TableHead>Гарантийный н/ч</TableHead>
                  <TableHead>Коммерческий н/ч</TableHead>
                  <TableHead />
                </TableRow>
              </TableHeader>
              <TableBody>
                {list.map((row) => (
                  <TableRow key={row.id}>
                    <TableCell>{brandName(row.brand_id)}</TableCell>
                    <TableCell>{pointName(row.dealer_point_id)}</TableCell>
                    <TableCell>{row.warranty_hour_price}</TableCell>
                    <TableCell>{row.commercial_hour_price}</TableCell>
                    <TableCell className="space-x-2 whitespace-nowrap">
                      <Button variant="link" className="h-auto p-0" onClick={() => startEdit(row)}>
                        Изменить
                      </Button>
                      <Button variant="link" className="h-auto p-0 text-destructive" onClick={() => handleDelete(row.id)}>
                        Удалить
                      </Button>
                    </TableCell>
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
