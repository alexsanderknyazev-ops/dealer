import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import type { Work } from './worksApi'
import * as api from './worksApi'
import { useAuth } from './auth'
import { EntityViewPage } from '@/components/common/EntityViewPage'

export function WorkView() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { logout } = useAuth()
  const [work, setWork] = useState<Work | null>(null)
  const [folderName, setFolderName] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return
    api
      .getWork(id)
      .then((w) => {
        setWork(w)
        if (w.folder_id) {
          api.getFolder(w.folder_id).then((f) => setFolderName(f.name)).catch(() => {})
        }
      })
      .catch(async (e) => {
        if (e instanceof Error && (e.message.includes('401') || e.message.includes('403'))) {
          await logout()
          navigate('/login', { replace: true })
          return
        }
        setError(e instanceof Error ? e.message : 'Ошибка загрузки')
      })
      .finally(() => setLoading(false))
  }, [id, logout, navigate])

  async function handleDelete() {
    if (!id || !work || !confirm(`Удалить работу ${work.code} — ${work.name}?`)) return
    try {
      await api.deleteWork(id)
      navigate('/works', { replace: true })
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка удаления')
    }
  }

  const fields = work
    ? [
        { label: 'Код', value: <span className="font-mono">{work.code || '—'}</span> },
        { label: 'Название', value: work.name || '—' },
        { label: 'Категория', value: work.category || '—' },
        ...(folderName != null ? [{ label: 'Папка', value: folderName }] : []),
        { label: 'Норма-час', value: work.labor_hours || '—' },
        {
          label: 'Цена за норма-час',
          value: work.unit_price ? Number(work.unit_price).toLocaleString('ru') : '—',
        },
        ...(work.notes ? [{ label: 'Примечания', value: work.notes }] : []),
      ]
    : []

  return (
    <EntityViewPage
      title={work ? `${work.code} — ${work.name}` : 'Работа'}
      backTo="/works"
      backLabel="Все работы"
      editTo={id ? `/works/${id}/edit` : undefined}
      onDelete={work ? handleDelete : undefined}
      fields={fields}
      loading={loading}
      error={error}
    />
  )
}
