# errors-ingest-service — API

Сервис приёма ошибок/телеметрии (порт 8092). HTTP + Kafka.

## Endpoints

| Метод | Путь | Описание |
|---|---|---|
| `POST` | `/api/telemetry/events` | Приём события телеметрии/ошибки (JSON) |
| `OPTIONS` | `/api/telemetry/events` | CORS preflight |

Данные сохраняются в ClickHouse (БД `analytics`). Также сервис потребляет события `platform.errors.v1` из Kafka.
