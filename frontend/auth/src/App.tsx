import type { ReactNode } from 'react'
import { Navigate, Route, Routes } from 'react-router-dom'
import { useAuth } from './auth'
import { CustomerForm } from './CustomerForm'
import { CustomerView } from './CustomerView'
import { Customers } from './Customers'
import { Dashboard } from './Dashboard'
import { Layout } from './Layout'
import { Login } from './Login'
import { Register } from './Register'
import { DealForm } from './DealForm'
import { DealView } from './DealView'
import { Deals } from './Deals'
import { PartForm } from './PartForm'
import { PartView } from './PartView'
import { Parts } from './Parts'
import { VehicleForm } from './VehicleForm'
import { VehicleView } from './VehicleView'
import { Vehicles } from './Vehicles'
import { Brands } from './Brands'
import { BrandForm } from './BrandForm'
import { DealerPoints } from './DealerPoints'
import { DealerPointForm } from './DealerPointForm'
import { LegalEntities } from './LegalEntities'
import { LegalEntityForm } from './LegalEntityForm'
import { Warehouses } from './Warehouses'
import { WarehouseForm } from './WarehouseForm'

function RequireAuth(props: { children: ReactNode }) {
  const { user, loading } = useAuth()
  if (loading) return <div className="layout"><div className="main loading">Загрузка…</div></div>
  if (!user) return <Navigate to="/login" replace />
  return <>{props.children}</>
}

function GuestOnly(props: { children: ReactNode }) {
  const { user, loading } = useAuth()
  if (loading) return <div className="layout"><div className="main loading">Загрузка…</div></div>
  if (user) return <Navigate to="/" replace />
  return <>{props.children}</>
}

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<RequireAuth><Dashboard /></RequireAuth>} />
        <Route path="customers" element={<RequireAuth><Customers /></RequireAuth>} />
        <Route path="customers/new" element={<RequireAuth><CustomerForm /></RequireAuth>} />
        <Route path="customers/:id" element={<RequireAuth><CustomerView /></RequireAuth>} />
        <Route path="customers/:id/edit" element={<RequireAuth><CustomerForm /></RequireAuth>} />
        <Route path="vehicles" element={<RequireAuth><Vehicles /></RequireAuth>} />
        <Route path="vehicles/new" element={<RequireAuth><VehicleForm /></RequireAuth>} />
        <Route path="vehicles/:id" element={<RequireAuth><VehicleView /></RequireAuth>} />
        <Route path="vehicles/:id/edit" element={<RequireAuth><VehicleForm /></RequireAuth>} />
        <Route path="deals" element={<RequireAuth><Deals /></RequireAuth>} />
        <Route path="deals/new" element={<RequireAuth><DealForm /></RequireAuth>} />
        <Route path="deals/:id" element={<RequireAuth><DealView /></RequireAuth>} />
        <Route path="deals/:id/edit" element={<RequireAuth><DealForm /></RequireAuth>} />
        <Route path="parts" element={<RequireAuth><Parts /></RequireAuth>} />
        <Route path="parts/new" element={<RequireAuth><PartForm /></RequireAuth>} />
        <Route path="parts/:id" element={<RequireAuth><PartView /></RequireAuth>} />
        <Route path="parts/:id/edit" element={<RequireAuth><PartForm /></RequireAuth>} />
        <Route path="brands" element={<RequireAuth><Brands /></RequireAuth>} />
        <Route path="brands/new" element={<RequireAuth><BrandForm /></RequireAuth>} />
        <Route path="brands/:id/edit" element={<RequireAuth><BrandForm /></RequireAuth>} />
        <Route path="dealer-points" element={<RequireAuth><DealerPoints /></RequireAuth>} />
        <Route path="dealer-points/new" element={<RequireAuth><DealerPointForm /></RequireAuth>} />
        <Route path="dealer-points/:id/edit" element={<RequireAuth><DealerPointForm /></RequireAuth>} />
        <Route path="legal-entities" element={<RequireAuth><LegalEntities /></RequireAuth>} />
        <Route path="legal-entities/new" element={<RequireAuth><LegalEntityForm /></RequireAuth>} />
        <Route path="legal-entities/:id/edit" element={<RequireAuth><LegalEntityForm /></RequireAuth>} />
        <Route path="warehouses" element={<RequireAuth><Warehouses /></RequireAuth>} />
        <Route path="warehouses/new" element={<RequireAuth><WarehouseForm /></RequireAuth>} />
        <Route path="warehouses/:id/edit" element={<RequireAuth><WarehouseForm /></RequireAuth>} />
        <Route path="login" element={<GuestOnly><Login /></GuestOnly>} />
        <Route path="register" element={<GuestOnly><Register /></GuestOnly>} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  )
}
