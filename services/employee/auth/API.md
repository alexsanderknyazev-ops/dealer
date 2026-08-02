# auth-service — API

gRPC-сервис `auth.v1.AuthService`. Методы без HTTP-маппинга доступны только по gRPC (service-to-service).

## Endpoints

| gRPC | HTTP | Описание |
|---|---|---|
| `Register` | `POST /api/register` | Регистрация сотрудника (роль по умолчанию) |
| `Login` | `POST /api/login` | Вход, выдача access + refresh токенов |
| `Refresh` | `POST /api/refresh` | Обновление токенов по refresh-токену |
| `Logout` | `POST /api/logout` | Выход, инвалидация refresh-токена |
| `Validate` | `GET /api/me` | Проверка access-токена, возврат профиля |
| `RegisterClient` | — (внутренний gRPC) | Регистрация владельца авто (role=client) для контура клиентов |

## Сообщения

### Register / RegisterClient
Request: `email`, `password`, `name`, `phone`
Response: `user_id`, `email`, `access_token`, `refresh_token`, `expires_at`

### Login
Request: `email`, `password`
Response: `user_id`, `email`, `access_token`, `refresh_token`, `expires_at`

### Refresh
Request: `refresh_token`
Response: `access_token`, `refresh_token`, `expires_at`

### Logout
Request: `refresh_token`
Response: пусто

### Validate
Request: `access_token`
Response: `user_id`, `email`, `valid`
