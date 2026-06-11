# dealer-points — тест-кейсы

**Сервис:** `dealer-points-service`  
**REST:** `/api/dealer-points`, `/api/legal-entities`, `/api/warehouses`  
**gRPC:** `:50057`

## Dealer Points

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-DP-001 | P0 | GET /api/dealer-points | Bearer | — | 200, list | DP-001 |
| TC-DP-002 | P0 | POST /api/dealer-points | Bearer sales+ | name, address, … | 200, dealer_point | DP-002 |
| TC-DP-003 | P1 | GET /api/dealer-points/{id} | Bearer | — | 200 | DP-003 |
| TC-DP-004 | P1 | PUT /api/dealer-points/{id} | Bearer | — | 200 | manual |
| TC-DP-005 | P1 | DELETE /api/dealer-points/{id} | Bearer | без зависимостей | 200/204 | manual |

## Legal Entities

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-DP-010 | P0 | GET /api/legal-entities | Bearer | — | 200 | DP-010 |
| TC-DP-011 | P1 | POST /api/legal-entities | Bearer | inn, name, … | 200 | DP-011 |
| TC-DP-012 | P1 | POST /api/dealer-points/{dp_id}/legal-entities | Bearer | link body | 200 | manual |
| TC-DP-013 | P1 | GET /api/dealer-points/{dp_id}/legal-entities | Bearer | — | 200 | manual |
| TC-DP-014 | P1 | DELETE .../legal-entities/{le_id} | Bearer | — | 200 | manual |

## Warehouses

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-DP-020 | P0 | GET /api/warehouses | Bearer | — | 200 | DP-020 |
| TC-DP-021 | P1 | POST /api/warehouses | Bearer | dealer_point_id, name | 200 | DP-021 |
| TC-DP-022 | P1 | GET /api/warehouses/{id} | Bearer | — | 200 | DP-022 |
| TC-DP-023 | P1 | PUT /api/warehouses/{id} | Bearer | — | 200 | manual |
| TC-DP-024 | P1 | DELETE /api/warehouses/{id} | Bearer | пустой склад | 200 | manual |

## Зависимости

- parts, vehicles, workorders — WarehouseExists / DealerPointExists
