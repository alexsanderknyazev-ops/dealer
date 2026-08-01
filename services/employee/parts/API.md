# parts-service — API

gRPC-сервис `parts.v1.PartsService`. Доступ защищён JWT.

## Endpoints

### Запчасти

| gRPC | HTTP | Описание |
|---|---|---|
| `CreatePart` | `POST /api/parts` | Создание запчасти |
| `GetPart` | `GET /api/parts/{id}` | Запчасть по id |
| `ListParts` | `GET /api/parts` | Список запчастей (фильтры/поиск) |
| `ListPartStock` | `GET /api/parts/{part_id}/stock` | Остатки запчасти по складам |
| `UpdatePart` | `PUT /api/parts/{id}` | Обновление запчасти (частичное) |
| `DeletePart` | `DELETE /api/parts/{id}` | Удаление запчасти |

### Папки запчастей

| gRPC | HTTP | Описание |
|---|---|---|
| `CreateFolder` | `POST /api/parts/folders` | Создание папки |
| `GetFolder` | `GET /api/parts/folders/{id}` | Папка по id |
| `ListFolders` | `GET /api/parts/folders` | Список папок |
| `UpdateFolder` | `PUT /api/parts/folders/{id}` | Обновление папки |
| `DeleteFolder` | `DELETE /api/parts/folders/{id}` | Удаление папки |

### Документы перемещения

| gRPC | HTTP | Описание |
|---|---|---|
| `CreateMovementDocument` | `POST /api/movement-documents` | Создание документа перемещения |
| `GetMovementDocument` | `GET /api/movement-documents/{id}` | Документ по id |
| `UpdateMovementDocument` | `PUT /api/movement-documents/{id}` | Обновление документа |
| `ListMovementDocuments` | `GET /api/movement-documents` | Список документов |
| `StartMovementDocument` | `POST /api/movement-documents/{id}/start` | Начать перемещение |
| `CloseMovementDocument` | `POST /api/movement-documents/{id}/close` | Закрыть документ (списание со склада) |
| `ConfirmMovementDocument` | `POST /api/movement-documents/{id}/confirm` | Устаревший alias закрытия |
| `CancelMovementDocument` | `POST /api/movement-documents/{id}/cancel` | Отмена документа |
| `CreateProductionExtraction` | `POST /api/movement-documents/{id}/create-production-extraction` | Производственное списание |

### Поставщики и закупки

| gRPC | HTTP | Описание |
|---|---|---|
| `ListSuppliers` | `GET /api/suppliers` | Список поставщиков |
| `CreateSupplierOrder` | `POST /api/supplier-orders` | Создание заказа поставщику |
| `GetSupplierOrder` | `GET /api/supplier-orders/{id}` | Заказ поставщику по id |
| `UpdateSupplierOrder` | `PUT /api/supplier-orders/{id}` | Обновление заказа поставщику |
| `ListSupplierOrders` | `GET /api/supplier-orders` | Список заказов поставщикам |
| `CancelSupplierOrder` | `POST /api/supplier-orders/{id}/cancel` | Отмена заказа поставщику |
| `CreateReceiptFromSupplierOrder` | `POST /api/supplier-orders/{id}/create-receipt` | Приёмка (поступление) по заказу |

### Клиентские заказы и продажи

| gRPC | HTTP | Описание |
|---|---|---|
| `CreateCustomerOrder` | `POST /api/customer-orders` | Создание клиентского заказа |
| `GetCustomerOrder` | `GET /api/customer-orders/{id}` | Клиентский заказ по id |
| `UpdateCustomerOrder` | `PUT /api/customer-orders/{id}` | Обновление клиентского заказа |
| `ListCustomerOrders` | `GET /api/customer-orders` | Список клиентских заказов |
| `CancelCustomerOrder` | `POST /api/customer-orders/{id}/cancel` | Отмена клиентского заказа |
| `CreateSaleFromCustomerOrder` | `POST /api/customer-orders/{id}/create-sale` | Продажа по клиентскому заказу |

### Связка с СТО (заказ-наряды)

| gRPC | HTTP | Описание |
|---|---|---|
| `CreateWorkOrderFromSupplierOrder` | `POST /api/supplier-orders/{id}/create-work-order` | Создать заказ-наряд из заказа поставщику |
| `CreateWorkOrderFromCustomerOrder` | `POST /api/customer-orders/{id}/create-work-order` | Создать заказ-наряд из клиентского заказа |
| `FulfillOrderFromWorkOrder` | — (внутренний gRPC) | Закрытие заказа после заказ-наряда (вызывается workorders-service) |

## Сообщения

