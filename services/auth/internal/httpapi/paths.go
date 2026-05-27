package httpapi

// Auth HTTP API path literals (single place; handlers and tests use these identifiers).
const (
	pathAPIRegister = "/api/register"
	pathAPILogin    = "/api/login"
	pathAPIRefresh  = "/api/refresh"
	pathAPILogout   = "/api/logout"
	pathAPIMe       = "/api/me"
)

// authPathsWithOptions lists paths that need both a handler route and an OPTIONS preflight route.
var authPathsWithOptions = []string{
	pathAPIRegister,
	pathAPILogin,
	pathAPIRefresh,
	pathAPILogout,
	pathAPIMe,
}
