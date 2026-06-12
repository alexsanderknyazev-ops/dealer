import { useCallback, useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Plus, Trash2 } from 'lucide-react'
import type { MovementDocumentForm as FormType, MovementDocumentLineInput } from './movementDocumentsApi'
import * as api from './movementDocumentsApi'
import * as partsApi from './partsApi'
import type { Warehouse } from './dealerPointsApi'
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
import { cn } from '@/lib/utils'

const emptyLine = (): MovementDocumentLineInput => ({
  part_id: '',
  warehouse_id: '',
  destination_warehouse_id: '',
  quantity: 1,
})

const MOVEMENT_DEST_LABEL: Record<string, string> = {
  transfer: 'Другой склад',
  to_production: 'Производство',
  work_order_issue: 'В работу (заказ-наряд)',
  from_production: 'Склад (возврат)',
}

export function MovementDocumentForm() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { logout, user } = useAuth()
  const isNew = id === 'new' || !id
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [parts, setParts] = useState<partsApi.Part[]>([])
  const [warehouses, setWarehouses] = useState<Warehouse[]>([])
  const [stockByLine, setStockByLine] = useState<Record<number, number | null>>({})
  const [form, setForm] = useState<FormType>({
    movement_type: 'transfer',
    lines: [emptyLine()],
  })

  const needsDestination = form.movement_type === 'transfer'
  const isFromProduction = form.movement_type === 'from_production'

  useEffect(() => {
    partsApi.listParts({ limit: 500 }).then((r) => setParts(r.parts)).catch(() => {})
    dealerPointsApi.listWarehouses({ limit: 500 }).then((r) => setWarehouses(r.warehouses)).catch(() => {})
  }, [])

  const refreshLineStock = useCallback(async (lineIndex: number, partId: string, warehouseId: string) => {
    if (!partId || !warehouseId) {
      setStockByLine((prev) => ({ ...prev, [lineIndex]: null }))
      return
    }
    try {
      const { stock } = await partsApi.listPartStock(partId)
      const row = stock.find((s) => s.warehouse_id === warehouseId)
      setStockByLine((prev) => ({ ...prev, [lineIndex]: row?.quantity ?? 0 }))
    } catch {
      setStockByLine((prev) => ({ ...prev, [lineIndex]: null }))
    }
  }, [])

  useEffect(() => {
    if (isNew) return
    api
      .getMovementDocument(id!)
      .then((doc) => {
        if (doc.status !== 'draft' && doc.status !== 'in_progress') {
          setError('Документ нельзя редактировать в текущем статусе')
          return
        }
        setForm({
          movement_type: doc.movement_type,
          notes: doc.notes,
          lines: doc.lines?.length
            ? doc.lines.map((l) => ({
                part_id: l.part_id,
                warehouse_id: l.warehouse_id,
                destination_warehouse_id: l.destination_warehouse_id || '',
                quantity: l.quantity,
                notes: l.notes,
                sort_order: l.sort_order,
              }))
            : [emptyLine()],
        })
        doc.lines?.forEach((l, i) => {
          if (l.part_id && l.warehouse_id) {
            void refreshLineStock(i, l.part_id, l.warehouse_id)
          }
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
  }, [id, isNew, logout, navigate, refreshLineStock])

  function updateLine(index: number, patch: Partial<MovementDocumentLineInput>) {
    setForm((f) => {
      const lines = [...(f.lines || [])]
      lines[index] = { ...lines[index], ...patch }
      return { ...f, lines }
    })
    const next = { ...(form.lines[index] || emptyLine()), ...patch }
    if (next.part_id && next.warehouse_id) {
      void refreshLineStock(index, next.part_id, next.warehouse_id)
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    const payload: FormType = {
      ...form,
      lines: form.lines.filter((l) => l.part_id && l.warehouse_id && l.quantity > 0),
    }
    for (let i = 0; i < payload.lines.length; i++) {
      const l = payload.lines[i]
      const available = stockByLine[i]
      if (
        form.movement_type !== 'from_production' &&
        available != null &&
        l.quantity > available
      ) {
        setError(`Строка ${i + 1}: количество ${l.quantity} больше остатка на складе (${available})`)
        setSubmitting(false)
        return
      }
      if (needsDestination && !l.destination_warehouse_id) {
        setError(`Строка ${i + 1}: укажите склад назначения`)
        setSubmitting(false)
        return
      }
    }
    try {
      const saved = isNew
        ? await api.createMovementDocument(payload, user?.userId)
        : await api.updateMovementDocument(id!, payload)
      navigate(`/movement-documents/${saved.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка сохранения')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <FormPage title={isNew ? 'Новый документ перемещения' : 'Редактирование документа'} loading={loading}>
      <form onSubmit={handleSubmit} className="space-y-6">
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        <div className="grid gap-4 md:grid-cols-2">
          <FormField label="Тип перемещения">
            {isFromProduction ? (
              <p className="text-sm text-muted-foreground">Извлечение из производства (создаётся из закрытого документа)</p>
            ) : (
              <NativeSelect
                value={form.movement_type}
                onChange={(e) =>
                  setForm({
                    ...form,
                    movement_type: e.target.value,
                    lines: (form.lines || []).map((l) => ({ ...l, destination_warehouse_id: '' })),
                  })
                }
              >
                <option value="transfer">Между складами</option>
                <option value="to_production">В производство</option>
                <option value="work_order_issue">В работу</option>
              </NativeSelect>
            )}
          </FormField>
          <FormField label="Куда">
            <p className="text-sm text-muted-foreground pt-2">
              {MOVEMENT_DEST_LABEL[form.movement_type] || '—'}
              {needsDestination && ' — выберите склад в каждой строке'}
            </p>
          </FormField>
        </div>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-base">Запчасти</CardTitle>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setForm((f) => ({ ...f, lines: [...(f.lines || []), emptyLine()] }))}
            >
              <Plus className="mr-1 h-4 w-4" /> Добавить
            </Button>
          </CardHeader>
          <CardContent className="space-y-3">
            {(form.lines || []).map((line, i) => {
              const stock = stockByLine[i]
              const overStock = stock != null && line.quantity > stock && !isFromProduction
              return (
                <div key={i} className="space-y-2 rounded-md border p-3">
                  <div className="grid gap-2 md:grid-cols-2 lg:grid-cols-3">
                    <NativeSelect
                      value={line.part_id}
                      onChange={(e) => {
                        const part = parts.find((x) => x.id === e.target.value)
                        updateLine(i, {
                          part_id: e.target.value,
                          warehouse_id: line.warehouse_id || part?.warehouse_id || '',
                        })
                      }}
                    >
                      <option value="">Запчасть</option>
                      {parts.map((part) => (
                        <option key={part.id} value={part.id}>
                          {part.sku} — {part.name}
                        </option>
                      ))}
                    </NativeSelect>
                    <NativeSelect
                      value={line.warehouse_id}
                      onChange={(e) => updateLine(i, { warehouse_id: e.target.value })}
                    >
                      <option value="">{isFromProduction ? 'Склад (куда)' : 'Склад (откуда)'}</option>
                      {warehouses.map((w) => (
                        <option key={w.id} value={w.id}>
                          {w.name}
                        </option>
                      ))}
                    </NativeSelect>
                    {needsDestination ? (
                      <NativeSelect
                        value={line.destination_warehouse_id || ''}
                        onChange={(e) => updateLine(i, { destination_warehouse_id: e.target.value })}
                      >
                        <option value="">Склад (куда)</option>
                        {warehouses
                          .filter((w) => w.id !== line.warehouse_id)
                          .map((w) => (
                            <option key={w.id} value={w.id}>
                              {w.name}
                            </option>
                          ))}
                      </NativeSelect>
                    ) : (
                      <div className="flex items-center rounded-md border bg-muted/40 px-3 text-sm text-muted-foreground">
                        {MOVEMENT_DEST_LABEL[form.movement_type]}
                      </div>
                    )}
                  </div>
                  <div className="grid gap-2 md:grid-cols-[minmax(120px,1fr)_auto_1fr_auto] md:items-center">
                    <div>
                      <Input
                        type="number"
                        min={1}
                        placeholder="Кол-во"
                        value={line.quantity}
                        onChange={(e) => updateLine(i, { quantity: Number(e.target.value) || 0 })}
                        className={cn(overStock && 'border-destructive')}
                      />
                      {line.part_id && line.warehouse_id && !isFromProduction && (
                        <p className={cn('mt-1 text-xs', overStock ? 'text-destructive' : 'text-muted-foreground')}>
                          Остаток на складе: {stock == null ? '…' : stock}
                        </p>
                      )}
                    </div>
                    <Input
                      placeholder="Примечание"
                      value={line.notes || ''}
                      onChange={(e) => updateLine(i, { notes: e.target.value })}
                      className="md:col-span-2"
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      onClick={() =>
                        setForm((f) => ({
                          ...f,
                          lines: (f.lines || []).filter((_, idx) => idx !== i),
                        }))
                      }
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              )
            })}
          </CardContent>
        </Card>

        <FormField label="Примечание к документу">
          <Textarea value={form.notes || ''} onChange={(e) => setForm({ ...form, notes: e.target.value })} />
        </FormField>

        <FormActions
          submitting={submitting}
          submitLabel={isNew ? 'Создать' : 'Сохранить'}
          onCancel={() => navigate('/movement-documents')}
        />
      </form>
    </FormPage>
  )
}
