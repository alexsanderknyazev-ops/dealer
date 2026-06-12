import { useEffect, useState } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { Plus, Trash2 } from 'lucide-react'
import * as api from './supplierOrdersApi'
import * as partsApi from './partsApi'
import * as suppliersApi from './suppliersApi'
import * as dealerPointsApi from './dealerPointsApi'
import { useAuth } from './auth'
import { FormPage } from '@/components/common/FormPage'
import { FormField } from '@/components/common/FormField'
import { FormActions } from '@/components/common/FormActions'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'
import { NativeSelect } from '@/components/ui/native-select'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

const emptyLine = (): api.PartOrderLineInput => ({ part_id: '', quantity: 1, unit_price: '' })

export function SupplierOrderForm() {
  const { id } = useParams()
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { logout, user } = useAuth()
  const isNew = id === 'new' || !id
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [parts, setParts] = useState<partsApi.Part[]>([])
  const [suppliers, setSuppliers] = useState<suppliersApi.Supplier[]>([])
  const [warehouses, setWarehouses] = useState<dealerPointsApi.Warehouse[]>([])
  const [form, setForm] = useState<api.SupplierOrderForm>({
    supplier_id: '',
    receipt_warehouse_id: '',
    lines: [emptyLine()],
  })

  useEffect(() => {
    const customerOrderId = searchParams.get('customer_order_id')
    if (isNew && customerOrderId) {
      import('./customerOrdersApi')
        .then((customerOrdersApi) => customerOrdersApi.getCustomerOrder(customerOrderId))
        .then((order) => {
          setForm((f) => ({
            ...f,
            customer_order_id: order.id,
            receipt_warehouse_id: order.issue_warehouse_id || f.receipt_warehouse_id,
            lines: order.lines?.length
              ? order.lines.map((l) => ({
                  part_id: l.part_id,
                  quantity: l.quantity,
                  unit_price: l.unit_price,
                  notes: l.notes,
                  sort_order: l.sort_order,
                }))
              : f.lines,
          }))
        })
        .catch(() => {})
    }
  }, [isNew, searchParams])

  useEffect(() => {
    partsApi.listParts({ limit: 500 }).then((r) => setParts(r.parts)).catch(() => {})
    suppliersApi.listSuppliers({ limit: 500 }).then((r) => setSuppliers(r.suppliers)).catch(() => {})
    dealerPointsApi.listWarehouses({ limit: 500 }).then((r) => setWarehouses(r.warehouses)).catch(() => {})
  }, [])

  useEffect(() => {
    if (isNew) return
    api
      .getSupplierOrder(id!)
      .then((o) => {
        if (o.status !== 'draft') {
          setError('Заказ нельзя редактировать в текущем статусе')
          return
        }
        setForm({
          supplier_id: o.supplier_id,
          receipt_warehouse_id: o.receipt_warehouse_id,
          notes: o.notes,
          lines: o.lines?.length
            ? o.lines.map((l) => ({
                part_id: l.part_id,
                quantity: l.quantity,
                unit_price: l.unit_price,
                notes: l.notes,
                sort_order: l.sort_order,
              }))
            : [emptyLine()],
        })
      })
      .catch(async (e) => {
        if (e instanceof api.ApiError && (e.status === 401 || e.status === 403)) {
          await logout()
          navigate('/login', { replace: true })
          return
        }
        setError(e instanceof Error ? e.message : 'Ошибка загрузки')
      })
      .finally(() => setLoading(false))
  }, [id, isNew, logout, navigate])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    const payload: api.SupplierOrderForm = {
      ...form,
      lines: form.lines.filter((l) => l.part_id && l.quantity > 0),
    }
    if (!payload.supplier_id || !payload.receipt_warehouse_id) {
      setError('Укажите поставщика и склад поступления')
      setSubmitting(false)
      return
    }
    for (let i = 0; i < payload.lines.length; i++) {
      if (!payload.lines[i].unit_price || Number(payload.lines[i].unit_price) <= 0) {
        setError(`Строка ${i + 1}: укажите цену заказа`)
        setSubmitting(false)
        return
      }
    }
    try {
      const saved = isNew
        ? await api.createSupplierOrder(payload, user?.userId)
        : await api.updateSupplierOrder(id!, payload)
      navigate(`/supplier-orders/${saved.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка сохранения')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <FormPage title={isNew ? 'Новый заказ поставщику' : 'Редактирование заказа'} loading={loading}>
      <form onSubmit={handleSubmit} className="space-y-6">
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        <div className="grid gap-4 md:grid-cols-2">
          <FormField label="Поставщик" required>
            <NativeSelect value={form.supplier_id} onChange={(e) => setForm({ ...form, supplier_id: e.target.value })}>
              <option value="">Выберите поставщика</option>
              {suppliers.map((s) => (
                <option key={s.id} value={s.id}>{s.name}</option>
              ))}
            </NativeSelect>
          </FormField>
          <FormField label="Склад поступления" required>
            <NativeSelect
              value={form.receipt_warehouse_id}
              onChange={(e) => setForm({ ...form, receipt_warehouse_id: e.target.value })}
            >
              <option value="">Выберите склад</option>
              {warehouses.map((w) => (
                <option key={w.id} value={w.id}>{w.name}</option>
              ))}
            </NativeSelect>
          </FormField>
        </div>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-base">Запчасти</CardTitle>
            <Button type="button" variant="outline" size="sm" onClick={() => setForm((f) => ({ ...f, lines: [...f.lines, emptyLine()] }))}>
              <Plus className="mr-1 h-4 w-4" /> Добавить
            </Button>
          </CardHeader>
          <CardContent className="space-y-3">
            {form.lines.map((line, i) => (
              <div key={i} className="grid gap-2 rounded-md border p-3 md:grid-cols-[2fr_1fr_1fr_1fr_auto] md:items-center">
                <NativeSelect
                  value={line.part_id}
                  onChange={(e) => {
                    const lines = [...form.lines]
                    lines[i] = { ...lines[i], part_id: e.target.value }
                    setForm({ ...form, lines })
                  }}
                >
                  <option value="">Запчасть</option>
                  {parts.map((p) => (
                    <option key={p.id} value={p.id}>{p.sku} — {p.name}</option>
                  ))}
                </NativeSelect>
                <Input type="number" min={1} value={line.quantity} onChange={(e) => {
                  const lines = [...form.lines]
                  lines[i] = { ...lines[i], quantity: Number(e.target.value) || 0 }
                  setForm({ ...form, lines })
                }} placeholder="Кол-во" />
                <Input type="number" min={0} step="0.01" value={line.unit_price} onChange={(e) => {
                  const lines = [...form.lines]
                  lines[i] = { ...lines[i], unit_price: e.target.value }
                  setForm({ ...form, lines })
                }} placeholder="Цена" />
                <Input value={line.notes || ''} onChange={(e) => {
                  const lines = [...form.lines]
                  lines[i] = { ...lines[i], notes: e.target.value }
                  setForm({ ...form, lines })
                }} placeholder="Примечание" />
                <Button type="button" variant="ghost" size="icon" onClick={() => setForm((f) => ({ ...f, lines: f.lines.filter((_, idx) => idx !== i) }))}>
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))}
          </CardContent>
        </Card>
        <FormField label="Примечание">
          <Textarea value={form.notes || ''} onChange={(e) => setForm({ ...form, notes: e.target.value })} />
        </FormField>
        <FormActions submitting={submitting} submitLabel={isNew ? 'Создать' : 'Сохранить'} onCancel={() => navigate('/supplier-orders')} />
      </form>
    </FormPage>
  )
}
