# client-statistics-service — API

gRPC-сервис `statistics.client.v1.ClientStatisticsService`. Доступ защищён JWT. Потребляет события `review.published.v1` и `client.registration.v1`.

## Endpoints

| gRPC | HTTP | Описание |
|---|---|---|
| `GetOverview` | `GET /api/stats/client/overview` | Сводная статистика по клиентской зоне |

## Сообщения

### GetOverview
Request: пусто

Response: `overview` (ClientOverview):
- `clients_count` — клиентов
- `client_vehicles_count` — автомобилей клиентов
- `registered_users_count` — зарегистрированных пользователей
- `reviews_count` — отзывов
- `average_rating` — средняя оценка
- `reviews_by_status[]` — `{status, count}`
