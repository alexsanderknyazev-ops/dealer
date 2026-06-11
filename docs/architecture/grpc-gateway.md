# gRPC + grpc-gateway (Variant A)

## Поток запросов

```
Browser (REST/JSON)
  → auth-service :8080 (SPA + proxy)
    → gateway-service :8090 (grpc-gateway)
      → *-service gRPC :5005x
```

- Контракт API описан в `api/proto/**` с `google.api.http` аннотациями (включая auth).
- `make proto` генерирует `*.pb.gw.go` в `pkg/pb/**`.
- `gateway-service` транслирует REST в gRPC и пробрасывает `Authorization` в metadata.
- Domain-сервисы проверяют JWT на gRPC через `pkg/grpcauth` interceptor.
- Domain-сервисы отдают только gRPC API; HTTP на портах 808x — health/metrics (`/healthz`, `/readyz`).

## Локальный запуск

```bash
make proto
docker compose up -d --build
```

После пересборки отдельных domain-сервисов перезапустите gateway (иначе возможен 503 из‑за устаревших gRPC-соединений):

```bash
docker compose up -d --build deals-service vehicles-service
docker compose restart gateway-service
```

Прямой доступ к gateway (без auth proxy): `http://localhost:8090/api/customers`.

## Auth endpoints

`/api/register`, `/api/login`, `/api/refresh`, `/api/logout`, `/api/me` — в `auth.proto`, REST через gateway.
`auth-service` проксирует эти пути на gateway вместе с domain API.
`GET /api/me` читает Bearer-токен из `Authorization` (metadata в gRPC `Validate`).

## Межсервисно (gRPC)

| Сервис | Env |
|--------|-----|
| gateway | `AUTH_GRPC_ADDR`, `*_GRPC_ADDR` для всех доменов |
| deals | `CUSTOMERS_GRPC_ADDR`, `VEHICLES_GRPC_ADDR` |
| vehicles | `BRANDS_GRPC_ADDR`, `DEALER_POINTS_GRPC_ADDR` |
| parts | `BRANDS_GRPC_ADDR`, `DEALER_POINTS_GRPC_ADDR` |

JWT пробрасывается через `pkg/grpclient` (из gRPC metadata или HTTP `Authorization`).
