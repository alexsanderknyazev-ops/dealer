# client-statistics-service

Сервис контура клиентов: статистика для клиента (overview). Считает агрегаты из событий регистраций и отзывов через Kafka. Доступ защищён JWT.

## Стек

- Go 1.22, gRPC, JWT (golang-jwt/v5)
- PostgreSQL (pgx/v5) — схема `client_statistics`
- Kafka — потребление событий

## Порты

| Протокол | Порт |
|----------|------|
| gRPC | 50062 |
| HTTP (health/metrics) | 8095 |

## gRPC API

`statistics.client.v1.ClientStatisticsService`:
- `GetOverview`

## Взаимодействия

- Исходящие gRPC: —
- Kafka: потребление ← `review.published.v1`, `client.registration.v1`
- Хранилища: PostgreSQL (`client_statistics`)

## Запуск

```bash
go run ./services/statistics/client   # make run-client-statistics
```

Docker: `build/client-statistics-service.Dockerfile`, compose-сервис `client-statistics-service`, версия в `VERSION`.

## API

Полное описание всех эндпоинтов — см. [API.md](API.md).
