# gateway-employee — детальные тест-кейсы

Gateway не имеет своей БД — проверяем **маршрутизацию**, **auth header forwarding**, **CORS**, **503 при падении backend**.

---

## TC-GW-D001 — Health (P0)

```bash
curl -sf http://127.0.0.1:8090/healthz
curl -sf http://127.0.0.1:8090/readyz
```

---

## TC-GW-D002 — Auth header forwarding (P0)

1. Login employee → TOKEN  
2. POST /api/customers с Bearer  
3. **БД:** row created in `customers.customers` — доказывает, что gateway передал JWT в gRPC metadata

Без Bearer → 401, БД без изменений.

---

## TC-GW-D003 — CORS preflight (P1)

```bash
curl -s -D- -X OPTIONS http://127.0.0.1:8090/api/vehicles \
  -H 'Origin: http://localhost:5173' \
  -H 'Access-Control-Request-Method: POST'
```
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Headers` contains Authorization

---

## TC-GW-D004 — Backend down → 503 (P1)

Stop `workorders-service`, then:

```bash
curl -s -w "\nHTTP:%{http_code}\n" -H "Authorization: Bearer $TOKEN" \
  http://127.0.0.1:8090/api/work-orders
```
- **503** (circuit breaker / connection refused)

Other services still **200** — изоляция backend-ов.

---

## TC-GW-D005 — JSON contract (P2)

GET /api/deals с пустым списком:

```bash
curl -s -H "Authorization: Bearer $TOKEN" http://127.0.0.1:8090/api/deals
```
- `deals: []` present (EmitUnpopulated=true в gateway)
- `total: 0`

---

## TC-GW-D006 — All registered routes smoke (P1)

Для каждого backend один authenticated GET:

| Path | Backend schema check |
|------|-------------------|
| /api/me | auth (no DB for validate) |
| /api/customers | customers |
| /api/vehicles | vehicles |
| /api/deals | deals |
| /api/parts | parts |
| /api/brands | brands |
| /api/dealer-points | dealerpoints |
| /api/work-orders | workorders |
| /api/movement-documents | parts |
| /api/reviews | employee_reviews |
| /api/stats/employee/overview | employee_statistics |
| /api/stats/client/overview | client_statistics |

Все **200** (или 503 для WO если down) — Postgres не меняется.

---

## TC-GW-D007 — Invalid UUID path (P2)

```bash
curl -s -w "\nHTTP:%{http_code}\n" -H "Authorization: Bearer $TOKEN" \
  http://127.0.0.1:8090/api/customers/not-a-uuid
```
- **400/404**, no DB side effects
