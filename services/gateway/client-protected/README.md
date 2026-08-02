# client-protected-gateway-service

Защищённый HTTP-шлюз контура клиентов: профиль, автомобили, отзывы, уведомления (требуется JWT клиента). Проксирует запросы по gRPC на client-auth, client-registration и client-reviews.

## Стек

- Go 1.22, HTTP (net/http), gRPC-клиенты, JWT (golang-jwt/v5)

## Порты

| Протокол | Порт |
|----------|------|
| HTTP | 8093 |

## Проксируемые сервисы

- client-auth (Validate, Refresh, Logout)
- client-registration
- client-reviews

## Взаимодействия

- Исходящие gRPC: client-auth, client-registration, client-reviews
- Хранилища: —

## Запуск

```bash
go run ./services/gateway/client-protected   # make run-client-protected-gateway
```

Docker: `build/client-protected-gateway-service.Dockerfile`, compose-сервис `client-protected-gateway-service`, версия в `VERSION`.

## API

Полное описание всех эндпоинтов — см. [API.md](API.md).
