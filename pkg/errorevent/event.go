package errorevent

import (
	"time"

	"github.com/google/uuid"
)

const SchemaVersion = 1

// Event — единый контракт ошибки/инцидента для Kafka и ClickHouse.
type Event struct {
	SchemaVersion int            `json:"schema_version"`
	EventID       string         `json:"event_id"`
	OccurredAt    string         `json:"occurred_at"`
	Source        string         `json:"source"`
	Kind          string         `json:"kind"`
	Severity      string         `json:"severity"`
	Message       string         `json:"message"`
	ErrorCode     string         `json:"error_code,omitempty"`
	HTTPStatus    int            `json:"http_status,omitempty"`
	GRPCCode      string         `json:"grpc_code,omitempty"`
	TraceID       string         `json:"trace_id,omitempty"`
	RequestID     string         `json:"request_id,omitempty"`
	UserID        string         `json:"user_id,omitempty"`
	Route         string         `json:"route,omitempty"`
	Service       string         `json:"service"`
	Environment   string         `json:"environment,omitempty"`
	Context       map[string]any `json:"context,omitempty"`
	Stack         string         `json:"stack,omitempty"`
	Client        map[string]any `json:"client,omitempty"`
}

// New создаёт событие с обязательными полями.
func New(service, kind, severity, message string) Event {
	return Event{
		SchemaVersion: SchemaVersion,
		EventID:       uuid.NewString(),
		OccurredAt:    time.Now().UTC().Format(time.RFC3339Nano),
		Source:        service,
		Service:       service,
		Kind:          kind,
		Severity:      severity,
		Message:       message,
	}
}
