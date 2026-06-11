# gRPC + grpc-gateway (Variant A)

## Поток запросов

```
Browser (REST/JSON)
  → auth-service :8080 (SPA + /api/login + proxy)
    → gateway-service :8090 (grpc-gateway)
      → *-service gRPC :5005x
```

- Контракт API описан в `api/proto/**` с `google.api.http` аннотациями.
- `make proto` генерирует `*.pb.gw.go` в `pkg/pb/**`.
- `gateway-service` транслирует REST в gRPC и пробрасывает `Authorization` в metadata.
- Domain-сервисы проверяют JWT на gRPC через `pkg/grpcauth` interceptor.

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

## Миграция

1. **Сейчас**: gateway + legacy HTTP handlers в сервисах (оба пути работают).
2. **Межсервисно (gRPC)**:
   - `deals-service` → `customers-service`, `vehicles-service` (`GetCustomer` / `GetVehicle` для referential integrity)
   - `vehicles-service` → `brands-service` (`GetBrand` при `brand_id`)
   - JWT пробрасывается через `pkg/grpclient` (из gRPC metadata или HTTP `Authorization`)
3. **Финал**: удалить `internal/httpapi` из domain-сервисов, оставить только gRPC server.

Переменные окружения:

| Сервис | Env |
|--------|-----|
| deals | `CUSTOMERS_GRPC_ADDR`, `VEHICLES_GRPC_ADDR` |
| vehicles | `BRANDS_GRPC_ADDR` |

## Auth endpoints

`/api/register`, `/api/login`, `/api/refresh`, `/api/logout`, `/api/me` пока остаются на HTTP в `auth-service`.
Их можно перенести в proto + gateway отдельным этапом.
