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
import { BrandLaborRates } from './BrandLaborRates'
import { DealerPoints } from './DealerPoints'
import { DealerPointForm } from './DealerPointForm'
import { LegalEntities } from './LegalEntities'
import { LegalEntityForm } from './LegalEntityForm'
import { Warehouses } from './Warehouses'
import { WarehouseForm } from './WarehouseForm'
import { Statistics } from './Statistics'
import { WorkOrders } from './WorkOrders'
import { WorkOrderForm } from './WorkOrderForm'
import { WorkOrderView } from './WorkOrderView'
import { MovementDocuments } from './MovementDocuments'
import { MovementDocumentView } from './MovementDocumentView'
import { MovementDocumentForm } from './MovementDocumentForm'
import { SupplierOrders } from './SupplierOrders'
import { SupplierOrderForm } from './SupplierOrderForm'
import { SupplierOrderView } from './SupplierOrderView'
import { CustomerOrders } from './CustomerOrders'
import { CustomerOrderForm } from './CustomerOrderForm'
import { CustomerOrderView } from './CustomerOrderView'
import { RepairAppointments } from './RepairAppointments'
import { RepairAppointmentForm } from './RepairAppointmentForm'
import { RepairAppointmentView } from './RepairAppointmentView'
import { Works } from './Works'
import { WorkForm } from './WorkForm'
import { WorkView } from './WorkView'
import { Employees } from './Employees'
import { EmployeeForm } from './EmployeeForm'
import { EmployeeView } from './EmployeeView'
import { Reviews } from './Reviews'
import { ReviewView } from './ReviewView'

function RequireAuth(props: { children: ReactNode }) {
  const { user, loading } = useAuth()
  if (loading) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center text-muted-foreground">Загрузка…</div>
    )
  }
  if (!user) return <Navigate to="/login" replace />
  return <>{props.children}</>
}

function GuestOnly(props: { children: ReactNode }) {
  const { user, loading } = useAuth()
  if (loading) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center text-muted-foreground">Загрузка…</div>
    )
  }
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
        <Route path="works" element={<RequireAuth><Works /></RequireAuth>} />
        <Route path="works/new" element={<RequireAuth><WorkForm /></RequireAuth>} />
        <Route path="works/:id/edit" element={<RequireAuth><WorkForm /></RequireAuth>} />
        <Route path="works/:id" element={<RequireAuth><WorkView /></RequireAuth>} />
        <Route path="employees" element={<RequireAuth><Employees /></RequireAuth>} />
        <Route path="employees/new" element={<RequireAuth><EmployeeForm /></RequireAuth>} />
        <Route path="employees/:id" element={<RequireAuth><EmployeeView /></RequireAuth>} />
        <Route path="employees/:id/edit" element={<RequireAuth><EmployeeForm /></RequireAuth>} />
        <Route path="brands" element={<RequireAuth><Brands /></RequireAuth>} />
        <Route path="brands/new" element={<RequireAuth><BrandForm /></RequireAuth>} />
        <Route path="brands/:id/edit" element={<RequireAuth><BrandForm /></RequireAuth>} />
        <Route path="brand-labor-rates" element={<RequireAuth><BrandLaborRates /></RequireAuth>} />
        <Route path="dealer-points" element={<RequireAuth><DealerPoints /></RequireAuth>} />
        <Route path="dealer-points/new" element={<RequireAuth><DealerPointForm /></RequireAuth>} />
        <Route path="dealer-points/:id/edit" element={<RequireAuth><DealerPointForm /></RequireAuth>} />
        <Route path="legal-entities" element={<RequireAuth><LegalEntities /></RequireAuth>} />
        <Route path="legal-entities/new" element={<RequireAuth><LegalEntityForm /></RequireAuth>} />
        <Route path="legal-entities/:id/edit" element={<RequireAuth><LegalEntityForm /></RequireAuth>} />
        <Route path="warehouses" element={<RequireAuth><Warehouses /></RequireAuth>} />
        <Route path="warehouses/new" element={<RequireAuth><WarehouseForm /></RequireAuth>} />
        <Route path="warehouses/:id/edit" element={<RequireAuth><WarehouseForm /></RequireAuth>} />
        <Route path="statistics" element={<RequireAuth><Statistics /></RequireAuth>} />
        <Route path="reviews" element={<RequireAuth><Reviews /></RequireAuth>} />
        <Route path="reviews/:id" element={<RequireAuth><ReviewView /></RequireAuth>} />
        <Route path="work-orders" element={<RequireAuth><WorkOrders /></RequireAuth>} />
        <Route path="work-orders/new" element={<RequireAuth><WorkOrderForm /></RequireAuth>} />
        <Route path="work-orders/:id/edit" element={<RequireAuth><WorkOrderForm /></RequireAuth>} />
        <Route path="movement-documents" element={<RequireAuth><MovementDocuments /></RequireAuth>} />
        <Route path="movement-documents/new" element={<RequireAuth><MovementDocumentForm /></RequireAuth>} />
        <Route path="movement-documents/:id/edit" element={<RequireAuth><MovementDocumentForm /></RequireAuth>} />
        <Route path="movement-documents/:id" element={<RequireAuth><MovementDocumentView /></RequireAuth>} />
        <Route path="supplier-orders" element={<RequireAuth><SupplierOrders /></RequireAuth>} />
        <Route path="supplier-orders/new" element={<RequireAuth><SupplierOrderForm /></RequireAuth>} />
        <Route path="supplier-orders/:id/edit" element={<RequireAuth><SupplierOrderForm /></RequireAuth>} />
        <Route path="supplier-orders/:id" element={<RequireAuth><SupplierOrderView /></RequireAuth>} />
        <Route path="customer-orders" element={<RequireAuth><CustomerOrders /></RequireAuth>} />
        <Route path="customer-orders/new" element={<RequireAuth><CustomerOrderForm /></RequireAuth>} />
        <Route path="customer-orders/:id/edit" element={<RequireAuth><CustomerOrderForm /></RequireAuth>} />
        <Route path="customer-orders/:id" element={<RequireAuth><CustomerOrderView /></RequireAuth>} />
        <Route path="repair-appointments" element={<RequireAuth><RepairAppointments /></RequireAuth>} />
        <Route path="repair-appointments/new" element={<RequireAuth><RepairAppointmentForm /></RequireAuth>} />
        <Route path="repair-appointments/:id/edit" element={<RequireAuth><RepairAppointmentForm /></RequireAuth>} />
        <Route path="repair-appointments/:id" element={<RequireAuth><RepairAppointmentView /></RequireAuth>} />
        <Route path="work-orders/:id" element={<RequireAuth><WorkOrderView /></RequireAuth>} />
        <Route path="login" element={<GuestOnly><Login /></GuestOnly>} />
        <Route path="register" element={<GuestOnly><Register /></GuestOnly>} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  )
}
