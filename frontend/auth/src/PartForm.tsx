import { useEffect, useState } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import type { PartForm as PartFormType, PartStockRow } from './partsApi'
import * as api from './partsApi'
import * as brandsApi from './brandsApi'
import * as dealerPointsApi from './dealerPointsApi'
import { FormPage } from '@/components/common/FormPage'
import { FormField } from '@/components/common/FormField'
import { FormActions } from '@/components/common/FormActions'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { NativeSelect } from '@/components/ui/native-select'
import { Textarea } from '@/components/ui/textarea'

export function PartForm() {
  const { id } = useParams()
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const isNew = id === 'new' || !id
  const defaultFolderId = searchParams.get('folder_id') || ''
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [folderOptions, setFolderOptions] = useState<{ id: string; name: string; level: number }[]>([])
  const [brands, setBrands] = useState<brandsApi.Brand[]>([])
  const [points, setPoints] = useState<dealerPointsApi.DealerPoint[]>([])
  const [legalEntities, setLegalEntities] = useState<dealerPointsApi.LegalEntity[]>([])
  const [warehouses, setWarehouses] = useState<dealerPointsApi.Warehouse[]>([])
  const [allPartsWarehouses, setAllPartsWarehouses] = useState<dealerPointsApi.Warehouse[]>([])
  const [form, setForm] = useState<PartFormType>({
    sku: '',
    name: '',
    category: '',
    folder_id: defaultFolderId || undefined,
    brand_id: undefined,
    dealer_point_id: undefined,
    legal_entity_id: undefined,
    warehouse_id: undefined,
    quantity: 0,
    unit: 'шт',
    price: '',
    location: '',
    notes: '',
    stock: [],
  })

  useEffect(() => {
    api.loadAllFoldersFlat().then(setFolderOptions).catch(() => setFolderOptions([]))
    brandsApi.listBrands({ limit: 500 }).then((r) => setBrands(r.brands)).catch(() => setBrands([]))
    dealerPointsApi.listDealerPoints({ limit: 200 }).then((r) => setPoints(r.dealer_points)).catch(() => setPoints([]))
    dealerPointsApi.listWarehouses({ limit: 200, type: 'parts' }).then((r) => setAllPartsWarehouses(r.warehouses)).catch(() => setAllPartsWarehouses([]))
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
      .listWarehouses({ limit: 200, dealer_point_id: form.dealer_point_id, legal_entity_id: form.legal_entity_id, type: 'parts' })
      .then((r) => setWarehouses(r.warehouses))
      .catch(() => setWarehouses([]))
  }, [form.dealer_point_id, form.legal_entity_id])

  useEffect(() => {
    if (isNew) {
      if (defaultFolderId) setForm((f) => ({ ...f, folder_id: defaultFolderId }))
      return
    }
    api
      .getPart(id!)
      .then((p) => {
        setForm({
          sku: p.sku,
          name: p.name || '',
          category: p.category || '',
          folder_id: p.folder_id || undefined,
          brand_id: p.brand_id || undefined,
          dealer_point_id: p.dealer_point_id || undefined,
          legal_entity_id: p.legal_entity_id || undefined,
          warehouse_id: p.warehouse_id || undefined,
          quantity: p.quantity ?? 0,
          unit: p.unit || 'шт',
          price: p.price || '',
          location: p.location || '',
          notes: p.notes || '',
          stock: p.stock && p.stock.length > 0 ? p.stock : [],
        })
      })
      .catch((e) => setError(e instanceof Error ? e.message : 'Ошибка загрузки'))
      .finally(() => setLoading(false))
  }, [id, isNew, defaultFolderId])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!form.sku?.trim()) {
      setError('Укажите артикул')
      return
    }
    setError(null)
    setSubmitting(true)
    const stockFiltered = (form.stock ?? []).filter((s) => s.warehouse_id && s.quantity >= 0)
    const payload: PartFormType & { stock?: PartStockRow[] } = {
      sku: form.sku.trim(),
      name: form.name || undefined,
      category: form.category || undefined,
      folder_id: form.folder_id || undefined,
      brand_id: form.brand_id ?? '',
      dealer_point_id: form.dealer_point_id ?? '',
      legal_entity_id: form.legal_entity_id ?? '',
      warehouse_id: form.warehouse_id ?? '',
      quantity: form.quantity,
      unit: form.unit || undefined,
      price: form.price || undefined,
      location: form.location || undefined,
      notes: form.notes || undefined,
    }
    if (stockFiltered.length > 0) payload.stock = stockFiltered
    const returnUrl = form.folder_id ? `/parts?folder_id=${form.folder_id}` : '/parts'
    const save = isNew ? api.createPart(payload) : api.updatePart(id!, payload)
    save
      .then(() => navigate(returnUrl, { replace: true }))
      .catch((err) => {
        setError(err instanceof Error ? err.message : 'Ошибка сохранения')
        setSubmitting(false)
      })
  }

  const cancelUrl = form.folder_id ? `/parts?folder_id=${form.folder_id}` : '/parts'

  return (
    <FormPage title={isNew ? 'Новая запчасть' : 'Редактирование запчасти'} loading={loading}>
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        <FormField label="Артикул (SKU)" htmlFor="sku" required>
          <Input id="sku" value={form.sku} onChange={(e) => setForm((f) => ({ ...f, sku: e.target.value }))} required placeholder="ABC-12345" />
        </FormField>
        <FormField label="Папка" htmlFor="folder_id">
          <NativeSelect id="folder_id" value={form.folder_id ?? ''} onChange={(e) => setForm((f) => ({ ...f, folder_id: e.target.value || undefined }))}>
            <option value="">Без папки</option>
            {folderOptions.map((opt) => (
              <option key={opt.id} value={opt.id}>
                {'—'.repeat(opt.level)} {opt.name}
              </option>
            ))}
          </NativeSelect>
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
          <FormField label="Склад запчастей" htmlFor="warehouse_id">
            <NativeSelect id="warehouse_id" value={form.warehouse_id ?? ''} disabled={!form.legal_entity_id} onChange={(e) => setForm((f) => ({ ...f, warehouse_id: e.target.value || undefined }))}>
              <option value="">— не выбран —</option>
              {warehouses.map((w) => (
                <option key={w.id} value={w.id}>{w.name}</option>
              ))}
            </NativeSelect>
          </FormField>
        </div>
        <div className="space-y-2">
          <p className="text-sm font-medium">Остатки по складам (запчасть может быть на нескольких складах)</p>
          {(form.stock ?? []).map((row, idx) => (
            <div key={idx} className="flex flex-wrap items-center gap-2">
              <NativeSelect
                className="min-w-[220px] flex-1"
                value={row.warehouse_id}
                onChange={(e) =>
                  setForm((f) => ({
                    ...f,
                    stock: (f.stock ?? []).map((s, i) => (i === idx ? { ...s, warehouse_id: e.target.value } : s)),
                  }))
                }
              >
                <option value="">— выберите склад —</option>
                {allPartsWarehouses.map((w) => (
                  <option key={w.id} value={w.id}>{w.name}</option>
                ))}
              </NativeSelect>
              <Input
                type="number"
                min={0}
                className="w-20"
                value={row.quantity}
                onChange={(e) =>
                  setForm((f) => ({
                    ...f,
                    stock: (f.stock ?? []).map((s, i) => (i === idx ? { ...s, quantity: parseInt(e.target.value, 10) || 0 } : s)),
                  }))
                }
              />
              <span className="text-sm text-muted-foreground">{form.unit || 'шт'}</span>
              <Button type="button" variant="outline" size="sm" onClick={() => setForm((f) => ({ ...f, stock: (f.stock ?? []).filter((_, i) => i !== idx) }))}>
                Удалить
              </Button>
            </div>
          ))}
          <Button type="button" variant="outline" size="sm" onClick={() => setForm((f) => ({ ...f, stock: [...(f.stock ?? []), { warehouse_id: '', quantity: 0 }] }))}>
            + Добавить склад
          </Button>
        </div>
        <FormField label="Название" htmlFor="name">
          <Input id="name" value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} placeholder="Масляный фильтр" />
        </FormField>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <FormField label="Категория" htmlFor="category">
            <Input id="category" value={form.category} onChange={(e) => setForm((f) => ({ ...f, category: e.target.value }))} placeholder="Фильтры, Тормоза..." />
          </FormField>
          <FormField label="Количество" htmlFor="quantity">
            <Input id="quantity" type="number" min={0} value={form.quantity ?? ''} onChange={(e) => setForm((f) => ({ ...f, quantity: e.target.value ? parseInt(e.target.value, 10) : 0 }))} />
          </FormField>
          <FormField label="Ед. изм." htmlFor="unit">
            <NativeSelect id="unit" value={form.unit} onChange={(e) => setForm((f) => ({ ...f, unit: e.target.value }))}>
              <option value="шт">шт</option>
              <option value="комплект">комплект</option>
              <option value="л">л</option>
              <option value="кг">кг</option>
            </NativeSelect>
          </FormField>
          <FormField label="Цена" htmlFor="price">
            <Input id="price" value={form.price} onChange={(e) => setForm((f) => ({ ...f, price: e.target.value }))} />
          </FormField>
        </div>
        <FormField label="Расположение (склад/полка)" htmlFor="location">
          <Input id="location" value={form.location} onChange={(e) => setForm((f) => ({ ...f, location: e.target.value }))} placeholder="Склад А, полка 12" />
        </FormField>
        <FormField label="Заметки" htmlFor="notes">
          <Textarea id="notes" value={form.notes} onChange={(e) => setForm((f) => ({ ...f, notes: e.target.value }))} rows={3} />
        </FormField>
        <FormActions submitting={submitting} submitLabel={isNew ? 'Создать' : 'Сохранить'} onCancel={() => navigate(cancelUrl)} />
      </form>
    </FormPage>
  )
}
