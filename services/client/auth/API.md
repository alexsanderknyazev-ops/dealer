# client-auth-service — API

gRPC-сервисы пакета `clientauth.v1`. Вызываются из client-public-gateway / client-protected-gateway.

## Endpoints

### ClientAuthPublicService (публичная авторизация)

| gRPC | HTTP | Описание |
|---|---|---|
| `Login` | `POST /api/login` | Вход клиента, выдача токенов |
| `Refresh` | `POST /api/refresh` | Обновление токенов |
| `Logout` | `POST /api/logout` | Выход, инвалидация refresh-токена |

### ClientAuthSessionService (сессия)

| gRPC | HTTP | Описание |
|---|---|---|
| `Validate` | `GET /api/me` | Проверка access-токена, возврат профиля |

### ClientAuthService (внутренний, без HTTP)

| gRPC | Описание |
|---|---|
| `IssueTokens` | Выдача токенов после регистрации (для client-registration-service) |

## Сообщения

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

### IssueTokens
Request: `user_id`
Response: `access_token`, `refresh_token`, `expires_at`
