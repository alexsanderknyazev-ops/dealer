# employee-reviews — детальные тест-кейсы

**Read-only API** — данные приходят из Kafka (`review.published.v1`), не создаются через HTTP.

**Схема:** `employee_reviews.reviews`

---

## TC-EREV-D001 — List all reviews (P0)

```bash
curl -s -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  "$EMPLOYEE_API/api/reviews?limit=20&offset=0"
```

```sql
SELECT COUNT(*) FROM employee_reviews.reviews;
-- response.total должен совпадать (если нет фильтров)
```

---

## TC-EREV-D002 — Filter by status (P1)

```bash
curl -s -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  "$EMPLOYEE_API/api/reviews?status=published"
```

```sql
SELECT COUNT(*) FROM employee_reviews.reviews WHERE status = 'published';
```

---

## TC-EREV-D003 — Reviews by client (P0)

После client create review (см. client-reviews TC-REV-D001):

```bash
curl -s -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  "$EMPLOYEE_API/api/clients/$CLIENT_ID/reviews"
```

```sql
SELECT review_id, rating, client_email, vehicle_vin
FROM employee_reviews.reviews WHERE client_id = '<CLIENT_ID>';
```
- Количество rows = len(response.reviews)

---

## TC-EREV-D004 — Stats endpoint (P0)

```bash
curl -s -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  "$EMPLOYEE_API/api/reviews/stats"
```

```sql
SELECT COUNT(*) AS total, ROUND(AVG(rating)::numeric, 2) AS avg_rating
FROM employee_reviews.reviews;
```

---

## TC-EREV-D005 — Kafka enrichment fields (P1)

После async replicate:

```sql
SELECT vehicle_vin, vehicle_make, vehicle_model, vehicle_year
FROM employee_reviews.reviews WHERE review_id = '<REVIEW_ID>';
```

Сверить с `vehicles.vehicles` для того же vehicle_id — VIN/make/model/year consistent.

---

## TC-EREV-D006 — No write API (P2)

Employee gateway **не** exposes POST /api/reviews — только GET.  
Попытка POST → **404/405**.

```sql
-- новый review только через client-reviews → Kafka
```