### Part (модель)
`id`, `sku`, `name`, `category`, `folder_id`, `quantity`, `unit` (шт/комплект/л/кг), `price` (decimal как строка), `location`, `notes`, `brand_id`, `dealer_point_id`, `legal_entity_id`, `warehouse_id`, `created_at`, `updated_at`

### CreatePart
Request: `sku`, `name`, `category`, `folder_id`, `quantity`, `unit`, `price`, `location`, `notes`, `brand_id?`, `dealer_point_id`, `legal_entity_id`, `warehouse_id`
Response: `part`

### ListParts
Request: `limit`, `offset`, `search` (sku/name), `category`, `folder_id`, `brand_id`, `dealer_point_id`, `legal_entity_id`, `warehouse_id`
Response: `parts[]`, `total`

### ListPartStock
Request: `part_id`
Response: `stock[]` (`warehouse_id`, `quantity`)

### UpdatePart
Request: `id`, optional все поля модели
Response: `part`

### PartFolder / папки
`id`, `name`, `parent_id` (пусто = корень). Запросы как у works-service.

### MovementDocument (модель)
`id`, `document_number`, `status`, `movement_type`, `reference_type`, `reference_id`, `notes`, `created_by`, `confirmed_by`, `created_at`, `confirmed_at`, `updated_at`, `lines[]`, `created_by_name`, `confirmed_by_name`, `reference_label`, `customer_name`, `vehicle_vin`, `vehicle_label`, `parent_document_id`, `parent_document_number`, `customer_id`, `vehicle_id`, `supplier_id`, `supplier_name`, `receipt_warehouse_id`, `receipt_warehouse_name`

### MovementDocumentLine
`id`, `part_id`, `warehouse_id`, `quantity`, `reference_line_id`, `notes`, `sort_order`, `part_name`, `part_sku`, `warehouse_name`, `destination_warehouse_id`, `destination_warehouse_name`, `source_stock_quantity`, `unit_cost`

### CreateMovementDocument
Request: `movement_type`, `reference_type`, `reference_id`, `notes`, `created_by`, `lines[]`, `customer_id`, `vehicle_id`, `vehicle_vin`, `supplier_id`, `receipt_warehouse_id`
Response: `document`

### ListMovementDocuments
Request: `limit`, `offset`, `status`, `reference_type`, `reference_id`
Response: `documents[]`, `total`

### Start/Close/Confirm/Cancel/ProductionExtraction
Request: `id` (+ `closed_by`/`confirmed_by`/`cancelled_by`/`created_by`)
Response: `document`

### Supplier (модель)
`id`, `name`, `inn`, `phone`, `email`, `notes`, `created_at`, `updated_at`

### ListSuppliers
Request: `limit`, `offset`, `search`
Response: `suppliers[]`, `total`

### SupplierOrder (модель)
`id`, `order_number`, `status`, `supplier_id`, `supplier_name`, `receipt_warehouse_id`, `receipt_warehouse_name`, `fulfillment_movement_document_id`, `fulfillment_movement_document_number`, `fulfillment_work_order_id`, `fulfillment_work_order_number`, `customer_order_id`, `customer_order_number`, `notes`, `created_by`, `created_by_name`, `created_at`, `updated_at`, `lines[]`

### CreateSupplierOrder
Request: `supplier_id`, `receipt_warehouse_id`, `notes`, `created_by`, `lines[]`, `customer_order_id`
Response: `order`

### ListSupplierOrders
Request: `limit`, `offset`, `status`
Response: `orders[]`, `total`

### CancelSupplierOrder
Request: `id`
Response: `order`

### CreateReceiptFromSupplierOrder
Request: `id`, `created_by`
Response: `document` (поступление)

### CustomerOrder (модель)
`id`, `order_number`, `status`, `customer_id`, `customer_name`, `vehicle_id`, `vehicle_vin`, `vehicle_label`, `issue_warehouse_id`, `issue_warehouse_name`, `fulfillment_movement_document_id`, `fulfillment_movement_document_number`, `fulfillment_work_order_id`, `fulfillment_work_order_number`, `notes`, `created_by`, `created_by_name`, `created_at`, `updated_at`, `lines[]`

### CreateCustomerOrder
Request: `customer_id`, `vehicle_id`, `vehicle_vin`, `issue_warehouse_id`, `notes`, `created_by`, `lines[]`
Response: `order`

### ListCustomerOrders
Request: `limit`, `offset`, `status`
Response: `orders[]`, `total`

### CreateSaleFromCustomerOrder
Request: `id`, `created_by`
Response: `document`

### CreateWorkOrderFrom{Supplier,Customer}Order
Request: `id`, `customer_id`/`vehicle_id`/`vehicle_vin`, `notes`
Response: `work_order_id`, `work_order_number`

### FulfillOrderFromWorkOrder
Request: `source_order_type`, `source_order_id`
Response: пусто
