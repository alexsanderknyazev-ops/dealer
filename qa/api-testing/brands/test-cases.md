# brands — тест-кейсы

**Сервис:** `brands-service`  
**REST:** `/api/brands`  
**gRPC:** `:50056`

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-BRD-001 | P0 | GET /api/brands | Bearer | — | 200, brands[], total≥1 | BRD-001 |
| TC-BRD-002 | P0 | GET /api/brands/{id} | Bearer | id из list | 200, brand | BRD-002 |
| TC-BRD-003 | P1 | POST /api/brands | Bearer sales+ | name уникальный | 200, brand.id | BRD-003 |
| TC-BRD-004 | P1 | PUT /api/brands/{id} | Bearer | name | 200 | BRD-004 |
| TC-BRD-005 | P1 | DELETE /api/brands/{id} | Bearer | brand без vehicles | 200/204 | manual |
| TC-BRD-006 | P1 | POST /api/brands | Bearer | Дубликат name | 500/409 unique idx_brands_name | BRD-006 |
| TC-BRD-007 | P1 | GET /api/brands/{id} | Bearer | random uuid | 404 | BRD-007 |
| TC-BRD-008 | P1 | POST /api/brands | — | Без auth | 401 | BRD-008 |

## Потребители gRPC

- vehicles-service, parts-service — проверка brand_id
