# parts — тест-кейсы

**Сервис:** `parts-service`  
**REST:** `/api/parts`, `/api/parts/folders`, `/api/movement-documents`  
**gRPC:** `:50055`

## Parts CRUD

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-PRT-001 | P0 | GET /api/parts | Bearer | — | 200, parts[] | PRT-001 |
| TC-PRT-002 | P0 | POST /api/parts | Bearer **parts role** | sku, name, quantity, unit, price | 200, part | PRT-002 |
| TC-PRT-003 | P1 | POST /api/parts | Bearer **sales** | те же поля | **403** RBAC | PRT-003 |
| TC-PRT-004 | P1 | GET /api/parts/{id} | Bearer | — | 200 | PRT-004 |
| TC-PRT-005 | P1 | PUT /api/parts/{id} | Bearer parts role | price | 200 | manual |
| TC-PRT-006 | P1 | DELETE /api/parts/{id} | Bearer parts role | — | 200/204 | manual |
| TC-PRT-007 | P1 | POST /api/parts | Bearer parts | invalid brand_id | 4xx | manual |

## Folders

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-PRT-010 | P1 | GET /api/parts/folders | Bearer | — | 200 | PRT-010 |
| TC-PRT-011 | P1 | POST /api/parts/folders | Bearer parts | name, parent_id | 200 | manual |
| TC-PRT-012 | P1 | GET /api/parts/folders/{id} | Bearer | — | 200/404 | manual |
| TC-PRT-013 | P1 | PUT /api/parts/folders/{id} | Bearer parts | — | 200 | manual |
| TC-PRT-014 | P1 | DELETE /api/parts/folders/{id} | Bearer parts | пустая папка | 200 | manual |

## Movement Documents

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-PRT-020 | P0 | GET /api/movement-documents | Bearer | — | 200 | PRT-020 |
| TC-PRT-021 | P1 | POST /api/movement-documents | Bearer parts | movement_type, lines[] | 200, status=draft | manual |
| TC-PRT-022 | P0 | GET /api/movement-documents/{id} | Bearer | — | 200 | manual |
| TC-PRT-023 | P0 | POST .../confirm | Bearer parts | draft doc, достаточный stock | 200, status=confirmed, stock↓ | manual |
| TC-PRT-024 | P1 | POST .../confirm | Bearer parts | insufficient stock | 4xx insufficient stock | manual |
| TC-PRT-025 | P1 | POST .../cancel | Bearer parts | draft doc | 200, status=cancelled | manual |
| TC-PRT-026 | P1 | POST .../confirm | Bearer parts | reference_type=work_order | gRPC ApplyMovementDocument → workorders | manual |
| TC-PRT-027 | P2 | POST .../confirm | Bearer parts | повторный confirm | 4xx not draft | manual |

## gRPC (internal)

| ID | P | Method | Caller | Expected |
|----|---|--------|--------|----------|
| TC-PRT-030 | P0 | CreateMovementDocument | workorders MovePartsToWork | draft doc, reference=work_order |

## Write roles

admin, manager, parts_manager, storekeeper, master, service_advisor
