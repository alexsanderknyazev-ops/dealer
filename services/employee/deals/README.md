# deals-service

Сервис контура сотрудников: сделки — клиент, автомобиль, сумма, этап (draft → in_progress → paid → completed), ответственный. При завершении сделки публикует событие `deal.completed.v1`. Доступ защищён JWT.

## Стек

- Go 1.22, gRPC, JWT (golang-jwt/v5)
- PostgreSQL (pgx/v5) — схема `deals`
- Kafka — публикация событий

## Порты

| Протокол | Порт |
|----------|------|
| gRPC | 50054 |
| HTTP (health/metrics) | 8083 |

## gRPC API

`deals.v1.DealsService`:
- `CreateDeal`, `GetDeal`, `ListDeals`, `UpdateDeal`, `DeleteDeal`

## Взаимодействия

- Исходящие gRPC: customers, vehicles
- Kafka: публикация → `deal.completed.v1`
- Хранилища: PostgreSQL (`deals`)

## Запуск

```bash
go run ./services/employee/deals   # make run-deals
```

Docker: `build/deals-service.Dockerfile`, compose-сервис `deals-service`, версия в `VERSION`.

## API

Полное описание всех эндпоинтов — см. [API.md](API.md).
