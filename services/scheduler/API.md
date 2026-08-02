# scheduler-service — API

Фоновый воркер (порт 8100). Отдельного gRPC/HTTP API (кроме health/metrics) не имеет: выполняет периодические задачи напрямую через SQL по нескольким схемам PostgreSQL.

## Endpoints

| Метод | Путь | Описание |
|---|---|---|
| `GET` | `/health` | Healthcheck (PostgreSQL) |
| `GET` | `/metrics` | Prometheus-метрики |
