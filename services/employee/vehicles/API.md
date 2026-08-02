# vehicles-service — API

gRPC-сервис `vehicles.v1.VehiclesService`. Доступ защищён JWT.

## Endpoints

| gRPC | HTTP | Описание |
|---|---|---|
| `CreateVehicle` | `POST /api/vehicles` | Добавление автомобиля |
| `GetVehicle` | `GET /api/vehicles/{id}` | Автомобиль по id |
| `ListVehicles` | `GET /api/vehicles` | Список автомобилей (фильтры/поиск) |
| `UpdateVehicle` | `PUT /api/vehicles/{id}` | Обновление автомобиля (частичное) |
| `DeleteVehicle` | `DELETE /api/vehicles/{id}` | Удаление автомобиля |
| `GetVehicleByVIN` | — (внутренний gRPC) | Точный поиск по VIN (для client-registration-service) |

## Сообщения

### Vehicle (модель)
`id`, `vin`, `make`, `model`, `year`, `mileage_km`, `price` (decimal как строка), `status` (`available`/`sold`/`reserved`), `color`, `notes`, `brand_id`, `dealer_point_id`, `legal_entity_id`, `warehouse_id`, `created_at`, `updated_at`

### CreateVehicle
Request: `vin`, `make`, `model`, `year`, `mileage_km`, `price`, `status`, `color`, `notes`, `brand_id?`, `dealer_point_id`, `legal_entity_id`, `warehouse_id`
Response: `vehicle`

### GetVehicle / GetVehicleByVIN
Request: `id` или `vin`
Response: `vehicle`

### ListVehicles
Request: `limit`, `offset`, `search` (vin/make/model), `status`, `brand_id`, `dealer_point_id`, `legal_entity_id`, `warehouse_id`
Response: `vehicles[]`, `total`

### UpdateVehicle
Request: `id`, optional все поля модели
Response: `vehicle`

### DeleteVehicle
Request: `id`
Response: пусто
