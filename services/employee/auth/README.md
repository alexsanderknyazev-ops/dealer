# auth-service

Сервис аутентификации и авторизации сотрудников дилера: регистрация, вход, выдача/обновление JWT-токенов, управление сессиями (refresh-токены в Redis). Дополнительно раздаёт SPA сотрудников (`frontend/auth`) и проксирует HTTP API на gateway-service и telemetry на errors-ingest.

## Стек

- Go 1.22, gRPC, JWT (golang-jwt/v5)
- PostgreSQL (pgx/v5) — схема `auth`
- Redis — refresh-токены сессий
- Kafka — публикация событий

## Порты

| Протокол | Порт |
|----------|------|
| gRPC | 50051 |
| HTTP (SPA + proxy + health/metrics) | 8080 |

## gRPC API

`auth.v1.AuthService`:
- `Register`, `Login`, `Refresh`, `Logout`, `Validate`, `RegisterClient`

## Взаимодействия

- Исходящие gRPC: —
- Kafka: публикация → `auth.events`
- Хранилища: PostgreSQL (`auth`), Redis

## Запуск

```bash
go run ./services/employee/auth   # make run-auth
```

Docker: `build/auth-service.Dockerfile`, compose-сервис `auth-service`, версия в `VERSION`.

## API

Полное описание всех эндпоинтов — см. [API.md](API.md).
