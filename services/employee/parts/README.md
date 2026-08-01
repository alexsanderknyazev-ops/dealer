# parts-service

Сервис контура сотрудников: склад запасных частей — артикулы (SKU), категории, количество, единицы, цены, расположение; папки; документы перемещения; производственное списание; поставщики и закупки; клиентские заказы и продажи; связка с заказ-нарядами. Доступ защищён JWT.

## Стек

- Go 1.22, gRPC, JWT (golang-jwt/v5)
- PostgreSQL (pgx/v5) — схема `parts`

## Порты

| Протокол | Порт |
|----------|------|
| gRPC | 50055 |
| HTTP (health/metrics) | 8084 |

## gRPC API

`parts.v1.PartsService`:
- Запчасти: `CreatePart`, `GetPart`, `ListParts`, `ListPartStock`, `UpdatePart`, `DeletePart`
- Папки: `CreateFolder`, `GetFolder`, `ListFolders`, `UpdateFolder`, `DeleteFolder`
- Движения: `CreateMovementDocument`, `GetMovementDocument`, `UpdateMovementDocument`, `ListMovementDocuments`, `StartMovementDocument`, `CloseMovementDocument`, `ConfirmMovementDocument`, `CancelMovementDocument`
- Производство: `CreateProductionExtraction`
- Поставщики: `ListSuppliers`, `CreateSupplierOrder`, `GetSupplierOrder`, `UpdateSupplierOrder`, `ListSupplierOrders`, `CancelSupplierOrder`, `CreateReceiptFromSupplierOrder`
- Клиентские заказы: `CreateCustomerOrder`, `GetCustomerOrder`, `UpdateCustomerOrder`, `ListCustomerOrders`, `CancelCustomerOrder`, `CreateSaleFromCustomerOrder`
- Связка с СТО: `CreateWorkOrderFromSupplierOrder`, `CreateWorkOrderFromCustomerOrder`, `FulfillOrderFromWorkOrder`

## Взаимодействия

- Исходящие gRPC: brands, dealer-points, workorders, employees
- Kafka: —
- Хранилища: PostgreSQL (`parts`)

## Запуск

```bash
go run ./services/employee/parts   # make run-parts
```

Docker: `build/parts-service.Dockerfile`, compose-сервис `parts-service`, версия в `VERSION`.

## API

Полное описание всех эндпоинтов — см. [API.md](API.md).
