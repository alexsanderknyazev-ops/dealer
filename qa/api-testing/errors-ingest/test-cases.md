# errors-ingest — тест-кейсы

**Сервис:** `errors-ingest-service`  
**HTTP:** `:8092`  
**Kafka consumer:** `platform.errors.v1` → ClickHouse

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-ERR-001 | P0 | POST /api/telemetry/events | — | `{"kind":"js_error","message":"test","at":123}` | 204 | ERR-001 |
| TC-ERR-002 | P1 | POST /api/telemetry/events | — | `kind=api_latency`, path, status, duration_ms | 204 | ERR-002 |
| TC-ERR-003 | P1 | POST /api/telemetry/events | — | invalid body | 400 | ERR-003 |
| TC-ERR-004 | P1 | OPTIONS /api/telemetry/events | — | CORS | 204 | ERR-004 |
| TC-ERR-005 | P2 | Kafka platform.errors.v1 | — | backend error report | row in ClickHouse | manual |
| TC-ERR-006 | P1 | Proxy via auth :8080 | — | POST /api/telemetry/events | 204 | ERR-006 |

## Event kinds (frontend)

- `js_error`, `promise_rejection`, `api_latency`
