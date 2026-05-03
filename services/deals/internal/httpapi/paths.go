package httpapi

const pathAPIDeals = "/api/deals"

func pathDealByID(id string) string {
	return pathAPIDeals + "/" + id
}
