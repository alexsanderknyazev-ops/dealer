# Трейсинг (OpenTelemetry)

Все сервисы экспортируют распределённые трейсы по **OTLP/HTTP** (OpenTelemetry).
Инструментирование реализовано в `pkg/observe`:

- **HTTP:** `observe.WrapHTTP` оборачивает каждый маршрут в span (`otelhttp`),
  probe-пути (`/healthz`, `/readyz`, `/metrics`) исключаются.
- **gRPC server:** `observe.GRPCServerOptions` добавляет server interceptor (`otelgrpc`).
- **gRPC client:** `grpclient.DefaultDialOptions` добавляет client interceptor —
  трейсы прозрачно проходят через всю цепочку вызовов.
- **Запуск:** каждая `main()` вызывает `observe.InitTracing(serviceName)` (no-op,
  если коллектор не настроен).

## Как включить

Трейсинг включается стандартными переменными окружения OpenTelemetry:

| Переменная | Значение | По умолчанию |
|---|---|---|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | URL коллектора OTLP/HTTP | пусто → трейсинг выключен |
| `OTEL_TRACES_EXPORTER` | `otlp` (или `none` для отключения) | `otlp` |

Если `OTEL_EXPORTER_OTLP_ENDPOINT` не задан (или `OTEL_TRACES_EXPORTER=none`) —
трейсинг no-op, дополнительных ресурсов сервисы не тратят.

## Локальный стенд (docker compose)

В `docker-compose.yml` добавлен сервис **Jaeger** (v2, native OTLP) и переменная
`OTEL_EXPORTER_OTLP_ENDPOINT: http://jaeger:4318` для всех Go-сервисов:

- **UI Jaeger:** http://localhost:16686
- **OTLP HTTP:** порт `4318`

Поднимите стенд как обычно (`make local-up`) — трейсы появятся в Jaeger сразу.

## НТ-стенд (minikube)

`scripts/kube-up.sh` и манифесты в `k8s/` прокладывают тот же
`OTEL_EXPORTER_OTLP_ENDPOINT` на Jaeger внутри кластера. См. `docs/deploy/nt-lan.md`.

## Проверка

1. Откройте http://localhost:16686.
2. Выберите сервис (например `gateway-service`), `Find Traces`.
3. Выполните запрос к API — появится цепочка span через gateway → gRPC-сервисы → Postgres.

Запрос к employee-API: `curl -i http://localhost:8080/api/...`.
