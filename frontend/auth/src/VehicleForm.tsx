import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import type { VehicleForm as VehicleFormType } from './vehiclesApi'
import * as api from './vehiclesApi'
import * as brandsApi from './brandsApi'
import * as dealerPointsApi from './dealerPointsApi'
import { FormPage } from '@/components/common/FormPage'
import { FormField } from '@/components/common/FormField'
import { FormActions } from '@/components/common/FormActions'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'
import { NativeSelect } from '@/components/ui/native-select'
import { Textarea } from '@/components/ui/textarea'

export function VehicleForm() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isNew = id === 'new' || !id
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [brands, setBrands] = useState<brandsApi.Brand[]>([])
  const [points, setPoints] = useState<dealerPointsApi.DealerPoint[]>([])
  const [legalEntities, setLegalEntities] = useState<dealerPointsApi.LegalEntity[]>([])
  const [warehouses, setWarehouses] = useState<dealerPointsApi.Warehouse[]>([])
  const [form, setForm] = useState<VehicleFormType>({
    vin: '',
    make: '',
    model: '',
    year: new Date().getFullYear(),
    mileage_km: 0,
    price: '',
    status: 'available',
    color: '',
    notes: '',
    brand_id: undefined,
    dealer_point_id: undefined,
    legal_entity_id: undefined,
    warehouse_id: undefined,
  })

  useEffect(() => {
    brandsApi.listBrands({ limit: 500 }).then((r) => setBrands(r.brands)).catch(() => setBrands([]))
    dealerPointsApi.listDealerPoints({ limit: 200 }).then((r) => setPoints(r.dealer_points)).catch(() => setPoints([]))
  }, [])

  useEffect(() => {
    if (!form.dealer_point_id) {
      setLegalEntities([])
      setWarehouses([])
      return
    }
    dealerPointsApi.listLegalEntitiesByDealerPoint(form.dealer_point_id).then(setLegalEntities).catch(() => setLegalEntities([]))
  }, [form.dealer_point_id])

  useEffect(() => {
    if (!form.dealer_point_id || !form.legal_entity_id) {
      setWarehouses([])
      return
    }
    dealerPointsApi
      .listWarehouses({ limit: 200, dealer_point_id: form.dealer_point_id, legal_entity_id: form.legal_entity_id, type: 'cars' })
      .then((r) => setWarehouses(r.warehouses))
      .catch(() => setWarehouses([]))
  }, [form.dealer_point_id, form.legal_entity_id])

  useEffect(() => {
    if (isNew) return
    api
      .getVehicle(id!)
      .then((v) => {
        setForm({
          vin: v.vin,
          make: v.make || '',
          model: v.model || '',
          year: v.year,
          mileage_km: v.mileage_km ?? 0,
          price: v.price || '',
          status: v.status || 'available',
          color: v.color || '',
          notes: v.notes || '',
          brand_id: v.brand_id || undefined,
          dealer_point_id: v.dealer_point_id || undefined,
          legal_entity_id: v.legal_entity_id || undefined,
          warehouse_id: v.warehouse_id || undefined,
        })
      })
      .catch((e) => setError(e instanceof Error ? e.message : 'Ошибка загрузки'))
      .finally(() => setLoading(false))
  }, [id, isNew])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    const payload = {
      vin: form.vin,
      make: form.make || undefined,
      model: form.model || undefined,
      year: form.year,
      mileage_km: form.mileage_km,
      price: form.price || undefined,
      status: form.status || undefined,
      color: form.color || undefined,
      notes: form.notes || undefined,
      brand_id: form.brand_id ?? '',
      dealer_point_id: form.dealer_point_id ?? '',
      legal_entity_id: form.legal_entity_id ?? '',
      warehouse_id: form.warehouse_id ?? '',
    }
    const save = isNew ? api.createVehicle(payload) : api.updateVehicle(id!, payload)
    save
      .then(() => navigate('/vehicles', { replace: true }))
      .catch((err) => {
        setError(err instanceof Error ? err.message : 'Ошибка сохранения')
        setSubmitting(false)
      })
  }

  return (
    <FormPage title={isNew ? 'Новый автомобиль' : 'Редактирование автомобиля'} loading={loading}>
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        <FormField label="VIN" htmlFor="vin" required>
          <Input id="vin" value={form.vin} onChange={(e) => setForm((f) => ({ ...f, vin: e.target.value }))} required placeholder="WVWZZZ3CZWE123456" />
        </FormField>
        <FormField label="Бренд" htmlFor="brand_id">
          <NativeSelect id="brand_id" className="max-w-xs" value={form.brand_id ?? ''} onChange={(e) => setForm((f) => ({ ...f, brand_id: e.target.value || undefined }))}>
            <option value="">— не выбран —</option>
            {brands.map((b) => (
              <option key={b.id} value={b.id}>{b.name}</option>
            ))}
          </NativeSelect>
        </FormField>
        <div className="grid gap-4 sm:grid-cols-3">
          <FormField label="Дилерская точка" htmlFor="dealer_point_id">
            <NativeSelect id="dealer_point_id" value={form.dealer_point_id ?? ''} onChange={(e) => setForm((f) => ({ ...f, dealer_point_id: e.target.value || undefined, legal_entity_id: undefined, warehouse_id: undefined }))}>
              <option value="">— не выбрана —</option>
              {points.map((p) => (
                <option key={p.id} value={p.id}>{p.name}</option>
              ))}
            </NativeSelect>
          </FormField>
          <FormField label="Юр. лицо" htmlFor="legal_entity_id">
            <NativeSelect id="legal_entity_id" value={form.legal_entity_id ?? ''} disabled={!form.dealer_point_id} onChange={(e) => setForm((f) => ({ ...f, legal_entity_id: e.target.value || undefined, warehouse_id: undefined }))}>
              <option value="">— не выбрано —</option>
              {legalEntities.map((ent) => (
                <option key={ent.id} value={ent.id}>{ent.name}</option>
              ))}
            </NativeSelect>
          </FormField>
          <FormField label="Склад автомобилей" htmlFor="warehouse_id">
            <NativeSelect id="warehouse_id" value={form.warehouse_id ?? ''} disabled={!form.legal_entity_id} onChange={(e) => setForm((f) => ({ ...f, warehouse_id: e.target.value || undefined }))}>
              <option value="">— не выбран —</option>
              {warehouses.map((w) => (
                <option key={w.id} value={w.id}>{w.name}</option>
              ))}
            </NativeSelect>
          </FormField>
        </div>
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField label="Марка" htmlFor="make">
            <Input id="make" value={form.make} onChange={(e) => setForm((f) => ({ ...f, make: e.target.value }))} />
          </FormField>
          <FormField label="Модель" htmlFor="model">
            <Input id="model" value={form.model} onChange={(e) => setForm((f) => ({ ...f, model: e.target.value }))} />
          </FormField>
        </div>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <FormField label="Год" htmlFor="year">
            <Input id="year" type="number" min={1900} max={2100} value={form.year ?? ''} onChange={(e) => setForm((f) => ({ ...f, year: e.target.value ? parseInt(e.target.value, 10) : undefined }))} />
          </FormField>
          <FormField label="Пробег (км)" htmlFor="mileage_km">
            <Input id="mileage_km" type="number" min={0} value={form.mileage_km ?? ''} onChange={(e) => setForm((f) => ({ ...f, mileage_km: e.target.value ? parseInt(e.target.value, 10) : 0 }))} />
          </FormField>
          <FormField label="Цена" htmlFor="price">
            <Input id="price" value={form.price} onChange={(e) => setForm((f) => ({ ...f, price: e.target.value }))} />
          </FormField>
          <FormField label="Статус" htmlFor="status">
            <NativeSelect id="status" value={form.status} onChange={(e) => setForm((f) => ({ ...f, status: e.target.value }))}>
              <option value="available">В наличии</option>
              <option value="reserved">Зарезервирован</option>
              <option value="sold">Продан</option>
            </NativeSelect>
          </FormField>
        </div>
        <FormField label="Цвет" htmlFor="color">
          <Input id="color" value={form.color} onChange={(e) => setForm((f) => ({ ...f, color: e.target.value }))} />
        </FormField>
        <FormField label="Заметки" htmlFor="notes">
          <Textarea id="notes" value={form.notes} onChange={(e) => setForm((f) => ({ ...f, notes: e.target.value }))} rows={3} />
        </FormField>
        <FormActions submitting={submitting} submitLabel={isNew ? 'Создать' : 'Сохранить'} onCancel={() => navigate('/vehicles')} />
      </form>
    </FormPage>
  )
}
