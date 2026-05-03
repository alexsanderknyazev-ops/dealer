package httpapi

const pathAPIBrands = "/api/brands"

func pathBrandByID(id string) string {
	return pathAPIBrands + "/" + id
}
