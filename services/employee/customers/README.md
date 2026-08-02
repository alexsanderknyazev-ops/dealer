# customers-service

Сервис контура сотрудников: CRUD клиентов дилера (физические и юридические лица), список, поиск. Доступ защищён JWT (роли admin/manager/sales).

## Стек

- Go 1.22, gRPC, JWT (golang-jwt/v5)
- PostgreSQL (pgx/v5) — схема `customers`

## Порты

| Протокол | Порт |
|----------|------|
| gRPC | 50052 |
| HTTP (health/metrics) | 8081 |

## gRPC API

`customers.v1.CustomersService`:
- `CreateCustomer`, `GetCustomer`, `ListCustomers`, `UpdateCustomer`, `DeleteCustomer`

## Взаимодействия

- Исходящие gRPC: —
- Kafka: —
- Хранилища: PostgreSQL (`customers`)

## Запуск

```bash
go run ./services/employee/customers   # make run-customers
```

Docker: `build/customers-service.Dockerfile`, compose-сервис `customers-service`, версия в `VERSION`.

## API

Полное описание всех эндпоинтов — см. [API.md](API.md).
