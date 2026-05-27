import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import App from './App'
import { AuthProvider } from './auth'
import './index.css'

class ErrorBoundary extends React.Component<{ children: React.ReactNode }, { error: Error | null }> {
  state = { error: null as Error | null }
  static getDerivedStateFromError(error: Error) { return { error } }
  render() {
    if (this.state.error) {
      return (
        <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', flexDirection: 'column', gap: 16, padding: 24, background: '#0d0f14', color: '#e6e9ef', fontFamily: 'system-ui' }}>
          <h2 style={{ margin: 0 }}>Ошибка загрузки</h2>
          <p style={{ color: '#8b91a0' }}>{this.state.error.message}</p>
          <button onClick={() => window.location.reload()} style={{ padding: '10px 20px', cursor: 'pointer' }}>Обновить страницу</button>
        </div>
      )
    }
    return this.props.children
  }
}

const root = document.getElementById('root')
if (!root) {
  document.body.innerHTML = '<div style="padding:24px;color:#e6e9ef;">Нет элемента #root</div>'
} else {
  ReactDOM.createRoot(root).render(
    <React.StrictMode>
      <ErrorBoundary>
        <BrowserRouter>
          <AuthProvider>
            <App />
          </AuthProvider>
        </BrowserRouter>
      </ErrorBoundary>
    </React.StrictMode>,
  )
}
