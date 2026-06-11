import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import type { WarehouseForm as FormType } from './dealerPointsApi'
import * as api from './dealerPointsApi'
import { FormPage } from '@/components/common/FormPage'
import { FormField } from '@/components/common/FormField'
import { FormActions } from '@/components/common/FormActions'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'
import { NativeSelect } from '@/components/ui/native-select'

export function WarehouseForm() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isNew = id === 'new' || !id
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [points, setPoints] = useState<api.DealerPoint[]>([])
  const [legalEntities, setLegalEntities] = useState<api.LegalEntity[]>([])
  const [form, setForm] = useState<FormType>({
    dealer_point_id: '',
    legal_entity_id: '',
    type: 'parts',
    name: '',
  })

  useEffect(() => {
    api.listDealerPoints({ limit: 200 }).then((r) => setPoints(r.dealer_points)).catch(() => setPoints([]))
  }, [])

  useEffect(() => {
    if (!form.dealer_point_id) {
      setLegalEntities([])
      return
    }
    api.listLegalEntitiesByDealerPoint(form.dealer_point_id).then(setLegalEntities).catch(() => setLegalEntities([]))
  }, [form.dealer_point_id])

  useEffect(() => {
    if (isNew) return
    api
      .getWarehouse(id!)
      .then((w) =>
        setForm({
          dealer_point_id: w.dealer_point_id,
          legal_entity_id: w.legal_entity_id,
          type: w.type,
          name: w.name,
        }),
      )
      .catch((e) => setError(e instanceof Error ? e.message : 'Ошибка загрузки'))
      .finally(() => setLoading(false))
  }, [id, isNew])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!form.dealer_point_id || !form.legal_entity_id || !form.name?.trim()) {
      setError('Укажите дилерскую точку, юр. лицо и название склада')
      return
    }
    setError(null)
    setSubmitting(true)
    const payload = {
      dealer_point_id: form.dealer_point_id,
      legal_entity_id: form.legal_entity_id,
      type: form.type,
      name: form.name.trim(),
    }
    const save = isNew ? api.createWarehouse(payload) : api.updateWarehouse(id!, { name: form.name.trim() })
    save
      .then(() => navigate('/warehouses', { replace: true }))
      .catch((err) => {
        setError(err instanceof Error ? err.message : 'Ошибка сохранения')
        setSubmitting(false)
      })
  }

  return (
    <FormPage title={isNew ? 'Новый склад' : 'Редактирование склада'} loading={loading}>
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        <FormField label="Дилерская точка" htmlFor="dealer_point_id" required>
          <NativeSelect
            id="dealer_point_id"
            value={form.dealer_point_id}
            disabled={!isNew}
            required
            onChange={(e) => setForm((f) => ({ ...f, dealer_point_id: e.target.value, legal_entity_id: '' }))}
          >
            <option value="">— выберите —</option>
            {points.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </NativeSelect>
        </FormField>
        <FormField label="Юридическое лицо" htmlFor="legal_entity_id" required>
          <NativeSelect
            id="legal_entity_id"
            value={form.legal_entity_id}
            disabled={!isNew}
            required
            onChange={(e) => setForm((f) => ({ ...f, legal_entity_id: e.target.value }))}
          >
            <option value="">— выберите —</option>
            {legalEntities.map((e) => (
              <option key={e.id} value={e.id}>
                {e.name} {e.inn ? `(ИНН ${e.inn})` : ''}
              </option>
            ))}
          </NativeSelect>
        </FormField>
        <FormField label="Тип склада" htmlFor="type" required>
          <NativeSelect
            id="type"
            value={form.type}
            disabled={!isNew}
            onChange={(e) => setForm((f) => ({ ...f, type: e.target.value as 'cars' | 'parts' }))}
          >
            <option value="cars">Автомобили</option>
            <option value="parts">Запчасти</option>
          </NativeSelect>
        </FormField>
        <FormField label="Название склада" htmlFor="name" required>
          <Input
            id="name"
            value={form.name}
            onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
            required
            placeholder="Например: Склад автомобилей — Москва"
          />
        </FormField>
        <FormActions submitting={submitting} submitLabel={isNew ? 'Создать' : 'Сохранить'} onCancel={() => navigate('/warehouses')} />
      </form>
    </FormPage>
  )
}
