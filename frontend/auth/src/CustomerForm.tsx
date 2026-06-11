import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import type { CustomerForm as CustomerFormType } from './customersApi'
import * as api from './customersApi'
import { FormPage } from '@/components/common/FormPage'
import { FormField } from '@/components/common/FormField'
import { FormActions } from '@/components/common/FormActions'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'
import { NativeSelect } from '@/components/ui/native-select'
import { Textarea } from '@/components/ui/textarea'

export function CustomerForm() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isNew = id === 'new' || !id
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [form, setForm] = useState<CustomerFormType>({
    name: '',
    email: '',
    phone: '',
    customer_type: 'individual',
    inn: '',
    address: '',
    notes: '',
  })

  useEffect(() => {
    if (isNew) return
    api
      .getCustomer(id!)
      .then((c) => {
        setForm({
          name: c.name,
          email: c.email || '',
          phone: c.phone || '',
          customer_type: c.customer_type || 'individual',
          inn: c.inn || '',
          address: c.address || '',
          notes: c.notes || '',
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
      name: form.name,
      email: form.email || undefined,
      phone: form.phone || undefined,
      customer_type: form.customer_type || undefined,
      inn: form.inn || undefined,
      address: form.address || undefined,
      notes: form.notes || undefined,
    }
    const save = isNew ? api.createCustomer(payload) : api.updateCustomer(id!, payload)
    save
      .then(() => navigate('/customers', { replace: true }))
      .catch((err) => {
        setError(err instanceof Error ? err.message : 'Ошибка сохранения')
        setSubmitting(false)
      })
  }

  return (
    <FormPage title={isNew ? 'Новый клиент' : 'Редактирование клиента'} loading={loading}>
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        <FormField label="Имя" htmlFor="name" required>
          <Input id="name" value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} required />
        </FormField>
        <FormField label="Email" htmlFor="email">
          <Input id="email" type="email" value={form.email} onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))} />
        </FormField>
        <FormField label="Телефон" htmlFor="phone">
          <Input id="phone" type="tel" value={form.phone} onChange={(e) => setForm((f) => ({ ...f, phone: e.target.value }))} />
        </FormField>
        <FormField label="Тип" htmlFor="customer_type">
          <NativeSelect
            id="customer_type"
            value={form.customer_type}
            onChange={(e) => setForm((f) => ({ ...f, customer_type: e.target.value }))}
          >
            <option value="individual">Физ. лицо</option>
            <option value="legal">Юр. лицо</option>
          </NativeSelect>
        </FormField>
        <FormField label="ИНН" htmlFor="inn">
          <Input id="inn" value={form.inn} onChange={(e) => setForm((f) => ({ ...f, inn: e.target.value }))} />
        </FormField>
        <FormField label="Адрес" htmlFor="address">
          <Input id="address" value={form.address} onChange={(e) => setForm((f) => ({ ...f, address: e.target.value }))} />
        </FormField>
        <FormField label="Заметки" htmlFor="notes">
          <Textarea id="notes" value={form.notes} onChange={(e) => setForm((f) => ({ ...f, notes: e.target.value }))} rows={3} />
        </FormField>
        <FormActions
          submitting={submitting}
          submitLabel={isNew ? 'Создать' : 'Сохранить'}
          onCancel={() => navigate('/customers')}
        />
      </form>
    </FormPage>
  )
}
