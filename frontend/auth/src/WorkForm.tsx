import { useEffect, useState } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import type { WorkForm as WorkFormType } from './worksApi'
import * as api from './worksApi'
import { useAuth } from './auth'
import { FormPage } from '@/components/common/FormPage'
import { FormField } from '@/components/common/FormField'
import { FormActions } from '@/components/common/FormActions'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'
import { NativeSelect } from '@/components/ui/native-select'
import { Textarea } from '@/components/ui/textarea'

const CATEGORY_OPTIONS = [
  'Техобслуживание',
  'Диагностика',
  'Ремонт',
  'Кузовные работы',
  'Электрика',
  'Шиномонтаж',
]

export function WorkForm() {
  const { id } = useParams()
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { logout } = useAuth()
  const isNew = id === 'new' || !id
  const defaultFolderId = searchParams.get('folder_id') || ''
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [folderOptions, setFolderOptions] = useState<{ id: string; name: string; level: number }[]>([])
  const [form, setForm] = useState<WorkFormType>({
    code: '',
    name: '',
    category: '',
    folder_id: defaultFolderId || undefined,
    labor_hours: '1',
    unit_price: '0',
    notes: '',
  })

  useEffect(() => {
    api.loadAllFoldersFlat().then(setFolderOptions).catch(() => setFolderOptions([]))
  }, [])

  useEffect(() => {
    if (isNew) {
      if (defaultFolderId) setForm((f) => ({ ...f, folder_id: defaultFolderId }))
      return
    }
    api
      .getWork(id!)
      .then((w) =>
        setForm({
          code: w.code,
          name: w.name,
          category: w.category,
          folder_id: w.folder_id || undefined,
          labor_hours: w.labor_hours || '1',
          unit_price: w.unit_price || '0',
          notes: w.notes,
        }),
      )
      .catch(async (e) => {
        if (e instanceof Error && (e.message.includes('401') || e.message.includes('403'))) {
          await logout()
          navigate('/login', { replace: true })
          return
        }
        setError(e instanceof Error ? e.message : 'Ошибка загрузки')
      })
      .finally(() => setLoading(false))
  }, [id, isNew, defaultFolderId, logout, navigate])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!form.code?.trim()) {
      setError('Укажите код работы')
      return
    }
    if (!form.name?.trim()) {
      setError('Укажите название')
      return
    }
    setError(null)
    setSubmitting(true)
    const payload: WorkFormType = {
      code: form.code.trim(),
      name: form.name.trim(),
      category: form.category?.trim() || undefined,
      folder_id: form.folder_id || '',
      labor_hours: form.labor_hours?.trim() || '1',
      unit_price: form.unit_price?.trim() || '0',
      notes: form.notes?.trim() || undefined,
    }
    const returnUrl = form.folder_id ? `/works?folder_id=${form.folder_id}` : '/works'
    const save = isNew ? api.createWork(payload) : api.updateWork(id!, payload)
    save
      .then((w) => navigate(isNew ? returnUrl : `/works/${w.id}`, { replace: true }))
      .catch((err) => {
        setError(err instanceof Error ? err.message : 'Ошибка сохранения')
        setSubmitting(false)
      })
  }

  return (
    <FormPage title={isNew ? 'Новая работа' : 'Редактирование работы'} loading={loading}>
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        <FormField label="Папка" htmlFor="folder_id">
          <NativeSelect
            id="folder_id"
            value={form.folder_id ?? ''}
            onChange={(e) => setForm((f) => ({ ...f, folder_id: e.target.value || undefined }))}
          >
            <option value="">Без папки</option>
            {folderOptions.map((opt) => (
              <option key={opt.id} value={opt.id}>
                {'—'.repeat(opt.level)} {opt.name}
              </option>
            ))}
          </NativeSelect>
        </FormField>
        <FormField label="Код" htmlFor="code" required>
          <Input
            id="code"
            value={form.code}
            onChange={(e) => setForm((f) => ({ ...f, code: e.target.value }))}
            required
            placeholder="WO-001"
          />
        </FormField>
        <FormField label="Название" htmlFor="name" required>
          <Input
            id="name"
            value={form.name}
            onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
            required
            placeholder="Замена масла двигателя"
          />
        </FormField>
        <FormField label="Категория" htmlFor="category">
          <NativeSelect
            id="category"
            value={CATEGORY_OPTIONS.includes(form.category || '') ? form.category : ''}
            onChange={(e) => setForm((f) => ({ ...f, category: e.target.value }))}
          >
            <option value="">— не выбрана —</option>
            {CATEGORY_OPTIONS.map((c) => (
              <option key={c} value={c}>
                {c}
              </option>
            ))}
          </NativeSelect>
          <Input
            className="mt-2"
            value={form.category || ''}
            onChange={(e) => setForm((f) => ({ ...f, category: e.target.value }))}
            placeholder="Или введите категорию вручную"
          />
        </FormField>
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField label="Норма-час" htmlFor="labor_hours" required>
            <Input
              id="labor_hours"
              value={form.labor_hours}
              onChange={(e) => setForm((f) => ({ ...f, labor_hours: e.target.value }))}
              placeholder="1.5"
              required
            />
          </FormField>
          <FormField label="Цена за норма-час" htmlFor="unit_price">
            <Input
              id="unit_price"
              value={form.unit_price}
              onChange={(e) => setForm((f) => ({ ...f, unit_price: e.target.value }))}
              placeholder="1500"
            />
          </FormField>
        </div>
        <FormField label="Примечания" htmlFor="notes">
          <Textarea
            id="notes"
            value={form.notes || ''}
            onChange={(e) => setForm((f) => ({ ...f, notes: e.target.value }))}
            rows={3}
          />
        </FormField>
        <FormActions
          submitting={submitting}
          submitLabel={isNew ? 'Создать' : 'Сохранить'}
          onCancel={() => navigate(form.folder_id ? `/works?folder_id=${form.folder_id}` : '/works')}
        />
      </form>
    </FormPage>
  )
}
