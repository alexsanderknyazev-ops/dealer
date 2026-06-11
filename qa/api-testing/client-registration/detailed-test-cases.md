# client-registration + public gateway — детальные тест-кейсы

**Public:** `:8091` — register, login  
**Protected:** `:8093` — profile, vehicles  
**Схемы:** `clients.*`, `clientauth.users`  
**Kafka:** `client.registration.v1`

---

## TC-CR-D001 — Register client with VIN (P0)

### Preconditions
```sql
SELECT id, vin FROM vehicles.vehicles WHERE vin = '<VIN>' AND status = 'available';
-- VEHICLE_ID exists
```

### До
```sql
SELECT COUNT(*) FROM clients.clients WHERE email = 'client-qa@test.local';
SELECT COUNT(*) FROM clientauth.users WHERE email = 'client-qa@test.local';
-- both 0
```

### API
```bash
curl -s -X POST "$CLIENT_PUBLIC/api/client/register" \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "client-qa@test.local",
    "password": "Test1234!",
    "full_name": "Петров Пётр",
    "phone": "+79005556677",
    "vin": "<VIN>"
  }'
```

Сохранить: `CLIENT_ID`, `USER_ID`, `CLIENT_ACCESS`, `CLIENT_REFRESH`

### HTTP
- 200, access_token, refresh_token, client_id, user_id

### БД — clients (сразу)
```sql
SELECT id, user_id, email, full_name, phone FROM clients.clients WHERE id = '<CLIENT_ID>';
```

### БД — client_vehicles
```sql
SELECT client_id, vehicle_id, vin, make, model, year
FROM clients.client_vehicles WHERE client_id = '<CLIENT_ID>';
-- vehicle_id = VEHICLE_ID, vin = <VIN>
```

### БД — clientauth (async ≤15 с после Kafka)
```sql
SELECT id, email, full_name, password_hash <> '' AS has_hash
FROM clientauth.users WHERE id = '<USER_ID>';
-- 1 row, has_hash true (hash from Kafka event, not plain password)
```

### Redis
```bash
redis-cli GET "client-auth:refresh:<CLIENT_REFRESH>"
-- exists, user_id match
```

### Kafka
- Topic `client.registration.v1`, event `client.registered`, password_hash in payload (NOT plain)

---

## TC-CR-D002 — Register without VIN (P1)

### API
- omit vin → **400** «email, password, full_name, phone and vin required»

### БД
- COUNT clients/clientsauth без изменений

---

## TC-CR-D003 — Register unknown VIN (P1)

- vin = 'UNKNOWNVIN123' → **4xx** vehicle not found
- БД: 0 новых clients

---

## TC-CR-D004 — GET profile (P0)

```bash
curl -s -H "Authorization: Bearer $CLIENT_ACCESS" \
  "$CLIENT_PROTECTED/api/client/profile"
```

### Cross-check БД
```sql
SELECT c.*, (SELECT json_agg(cv) FROM clients.client_vehicles cv WHERE cv.client_id = c.id) AS vehicles
FROM clients.clients c WHERE c.id = '<CLIENT_ID>';
```
- Response vehicles[] match client_vehicles rows

---

## TC-CR-D005 — List / add vehicle (P1)

### List
```bash
curl -s -H "Authorization: Bearer $CLIENT_ACCESS" \
  "$CLIENT_PROTECTED/api/client/vehicles"
```

### Add second VIN
```bash
curl -s -X POST "$CLIENT_PROTECTED/api/client/vehicles" \
  -H "Authorization: Bearer $CLIENT_ACCESS" \
  -d '{"vin": "<VIN2>"}'
```

### БД
```sql
SELECT COUNT(*) FROM clients.client_vehicles WHERE client_id = '<CLIENT_ID>';
-- 2
```

---

## TC-CR-D006 — Duplicate client email (P1)

Повторный register с тем же email → **4xx**, COUNT clients = 1

---

## TC-CR-D007 — client_statistics event (P1)

После register, wait 15s:

```sql
SELECT user_id, email, vehicle_id FROM client_statistics.client_registration_events
WHERE user_id = '<USER_ID>';
-- 1 row
```

Cross-check API:
```bash
curl -s -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  "$EMPLOYEE_API/api/stats/client/overview"
```
- overview.registrations или аналогичная метрика выросла

---

## TC-CR-D008 — IssueTokens retry (P2, manual)

Если client-auth временно down при register:
- clients row может существовать
- API returns ErrAuthNotReady
- Проверить: нет orphan без retry policy (document actual behavior)
