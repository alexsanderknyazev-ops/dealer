package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dealer/dealer/pkg/errorevent"
	"github.com/dealer/dealer/services/errors-ingest/internal/clickhouse"
)

const pathTelemetryEvents = "/api/telemetry/events"

// Handler принимает HTTP-события (frontend telemetry) и пишет в ClickHouse.
type Handler struct {
	store       *clickhouse.Store
	environment string
}

// NewHandler создаёт HTTP handler.
func NewHandler(store *clickhouse.Store, environment string) *Handler {
	return &Handler{store: store, environment: environment}
}

// RegisterRoutes вешает маршруты на mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc(http.MethodPost+" "+pathTelemetryEvents, h.handleTelemetry)
	mux.HandleFunc(http.MethodOptions+" "+pathTelemetryEvents, h.handleOptions)
}

func (h *Handler) handleOptions(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleTelemetry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var raw map[string]any
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	ev, ok := telemetryToEvent(raw, h.environment)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported event"})
		return
	}
	if err := h.store.InsertEvents(r.Context(), []errorevent.Event{ev}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "store failed"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func telemetryToEvent(raw map[string]any, environment string) (errorevent.Event, bool) {
	kind, _ := raw["kind"].(string)
	message, _ := raw["message"].(string)
	switch kind {
	case "js_error", "promise_rejection":
		if message == "" {
			return errorevent.Event{}, false
		}
		ev := errorevent.New("frontend", "frontend_"+kind, "error", message)
		ev.Environment = environment
		if src, ok := raw["source"].(string); ok && src != "" {
			ev.Route = src
		}
		if at := parseAtMS(raw["at"]); !at.IsZero() {
			ev.OccurredAt = at.UTC().Format(time.RFC3339Nano)
		}
		return ev, true
	case "api_latency":
		path, _ := raw["path"].(string)
		status := intFromAny(raw["status"])
		duration := intFromAny(raw["duration_ms"])
		msg := "slow api"
		severity := "info"
		if status >= 500 {
			msg = "api server error"
			severity = "error"
		} else if status >= 400 {
			msg = "api client error"
			severity = "warn"
		}
		ev := errorevent.New("frontend", "frontend_api", severity, msg)
		ev.Environment = environment
		ev.Route = path
		ev.HTTPStatus = status
		ev.Context = map[string]any{"duration_ms": duration, "status": status}
		if at := parseAtMS(raw["at"]); !at.IsZero() {
			ev.OccurredAt = at.UTC().Format(time.RFC3339Nano)
		}
		if status < 400 && duration < 800 {
			ev.Severity = "info"
		}
		return ev, true
	default:
		body, err := json.Marshal(raw)
		if err != nil {
			return errorevent.Event{}, false
		}
		var ev errorevent.Event
		if err := json.Unmarshal(body, &ev); err != nil || ev.Message == "" {
			return errorevent.Event{}, false
		}
		if ev.EventID == "" {
			built := errorevent.New(ev.Service, ev.Kind, ev.Severity, ev.Message)
			ev.EventID = built.EventID
			ev.OccurredAt = built.OccurredAt
			ev.SchemaVersion = built.SchemaVersion
		}
		if ev.Environment == "" {
			ev.Environment = environment
		}
		return ev, true
	}
}

func parseAtMS(v any) time.Time {
	switch n := v.(type) {
	case float64:
		return time.UnixMilli(int64(n))
	case int:
		return time.UnixMilli(int64(n))
	case int64:
		return time.UnixMilli(n)
	default:
		return time.Time{}
	}
}

func intFromAny(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	default:
		return 0
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
