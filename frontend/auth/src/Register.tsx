import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from './auth'
import './Form.css'

export function Register() {
  const { register } = useAuth()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [name, setName] = useState('')
  const [phone, setPhone] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      await register(email, password, name || undefined, phone || undefined)
      navigate('/customers', { replace: true })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка регистрации')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="form-card">
      <h1 className="form-title">Регистрация</h1>
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
            autoComplete="new-password"
            className="form-input"
          />
        </label>
        <label className="form-label">
          Имя
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            autoComplete="name"
            className="form-input"
            placeholder="Необязательно"
          />
        </label>
        <label className="form-label">
          Телефон
          <input
            type="tel"
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            autoComplete="tel"
            className="form-input"
            placeholder="Необязательно"
          />
        </label>
        <button type="submit" disabled={submitting} className="form-submit">
          {submitting ? 'Регистрация…' : 'Зарегистрироваться'}
        </button>
      </form>
      <p className="form-footer">
        Уже есть аккаунт? <Link to="/login">Войти</Link>
      </p>
    </div>
  )
}
