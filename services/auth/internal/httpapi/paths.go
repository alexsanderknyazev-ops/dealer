package httpapi

// Константы путей auth HTTP API (Sonar: без дублирования строковых литералов в mux и тестах).
const (
	pathAPIRegister = "/api/register"
	pathAPILogin    = "/api/login"
	pathAPIRefresh  = "/api/refresh"
	pathAPILogout   = "/api/logout"
	pathAPIMe       = "/api/me"
)
