# employee-auth — детальные тест-кейсы (API + БД + Redis)

**Base URL:** `http://127.0.0.1:8090` или `:8080`  
**Схема БД:** `auth.users`  
**Redis:** `auth:refresh:{token}`

---

## TC-AUTH-D001 — Register employee (P0)

### До
```sql
SELECT COUNT(*) AS before_cnt FROM auth.users WHERE email = 'qa-auth-d001@test.local';
-- ожидание: 0
```

### API
```bash
curl -s -w "\nHTTP:%{http_code}\n" -X POST "$EMPLOYEE_API/api/register" \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "qa-auth-d001@test.local",
    "password": "Test1234!",
    "name": "Auth QA",
    "phone": "+79001112233"
  }'
```

### HTTP
- `200`, body: `user_id`, `access_token`, `refresh_token`, `expires_at`

Сохранить: `USER_ID`, `ACCESS`, `REFRESH`

### БД
```sql
SELECT id, email, name, phone, role, password_hash <> '' AS has_hash
FROM auth.users WHERE id = '<USER_ID>';
```
| Поле | Ожидание |
|------|----------|
| email | qa-auth-d001@test.local |
| name | Auth QA |
| role | sales |
| has_hash | true |

### Redis
```bash
redis-cli GET "auth:refresh:<REFRESH>"
```
- JSON с `user_id` = USER_ID
- `TTL` > 0

### JWT
- decode access_token → `role` = sales

---

## TC-AUTH-D002 — Login (P0)

### API
```bash
curl -s -X POST "$EMPLOYEE_API/api/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"qa-auth-d001@test.local","password":"Test1234!"}'
```

### БД
- Новых строк в `auth.users` **нет** (COUNT не меняется)

### Redis
- Новый ключ `auth:refresh:{new_refresh}` (старый от register может оставаться до TTL)

---

## TC-AUTH-D003 — GET /api/me (P0)

```bash
curl -s -H "Authorization: Bearer $ACCESS" "$EMPLOYEE_API/api/me"
```

### HTTP
- `200`, `user_id` = USER_ID, `valid: true`

### БД
- Без изменений

---

## TC-AUTH-D004 — Refresh rotation (P1)

### До
```bash
redis-cli EXISTS "auth:refresh:$OLD_REFRESH"   # 1
```

### API
```bash
curl -s -X POST "$EMPLOYEE_API/api/refresh" \
  -H 'Content-Type: application/json' \
  -d "{\"refresh_token\":\"$OLD_REFRESH\"}"
```

### HTTP
- `200`, новый `access_token`, новый `refresh_token`

### Redis
- `EXISTS auth:refresh:$OLD_REFRESH` → **0**
- `EXISTS auth:refresh:$NEW_REFRESH` → **1**

---

## TC-AUTH-D005 — Logout (P1)

### API
```bash
curl -s -X POST "$EMPLOYEE_API/api/logout" \
  -H 'Content-Type: application/json' \
  -d "{\"refresh_token\":\"$REFRESH\"}"
```

### Redis
- Key удалён

### API после logout
```bash
curl -s -X POST "$EMPLOYEE_API/api/refresh" -d "{\"refresh_token\":\"$REFRESH\"}"
```
- **401**

---

## TC-AUTH-D006 — Duplicate email (P1)

### До/После
```sql
SELECT COUNT(*) FROM auth.users WHERE email = 'qa-auth-d001@test.local';
-- delta: 0
```

### API
- Повторный register с тем же email → **400/409/500**, без новой строки

---

## TC-AUTH-D007 — Wrong password (P1)

- Login → **401/403**
- `auth.users` без изменений
- Redis: новый refresh **не** создаётся

---

## TC-AUTH-D008 — No auth on /api/me (P1)

```bash
curl -s -w "\nHTTP:%{http_code}\n" "$EMPLOYEE_API/api/me"
```
- **401**, БД без изменений

---

## TC-AUTH-D009 — Proxy через :8080 (P1)

```bash
curl -s -H "Authorization: Bearer $ACCESS" "http://127.0.0.1:8080/api/customers"
```
- **200** (reverse proxy → gateway)
- БД customers без изменений (read-only)

---

## TC-AUTH-D010 — Kafka auth.events (P2, manual)

После register проверить topic `auth.events` (consumer отсутствует — только наличие сообщения):

```json
{"event":"user.registered","user_id":"...","email":"..."}
```

БД: без дополнительных таблиц.
