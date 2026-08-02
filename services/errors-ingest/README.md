# errors-ingest-service

Инфраструктурный сервис: приём ошибок/телеметрии с платформы. Потребляет события `platform.errors.v1` из Kafka и пишет их в ClickHouse (БД `analytics`) для мониторинга/аналитики. Также принимает HTTP-телеметрию (через proxy от auth-service).

## Стек

- Go 1.22, HTTP (net/http)
- ClickHouse (24.8)
- Kafka — потребление событий

## Порты

| Протокол | Порт |
|----------|------|
| HTTP | 8092 |

## Взаимодействия

- Исходящие gRPC: —
- Kafka: потребление ← `platform.errors.v1`
- Хранилища: ClickHouse (`analytics`)

## Запуск

Docker: `build/errors-ingest-service.Dockerfile`, compose-сервис `errors-ingest-service`, версия в `VERSION`.

## API

Полное описание всех эндпоинтов — см. [API.md](API.md).
