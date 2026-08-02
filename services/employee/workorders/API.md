# workorders-service — API

gRPC-сервис `workorders.v1.WorkOrdersService`. Доступ защищён JWT.

## Endpoints

| gRPC | HTTP | Описание |
|---|---|---|
| `CreateWorkOrder` | `POST /api/work-orders` | Создание заказ-наряда |
| `GetWorkOrder` | `GET /api/work-orders/{id}` | Заказ-наряд по id |
| `ListWorkOrders` | `GET /api/work-orders` | Список заказ-нарядов (фильтры) |
| `UpdateWorkOrder` | `PUT /api/work-orders/{id}` | Обновление заказ-наряда (частичное) |
| `DeleteWorkOrder` | `DELETE /api/work-orders/{id}` | Удаление заказ-наряда |
| `MovePartsToWork` | `POST /api/work-orders/{id}/move-parts-to-work` | Передача запчастей в работу (выдача) |
| `ApplyMovementDocument` | — (внутренний gRPC) | Применение документа перемещения (вызывается parts-service) |

## Сообщения

### WorkOrder (модель)
`id`, `order_number`, `customer_id`, `vehicle_id`, `dealer_point_id`, `warehouse_id`, `repair_type`, `status`, `service_advisor_id`, `complaint`, `diagnosis`, `mileage_km`, `labor_cost`, `parts_cost`, `total_cost`, `opened_at`, `closed_at`, `parts_issued`, `parts_issued_at`, `notes`, `created_at`, `updated_at`, `labor[]`, `parts[]`, `movement_document_id`, `movement_document_status`, `service_advisor_name`, `customer_name`, `vehicle_vin`, `vehicle_label`, `source_order_type`, `source_order_id`

### WorkOrderLabor (позиция работы)
`id`, `work_id`, `description`, `quantity`, `unit_price`, `amount`, `executor_id`, `sort_order`, `executor_name`, `work_code`, `work_name`, `labor_hours`

### WorkOrderPart (позиция запчасти)
`id`, `part_id`, `warehouse_id`, `description`, `quantity`, `unit_price`, `amount`, `issued`, `sort_order`, `part_sku`, `warehouse_name`, `part_name`

### CreateWorkOrder
Request: `customer_id`, `vehicle_id`, `dealer_point_id`, `warehouse_id`, `repair_type`, `status`, `service_advisor_id`, `complaint`, `diagnosis`, `mileage_km`, `opened_at`, `notes`, `labor[]` (WorkOrderLaborInput), `parts[]` (WorkOrderPartInput), `source_order_type`, `source_order_id`
Response: `work_order`

### ListWorkOrders
Request: `limit`, `offset`, `status`, `repair_type`, `customer_id`, `vehicle_id`
Response: `work_orders[]`, `total`

### UpdateWorkOrder
Request: `id`, optional поля + `labor[]`, `parts[]`
Response: `work_order`

### MovePartsToWork
Request: `id`, `issued_by`
Response: `work_order`

### ApplyMovementDocument
Request: `work_order_id`, `movement_document_id`, `movement_document_status`
Response: `work_order`
