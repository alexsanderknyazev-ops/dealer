# employee-statistics-service

Сервис контура сотрудников: статистика продаж/сделок (overview). Считает агрегаты из событий сделок через Kafka. Доступ защищён JWT.

## Стек

- Go 1.22, gRPC, JWT (golang-jwt/v5)
- PostgreSQL (pgx/v5) — схема `employee_statistics`
- Kafka — потребление событий

## Порты

| Протокол | Порт |
|----------|------|
| gRPC | 50061 |
| HTTP (health/metrics) | 8094 |

## gRPC API

`statistics.employee.v1.EmployeeStatisticsService`:
- `GetOverview`

## Взаимодействия

- Исходящие gRPC: —
- Kafka: потребление ← `deal.completed.v1`
- Хранилища: PostgreSQL (`employee_statistics`)

## Запуск

```bash
go run ./services/statistics/employee   # make run-employee-statistics
```

Docker: `build/employee-statistics-service.Dockerfile`, compose-сервис `employee-statistics-service`, версия в `VERSION`.

## API

Полное описание всех эндпоинтов — см. [API.md](API.md).
