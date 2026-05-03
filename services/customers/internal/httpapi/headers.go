package httpapi

import "net/http"

const (
	headerContentType     = "Content-Type"
	mimeApplicationJSON   = "application/json"
	headerCORSAllowHeader = headerContentType + ", Authorization"
)

func setRequestJSONContentType(r *http.Request) {
	r.Header.Set(headerContentType, mimeApplicationJSON)
}
