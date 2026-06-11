# integration — сквозные E2E-потоки

Комбинации нескольких сервисов. Все через HTTP unless noted.

## INT-001: Employee onboarding → CRUD customer → deal (P0)

| Step | Action | Expected |
|------|--------|----------|
| 1 | POST /api/register (8090) | tokens, role=sales |
| 2 | POST /api/customers | customer.id |
| 3 | GET /api/brands → brand_id | 200 |
| 4 | POST /api/vehicles | vehicle.id |
| 5 | POST /api/deals {customer, vehicle, amount, stage=draft} | deal.id |
| 6 | PUT /api/deals/{id} stage=completed | 200, Kafka deal.completed |
| 7 | GET /api/stats/employee/overview | metrics updated (async) |

**Auto:** INT-001

---

## INT-002: Client registration full path (P0)

| Step | Action | Expected |
|------|--------|----------|
| 1 | Employee: create vehicle with VIN | vehicle in stock |
| 2 | POST /api/client/register (8091) with vin | 200, client tokens |
| 3 | GET /api/client/profile (8093) | profile + vehicle |
| 4 | GET /api/stats/client/overview (8090 employee token) | registration counted |

**Auto:** INT-002

---

## INT-003: Client review → employee visibility (P0)

| Step | Action | Expected |
|------|--------|----------|
| 1 | Client login/register | client token |
| 2 | POST /api/client/reviews | review id |
| 3 | Wait ≤10s Kafka | — |
| 4 | GET /api/reviews (8090 employee) | review in list |
| 5 | GET /api/reviews/stats | count ≥1 |

**Auto:** partial (INT-003)

---

## INT-004: Work order + parts movement (P0)

**Precondition:** user with parts/WO write role (admin/master)

| Step | Action | Expected |
|------|--------|----------|
| 1 | Create customer, vehicle, dealer_point, warehouse, part+stock | refs OK |
| 2 | POST /api/work-orders with parts line | WO draft |
| 3 | POST /api/work-orders/{id}/move-parts-to-work | movement_document_id |
| 4 | GET /api/movement-documents/{id} | status=draft |
| 5 | POST .../confirm | status=confirmed, stock reduced |
| 6 | GET /api/work-orders/{id} | parts_issued=true |

**Auto:** manual (RBAC)

---

## INT-005: Reference integrity deals (P1)

| Step | Action | Expected |
|------|--------|----------|
| 1 | POST /api/deals fake customer uuid | 4xx |
| 2 | POST /api/deals fake vehicle uuid | 4xx |

**Auto:** DEL-007, DEL-008

---

## INT-006: RBAC sales vs parts (P1)

| Step | Action | Expected |
|------|--------|----------|
| 1 | Register new employee (sales) | — |
| 2 | POST /api/parts | 403 |
| 3 | POST /api/customers | 200 |
| 4 | POST /api/work-orders | 403 |

**Auto:** PRT-003, WO-003

---

## INT-007: Token lifecycle (P1)

| Step | Action | Expected |
|------|--------|----------|
| 1 | login | tokens |
| 2 | refresh | new access |
| 3 | logout | — |
| 4 | refresh again | 401 |

**Auto:** AUTH-004, AUTH-005

---

## INT-008: Cross-gateway isolation (P1)

| Step | Action | Expected |
|------|--------|----------|
| 1 | Employee token on 8093 /api/client/profile | 401/403 |
| 2 | Client token on 8090 /api/customers | 401/403 |

**Auto:** CPP-005, manual
