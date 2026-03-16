import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from './auth'
import './Form.css'

export function Login() {
  const { login } = useAuth()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      await login(email, password)
      navigate('/customers', { replace: true })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка входа')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="form-card">
      <h1 className="form-title">Вход</h1>
      <form onSubmit={handleSubmit} className="form">
        {error && <div className="form-error">{error}</div>}
        <label className="form-label">
          Email
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            autoComplete="email"
            className="form-input"
          />
        </label>
        <label className="form-label">
          Пароль
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            autoComplete="current-password"
            className="form-input"
          />
        </label>
        <button type="submit" disabled={submitting} className="form-submit">
          {submitting ? 'Вход…' : 'Войти'}
        </button>
      </form>
      <p className="form-footer">
        Нет аккаунта? <Link to="/register">Регистрация</Link>
      </p>
    </div>
  )
}
