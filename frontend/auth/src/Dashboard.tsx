import { useNavigate } from 'react-router-dom'
import { useAuth } from './auth'
import './Dashboard.css'

export function Dashboard() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  async function handleLogout() {
    await logout()
    navigate('/login', { replace: true })
  }

  return (
    <div className="dashboard">
      <div className="dashboard-card">
        <p className="dashboard-welcome">Вы вошли как</p>
        <p className="dashboard-email">{user?.email}</p>
        <div className="dashboard-actions">
          <button type="button" onClick={() => navigate('/customers')} className="dashboard-primary">
            Клиенты
          </button>
          <button type="button" onClick={() => navigate('/vehicles')} className="dashboard-primary">
            Автомобили
          </button>
          <button type="button" onClick={() => navigate('/deals')} className="dashboard-primary">
            Сделки
          </button>
          <button type="button" onClick={() => navigate('/parts')} className="dashboard-primary">
            Запчасти
          </button>
          <button type="button" onClick={handleLogout} className="dashboard-logout">
            Выйти
          </button>
        </div>
      </div>
    </div>
  )
}
