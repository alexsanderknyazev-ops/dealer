# client-public-gateway-service

Публичный HTTP-шлюз контура клиентов: вход и регистрация клиентов (без JWT). Проксирует запросы по gRPC на client-auth и client-registration.

## Стек

- Go 1.22, HTTP (net/http), gRPC-клиенты

## Порты

| Протокол | Порт |
|----------|------|
| HTTP | 8091 |

## Проксируемые сервисы

- client-auth (Login/Refresh/Logout)
- client-registration (RegisterClient)

## Взаимодействия

- Исходящие gRPC: client-auth, client-registration
- Хранилища: —

## Запуск

```bash
go run ./services/gateway/client-public   # make run-client-public-gateway
```

Docker: `build/client-public-gateway-service.Dockerfile`, compose-сервис `client-public-gateway-service`, версия в `VERSION`.

## API

Полное описание всех эндпоинтов — см. [API.md](API.md).
