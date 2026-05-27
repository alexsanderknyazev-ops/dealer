package httpapi

import "net/http"

const (
	headerContentType     = "Content-Type"
	mimeApplicationJSON   = "application/json"
	mimeHTMLUTF8          = "text/html; charset=utf-8"
	headerCORSAllowHeader = headerContentType + ", Authorization"
)

func setRequestJSONContentType(r *http.Request) {
	r.Header.Set(headerContentType, mimeApplicationJSON)
}
