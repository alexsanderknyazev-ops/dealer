# client-reviews-service — API

gRPC-сервис `reviews.v1.ReviewsService`. Вызывается из client-protected-gateway. Публикует событие `review.published.v1`.

## Endpoints

| gRPC | HTTP | Описание |
|---|---|---|
| `CreateReview` | `POST /api/client/reviews` | Создание отзыва |
| `ListMyReviews` | `GET /api/client/reviews` | Мои отзывы |
| `GetReview` | `GET /api/client/reviews/{id}` | Отзыв по id |
| `ListReviewInvitations` | `GET /api/client/review-invitations` | Приглашения на отзыв |
| `DismissReviewInvitation` | `POST /api/client/review-invitations/{id}/dismiss` | Скрыть/отклонить приглашение |

## Сообщения

### Review (модель)
`id`, `client_id`, `dealer_point_id`, `vehicle_id`, `rating`, `text`, `status`, `created_at`, `updated_at`

### CreateReview
Request: `vehicle_id`, `rating`, `text`
Response: `review`

### ListMyReviews
Request: пусто
Response: `reviews[]`

### GetReview
Request: `id`
Response: `review`

### ReviewInvitation (модель)
`id`, `client_id`, `vehicle_id`, `dealer_point_id`, `source_type`, `source_id`, `service_kind`, `status`, `created_at`

### ListReviewInvitations
Request: пусто
Response: `invitations[]`

### DismissReviewInvitation
Request: `id`
Response: пусто
