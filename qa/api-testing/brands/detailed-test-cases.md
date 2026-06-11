# brands — детальные тест-кейсы (API + БД)

**Схема:** `brands.brands`  
**Unique:** `idx_brands_name` on LOWER(TRIM(name))

---

## TC-BRD-D001 — List + Get (P0)

```bash
curl -s -H "Authorization: Bearer $EMPLOYEE_TOKEN" "$EMPLOYEE_API/api/brands?limit=5"
```

```sql
SELECT COUNT(*) FROM brands.brands;
-- total in response
```

```bash
curl -s -H "Authorization: Bearer $EMPLOYEE_TOKEN" "$EMPLOYEE_API/api/brands/<BRAND_ID>"
```

---

## TC-BRD-D002 — Create unique brand (P1)

### До
```sql
SELECT COUNT(*) FROM brands.brands WHERE LOWER(TRIM(name)) = LOWER('QA Brand X');
```

### API
```bash
curl -s -X POST "$EMPLOYEE_API/api/brands" \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -d '{"name": "QA Brand X"}'
```

### БД
```sql
SELECT id, name, created_at FROM brands.brands WHERE name = 'QA Brand X';
```

---

## TC-BRD-D003 — Duplicate name (P1)

POST name='Hyundai' (seed exists) → **500/409**

```sql
SELECT COUNT(*) FROM brands.brands WHERE name ILIKE 'hyundai';
-- 1
```

---

## TC-BRD-D004 — Update (P2)

PUT name → проверить `updated_at` изменился

---

## TC-BRD-D005 — Delete unused brand (P2)

DELETE brand без vehicles/parts refs:

```sql
SELECT COUNT(*) FROM vehicles.vehicles WHERE brand_id = '<BRAND_ID>';
SELECT COUNT(*) FROM parts.parts WHERE brand_id = '<BRAND_ID>';
-- both 0 before delete
```

После DELETE — row отсутствует

---

## TC-BRD-D006 — FK from vehicles (P1)

Create vehicle с brand_id → vehicles.vehicles.brand_id = BRAND_ID (см. vehicles TC-VEH-D001)
