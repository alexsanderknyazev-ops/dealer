# errors-ingest — детальные тест-кейсы

**HTTP:** `POST /api/telemetry/events` (:8092 или proxy :8080)  
**Sink:** ClickHouse `analytics.events` (если поднят)  
**Kafka consumer:** `platform.errors.v1`

---

## TC-ERR-D001 — js_error event (P0)

### API
```bash
curl -s -w "\nHTTP:%{http_code}\n" -X POST "$ERRORS_INGEST/api/telemetry/events" \
  -H 'Content-Type: application/json' \
  -d '{
    "kind": "js_error",
    "message": "QA test TypeError",
    "source": "/dashboard",
    "at": 1710000000000
  }'
```

### HTTP
- **204**

### ClickHouse (optional)
```bash
curl -s "http://127.0.0.1:8123/?query=SELECT+message,route,severity+FROM+analytics.events+WHERE+message+LIKE+'%QA+test%25'+ORDER+BY+occurred_at+DESC+LIMIT+1"
```
- message contains "QA test"
- route = /dashboard or source mapped

---

## TC-ERR-D002 — api_latency event (P1)

```json
{
  "kind": "api_latency",
  "path": "/api/deals",
  "status": 500,
  "duration_ms": 1200,
  "at": 1710000000000
}
```

### Expected severity mapping
- status ≥ 500 → severity error
- status 4xx → warn
- slow + 2xx → info

---

## TC-ERR-D003 — Invalid event (P1)

```json
{"kind": "js_error", "message": ""}
```
- **400** invalid body
- ClickHouse: no new row

---

## TC-ERR-D004 — Proxy via auth-service (P1)

```bash
curl -s -w "\nHTTP:%{http_code}\n" -X POST "http://127.0.0.1:8080/api/telemetry/events" \
  -H 'Content-Type: application/json' \
  -d '{"kind":"js_error","message":"via proxy","at":123}'
```
- **204** if errors-ingest up, else 502/503

---

## TC-ERR-D005 — CORS OPTIONS (P2)

```bash
curl -s -X OPTIONS "$ERRORS_INGEST/api/telemetry/events" -H 'Origin: http://localhost'
```
- **204**, Access-Control-Allow-Origin: *

---

## TC-ERR-D006 — Backend error via Kafka (P2)

Trigger backend error (invalid request causing logged error with KAFKA_TOPIC_ERRORS):

- Check errors-ingest consumer processed message
- ClickHouse row with service name, kind, severity

БД Postgres **не** используется для errors-ingest.
