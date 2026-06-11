import { useCallback, useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { ChevronDown, ChevronRight, Folder, Plus, X } from 'lucide-react'
import type { Part, PartFolder } from './partsApi'
import * as api from './partsApi'
import { PageHeader } from '@/components/common/PageHeader'
import { ErrorAlert } from '@/components/common/ErrorAlert'
import { LoadingState } from '@/components/common/LoadingState'
import { EmptyState } from '@/components/common/EmptyState'
import { Pagination } from '@/components/common/Pagination'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { NativeSelect } from '@/components/ui/native-select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { cn } from '@/lib/utils'

type BreadcrumbItem = { id: string | null; name: string }

export function Parts() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [folderStack, setFolderStack] = useState<BreadcrumbItem[]>(() => [{ id: null, name: 'Корень' }])
  const currentFolderId = folderStack[folderStack.length - 1].id
  const folderIdFromQuery = searchParams.get('folder_id')

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
    api
      .listFolders(undefined)
      .then((r) => setFoldersByParent((prev) => ({ ...prev, root: r.folders })))
      .catch(() => setFoldersByParent((prev) => ({ ...prev, root: [] })))
  }, [])

  const loadChildFolders = useCallback((parentId: string) => {
    api
      .listFolders(parentId)
      .then((r) => setFoldersByParent((prev) => ({ ...prev, [parentId]: r.folders })))
      .catch(() => setFoldersByParent((prev) => ({ ...prev, [parentId]: [] })))
  }, [])

  const loadParts = useCallback(() => {
    setLoading(true)
    setError(null)
    api
      .listParts({
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
    if (!folderIdFromQuery && currentFolderId !== null) {
      setFolderStack([{ id: null, name: 'Корень' }])
      return
    }
    if (!folderIdFromQuery || folderIdFromQuery === currentFolderId) return
    api
      .getFolder(folderIdFromQuery)
      .then((f) => setFolderStack([{ id: null, name: 'Корень' }, { id: f.id, name: f.name }]))
      .catch(() => {})
  }, [folderIdFromQuery, currentFolderId])

  function goToFolder(folder: PartFolder) {
    setFolderStack((prev) => {
      const existingIndex = prev.findIndex((item) => item.id === folder.id)
      if (existingIndex >= 0) return prev.slice(0, existingIndex + 1)
      return [...prev, { id: folder.id, name: folder.name }]
    })
    setSearchParams((p) => {
      const next = new URLSearchParams(p)
      next.set('folder_id', folder.id)
      return next
    })
    setPage(0)
  }

  function goToBreadcrumb(index: number) {
    const item = folderStack[index]
    setFolderStack((prev) => prev.slice(0, index + 1))
    setSearchParams((p) => {
      const next = new URLSearchParams(p)
      if (item?.id) next.set('folder_id', item.id)
      else next.delete('folder_id')
      return next
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
    api
      .createFolder({ name, parent_id: currentFolderId ?? undefined })
      .then(() => {
        setNewFolderName('')
        loadRootFolders()
        const parentKey = currentFolderId ?? 'root'
        api.listFolders(currentFolderId ?? undefined).then((r) =>
          setFoldersByParent((prev) => ({ ...prev, [parentKey]: r.folders })),
        )
      })
      .catch((err) => setError(err instanceof Error ? err.message : 'Ошибка создания папки'))
      .finally(() => setCreatingFolder(false))
  }

  function handleDeleteFolder(e: React.MouseEvent, folderId: string, folderName: string, parentKey: string) {
    e.preventDefault()
    e.stopPropagation()
    if (!confirm(`Удалить папку «${folderName}»? Запчасти в ней останутся без папки.`)) return
    api
      .deleteFolder(folderId)
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
        api
          .listFolders(parentKey === 'root' ? undefined : parentKey)
          .then((r) => setFoldersByParent((prev) => ({ ...prev, [parentKey]: r.folders })))
          .catch(() => {})
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
        <div key={f.id} style={{ paddingLeft: level * 12 }}>
          <div className="flex items-center gap-1">
            <Button type="button" variant="ghost" size="icon" className="h-6 w-6 shrink-0" onClick={() => toggleExpand(f.id)} aria-label={isExpanded ? 'Свернуть' : 'Развернуть'}>
              {isExpanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
            </Button>
            <button
              type="button"
              className={cn(
                'flex min-w-0 flex-1 items-center gap-1 rounded-md px-2 py-1 text-left text-sm',
                isSelected ? 'bg-accent font-medium text-accent-foreground' : 'hover:bg-muted',
              )}
              onClick={() => goToFolder(f)}
            >
              <Folder className="h-3.5 w-3.5 shrink-0" />
              <span className="truncate">{f.name}</span>
            </button>
            <Button type="button" variant="ghost" size="icon" className="h-6 w-6 shrink-0 text-muted-foreground hover:text-destructive" onClick={(e) => handleDeleteFolder(e, f.id, f.name, parentKey)} aria-label="Удалить папку">
              <X className="h-3 w-3" />
            </Button>
          </div>
          {isExpanded && (childrenLoaded ? renderFolderTree(f.id, level + 1) : <p className="py-1 pl-8 text-xs text-muted-foreground">…</p>)}
        </div>
      )
    })
  }

  return (
    <div className="mx-auto w-full max-w-6xl">
      <PageHeader
        title="Запасные части"
        action={
          <Button asChild>
            <Link to={addPartLink}>
              <Plus className="mr-2 h-4 w-4" />
              Добавить запчасть
            </Link>
          </Button>
        }
      />

      <div className="flex flex-col gap-4 lg:flex-row">
        <Card className="w-full shrink-0 lg:w-64">
          <CardContent className="space-y-3 p-4">
            <div className="flex flex-wrap gap-1 text-sm">
              {folderStack.map((item, i) => (
                <span key={item.id ?? 'root'} className="flex items-center">
                  {i > 0 && <span className="mx-1 text-muted-foreground">/</span>}
                  <button type="button" className="text-primary hover:underline" onClick={() => goToBreadcrumb(i)}>
                    {item.name}
                  </button>
                </span>
              ))}
            </div>
            <div className="space-y-0.5">
              <button
                type="button"
                className={cn(
                  'flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm',
                  currentFolderId === null ? 'bg-accent font-medium text-accent-foreground' : 'hover:bg-muted',
                )}
                onClick={() => goToBreadcrumb(0)}
              >
                <Folder className="h-3.5 w-3.5" />
                Корень
              </button>
              {renderFolderTree('root', 0)}
            </div>
            <form onSubmit={(e) => { e.preventDefault(); handleCreateFolder(e) }} className="flex gap-2">
              <Input
                type="text"
                placeholder="Новая папка..."
                value={newFolderName}
                onChange={(e) => setNewFolderName(e.target.value)}
                disabled={creatingFolder}
              />
              <Button type="button" size="sm" disabled={creatingFolder || !newFolderName.trim()} onClick={() => handleCreateFolder()}>
                +
              </Button>
            </form>
          </CardContent>
        </Card>

        <div className="min-w-0 flex-1">
          <div className="mb-4 flex flex-col gap-3 sm:flex-row">
            <Input
              type="search"
              placeholder="Поиск по артикулу, названию..."
              value={search}
              onChange={(e) => {
                setSearch(e.target.value)
                setPage(0)
              }}
            />
            <NativeSelect
              className="sm:max-w-[200px]"
              value={categoryFilter}
              onChange={(e) => {
                setCategoryFilter(e.target.value)
                setPage(0)
              }}
            >
              <option value="">Все категории</option>
              <option value="Фильтры">Фильтры</option>
              <option value="Тормоза">Тормоза</option>
              <option value="Масла">Масла</option>
              <option value="Расходники">Расходники</option>
            </NativeSelect>
          </div>
          {error && <ErrorAlert message={error} />}
          {loading ? (
            <LoadingState />
          ) : list.length === 0 && !error ? (
            <EmptyState>В этой папке нет запчастей. Добавьте запчасть или выберите другую папку.</EmptyState>
          ) : (
            <>
              <Card>
                <CardContent className="p-0">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Артикул</TableHead>
                        <TableHead>Название</TableHead>
                        <TableHead>Категория</TableHead>
                        <TableHead>Кол-во</TableHead>
                        <TableHead>Ед.</TableHead>
                        <TableHead>Цена</TableHead>
                        <TableHead>Расположение</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {list.map((p) => (
                        <TableRow key={p.id}>
                          <TableCell className="font-mono text-xs">{p.sku}</TableCell>
                          <TableCell>
                            <Button variant="link" className="h-auto p-0" asChild>
                              <Link to={`/parts/${p.id}`}>{p.name || '—'}</Link>
                            </Button>
                          </TableCell>
                          <TableCell>{p.category || '—'}</TableCell>
                          <TableCell>{p.quantity}</TableCell>
                          <TableCell>{p.unit || 'шт'}</TableCell>
                          <TableCell>{p.price ? Number(p.price).toLocaleString('ru') : '—'}</TableCell>
                          <TableCell>{p.location || '—'}</TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </CardContent>
              </Card>
              <Pagination page={page} total={total} limit={limit} onPageChange={setPage} showTotal />
            </>
          )}
        </div>
      </div>
    </div>
  )
}
