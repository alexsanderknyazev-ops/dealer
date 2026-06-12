import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import type { EmployeeForm as FormType } from './employeesApi'
import * as api from './employeesApi'
import { FormPage } from '@/components/common/FormPage'
import { FormField } from '@/components/common/FormField'
import { FormActions } from '@/components/common/FormActions'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'
import { NativeSelect } from '@/components/ui/native-select'

export function EmployeeForm() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isNew = id === 'new' || !id
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [form, setForm] = useState<FormType>({
    full_name: '',
    position: 'master',
    department: '',
    phone: '',
    user_id: '',
    active: true,
  })

  useEffect(() => {
    if (isNew) return
    api
      .getEmployee(id!)
      .then((e) => {
        setForm({
          full_name: e.full_name,
          position: e.position || 'master',
          department: e.department || '',
          phone: e.phone || '',
          user_id: e.user_id || '',
          active: e.active,
        })
      })
      .catch((e) => setError(e instanceof Error ? e.message : 'Ошибка загрузки'))
      .finally(() => setLoading(false))
  }, [id, isNew])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    const payload: FormType = {
      full_name: form.full_name.trim(),
      position: form.position,
      department: form.department?.trim() || undefined,
      phone: form.phone?.trim() || undefined,
      user_id: form.user_id?.trim() || undefined,
      active: form.active ?? true,
    }
    const save = isNew ? api.createEmployee(payload) : api.updateEmployee(id!, payload)
    save
      .then((saved) => navigate(`/employees/${saved.id}`, { replace: true }))
      .catch((err) => {
        setError(err instanceof Error ? err.message : 'Ошибка сохранения')
        setSubmitting(false)
      })
  }

  return (
    <FormPage title={isNew ? 'Новый сотрудник' : 'Редактирование сотрудника'} loading={loading}>
      <form onSubmit={handleSubmit} className="mx-auto max-w-lg space-y-4">
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        <FormField label="ФИО" htmlFor="full_name" required>
          <Input
            id="full_name"
            value={form.full_name}
            onChange={(e) => setForm((f) => ({ ...f, full_name: e.target.value }))}
            required
          />
        </FormField>

        <FormField label="Должность" htmlFor="position" required>
          <NativeSelect
            id="position"
            value={form.position}
            onChange={(e) => setForm((f) => ({ ...f, position: e.target.value }))}
            required
          >
            {Object.entries(api.POSITION_LABEL).map(([value, label]) => (
              <option key={value} value={value}>{label}</option>
            ))}
          </NativeSelect>
        </FormField>

        <FormField label="Отдел" htmlFor="department">
          <Input
            id="department"
            value={form.department || ''}
            onChange={(e) => setForm((f) => ({ ...f, department: e.target.value }))}
          />
        </FormField>

        <FormField label="Телефон" htmlFor="phone">
          <Input
            id="phone"
            value={form.phone || ''}
            onChange={(e) => setForm((f) => ({ ...f, phone: e.target.value }))}
          />
        </FormField>

        <FormField label="User ID (auth.users)" htmlFor="user_id">
          <Input
            id="user_id"
            className="font-mono text-sm"
            placeholder="UUID учётной записи"
            value={form.user_id || ''}
            onChange={(e) => setForm((f) => ({ ...f, user_id: e.target.value }))}
          />
        </FormField>

        <FormField label="Статус" htmlFor="active">
          <NativeSelect
            id="active"
            value={form.active ? 'true' : 'false'}
            onChange={(e) => setForm((f) => ({ ...f, active: e.target.value === 'true' }))}
          >
            <option value="true">Активен</option>
            <option value="false">Неактивен</option>
          </NativeSelect>
        </FormField>

        <FormActions
          submitting={submitting}
          submitLabel={isNew ? 'Создать' : 'Сохранить'}
          onCancel={() => navigate(isNew ? '/employees' : `/employees/${id}`)}
        />
      </form>
    </FormPage>
  )
}
