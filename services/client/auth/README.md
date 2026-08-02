# client-auth-service

Сервис контура клиентов: аутентификация клиентов — вход, refresh, выход, выпуск токенов (для внутренних сервисов), валидация сессий. Сессии и refresh-токены в Redis. Потребляет событие `client.registration.v1`.

## Стек

- Go 1.22, gRPC, JWT (golang-jwt/v5)
- PostgreSQL (pgx/v5) — схема `clientauth`
- Redis — сессии/refresh-токены
- Kafka — потребление событий

## Порты

| Протокол | Порт |
|----------|------|
| gRPC | 50059 |
| HTTP (health/metrics) | 8088 |

## gRPC API

- `clientauth.v1.ClientAuthPublicService`: `Login`, `Refresh`, `Logout`
- `clientauth.v1.ClientAuthSessionService`: `Validate`
- `clientauth.v1.ClientAuthService`: `IssueTokens`

## Взаимодействия

- Исходящие gRPC: —
- Kafka: потребление ← `client.registration.v1`
- Хранилища: PostgreSQL (`clientauth`), Redis

## Запуск

```bash
go run ./services/client/auth   # make run-client-auth
```

Docker: `build/client-auth-service.Dockerfile`, compose-сервис `client-auth-service`, версия в `VERSION`.

## API

Полное описание всех эндпоинтов — см. [API.md](API.md).
