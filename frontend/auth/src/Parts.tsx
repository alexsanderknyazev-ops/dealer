import { useCallback, useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import type { Part, PartFolder } from './partsApi'
import * as api from './partsApi'
import './Parts.css'

type BreadcrumbItem = { id: string | null; name: string }

export function Parts() {
  const [searchParams, setSearchParams] = useSearchParams()
  const initialFolderId = searchParams.get('folder_id') || null
  const [folderStack, setFolderStack] = useState<BreadcrumbItem[]>(() => [{ id: null, name: 'Корень' }])
  const currentFolderId = folderStack[folderStack.length - 1].id

  const [foldersByParent, setFoldersByParent] = useState<Record<string, PartFolder[]>>({ root: [] })
  const [expandedIds, setExpandedIds] = useState<Set<string>>(() => new Set())
  const [list, setList] = useState<Part[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [categoryFilter, setCategoryFilter] = useState('')
  const [page, setPage] = useState(0)
  const [newFolderName, setNewFolderName] = useState('')
  const [creatingFolder, setCreatingFolder] = useState(false)
  const limit = 20

  const loadRootFolders = useCallback(() => {
    api.listFolders(undefined)
      .then((r) => setFoldersByParent((prev) => ({ ...prev, root: r.folders })))
      .catch(() => setFoldersByParent((prev) => ({ ...prev, root: [] })))
  }, [])

  const loadChildFolders = useCallback((parentId: string) => {
    api.listFolders(parentId)
      .then((r) => setFoldersByParent((prev) => ({ ...prev, [parentId]: r.folders })))
      .catch(() => setFoldersByParent((prev) => ({ ...prev, [parentId]: [] })))
  }, [])

  const loadParts = useCallback(() => {
    setLoading(true)
    setError(null)
    api.listParts({
      limit,
      offset: page * limit,
      search: search || undefined,
      category: categoryFilter || undefined,
      folder_id: currentFolderId ?? undefined,
    })
      .then((r) => {
        setList(r.parts)
        setTotal(r.total)
      })
      .catch((err) => {
        setList([])
        setError(err instanceof Error ? err.message : 'Ошибка загрузки')
      })
      .finally(() => setLoading(false))
  }, [currentFolderId, page, search, categoryFilter])

  useEffect(() => {
    loadRootFolders()
  }, [loadRootFolders])

  useEffect(() => {
    loadParts()
  }, [loadParts])

  useEffect(() => {
    if (initialFolderId && folderStack.length === 1 && folderStack[0].id === null) {
      api.getFolder(initialFolderId)
        .then((f) => setFolderStack([{ id: null, name: 'Корень' }, { id: f.id, name: f.name }]))
        .catch(() => {})
    }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  function goToFolder(folder: PartFolder) {
    setFolderStack((prev) => [...prev, { id: folder.id, name: folder.name }])
    setSearchParams((p) => {
      p.set('folder_id', folder.id)
      return p
    })
    setPage(0)
  }

  function goToBreadcrumb(index: number) {
    const item = folderStack[index]
    setFolderStack((prev) => prev.slice(0, index + 1))
    setSearchParams((p) => {
      if (item?.id) p.set('folder_id', item.id)
      else p.delete('folder_id')
      return p
    })
    setPage(0)
  }

  function toggleExpand(folderId: string) {
    setExpandedIds((prev) => {
      const next = new Set(prev)
      if (next.has(folderId)) {
        next.delete(folderId)
      } else {
        next.add(folderId)
        if (!(folderId in foldersByParent)) loadChildFolders(folderId)
      }
      return next
    })
  }

  function handleCreateFolder(e?: React.FormEvent) {
    e?.preventDefault()
    const name = newFolderName.trim()
    if (!name) return
    setCreatingFolder(true)
    setError(null)
    api.createFolder({ name, parent_id: currentFolderId ?? undefined })
      .then(() => {
        setNewFolderName('')
        loadRootFolders()
        const parentKey = currentFolderId ?? 'root'
        api.listFolders(currentFolderId ?? undefined).then((r) =>
          setFoldersByParent((prev) => ({ ...prev, [parentKey]: r.folders }))
        )
      })
      .catch((err) => setError(err instanceof Error ? err.message : 'Ошибка создания папки'))
      .finally(() => setCreatingFolder(false))
  }

  function handleDeleteFolder(e: React.MouseEvent, folderId: string, folderName: string, parentKey: string) {
    e.preventDefault()
    e.stopPropagation()
    if (!confirm(`Удалить папку «${folderName}»? Запчасти в ней останутся без папки.`)) return
    api.deleteFolder(folderId)
      .then(() => {
        setExpandedIds((prev) => {
          const next = new Set(prev)
          next.delete(folderId)
          return next
        })
        setFoldersByParent((prev) => {
          const next = { ...prev }
          delete next[folderId]
          return next
        })
        const refreshParent = () => api.listFolders(parentKey === 'root' ? undefined : parentKey).then((r) =>
          setFoldersByParent((prev) => ({ ...prev, [parentKey]: r.folders }))
        ).catch(() => {})
        refreshParent()
        if (parentKey === 'root') loadRootFolders()
        if (currentFolderId === folderId) goToBreadcrumb(0)
      })
      .catch((err) => setError(err instanceof Error ? err.message : 'Ошибка удаления папки'))
  }

  const addPartLink = currentFolderId ? `/parts/new?folder_id=${currentFolderId}` : '/parts/new'

  function renderFolderTree(parentKey: string, level: number) {
    const folders = foldersByParent[parentKey] ?? []
    return folders.map((f) => {
      const isExpanded = expandedIds.has(f.id)
      const childrenLoaded = f.id in foldersByParent
      const isSelected = currentFolderId === f.id
      return (
        <div key={f.id} className="parts-tree-folder" style={{ paddingLeft: level * 12 }}>
          <span className="parts-tree-row">
            <button
              type="button"
              className="parts-tree-expand"
              onClick={() => toggleExpand(f.id)}
              aria-label={isExpanded ? 'Свернуть' : 'Развернуть'}
            >
              {isExpanded ? '▼' : '▶'}
            </button>
            <button
              type="button"
              className={`parts-folder-item ${isSelected ? 'parts-folder-item--selected' : ''}`}
              onClick={() => goToFolder(f)}
            >
              📁 {f.name}
            </button>
            <button
              type="button"
              className="parts-folder-delete"
              onClick={(e) => handleDeleteFolder(e, f.id, f.name, parentKey)}
              title="Удалить папку"
              aria-label="Удалить папку"
            >
              ×
            </button>
          </span>
          {isExpanded && (childrenLoaded ? renderFolderTree(f.id, level + 1) : <div className="parts-tree-loading">…</div>)}
        </div>
      )
    })
  }

  return (
    <div className="parts">
      <div className="parts-header">
        <h1 className="parts-title">Запасные части</h1>
        <Link to={addPartLink} className="parts-add">+ Добавить запчасть</Link>
      </div>

      <div className="parts-layout">
        <aside className="parts-sidebar">
          <div className="parts-breadcrumb">
            {folderStack.map((item, i) => (
              <span key={item.id ?? 'root'}>
                {i > 0 && <span className="parts-breadcrumb-sep"> / </span>}
                <button
                  type="button"
                  className="parts-breadcrumb-link"
                  onClick={() => goToBreadcrumb(i)}
                >
                  {item.name}
                </button>
              </span>
            ))}
          </div>
          <div className="parts-tree">
            <button
              type="button"
              className={`parts-folder-item ${currentFolderId === null ? 'parts-folder-item--selected' : ''}`}
              onClick={() => goToBreadcrumb(0)}
            >
              📁 Корень
            </button>
            {renderFolderTree('root', 0)}
          </div>
          <form onSubmit={(e) => { e.preventDefault(); handleCreateFolder(e); }} className="parts-new-folder">
            <input
              type="text"
              placeholder="Новая папка..."
              value={newFolderName}
              onChange={(e) => setNewFolderName(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleCreateFolder()}
              className="parts-new-folder-input"
              disabled={creatingFolder}
            />
            <button type="button" className="parts-new-folder-btn" disabled={creatingFolder || !newFolderName.trim()} onClick={() => handleCreateFolder()}>
              + Папка
            </button>
          </form>
        </aside>

        <div className="parts-main">
          <div className="parts-toolbar">
            <input
              type="search"
              placeholder="Поиск по артикулу, названию..."
              value={search}
              onChange={(e) => { setSearch(e.target.value); setPage(0) }}
              className="parts-search"
            />
            <select
              value={categoryFilter}
              onChange={(e) => { setCategoryFilter(e.target.value); setPage(0) }}
              className="parts-category-filter"
            >
              <option value="">Все категории</option>
              <option value="Фильтры">Фильтры</option>
              <option value="Тормоза">Тормоза</option>
              <option value="Масла">Масла</option>
              <option value="Расходники">Расходники</option>
            </select>
          </div>
          {error && <div className="parts-error">{error}</div>}
          {loading && <div className="parts-loading">Загрузка…</div>}
          {!loading && !error && list.length === 0 && (
            <div className="parts-empty">
              В этой папке нет запчастей. Добавьте запчасть или выберите другую папку.
            </div>
          )}
          {!loading && list.length > 0 && (
            <>
              <table className="parts-table">
                <thead>
                  <tr>
                    <th>Артикул</th>
                    <th>Название</th>
                    <th>Категория</th>
                    <th>Кол-во</th>
                    <th>Ед.</th>
                    <th>Цена</th>
                    <th>Расположение</th>
                  </tr>
                </thead>
                <tbody>
                  {list.map((p) => (
                    <tr key={p.id}>
                      <td className="parts-sku">{p.sku}</td>
                      <td>
                        <Link to={`/parts/${p.id}`} className="parts-link">{p.name || '—'}</Link>
                      </td>
                      <td>{p.category || '—'}</td>
                      <td>{p.quantity}</td>
                      <td>{p.unit || 'шт'}</td>
                      <td>{p.price ? Number(p.price).toLocaleString('ru') : '—'}</td>
                      <td>{p.location || '—'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
              <div className="parts-pagination">
                <span>Всего: {total}</span>
                <button type="button" disabled={page === 0} onClick={() => setPage((p) => p - 1)}>Назад</button>
                <span>Стр. {page + 1}</span>
                <button type="button" disabled={(page + 1) * limit >= total} onClick={() => setPage((p) => p + 1)}>Вперёд</button>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
