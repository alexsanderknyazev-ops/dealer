import { useEffect, useRef, useState } from 'react'
import { ChevronDown, Loader2, Search, X } from 'lucide-react'
import { cn } from '@/lib/utils'

export type SearchSelectItem<T = unknown> = {
  value: string
  label: string
  sublabel?: string
  data?: T
}

type SearchSelectProps<T = unknown> = {
  value: string
  displayValue?: string
  placeholder?: string
  disabled?: boolean
  clearable?: boolean
  presetItems?: SearchSelectItem<T>[]
  presetHeader?: string
  onChange: (value: string, item?: SearchSelectItem<T>) => void
  onSearch: (query: string) => Promise<SearchSelectItem<T>[]>
  className?: string
}

function matches<T>(item: SearchSelectItem<T>, q: string): boolean {
  const hay = `${item.label} ${item.sublabel || ''}`.toLowerCase()
  return hay.includes(q.toLowerCase())
}

export function SearchSelect<T = unknown>({
  value,
  displayValue,
  placeholder = 'Начните вводить…',
  disabled,
  clearable,
  presetItems = [],
  presetHeader,
  onChange,
  onSearch,
  className,
}: SearchSelectProps<T>) {
  const [query, setQuery] = useState('')
  const [items, setItems] = useState<SearchSelectItem<T>[]>([])
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(false)
  const [highlighted, setHighlighted] = useState(0)
  const rootRef = useRef<HTMLDivElement>(null)
  const searchIdRef = useRef(0)

  const shownValue = open ? query : (displayValue || '')

  useEffect(() => {
    if (!open) return
    function onDocClick(e: MouseEvent) {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', onDocClick)
    return () => document.removeEventListener('mousedown', onDocClick)
  }, [open])

  async function runSearch(q: string, local: SearchSelectItem<T>[]) {
    const id = ++searchIdRef.current
    setLoading(true)
    setError(false)
    try {
      const result = await onSearch(q.trim())
      if (id !== searchIdRef.current) return
      const seen = new Set<string>()
      const merged = [...local, ...result].filter((it) => {
        if (seen.has(it.value)) return false
        seen.add(it.value)
        return true
      })
      setItems(merged)
      setHighlighted(0)
      setOpen(true)
    } catch {
      if (id !== searchIdRef.current) return
      if (local.length > 0) {
        setItems(local)
        setOpen(true)
      } else {
        setItems([])
        setError(true)
        setOpen(true)
      }
    } finally {
      if (id === searchIdRef.current) setLoading(false)
    }
  }

  function handleInput(v: string) {
    setQuery(v)
    if (!v) {
      onChange('', undefined)
      setItems([])
      setOpen(false)
      return
    }
    const local = presetItems.filter((it) => matches(it, v))
    if (local.length > 0) {
      setItems(local)
      setHighlighted(0)
      setOpen(true)
    } else {
      setItems([])
    }
    void runSearch(v, local)
  }

  function handleFocus() {
    if (open) return
    setQuery(shownValue)
    if (!shownValue && presetItems.length > 0) {
      setItems(presetItems)
      setHighlighted(0)
      setOpen(true)
    } else {
      setOpen(false)
    }
  }

  function selectItem(item: SearchSelectItem<T>) {
    setQuery('')
    setItems([])
    setOpen(false)
    onChange(item.value, item)
  }

  function clear() {
    setQuery('')
    setItems([])
    setOpen(false)
    onChange('', undefined)
  }

  function onKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'ArrowDown' && items.length > 0) {
      e.preventDefault()
      setHighlighted((h) => (h + 1) % items.length)
    } else if (e.key === 'ArrowUp' && items.length > 0) {
      e.preventDefault()
      setHighlighted((h) => (h - 1 + items.length) % items.length)
    } else if (e.key === 'Enter' && open && items.length > 0) {
      e.preventDefault()
      const item = items[Math.min(highlighted, items.length - 1)]
      if (item) selectItem(item)
    } else if (e.key === 'Escape') {
      setOpen(false)
    }
  }

  const hasValue = Boolean(value)

  return (
    <div ref={rootRef} className={cn('relative', className)}>
      <div className="relative">
        <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <input
          className="flex h-9 w-full rounded-md border border-input bg-transparent py-1 pl-9 pr-8 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
          value={shownValue}
          placeholder={placeholder}
          disabled={disabled}
          onChange={(e) => handleInput(e.target.value)}
          onKeyDown={onKeyDown}
          onFocus={handleFocus}
        />
        <div className="absolute right-2.5 top-1/2 -translate-y-1/2">
          {loading ? (
            <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
          ) : clearable && hasValue && !open ? (
            <button
              type="button"
              tabIndex={-1}
              onClick={clear}
              className="rounded-sm text-muted-foreground hover:text-foreground"
              aria-label="Очистить"
            >
              <X className="h-4 w-4" />
            </button>
          ) : (
            <ChevronDown className="h-4 w-4 text-muted-foreground" />
          )}
        </div>
      </div>

      {open && (
        <div className="absolute z-50 mt-1 max-h-64 w-full overflow-auto rounded-md border bg-popover text-popover-foreground shadow-lg">
          {presetHeader && items.length > 0 && query.trim() === '' && (
            <div className="px-3 pt-2 pb-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">
              {presetHeader}
            </div>
          )}
          {error && (
            <div className="px-3 py-2 text-sm text-destructive">Не удалось выполнить поиск</div>
          )}
          {!error && items.length === 0 && !loading && (
            <div className="px-3 py-2 text-sm text-muted-foreground">Ничего не найдено</div>
          )}
          {items.map((item, idx) => (
            <button
              key={item.value}
              type="button"
              onMouseEnter={() => setHighlighted(idx)}
              onClick={() => selectItem(item)}
              className={cn(
                'flex w-full items-center justify-between gap-2 px-3 py-2 text-left text-sm',
                idx === highlighted && 'bg-accent text-accent-foreground',
              )}
            >
              <span className="min-w-0">
                <span className="block truncate">{item.label}</span>
                {item.sublabel && (
                  <span className="block truncate text-xs text-muted-foreground">{item.sublabel}</span>
                )}
              </span>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
