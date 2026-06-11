import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import type { Part } from './partsApi'
import * as api from './partsApi'
import * as dealerPointsApi from './dealerPointsApi'
import { EntityViewPage } from '@/components/common/EntityViewPage'

export function PartView() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [part, setPart] = useState<Part | null>(null)
  const [folderName, setFolderName] = useState<string | null>(null)
  const [pointName, setPointName] = useState<string | null>(null)
  const [legalEntityName, setLegalEntityName] = useState<string | null>(null)
  const [warehouseName, setWarehouseName] = useState<string | null>(null)
  const [warehouseNames, setWarehouseNames] = useState<Record<string, string>>({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return
    api
      .getPart(id)
      .then((p) => {
        setPart(p)
        if (p.folder_id) {
          api.getFolder(p.folder_id).then((f) => setFolderName(f.name)).catch(() => {})
        }
        if (p.dealer_point_id) {
          dealerPointsApi.getDealerPoint(p.dealer_point_id).then((d) => setPointName(d.name)).catch(() => {})
        }
        if (p.legal_entity_id) {
          dealerPointsApi.getLegalEntity(p.legal_entity_id).then((e) => setLegalEntityName(e.name)).catch(() => {})
        }
        if (p.warehouse_id) {
          dealerPointsApi.getWarehouse(p.warehouse_id).then((w) => setWarehouseName(w.name)).catch(() => {})
        }
        if (p.stock && p.stock.length > 0) {
          const ids = [...new Set(p.stock.map((s) => s.warehouse_id))]
          Promise.all(ids.map((wid) => dealerPointsApi.getWarehouse(wid).then((w) => ({ id: wid, name: w.name }))))
            .then((pairs) => setWarehouseNames(Object.fromEntries(pairs.map(({ id: i, name }) => [i, name]))))
            .catch(() => {})
        }
      })
      .catch((e) => setError(e instanceof Error ? e.message : 'Ошибка загрузки'))
      .finally(() => setLoading(false))
  }, [id])

  const hasLocation = part?.dealer_point_id || part?.legal_entity_id || part?.warehouse_id
  const hasStock = part?.stock && part.stock.length > 0

  async function handleDelete() {
    if (!id || !part || !confirm(`Удалить запчасть ${part.sku} — ${part.name || 'без названия'}?`)) return
    try {
      await api.deletePart(id)
      navigate('/parts', { replace: true })
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка удаления')
    }
  }

  const fields = part
    ? [
        { label: 'Артикул (SKU)', value: <span className="font-mono">{part.sku}</span> },
        { label: 'Название', value: part.name || '—' },
        { label: 'Категория', value: part.category || '—' },
        ...(folderName != null ? [{ label: 'Папка', value: folderName }] : []),
        {
          label: 'Количество',
          value: `${part.quantity} ${part.unit || 'шт'}${hasStock ? ' (всего по складам)' : ''}`,
        },
        { label: 'Цена', value: part.price ? Number(part.price).toLocaleString('ru') : '—' },
        ...(part.location ? [{ label: 'Расположение', value: part.location }] : []),
        ...(part.notes ? [{ label: 'Заметки', value: part.notes }] : []),
        ...(hasStock
          ? [
              {
                label: 'Наличие по складам',
                value: (
                  <div className="space-y-2">
                    <ul className="list-disc pl-5">
                      {part.stock!.map((s) => (
                        <li key={s.warehouse_id}>
                          <strong>{warehouseNames[s.warehouse_id] ?? s.warehouse_id}</strong>: {s.quantity}{' '}
                          {part.unit || 'шт'}
                        </li>
                      ))}
                    </ul>
                    <p className="text-sm text-muted-foreground">
                      Запчасть может быть на нескольких складах. Изменить остатки —{' '}
                      <Link to={`/parts/${id}/edit`} className="text-primary hover:underline">
                        Редактировать
                      </Link>
                      .
                    </p>
                  </div>
                ),
              },
            ]
          : []),
        ...(hasLocation && !hasStock
          ? [
              {
                label: 'Расположение (склад)',
                value: (
                  <div className="space-y-1">
                    <div>
                      <strong>Дилерская точка:</strong> {pointName ?? part.dealer_point_id ?? '—'}
                    </div>
                    <div>
                      <strong>Юр. лицо:</strong> {legalEntityName ?? part.legal_entity_id ?? '—'}
                    </div>
                    <div>
                      <strong>Склад запчастей:</strong>{' '}
                      {warehouseName ?? (part.warehouse_id ? `склад ${part.warehouse_id.slice(0, 8)}…` : '—')}
                    </div>
                    <p className="text-sm text-muted-foreground">
                      Чтобы указать несколько складов или изменить остатки —{' '}
                      <Link to={`/parts/${id}/edit`} className="text-primary hover:underline">
                        Редактировать
                      </Link>
                      .
                    </p>
                  </div>
                ),
              },
            ]
          : []),
      ]
    : []

  return (
    <EntityViewPage
      title={part?.name || part?.sku || 'Запчасть'}
      backTo="/parts"
      backLabel="К списку запчастей"
      editTo={id ? `/parts/${id}/edit` : undefined}
      onDelete={part ? handleDelete : undefined}
      loading={loading}
      error={error || (!part && !loading ? 'Не найдено' : null)}
      fields={fields}
    />
  )
}
