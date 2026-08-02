# workorders-service

Сервис контура сотрудников: заказ-наряды СТО — работы, запчасти, движение запасов по заказ-нарядам, применение документов перемещения. Доступ защищён JWT.

## Стек

- Go 1.22, gRPC, JWT (golang-jwt/v5)
- PostgreSQL (pgx/v5) — схема `workorders`

## Порты

| Протокол | Порт |
|----------|------|
| gRPC | 50064 |
| HTTP (health/metrics) | 8097 |

## gRPC API

`workorders.v1.WorkOrdersService`:
- `CreateWorkOrder`, `GetWorkOrder`, `ListWorkOrders`, `UpdateWorkOrder`, `DeleteWorkOrder`
- `MovePartsToWork`, `ApplyMovementDocument`

## Взаимодействия

- Исходящие gRPC: customers, vehicles, dealer-points, parts, works, employees
- Kafka: —
- Хранилища: PostgreSQL (`workorders`)

## Запуск

```bash
go run ./services/employee/workorders   # make run-workorders
```

Docker: `build/workorders-service.Dockerfile`, compose-сервис `workorders-service`, версия в `VERSION`.

## API

Полное описание всех эндпоинтов — см. [API.md](API.md).
