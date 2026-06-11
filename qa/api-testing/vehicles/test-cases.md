# vehicles — тест-кейсы

**Сервис:** `vehicles-service`  
**REST:** `/api/vehicles`  
**gRPC:** `:50053`

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-VEH-001 | P0 | POST /api/vehicles | Bearer sales+ | vin уникальный, brand_id, model, year, price, status=available | 200, vehicle | VEH-001 |
| TC-VEH-002 | P0 | GET /api/vehicles/{id} | Bearer | — | 200 | VEH-002 |
| TC-VEH-003 | P0 | GET /api/vehicles | Bearer | status=available | 200, list | VEH-003 |
| TC-VEH-004 | P1 | PUT /api/vehicles/{id} | Bearer | status=sold | 200 | VEH-004 |
| TC-VEH-005 | P1 | DELETE /api/vehicles/{id} | Bearer | — | 200/204 | VEH-005 |
| TC-VEH-006 | P1 | POST /api/vehicles | Bearer | Дубликат VIN | 409/500 unique constraint | VEH-006 |
| TC-VEH-007 | P1 | POST /api/vehicles | Bearer | Несуществующий brand_id | 4xx brand not found | VEH-007 |
| TC-VEH-008 | P1 | GET /api/vehicles/{id} | — | Без JWT | 401 | VEH-008 |
| TC-VEH-009 | P0 | gRPC GetVehicleByVIN | internal/public | vin существующего авто | 200 (client-registration) | VEH-009 |
| TC-VEH-010 | P1 | gRPC GetVehicleByVIN | — | Несуществующий vin | NotFound | manual |
| TC-VEH-011 | P2 | PUT /api/vehicles/{id} | Bearer | dealer_point_id валидный | 200, FK check via dealer-points gRPC | manual |

## Зависимости

- **brands-service** — BrandExists при create/update
- **dealer-points-service** — DealerPointExists
- **client-registration** — GetVehicleByVIN (public gRPC method)
- **client-reviews**, **employee-reviews** — GetVehicle enrich
