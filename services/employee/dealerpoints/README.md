# dealer-points-service

Сервис контура сотрудников: дилерские точки, юридические лица и их привязка, склады. Доступ защищён JWT.

## Стек

- Go 1.22, gRPC, JWT (golang-jwt/v5)
- PostgreSQL (pgx/v5) — схема `dealerpoints`

## Порты

| Протокол | Порт |
|----------|------|
| gRPC | 50057 |
| HTTP (health/metrics) | 8086 |

## gRPC API

`dealerpoints.v1.DealerPointsService`:
- Дилерские точки: `CreateDealerPoint`, `GetDealerPoint`, `ListDealerPoints`, `UpdateDealerPoint`, `DeleteDealerPoint`
- Юр. лица: `CreateLegalEntity`, `GetLegalEntity`, `ListLegalEntities`, `UpdateLegalEntity`, `DeleteLegalEntity`, `LinkLegalEntityToDealerPoint`, `UnlinkLegalEntityFromDealerPoint`, `ListLegalEntitiesByDealerPoint`
- Склады: `CreateWarehouse`, `GetWarehouse`, `ListWarehouses`, `UpdateWarehouse`, `DeleteWarehouse`

## Взаимодействия

- Исходящие gRPC: —
- Kafka: —
- Хранилища: PostgreSQL (`dealerpoints`)

## Запуск

```bash
go run ./services/employee/dealerpoints   # make run-dealer-points
```

Docker: `build/dealer-points-service.Dockerfile`, compose-сервис `dealer-points-service`, версия в `VERSION`.

## API

Полное описание всех эндпоинтов — см. [API.md](API.md).
