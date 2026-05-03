package httpapi

const pathAPICustomers = "/api/customers"

func pathCustomerByID(id string) string {
	return pathAPICustomers + "/" + id
}
