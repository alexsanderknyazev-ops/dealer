type MonitoringEvent =
  | { kind: 'js_error'; message: string; source?: string; at: number }
  | { kind: 'promise_rejection'; message: string; at: number }
  | { kind: 'api_latency'; path: string; status: number; duration_ms: number; at: number }

type MonitoringSnapshot = {
  jsErrors: number
  promiseRejections: number
  apiCalls: number
  apiSlowCalls: number
  lastEvents: MonitoringEvent[]
}

declare global {
  interface Window {
    __dealerMonitoring?: { getSnapshot: () => MonitoringSnapshot }
  }
}

const MONITORING_ENDPOINT = import.meta.env.VITE_FRONTEND_MONITORING_URL as string | undefined
const SLOW_API_THRESHOLD_MS = 800
const MAX_EVENTS = 50

const events: MonitoringEvent[] = []
let jsErrors = 0
let promiseRejections = 0
let apiCalls = 0
let apiSlowCalls = 0
let initialized = false

function pushEvent(event: MonitoringEvent) {
  events.unshift(event)
  if (events.length > MAX_EVENTS) events.pop()
  if (!MONITORING_ENDPOINT) return
  fetch(MONITORING_ENDPOINT, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(event),
    keepalive: true,
  }).catch(() => {
    // Monitoring is best-effort and should never break UI.
  })
}

export function initMonitoring() {
  if (initialized || typeof window === 'undefined') return
  initialized = true

  window.addEventListener('error', (event) => {
    jsErrors += 1
    pushEvent({
      kind: 'js_error',
      message: event.message || 'Unknown JS error',
      source: event.filename,
      at: Date.now(),
    })
  })

  window.addEventListener('unhandledrejection', (event) => {
    promiseRejections += 1
    const message = event.reason instanceof Error ? event.reason.message : String(event.reason || 'Unhandled rejection')
    pushEvent({ kind: 'promise_rejection', message, at: Date.now() })
  })

  const originalFetch = window.fetch.bind(window)
  window.fetch = async (...args) => {
    const startedAt = performance.now()
    const response = await originalFetch(...args)
    const duration_ms = Math.round(performance.now() - startedAt)
    apiCalls += 1
    if (duration_ms >= SLOW_API_THRESHOLD_MS) apiSlowCalls += 1
    const url = typeof args[0] === 'string' ? args[0] : args[0] instanceof Request ? args[0].url : ''
    const path = (() => {
      try {
        return new URL(url, window.location.origin).pathname
      } catch {
        return url
      }
    })()
    pushEvent({
      kind: 'api_latency',
      path,
      status: response.status,
      duration_ms,
      at: Date.now(),
    })
    return response
  }

  window.__dealerMonitoring = {
    getSnapshot: () => ({
      jsErrors,
      promiseRejections,
      apiCalls,
      apiSlowCalls,
      lastEvents: [...events],
    }),
  }
}
