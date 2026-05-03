package httpapi

import "net/http"

const (
	headerContentType      = "Content-Type"
	mimeApplicationJSON    = "application/json"
	headerCORSAllowHeaders = headerContentType + ", Authorization"
)

func setRequestJSONContentType(r *http.Request) {
	r.Header.Set(headerContentType, mimeApplicationJSON)
}
