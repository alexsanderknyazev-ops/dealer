# deals-service — API

gRPC-сервис `deals.v1.DealsService`. Доступ защищён JWT. При завершении сделки публикуется событие `deal.completed.v1`.

## Endpoints

| gRPC | HTTP | Описание |
|---|---|---|
| `CreateDeal` | `POST /api/deals` | Создание сделки |
| `GetDeal` | `GET /api/deals/{id}` | Сделка по id |
| `ListDeals` | `GET /api/deals` | Список сделок (фильтры) |
| `UpdateDeal` | `PUT /api/deals/{id}` | Обновление сделки (частичное) |
| `DeleteDeal` | `DELETE /api/deals/{id}` | Удаление сделки |

## Сообщения

### Deal (модель)
`id`, `customer_id`, `vehicle_id`, `amount` (decimal как строка), `stage` (`draft`/`in_progress`/`paid`/`completed`/`cancelled`), `assigned_to` (user_id из auth), `notes`, `created_at`, `updated_at`

### CreateDeal
Request: `customer_id`, `vehicle_id`, `amount`, `stage`, `assigned_to`, `notes`
Response: `deal`

### GetDeal
Request: `id`
Response: `deal`

### ListDeals
Request: `limit`, `offset`, `stage`, `customer_id`
Response: `deals[]`, `total`

### UpdateDeal
Request: `id`, optional `customer_id`, `vehicle_id`, `amount`, `stage`, `assigned_to`, `notes`
Response: `deal`

### DeleteDeal
Request: `id`
Response: пусто
