# brands-service

Сервис контура сотрудников: справочник брендов (марок) автомобилей и запчастей, включая нормо-часы/ставки по брендам (`brand_labor_rates`). Доступ защищён JWT.

## Стек

- Go 1.22, gRPC, JWT (golang-jwt/v5)
- PostgreSQL (pgx/v5) — схема `brands`

## Порты

| Протокол | Порт |
|----------|------|
| gRPC | 50056 |
| HTTP (health/metrics) | 8085 |

## gRPC API

`brands.v1.BrandsService`:
- `CreateBrand`, `GetBrand`, `ListBrands`, `UpdateBrand`, `DeleteBrand`
- `ListBrandLaborRates`, `UpdateBrandLaborRate`, `DeleteBrandLaborRate`, `ResolveBrandLaborRate`

## Взаимодействия

- Исходящие gRPC: —
- Kafka: —
- Хранилища: PostgreSQL (`brands`)

## Запуск

```bash
go run ./services/employee/brands   # make run-brands
```

Docker: `build/brands-service.Dockerfile`, compose-сервис `brands-service`, версия в `VERSION`.

## API

Полное описание всех эндпоинтов — см. [API.md](API.md).
