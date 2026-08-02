# client-protected-gateway-service — API

Защищённый HTTP-шлюз контура клиентов (порт 8093). Требуется JWT клиента. grpc-gateway reverse proxy на client-auth, client-registration (account) и client-reviews.

## Маршруты

| HTTP | Описание | Бэкенд |
|---|---|---|
| `GET /api/me` | Текущий профиль клиента (валидация токена) | client-auth-service |
| `POST /api/client/vehicles` | Привязать автомобиль по VIN | client-registration-service (account) |
| `GET /api/client/vehicles` | Мои автомобили | client-registration-service (account) |
| `GET /api/client/profile` | Профиль клиента | client-registration-service (account) |
| `GET /api/client/notifications` | Уведомления клиента | client-registration-service (account) |
| `POST /api/client/notifications/{id}/dismiss` | Скрыть уведомление | client-registration-service (account) |
| `POST /api/client/reviews` | Создать отзыв | client-reviews-service |
| `GET /api/client/reviews` | Мои отзывы | client-reviews-service |
| `GET /api/client/reviews/{id}` | Отзыв по id | client-reviews-service |
| `GET /api/client/review-invitations` | Приглашения на отзыв | client-reviews-service |
| `POST /api/client/review-invitations/{id}/dismiss` | Отклонить приглашение | client-reviews-service |
