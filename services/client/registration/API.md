# client-registration-service — API

gRPC-сервис `clients.v1.ClientRegistrationPublicService`. Вызывается из client-public-gateway / client-protected-gateway.

## Endpoints

| gRPC | HTTP | Описание |
|---|---|---|
| `RegisterClient` | `POST /api/client/register` | Публичная регистрация клиента (email, пароль, VIN) |

При регистрации сервис проверяет VIN через vehicles-service, выпускает токены через client-auth (IssueTokens) и публикует событие `client.registration.v1`.

## Сообщения

### RegisterClient
Request: `email`, `password`, `full_name`, `phone`, `vin`
Response: `client_id`, `user_id`, `email`, `access_token`, `refresh_token`, `expires_at`
