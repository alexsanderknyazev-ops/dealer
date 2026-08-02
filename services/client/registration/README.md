# client-registration-service

Сервис контура клиентов: публичная регистрация клиента — проверка VIN через vehicles-service, выпуск токенов через client-auth, создание записи и публикация события `client.registration.v1`.

## Стек

- Go 1.22, gRPC, JWT (golang-jwt/v5)
- PostgreSQL (pgx/v5) — схема `clients`
- Kafka — публикация событий

## Порты

| Протокол | Порт |
|----------|------|
| gRPC | 50058 |
| HTTP (health/metrics) | 8087 |

## gRPC API

`clients.v1.ClientRegistrationPublicService`:
- `RegisterClient`

## Взаимодействия

- Исходящие gRPC: vehicles (по VIN), client-auth (IssueTokens)
- Kafka: публикация → `client.registration.v1`
- Хранилища: PostgreSQL (`clients`)

## Запуск

```bash
go run ./services/client/registration   # make run-client-registration
```

Docker: `build/client-registration-service.Dockerfile`, compose-сервис `client-registration-service`, версия в `VERSION`.

## API

Полное описание всех эндпоинтов — см. [API.md](API.md).
