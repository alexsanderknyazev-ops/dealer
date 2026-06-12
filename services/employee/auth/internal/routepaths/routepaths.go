// Package routepaths holds reverse-proxy path prefixes for auth-service (single definition for main).
package routepaths

const (
	APIRegister = "/api/register"
	APILogin    = "/api/login"
	APIRefresh  = "/api/refresh"
	APILogout   = "/api/logout"
	APIMe       = "/api/me"

	APICustomers        = "/api/customers"
	APICustomersPrefix  = "/api/customers/"
	APIVehicles         = "/api/vehicles"
	APIVehiclesPrefix   = "/api/vehicles/"
	APIDeals            = "/api/deals"
	APIDealsPrefix      = "/api/deals/"
	APIParts            = "/api/parts"
	APIPartsPrefix      = "/api/parts/"
	APIBrands           = "/api/brands"
	APIBrandsPrefix     = "/api/brands/"
	APIBrandLaborRates  = "/api/brand-labor-rates"
	APIDealerPoints     = "/api/dealer-points"
	APIDealerPointsPre  = "/api/dealer-points/"
	APILegalEntities    = "/api/legal-entities"
	APILegalEntitiesPre = "/api/legal-entities/"
	APIWarehouses       = "/api/warehouses"
	APIWarehousesPrefix = "/api/warehouses/"
	APIStats            = "/api/stats"
	APIReviews          = "/api/reviews"
	APIClients          = "/api/clients"
	APIWorkOrders       = "/api/work-orders"
	APIWorks            = "/api/works"
	APIWorksPrefix      = "/api/works/"
	APIEmployees        = "/api/employees"
	APIEmployeesPrefix  = "/api/employees/"
	APIMovementDocuments    = "/api/movement-documents"
	APIRepairAppointments   = "/api/repair-appointments"
	APITelemetry        = "/api/telemetry"
	APITelemetryPrefix  = "/api/telemetry/"
)

// GatewayProxyPrefixes — REST paths served by grpc-gateway (Variant A).
func GatewayProxyPrefixes() []string {
	return []string{
		APIRegister,
		APILogin,
		APIRefresh,
		APILogout,
		APIMe,
		APICustomers,
		APIVehicles,
		APIDeals,
		APIParts,
		APIBrands,
		APIBrandLaborRates,
		APIDealerPoints,
		APILegalEntities,
		APIWarehouses,
		APIStats,
		APIReviews,
		APIClients,
		APIWorkOrders,
		APIWorks,
		APIEmployees,
		APIMovementDocuments,
		APIRepairAppointments,
	}
