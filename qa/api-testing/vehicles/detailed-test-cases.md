# vehicles — детальные тест-кейсы (API + БД)

**REST:** `/api/vehicles`  
**Схема:** `vehicles.vehicles`  
**gRPC-only:** `GetVehicleByVIN` (для client-registration)

---

## TC-VEH-D001 — Create vehicle (P0)

### Preconditions
```sql
SELECT id, name FROM brands.brands LIMIT 1;
-- BRAND_ID
```

### До
```sql
SELECT COUNT(*) FROM vehicles.vehicles WHERE vin = 'VINQA001TEST';
-- 0
```

### API
```bash
curl -s -X POST "$EMPLOYEE_API/api/vehicles" \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "vin": "VINQA001TEST",
    "brand_id": "<BRAND_ID>",
    "model": "Solaris",
    "year": 2023,
    "mileage_km": 15000,
    "price": "1850000",
    "status": "available",
    "color": "white"
  }'
```

Сохранить: `VEHICLE_ID`

### БД
```sql
SELECT id, vin, model, year, mileage_km, price, status, brand_id, color
FROM vehicles.vehicles WHERE id = '<VEHICLE_ID>';
```
- vin = VINQA001TEST (unique index `idx_vehicles_vin`)
- status IN ('available','sold','reserved')

---

## TC-VEH-D002 — Duplicate VIN (P1)

### API
- Повторный POST с тем же vin → **409/500**

### БД
```sql
SELECT COUNT(*) FROM vehicles.vehicles WHERE vin = 'VINQA001TEST';
-- 1 (не 2)
```

---

## TC-VEH-D003 — Update status sold (P1)

### API
```bash
curl -s -X PUT "$EMPLOYEE_API/api/vehicles/$VEHICLE_ID" \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"status": "sold"}'
```

### БД
```sql
SELECT status, updated_at FROM vehicles.vehicles WHERE id = '<VEHICLE_ID>';
-- status = 'sold'
```

---

## TC-VEH-D004 — Invalid brand_id (P1)

### API
- brand_id = `00000000-0000-4000-8000-000000000099` → **4xx**

### БД
- Новых строк нет

---

## TC-VEH-D005 — GetVehicleByVIN (gRPC, P0)

Вызывается из **client-registration** при register. Косвенная проверка:

1. Create vehicle с VIN
2. Client register с этим VIN → 200

### БД после client register
```sql
SELECT cv.vin, cv.vehicle_id, v.id
FROM clients.client_vehicles cv
JOIN vehicles.vehicles v ON v.id = cv.vehicle_id
WHERE cv.vin = 'VINQA001TEST';
-- vehicle_id = VEHICLE_ID
```

---

## TC-VEH-D006 — List filters (P2)

```bash
curl -s -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  "$EMPLOYEE_API/api/vehicles?status=available&limit=10"
```

### БД
```sql
SELECT COUNT(*) FROM vehicles.vehicles WHERE status = 'available';
-- total в response ≈ count (если нет других фильтров)
```

---

## TC-VEH-D007 — Delete (P1)

### БД до
```sql
SELECT COUNT(*) FROM vehicles.vehicles WHERE id = '<VEHICLE_ID>';
```

### API + БД после
- Row удалена, COUNT -1
- Проверить: client_vehicles с этим vehicle_id (orphan policy)

---

## TC-VEH-D008 — dealer_point / warehouse binding (P2)

При указании `dealer_point_id`, `warehouse_id`:

### БД
```sql
SELECT dealer_point_id, warehouse_id FROM vehicles.vehicles WHERE id = '<ID>';
```

### Cross-service
- warehouse должен существовать в `dealerpoints.warehouses` (gRPC check)
