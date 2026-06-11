# client-protected-gateway — детальные тест-кейсы

**HTTP:** `:8093` — только authenticated client routes.

---

## TC-CPP-D001 — Route table (P0)

| Method | Path | Backend | DB on success |
|--------|------|---------|---------------|
| GET | /api/me | client-auth session | Redis only |
| GET | /api/client/profile | client-registration | clients.clients |
| GET | /api/client/vehicles | client-registration | clients.client_vehicles |
| GET/POST | /api/client/reviews | client-reviews | reviews.reviews |

---

## TC-CPP-D002 — Valid client token (P0)

После client register:

```bash
for path in /api/me /api/client/profile /api/client/vehicles /api/client/reviews; do
  curl -s -o /dev/null -w "$path -> %{http_code}\n" \
    -H "Authorization: Bearer $CLIENT_ACCESS" \
    "http://127.0.0.1:8093$path"
done
```
Все **200**.

Cross-check profile с БД (см. client-registration TC-CR-D004).

---

## TC-CPP-D003 — Employee token rejected (P0)

```bash
curl -s -w "\nHTTP:%{http_code}\n" \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  http://127.0.0.1:8093/api/client/profile
```
- **401/403**
- `clients.clients` COUNT unchanged

---

## TC-CPP-D004 — No Authorization (P1)

Все routes → **401**

---

## TC-CPP-D005 — CORS + Authorization header (P1)

OPTIONS + POST review с Bearer — preflight passes, POST creates DB row.

---

## TC-CPP-D006 — Expired client JWT (P2)

Использовать access_token с прошедшим `exp` → **401** на /api/me

Redis refresh всё ещё может работать для renewal через :8091.
