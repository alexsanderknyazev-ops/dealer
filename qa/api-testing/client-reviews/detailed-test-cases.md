# client-reviews + employee-reviews — детальные тест-кейсы

**Client API:** `POST/GET /api/client/reviews` (:8093)  
**Employee API:** `GET /api/reviews`, `/api/reviews/stats` (:8090)  
**Схемы:** `reviews.reviews`, `employee_reviews.reviews`, `client_statistics.review_events`  
**Kafka:** `review.published.v1`

---

## TC-REV-D001 — Create review (P0)

### Preconditions
- CLIENT_TOKEN, VEHICLE_ID привязан к client через `clients.client_vehicles`

### До
```sql
SELECT COUNT(*) FROM reviews.reviews WHERE client_id = '<CLIENT_ID>';
```

### API
```bash
curl -s -X POST "$CLIENT_PROTECTED/api/client/reviews" \
  -H "Authorization: Bearer $CLIENT_ACCESS" \
  -H 'Content-Type: application/json' \
  -d '{
    "vehicle_id": "<VEHICLE_ID>",
    "rating": 5,
    "text": "Отличный сервис, всё понравилось"
  }'
```

Сохранить: `REVIEW_ID`

### БД — reviews (client DB)
```sql
SELECT id, client_id, user_id, vehicle_id, dealer_point_id, rating, text, status
FROM reviews.reviews WHERE id = '<REVIEW_ID>';
```
- rating BETWEEN 1 AND 5
- status = 'published' (default)
- dealer_point_id заполнен (из vehicle via gRPC GetVehicle)

### Unique constraint
```sql
-- idx_reviews_client_vehicle: один отзыв на client+vehicle
SELECT COUNT(*) FROM reviews.reviews WHERE client_id = '<CLIENT_ID>' AND vehicle_id = '<VEHICLE_ID>';
-- 1
```

---

## TC-REV-D002 — Duplicate review same vehicle (P1)

Второй POST с тем же vehicle_id → **4xx** (unique violation)

### БД
```sql
SELECT COUNT(*) FROM reviews.reviews WHERE client_id = '<CLIENT_ID>' AND vehicle_id = '<VEHICLE_ID>';
-- still 1
```

---

## TC-REV-D003 — List my reviews (P0)

```bash
curl -s -H "Authorization: Bearer $CLIENT_ACCESS" \
  "$CLIENT_PROTECTED/api/client/reviews"
```

Count в response = COUNT в БД для client_id

---

## TC-REV-D004 — Kafka → employee_reviews (P0)

Подождать ≤15 с после create:

### БД
```sql
SELECT review_id, client_email, client_full_name, rating, text, status,
       vehicle_vin, vehicle_make, vehicle_model
FROM employee_reviews.reviews WHERE review_id = '<REVIEW_ID>';
-- 1 row, vehicle_* enriched from vehicles-service
```

### Employee API
```bash
curl -s -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  "$EMPLOYEE_API/api/reviews?limit=20"
```
- review_id присутствует в list

```bash
curl -s -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  "$EMPLOYEE_API/api/clients/$CLIENT_ID/reviews"
```

---

## TC-REV-D005 — client_statistics (P1)

```sql
SELECT review_id, rating, status FROM client_statistics.review_events
WHERE review_id = '<REVIEW_ID>';
```

```bash
curl -s -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  "$EMPLOYEE_API/api/stats/client/overview"
```

---

## TC-REV-D006 — GET /api/reviews/stats (P1)

```bash
curl -s -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  "$EMPLOYEE_API/api/reviews/stats"
```

Cross-check:
```sql
SELECT COUNT(*), AVG(rating) FROM employee_reviews.reviews;
```

---

## TC-REV-D007 — Invalid rating (P1)

rating=0 или 6 → **400**, БД без новых rows

---

## TC-REV-D008 — Review чужого vehicle (P1)

vehicle_id не из client_vehicles → **4xx**

```sql
SELECT COUNT(*) FROM reviews.reviews WHERE client_id = '<CLIENT_ID>';
-- unchanged
```

---

## TC-REV-D009 — No auth (P1)

GET/POST без token → **401**
