# employee-reviews-service

Сервис контура сотрудников: отзывы клиентов с точки зрения дилера — просмотр, статистика, выборка по клиенту. Потребляет событие `review.published.v1`. Доступ защищён JWT.

## Стек

- Go 1.22, gRPC, JWT (golang-jwt/v5)
- PostgreSQL (pgx/v5) — схема `employee_reviews`
- Kafka — потребление событий

## Порты

| Протокол | Порт |
|----------|------|
| gRPC | 50063 |
| HTTP (health/metrics) | 8096 |

## gRPC API

`reviews.v1.EmployeeReviewsService`:
- `ListReviewsByClient`, `ListReviews`, `GetEmployeeReview`, `GetReviewStats`

## Взаимодействия

- Исходящие gRPC: vehicles
- Kafka: потребление ← `review.published.v1`
- Хранилища: PostgreSQL (`employee_reviews`)

## Запуск

```bash
go run ./services/employee/reviews   # make run-employee-reviews
```

Docker: `build/employee-reviews-service.Dockerfile`, compose-сервис `employee-reviews-service`, версия в `VERSION`.

## API

Полное описание всех эндпоинтов — см. [API.md](API.md).
