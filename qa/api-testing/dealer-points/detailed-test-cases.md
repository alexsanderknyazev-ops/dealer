# dealer-points — детальные тест-кейсы (API + БД)

**Схемы:** `dealerpoints.dealer_points`, `legal_entities`, `dealer_point_legal_entities`, `warehouses`

---

## TC-DP-D001 — Create dealer point (P0)

### API
```bash
curl -s -X POST "$EMPLOYEE_API/api/dealer-points" \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -d '{"name": "QA Point North", "address": "ул. Тестовая 1"}'
```

### БД
```sql
SELECT id, name, address FROM dealerpoints.dealer_points WHERE name = 'QA Point North';
```

---

## TC-DP-D002 — Legal entity + link (P1)

### Create LE
```bash
curl -s -X POST "$EMPLOYEE_API/api/legal-entities" \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -d '{"name": "ООО QA Auto", "inn": "500100732259", "address": "Москва"}'
```

### БД
```sql
SELECT id, inn FROM dealerpoints.legal_entities WHERE inn = '500100732259';
-- unique idx_legal_entities_inn
```

### Link
```bash
curl -s -X POST "$EMPLOYEE_API/api/dealer-points/$DP_ID/legal-entities" \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -d '{"legal_entity_id": "<LE_ID>"}'
```

### БД M2M
```sql
SELECT * FROM dealerpoints.dealer_point_legal_entities
WHERE dealer_point_id = '<DP_ID>' AND legal_entity_id = '<LE_ID>';
-- 1 row
```

---

## TC-DP-D003 — Create warehouse parts (P0)

```bash
curl -s -X POST "$EMPLOYEE_API/api/warehouses" \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -d '{
    "dealer_point_id": "<DP_ID>",
    "legal_entity_id": "<LE_ID>",
    "type": "parts",
    "name": "Склад запчастей QA"
  }'
```

### БД
```sql
SELECT id, dealer_point_id, legal_entity_id, type, name
FROM dealerpoints.warehouses WHERE id = '<WH_ID>';
-- type = 'parts'
```

Используется в: parts.part_stock, workorders.warehouse_id, vehicles.warehouse_id

---

## TC-DP-D004 — List legal entities by dealer point (P1)

```bash
curl -s -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  "$EMPLOYEE_API/api/dealer-points/$DP_ID/legal-entities"
```

```sql
SELECT le.* FROM dealerpoints.legal_entities le
JOIN dealerpoints.dealer_point_legal_entities dple ON dple.legal_entity_id = le.id
WHERE dple.dealer_point_id = '<DP_ID>';
```

---

## TC-DP-D005 — Unlink legal entity (P2)

DELETE link → row removed from `dealer_point_legal_entities`, LE и DP остаются

---

## TC-DP-D006 — Warehouse type cars (P2)

type='cars' — для vehicles warehouse binding

```sql
SELECT type FROM dealerpoints.warehouses WHERE id = '<WH_ID>';
```

---

## TC-DP-D007 — Delete dealer point cascade (P2)

DELETE DP с warehouses → CASCADE warehouses (FK ON DELETE CASCADE)

```sql
SELECT COUNT(*) FROM dealerpoints.warehouses WHERE dealer_point_id = '<DP_ID>';
-- 0 after delete
```
