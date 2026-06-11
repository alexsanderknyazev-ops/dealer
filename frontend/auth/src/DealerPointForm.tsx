import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import type { DealerPointForm as FormType } from './dealerPointsApi'
import * as api from './dealerPointsApi'
import { FormPage } from '@/components/common/FormPage'
import { FormField } from '@/components/common/FormField'
import { FormActions } from '@/components/common/FormActions'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'

export function DealerPointForm() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isNew = id === 'new' || !id
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [form, setForm] = useState<FormType>({ name: '', address: '' })

  useEffect(() => {
    if (isNew) return
    api
      .getDealerPoint(id!)
      .then((d) => setForm({ name: d.name, address: d.address || '' }))
      .catch((e) => setError(e instanceof Error ? e.message : 'Ошибка загрузки'))
      .finally(() => setLoading(false))
  }, [id, isNew])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!form.name?.trim()) {
      setError('Укажите название')
      return
    }
    setError(null)
    setSubmitting(true)
    const payload = { name: form.name.trim(), address: form.address || undefined }
    const save = isNew ? api.createDealerPoint(payload) : api.updateDealerPoint(id!, payload)
    save
      .then(() => navigate('/dealer-points', { replace: true }))
      .catch((err) => {
        setError(err instanceof Error ? err.message : 'Ошибка сохранения')
        setSubmitting(false)
      })
  }

  return (
    <FormPage title={isNew ? 'Новая дилерская точка' : 'Редактирование дилерской точки'} loading={loading}>
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        <FormField label="Название" htmlFor="name" required>
          <Input id="name" value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} required placeholder="Например: ДЦ Москва Ленинградское ш." />
        </FormField>
        <FormField label="Адрес" htmlFor="address">
          <Input id="address" value={form.address ?? ''} onChange={(e) => setForm((f) => ({ ...f, address: e.target.value }))} placeholder="г. Москва, ул. Примерная, 1" />
        </FormField>
        <FormActions submitting={submitting} submitLabel={isNew ? 'Создать' : 'Сохранить'} onCancel={() => navigate('/dealer-points')} />
      </form>
    </FormPage>
  )
}
