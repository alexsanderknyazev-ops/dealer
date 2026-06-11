# Redis и Kafka — проверки side-effects

## Redis (refresh tokens)

| Контур | Key prefix | Пример |
|--------|------------|--------|
| Employee auth | `auth:refresh:` | `auth:refresh:<refresh_token_uuid>` |
| Client auth | `client-auth:refresh:` | `client-auth:refresh:<refresh_token_uuid>` |

```bash
redis-cli -h 127.0.0.1 -p 6379

# после login — ключ должен существовать
GET auth:refresh:<REFRESH_TOKEN>

# после logout — ключ удалён
EXISTS auth:refresh:<REFRESH_TOKEN>   # → 0

# TTL (секунды)
TTL auth:refresh:<REFRESH_TOKEN>
```

**Ожидания:**
- Login/Register → `SET` key, JSON `{"user_id":"...","email":"..."}`
- Refresh → старый key `DEL`, новый `SET`
- Logout → `DEL`
- Refresh с revoked token → API 401, key отсутствует

---

## Kafka topics

| Topic | Producer | Consumer(s) | Задержка |
|-------|----------|-------------|----------|
| `client.registration.v1` | client-registration | client-auth, client-statistics | 1–10 с |
| `review.published.v1` | client-reviews | employee-reviews, client-statistics | 1–10 с |
| `deal.completed.v1` | deals | employee-statistics | 1–10 с |
| `auth.events` | employee auth | *(нет consumer в коде)* | — |
| `platform.errors.v1` | все сервисы | errors-ingest | 1–5 с |

### Чтение последних сообщений (локально)

```bash
# если kafka в docker
docker exec -it dealer-kafka-1 kafka-console-consumer \
  --bootstrap-server localhost:29092 \
  --topic client.registration.v1 \
  --from-beginning --max-messages 1

# или kcat
kcat -C -b 127.0.0.1:9092 -t review.published.v1 -o -5 -e
```

### Проверка через БД (надёжнее для QA)

После Kafka event дождаться **≤15 с**, затем SQL:

```sql
-- client registered
SELECT * FROM clientauth.users WHERE email = '<EMAIL>';
SELECT * FROM client_statistics.client_registration_events WHERE email = '<EMAIL>';

-- review published
SELECT * FROM employee_reviews.reviews WHERE review_id = '<REVIEW_ID>';
SELECT * FROM client_statistics.review_events WHERE review_id = '<REVIEW_ID>';

-- deal completed
SELECT * FROM employee_statistics.deal_events WHERE deal_id = '<DEAL_ID>';
```

---

## JWT (access token)

Декодировать payload (без verify для QA):

```bash
echo '<ACCESS_TOKEN>' | cut -d. -f2 | base64 -d 2>/dev/null | python3 -m json.tool
```

Проверить: `user_id`, `email`, `role`, `exp`.
