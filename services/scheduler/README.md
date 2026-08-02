# scheduler-service

Инфраструктурный сервис: фоновый воркер, выполняющий периодические задачи напрямую через SQL по нескольким схемам (клиенты, клиентские отзывы, сделки, заказ-наряды, автомобили, приглашения на отзывы и т.д.). Не имеет собственных gRPC-методов.

## Стек

- Go 1.22, фоновый worker
- PostgreSQL (pgx/v5) — кросс-схемные запросы
- Kafka — публикация ошибок

## Порты

| Протокол | Порт |
|----------|------|
| HTTP (health/metrics) | 8100 |

## Затрагиваемые схемы

- `clients`, `customers`, `deals`, `reviews`, `vehicles`, `workorders`

## Взаимодействия

- Исходящие gRPC: —
- Kafka: публикация → `platform.errors.v1` (при ошибках)
- Хранилища: PostgreSQL (cross-schema JOIN)

## Запуск

```bash
go run ./services/scheduler   # make run-scheduler
```

Docker: `build/scheduler-service.Dockerfile`, compose-сервис `scheduler-service`, версия в `VERSION`.

## API

Полное описание всех эндпоинтов — см. [API.md](API.md).
