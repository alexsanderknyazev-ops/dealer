import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Plus, Trash2 } from 'lucide-react'
import type { MovementDocumentForm as FormType, MovementDocumentLineInput } from './movementDocumentsApi'
import * as api from './movementDocumentsApi'
import * as partsApi from './partsApi'
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

const emptyLine = (): MovementDocumentLineInput => ({
  part_id: '',
  warehouse_id: '',
  quantity: 1,
})

export function MovementDocumentForm() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { logout, user } = useAuth()
  const isNew = id === 'new' || !id
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [parts, setParts] = useState<partsApi.Part[]>([])
  const [warehouses, setWarehouses] = useState<dealerPointsApi.Warehouse[]>([])
  const [form, setForm] = useState<FormType>({
    movement_type: 'manual_transfer',
    lines: [emptyLine()],
  })

  useEffect(() => {
    partsApi.listParts({ limit: 500 }).then((r) => setParts(r.parts)).catch(() => {})
    dealerPointsApi.listWarehouses({ limit: 500 }).then((r) => setWarehouses(r.warehouses)).catch(() => {})
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
                quantity: l.quantity,
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

  function updateLine(index: number, patch: Partial<MovementDocumentLineInput>) {
    setForm((f) => {
      const lines = [...(f.lines || [])]
      lines[index] = { ...lines[index], ...patch }
      return { ...f, lines }
    })
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    const payload: FormType = {
      ...form,
      lines: form.lines.filter((l) => l.part_id && l.warehouse_id && l.quantity > 0),
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
            <NativeSelect
              value={form.movement_type}
              onChange={(e) => setForm({ ...form, movement_type: e.target.value })}
            >
              <option value="manual_transfer">Ручное перемещение</option>
              <option value="work_order_issue">Выдача в работу</option>
            </NativeSelect>
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
            {(form.lines || []).map((line, i) => (
              <div key={i} className="grid gap-2 rounded-md border p-3 md:grid-cols-5">
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
                  <option value="">Склад</option>
                  {warehouses.map((w) => (
                    <option key={w.id} value={w.id}>
                      {w.name}
                    </option>
                  ))}
                </NativeSelect>
                <Input
                  type="number"
                  min={1}
                  placeholder="Кол-во"
                  value={line.quantity}
                  onChange={(e) => updateLine(i, { quantity: Number(e.target.value) || 0 })}
                />
                <Input
                  placeholder="Примечание"
                  value={line.notes || ''}
                  onChange={(e) => updateLine(i, { notes: e.target.value })}
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
            ))}
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
