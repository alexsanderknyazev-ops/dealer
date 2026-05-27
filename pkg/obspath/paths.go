package obspath

// Пути health/metrics, исключаемые из access-логов и HTTP-метрик.
const (
	Healthz = "/healthz"
	Readyz  = "/readyz"
	Metrics = "/metrics"
)

// Probes — все probe-пути (для тестов и итерации).
var Probes = []string{Healthz, Readyz, Metrics}

// IsProbe сообщает, нужно ли пропустить логирование/метрики для path.
func IsProbe(path string) bool {
	switch path {
	case Healthz, Readyz, Metrics:
		return true
	default:
		return false
	}
}
