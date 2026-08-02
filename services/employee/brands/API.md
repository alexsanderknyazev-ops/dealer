# brands-service — API

gRPC-сервис `brands.v1.BrandsService`. Доступ защищён JWT.

## Endpoints

### Бренды

| gRPC | HTTP | Описание |
|---|---|---|
| `CreateBrand` | `POST /api/brands` | Создание бренда |
| `GetBrand` | `GET /api/brands/{id}` | Бренд по id |
| `ListBrands` | `GET /api/brands` | Список брендов |
| `UpdateBrand` | `PUT /api/brands/{id}` | Переименование бренда |
| `DeleteBrand` | `DELETE /api/brands/{id}` | Удаление бренда |

### Нормо-часы по брендам (brand_labor_rates)

| gRPC | HTTP | Описание |
|---|---|---|
| `ListBrandLaborRates` | `GET /api/brand-labor-rates` | Список ставок (по brand_id / dealer_point_id) |
| `UpdateBrandLaborRate` | `PUT /api/brand-labor-rates` | Создание/обновление ставки |
| `DeleteBrandLaborRate` | `DELETE /api/brand-labor-rates/{id}` | Удаление ставки |
| `ResolveBrandLaborRate` | `GET /api/brand-labor-rates/resolve` | Разрешение ставки по бренду+точке+типу ремонта |

## Сообщения

### Brand (модель)
`id`, `name`, `created_at`, `updated_at`

### BrandLaborRate (модель)
`id`, `brand_id`, `dealer_point_id`, `warranty_hour_price`, `commercial_hour_price`, `created_at`, `updated_at`

### CreateBrand
Request: `name`
Response: `brand`

### GetBrand / UpdateBrand / DeleteBrand
Request: `id` (+ optional `name` для update)
Response: `brand` / пусто

### ListBrands
Request: `limit`, `offset`, `search`
Response: `brands[]`, `total`

### ListBrandLaborRates
Request: `limit`, `offset`, `brand_id`, `dealer_point_id`
Response: `brand_labor_rates[]`, `total`

### UpdateBrandLaborRate
Request: `brand_id`, `dealer_point_id`, `warranty_hour_price`, `commercial_hour_price`
Response: `brand_labor_rate`

### DeleteBrandLaborRate
Request: `id`
Response: пусто

### ResolveBrandLaborRate
Request: `brand_id`, `dealer_point_id`, `repair_type`
Response: `warranty_hour_price`, `commercial_hour_price`, `hour_price`, `found`
