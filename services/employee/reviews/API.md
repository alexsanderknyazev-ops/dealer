# employee-reviews-service — API

gRPC-сервис `reviews.v1.EmployeeReviewsService`. Доступ защищён JWT. Потребляет событие `review.published.v1`.

## Endpoints

| gRPC | HTTP | Описание |
|---|---|---|
| `ListReviewsByClient` | `GET /api/clients/{client_id}/reviews` | Отзывы по клиенту |
| `ListReviews` | `GET /api/reviews` | Список отзывов (фильтры) |
| `GetEmployeeReview` | `GET /api/reviews/{id}` | Отзыв по id |
| `GetReviewStats` | `GET /api/reviews/stats` | Статистика отзывов |

## Сообщения

### EmployeeReview (модель)
`id`, `review_id`, `client_id`, `user_id`, `client_email`, `client_full_name`, `dealer_point_id`, `vehicle_id`, `vehicle_vin`, `vehicle_make`, `vehicle_model`, `vehicle_year`, `rating`, `text`, `status`, `occurred_at`, `created_at`

### ListReviewsByClient
Request: `client_id`, `limit`, `offset`
Response: `reviews[]`, `total`

### ListReviews
Request: `limit`, `offset`, `client_id`, `dealer_point_id`, `status`
Response: `reviews[]`, `total`

### GetEmployeeReview
Request: `id`
Response: `review`

### GetReviewStats
Request: `client_id`, `dealer_point_id`
Response: `total_count`, `average_rating`, `by_status[]` (`status`, `count`)
