# client-auth — детальные тест-кейсы (API + Redis + Kafka consumer)

**Public gw:** login, refresh, logout (:8091)  
**Protected gw:** GET /api/me (:8093)  
**Схема:** `clientauth.users`  
**Redis:** `client-auth:refresh:{token}`

---

## TC-CA-D001 — Login after register (P0)

### API
```bash
curl -s -X POST "$CLIENT_PUBLIC/api/login" \
  -d '{"email":"client-qa@test.local","password":"Test1234!"}'
```

### БД
- Без новых rows (login не создаёт user)

### Redis
- Новый refresh key

### JWT
- role = **client** (не sales)

---

## TC-CA-D002 — GET /api/me (P0)

```bash
curl -s -H "Authorization: Bearer $CLIENT_ACCESS" "$CLIENT_PROTECTED/api/me"
```

- user_id = clientauth.users.id
- valid = true

---

## TC-CA-D003 — Kafka consumer creates user (P0)

Сценарий: register **до** того как client-auth consumer обработал event:

1. Register → 200 с tokens (IssueTokens sync)
2. ИЛИ simulate: только Kafka event без sync path

Проверка consumer path:
```sql
-- после register подождать 5s если login сразу не работал
SELECT id, email FROM clientauth.users WHERE email = '<EMAIL>';
```

Password в БД = **hash из Kafka** (`password_hash` field event), не plain text:

```sql
SELECT password_hash FROM clientauth.users WHERE email = '<EMAIL>';
-- bcrypt/argon hash, NOT 'Test1234!'
```

---

## TC-CA-D004 — Wrong password (P1)

Login → 401, Redis без нового key

---

## TC-CA-D005 — Employee token on /api/me (P1)

```bash
curl -s -w "\nHTTP:%{http_code}\n" \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  "$CLIENT_PROTECTED/api/me"
```
- **401/403**

---

## TC-CA-D006 — Refresh + logout cycle (P1)

Same as employee-auth TC-AUTH-D004/D005 but prefix `client-auth:refresh:`

---

## TC-CA-D007 — gRPC IssueTokens (internal, P2)

Вызывается из client-registration после Kafka publish.
Проверка: register response содержит tokens до истечения retry window.
