import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import type { LegalEntityForm as FormType } from './dealerPointsApi'
import * as api from './dealerPointsApi'
import { FormPage } from '@/components/common/FormPage'
import { FormField } from '@/components/common/FormField'
import { FormActions } from '@/components/common/FormActions'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'

export function LegalEntityForm() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isNew = id === 'new' || !id
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [form, setForm] = useState<FormType>({ name: '', inn: '', address: '' })

  useEffect(() => {
    if (isNew) return
    api
      .getLegalEntity(id!)
      .then((e) => setForm({ name: e.name, inn: e.inn || '', address: e.address || '' }))
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
    const payload = {
      name: form.name.trim(),
      inn: form.inn?.trim() || undefined,
      address: form.address?.trim() || undefined,
    }
    const save = isNew ? api.createLegalEntity(payload) : api.updateLegalEntity(id!, payload)
    save
      .then(() => navigate('/legal-entities', { replace: true }))
      .catch((err) => {
        setError(err instanceof Error ? err.message : 'Ошибка сохранения')
        setSubmitting(false)
      })
  }

  return (
    <FormPage title={isNew ? 'Новое юридическое лицо' : 'Редактирование юр. лица'} loading={loading}>
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        <FormField label="Название" htmlFor="name" required>
          <Input id="name" value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} required placeholder="ООО «Компания»" />
        </FormField>
        <FormField label="ИНН" htmlFor="inn">
          <Input id="inn" value={form.inn ?? ''} onChange={(e) => setForm((f) => ({ ...f, inn: e.target.value }))} placeholder="7707123456" />
        </FormField>
        <FormField label="Адрес" htmlFor="address">
          <Input id="address" value={form.address ?? ''} onChange={(e) => setForm((f) => ({ ...f, address: e.target.value }))} />
        </FormField>
        <FormActions submitting={submitting} submitLabel={isNew ? 'Создать' : 'Сохранить'} onCancel={() => navigate('/legal-entities')} />
      </form>
    </FormPage>
  )
}
