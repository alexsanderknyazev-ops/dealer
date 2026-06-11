package httplog

import (
	"fmt"
	"net/http"

	"github.com/dealer/dealer/pkg/errorevent"
	"github.com/dealer/dealer/pkg/errorreport"
)

func shouldReportHTTP(status int) bool {
	return status >= http.StatusInternalServerError
}

func reportHTTP(reporter *errorreport.Reporter, service, method, path string, status int) {
	if reporter == nil || !shouldReportHTTP(status) {
		return
	}
	ev := errorevent.New(service, "http_error", httpSeverity(status), fmt.Sprintf("HTTP %d %s", status, http.StatusText(status)))
	ev.HTTPStatus = status
	ev.Route = method + " " + path
	reporter.Report(ev)
}

func httpSeverity(status int) string {
	if status >= http.StatusInternalServerError {
		return "error"
	}
	return "warn"
}
