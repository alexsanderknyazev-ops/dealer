# GET endpoints probe — probe-20260612-142645

| Field | Value |
|-------|-------|
| Base URL | `http://192.168.0.27:9080` |
| User | `qa.master@test.local` |
| Passed | 42 |
| Warn (HTML not JSON) | 4 |
| Failed | 0 |
| Skipped | 2 |

## Results

| ID | Method | Path | HTTP | Status | Body (truncated) |
|----|--------|------|------|--------|------------------|
| healthz | GET | `/healthz` | 200 | PASS | ok |
| me | GET | `/api/me` | 200 | PASS | {"user_id":"a1100001-0000-4000-8000-000000000003", "email":"qa.master@test.local", "valid":true} |
| customers-list | GET | `/api/customers?limit=20` | 200 | PASS | {"customers":[{"id":"a2200001-0000-4000-8000-000000000099", "name":"QA Client User", "email":"qa.client@test.local", "ph |
| vehicles-list | GET | `/api/vehicles?limit=20` | 200 | PASS | {"vehicles":[{"id":"a3300001-0000-4000-8000-000000000003", "vin":"QAVINDEAL001", "make":"Volkswagen", "model":"QA Polo", |
| deals-list | GET | `/api/deals?limit=20` | 200 | PASS | {"deals":[{"id":"a5500001-0000-4000-8000-000000000002", "customer_id":"a2200001-0000-4000-8000-000000000002", "vehicle_i |
| parts-list | GET | `/api/parts?limit=20` | 200 | PASS | {"parts":[{"id":"a4400001-0000-4000-8000-000000000002", "sku":"QA-PART-LOW", "name":"QA Low Stock Gasket", "category":"Р |
| parts-folders-list | GET | `/api/parts/folders?limit=20` | 200 | PASS | {"folders":[{"id":"50000000-0000-4000-8000-000000000003", "name":"Масла", "parent_id":"", "created_at":"1781268251", "up |
| brands-list | GET | `/api/brands?limit=20` | 200 | PASS | {"brands":[{"id":"40000000-0000-4000-8000-000000000003", "name":"Hyundai", "created_at":"1781268251", "updated_at":"1781 |
| brand-labor-rates | GET | `/api/brand-labor-rates` | 200 | PASS | {"brand_labor_rates":[{"id":"60000000-0000-4000-8000-000000000001", "brand_id":"40000000-0000-4000-8000-000000000001", " |
| dealer-points-list | GET | `/api/dealer-points?limit=20` | 200 | PASS | {"dealer_points":[{"id":"90000003-0000-4000-8000-000000000001", "name":"Volume Dealer Point 1", "address":"г. Тестград,  |
| legal-entities-list | GET | `/api/legal-entities?limit=20` | 200 | PASS | {"legal_entities":[{"id":"90000013-0000-4000-8000-000000000001", "name":"ООО Volume Auto 1", "inn":"7700000001", "addres |
| warehouses-list | GET | `/api/warehouses?limit=20` | 200 | PASS | {"warehouses":[{"id":"90000023-0000-4000-8000-000000000001", "dealer_point_id":"90000003-0000-4000-8000-000000000001", " |
| work-orders-list | GET | `/api/work-orders?limit=20` | 200 | PASS | {"work_orders":[{"id":"a6600001-0000-4000-8000-000000000001", "order_number":"WO-QA-0001", "customer_id":"a2200001-0000- |
| works-list | GET | `/api/works?limit=20` | 200 | PASS | {"works":[{"id":"90000008-0000-4000-8000-00000000000e", "code":"VOL-LAB-0014", "name":"Volume Work 14", "category":"Двиг |
| works-folders-list | GET | `/api/works/folders?limit=20` | 200 | PASS | {"folders":[{"id":"60000000-0000-4000-8000-000000000004", "name":"Двигатель", "parent_id":"", "created_at":"1781268251", |
| employees-list | GET | `/api/employees?limit=20` | 200 | PASS | {"employees":[{"id":"ac022b64-5cf6-431a-bce6-9fa1a0bd1442", "user_id":"a1100001-0000-4000-8000-000000000001", "full_name |
| reviews-list | GET | `/api/reviews?limit=20` | 200 | PASS | {"reviews":[{"id":"9000000e-0000-4000-8000-000000000001", "review_id":"9000000d-0000-4000-8000-000000000001", "client_id |
| movement-documents-list | GET | `/api/movement-documents?limit=20` | 200 | PASS | {"documents":[{"id":"ae0a141f-74ce-4732-9c6d-3ccda483e076", "document_number":"MOV-2026-00001", "status":"closed", "move |
| suppliers-list | GET | `/api/suppliers?limit=20` | 200 | WARN | <!DOCTYPE html> <html lang="ru">   <head>     <meta charset="UTF-8" />     <meta name="viewport" content="width=device-w |
| supplier-orders-list | GET | `/api/supplier-orders?limit=20` | 200 | WARN | <!DOCTYPE html> <html lang="ru">   <head>     <meta charset="UTF-8" />     <meta name="viewport" content="width=device-w |
| customer-orders-list | GET | `/api/customer-orders?limit=20` | 200 | WARN | <!DOCTYPE html> <html lang="ru">   <head>     <meta charset="UTF-8" />     <meta name="viewport" content="width=device-w |
| repair-appointments-list | GET | `/api/repair-appointments?limit=20` | 200 | PASS | {"appointments":[{"id":"6fe8e56d-9056-4679-8b44-9bb5bc206091", "appointment_number":"RA-2026-00001", "customer_id":"9000 |
| stats-employee | GET | `/api/stats/employee/overview` | 200 | PASS | {"customers_count":"0", "vehicles_count":"0", "deals_count":"80", "deals_by_stage":[{"stage":"completed", "count":"40"}, |
| stats-client | GET | `/api/stats/client/overview` | 200 | PASS | {"clients_count":"81", "client_vehicles_count":"81", "registered_users_count":"81", "reviews_count":"80", "average_ratin |
| reviews-stats | GET | `/api/reviews/stats` | 200 | PASS | {"total_count":"80", "average_rating":3, "by_status":[{"status":"published", "count":"80"}]} |
| client-reviews | GET | `/api/clients/a7700001-0000-4000-8000-000000000002/reviews` | 200 | PASS | {"reviews":[], "total":"0"} |
| brand-labor-resolve | GET | `/api/brand-labor-rates/resolve?brand_id=40000000-0000-4000-8000-000000000003` | 200 | PASS | {"warranty_hour_price":"", "commercial_hour_price":"", "hour_price":"", "found":false} |
| repair-slots | GET | `/api/repair-appointment-slots?dealer_point_id=90000003-0000-4000-8000-000000000001&limit=10` | 200 | WARN | <!DOCTYPE html> <html lang="ru">   <head>     <meta charset="UTF-8" />     <meta name="viewport" content="width=device-w |
| dp-legal | GET | `/api/dealer-points/90000003-0000-4000-8000-000000000001/legal-entities` | 200 | PASS | {"legal_entities":[], "total":1} |
| customers-by-id | GET | `/api/customers/a2200001-0000-4000-8000-000000000099` | 200 | PASS | {"id":"a2200001-0000-4000-8000-000000000099", "name":"QA Client User", "email":"qa.client@test.local", "phone":"+7900300 |
| vehicles-by-id | GET | `/api/vehicles/a3300001-0000-4000-8000-000000000003` | 200 | PASS | {"id":"a3300001-0000-4000-8000-000000000003", "vin":"QAVINDEAL001", "make":"Volkswagen", "model":"QA Polo", "year":2022, |
| deals-by-id | GET | `/api/deals/a5500001-0000-4000-8000-000000000002` | 200 | PASS | {"id":"a5500001-0000-4000-8000-000000000002", "customer_id":"a2200001-0000-4000-8000-000000000002", "vehicle_id":"a33000 |
| parts-by-id | GET | `/api/parts/a4400001-0000-4000-8000-000000000002` | 200 | PASS | {"id":"a4400001-0000-4000-8000-000000000002", "sku":"QA-PART-LOW", "name":"QA Low Stock Gasket", "category":"Расходники" |
| parts-stock | GET | `/api/parts/a4400001-0000-4000-8000-000000000002/stock` | 200 | PASS | {"stock":[{"warehouse_id":"30000000-0000-4000-8000-000000000002", "quantity":5}]} |
| parts-folders-by-id | GET | `/api/parts/folders/50000000-0000-4000-8000-000000000003` | 200 | PASS | {"id":"50000000-0000-4000-8000-000000000003", "name":"Масла", "parent_id":"", "created_at":"1781268251", "updated_at":"1 |
| brands-by-id | GET | `/api/brands/40000000-0000-4000-8000-000000000003` | 200 | PASS | {"id":"40000000-0000-4000-8000-000000000003", "name":"Hyundai", "created_at":"1781268251", "updated_at":"1781268251"} |
| dealer-points-by-id | GET | `/api/dealer-points/90000003-0000-4000-8000-000000000001` | 200 | PASS | {"id":"90000003-0000-4000-8000-000000000001", "name":"Volume Dealer Point 1", "address":"г. Тестград, ул. Volume, д. 1", |
| legal-entities-by-id | GET | `/api/legal-entities/90000013-0000-4000-8000-000000000001` | 200 | PASS | {"id":"90000013-0000-4000-8000-000000000001", "name":"ООО Volume Auto 1", "inn":"7700000001", "address":"г. Тестград, юр |
| warehouses-by-id | GET | `/api/warehouses/90000023-0000-4000-8000-000000000001` | 200 | PASS | {"id":"90000023-0000-4000-8000-000000000001", "dealer_point_id":"90000003-0000-4000-8000-000000000001", "legal_entity_id |
| work-orders-by-id | GET | `/api/work-orders/a6600001-0000-4000-8000-000000000001` | 200 | PASS | {"id":"a6600001-0000-4000-8000-000000000001", "order_number":"WO-QA-0001", "customer_id":"a2200001-0000-4000-8000-000000 |
| works-by-id | GET | `/api/works/90000008-0000-4000-8000-00000000000e` | 200 | PASS | {"id":"90000008-0000-4000-8000-00000000000e", "code":"VOL-LAB-0014", "name":"Volume Work 14", "category":"Двигатель", "l |
| works-folders-by-id | GET | `/api/works/folders/60000000-0000-4000-8000-000000000004` | 200 | PASS | {"id":"60000000-0000-4000-8000-000000000004", "name":"Двигатель", "parent_id":"", "created_at":"1781268251", "updated_at |
| employees-by-id | GET | `/api/employees/ac022b64-5cf6-431a-bce6-9fa1a0bd1442` | 200 | PASS | {"id":"ac022b64-5cf6-431a-bce6-9fa1a0bd1442", "user_id":"a1100001-0000-4000-8000-000000000001", "full_name":"QA Admin",  |
| reviews-by-id | GET | `/api/reviews/9000000e-0000-4000-8000-000000000001` | 200 | PASS | {"id":"9000000e-0000-4000-8000-000000000001", "review_id":"9000000d-0000-4000-8000-000000000001", "client_id":"9000000c- |
| movement-documents-by-id | GET | `/api/movement-documents/ae0a141f-74ce-4732-9c6d-3ccda483e076` | 200 | PASS | {"id":"ae0a141f-74ce-4732-9c6d-3ccda483e076", "document_number":"MOV-2026-00001", "status":"closed", "movement_type":"wo |
| supplier-orders-by-id | GET | `—` | 0 | SKIP | no id from list |
| customer-orders-by-id | GET | `—` | 0 | SKIP | no id from list |
| repair-appointments-by-id | GET | `/api/repair-appointments/6fe8e56d-9056-4679-8b44-9bb5bc206091` | 200 | PASS | {"id":"6fe8e56d-9056-4679-8b44-9bb5bc206091", "appointment_number":"RA-2026-00001", "customer_id":"90000005-0000-4000-80 |
