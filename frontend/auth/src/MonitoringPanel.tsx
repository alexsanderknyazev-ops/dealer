import { useEffect, useState } from 'react'

type MonitoringSnapshot = {
  jsErrors: number
  promiseRejections: number
  apiCalls: number
  apiSlowCalls: number
}

export function MonitoringPanel() {
  const [snapshot, setSnapshot] = useState<MonitoringSnapshot>({
    jsErrors: 0,
    promiseRejections: 0,
    apiCalls: 0,
    apiSlowCalls: 0,
  })

  useEffect(() => {
    const update = () => {
      const getter = window.__dealerMonitoring?.getSnapshot
      if (!getter) return
      const data = getter()
      setSnapshot({
        jsErrors: data.jsErrors,
        promiseRejections: data.promiseRejections,
        apiCalls: data.apiCalls,
        apiSlowCalls: data.apiSlowCalls,
      })
    }
    update()
    const timer = window.setInterval(update, 3000)
    return () => window.clearInterval(timer)
  }, [])

  return (
    <div className="dashboard-monitoring">
      <h2 className="dashboard-monitoring-title">Frontend мониторинг</h2>
      <div className="dashboard-monitoring-grid">
        <span>JS ошибки: {snapshot.jsErrors}</span>
        <span>Unhandled promises: {snapshot.promiseRejections}</span>
        <span>API запросы: {snapshot.apiCalls}</span>
        <span>Медленные API (&gt; 800мс): {snapshot.apiSlowCalls}</span>
      </div>
    </div>
  )
}
