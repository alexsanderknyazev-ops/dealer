# vehicles-service

Сервис контура сотрудников: автомобили на складе — VIN, марка, модель, год, пробег, цена, статус (available/sold/reserved), поиск по VIN. Доступ защищён JWT.

## Стек

- Go 1.22, gRPC, JWT (golang-jwt/v5)
- PostgreSQL (pgx/v5) — схема `vehicles`

## Порты

| Протокол | Порт |
|----------|------|
| gRPC | 50053 |
| HTTP (health/metrics) | 8082 |

## gRPC API

`vehicles.v1.VehiclesService`:
- `CreateVehicle`, `GetVehicle`, `ListVehicles`, `UpdateVehicle`, `DeleteVehicle`, `GetVehicleByVIN`

## Взаимодействия

- Исходящие gRPC: brands, dealer-points
- Kafka: —
- Хранилища: PostgreSQL (`vehicles`)

## Запуск

```bash
go run ./services/employee/vehicles   # make run-vehicles
```

Docker: `build/vehicles-service.Dockerfile`, compose-сервис `vehicles-service`, версия в `VERSION`.

## API

Полное описание всех эндпоинтов — см. [API.md](API.md).
