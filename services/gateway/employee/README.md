# gateway-service

HTTP-шлюз контура сотрудников: принимает HTTP-запросы (frontend/auth и внешние клиенты) и проксирует их по gRPC на доменные сервисы контура сотрудников и сервисы статистики. Ключевая точка входа для SPA сотрудников.

## Стек

- Go 1.22, HTTP (net/http), gRPC-клиенты (grpc-gateway)

## Порты

| Протокол | Порт |
|----------|------|
| HTTP | 8090 |

## Проксируемые сервисы

- auth, customers, vehicles, deals, parts, brands, dealer-points
- workorders, works, employees, appointments, employee-reviews
- employee-statistics, client-statistics

## Взаимодействия

- Исходящие gRPC: все сервисы контура сотрудников + статистика
- Хранилища: —

## Запуск

```bash
go run ./services/gateway/employee   # make run-gateway
```

Docker: `build/gateway-service.Dockerfile`, compose-сервис `gateway-service`, версия в `VERSION`.

## API

Полное описание всех эндпоинтов — см. [API.md](API.md).
