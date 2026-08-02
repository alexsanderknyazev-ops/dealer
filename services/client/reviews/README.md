# client-reviews-service

Сервис контура клиентов: отзывы клиентов — создание, мои отзывы, приглашения на отзыв, просмотр. Публикует событие `review.published.v1`.

## Стек

- Go 1.22, gRPC, JWT (golang-jwt/v5)
- PostgreSQL (pgx/v5) — схемы `reviews`, `clients`
- Kafka — публикация событий

## Порты

| Протокол | Порт |
|----------|------|
| gRPC | 50060 |
| HTTP (health/metrics) | 8089 |

## gRPC API

`reviews.v1.ReviewsService`:
- `CreateReview`, `ListMyReviews`, `GetReview`
- `ListReviewInvitations`, `DismissReviewInvitation`

## Взаимодействия

- Исходящие gRPC: vehicles
- Kafka: публикация → `review.published.v1`
- Хранилища: PostgreSQL (`reviews`, `clients`)

## Запуск

```bash
go run ./services/client/reviews   # make run-client-reviews
```

Docker: `build/client-reviews-service.Dockerfile`, compose-сервис `client-reviews-service`, версия в `VERSION`.

## API

Полное описание всех эндпоинтов — см. [API.md](API.md).
