package httpapi

const pathAPIVehicles = "/api/vehicles"

func pathVehicleByID(id string) string {
	return pathAPIVehicles + "/" + id
}
