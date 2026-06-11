import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import type { BrandForm as BrandFormType } from './brandsApi'
import * as api from './brandsApi'
import { FormPage } from '@/components/common/FormPage'
import { FormField } from '@/components/common/FormField'
import { FormActions } from '@/components/common/FormActions'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'

export function BrandForm() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isNew = id === 'new' || !id
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [form, setForm] = useState<BrandFormType>({ name: '' })

  useEffect(() => {
    if (isNew) return
    api
      .getBrand(id!)
      .then((b) => setForm({ name: b.name }))
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
    const payload = { name: form.name.trim() }
    const save = isNew ? api.createBrand(payload) : api.updateBrand(id!, payload)
    save
      .then(() => navigate('/brands', { replace: true }))
      .catch((err) => {
        setError(err instanceof Error ? err.message : 'Ошибка сохранения')
        setSubmitting(false)
      })
  }

  return (
    <FormPage title={isNew ? 'Новый бренд' : 'Редактирование бренда'} loading={loading}>
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        <FormField label="Название" htmlFor="name" required>
          <Input id="name" value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} required placeholder="Например: BMW" />
        </FormField>
        <FormActions submitting={submitting} submitLabel={isNew ? 'Создать' : 'Сохранить'} onCancel={() => navigate('/brands')} />
      </form>
    </FormPage>
  )
}
