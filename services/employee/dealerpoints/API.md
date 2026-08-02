# dealer-points-service — API

gRPC-сервис `dealerpoints.v1.DealerPointsService`. Доступ защищён JWT.

## Endpoints

### Дилерские точки

| gRPC | HTTP | Описание |
|---|---|---|
| `CreateDealerPoint` | `POST /api/dealer-points` | Создание точки |
| `GetDealerPoint` | `GET /api/dealer-points/{id}` | Точка по id |
| `ListDealerPoints` | `GET /api/dealer-points` | Список точек |
| `UpdateDealerPoint` | `PUT /api/dealer-points/{id}` | Обновление точки |
| `DeleteDealerPoint` | `DELETE /api/dealer-points/{id}` | Удаление точки |

### Юридические лица

| gRPC | HTTP | Описание |
|---|---|---|
| `CreateLegalEntity` | `POST /api/legal-entities` | Создание юр. лица |
| `GetLegalEntity` | `GET /api/legal-entities/{id}` | Юр. лицо по id |
| `ListLegalEntities` | `GET /api/legal-entities` | Список юр. лиц |
| `UpdateLegalEntity` | `PUT /api/legal-entities/{id}` | Обновление юр. лица |
| `DeleteLegalEntity` | `DELETE /api/legal-entities/{id}` | Удаление юр. лица |

### Привязка юр. лиц к точкам

| gRPC | HTTP | Описание |
|---|---|---|
| `LinkLegalEntityToDealerPoint` | `POST /api/dealer-points/{dealer_point_id}/legal-entities` | Привязать юр. лицо к точке |
| `UnlinkLegalEntityFromDealerPoint` | `DELETE /api/dealer-points/{dealer_point_id}/legal-entities/{legal_entity_id}` | Отвязать юр. лицо |
| `ListLegalEntitiesByDealerPoint` | `GET /api/dealer-points/{dealer_point_id}/legal-entities` | Юр. лица точки |

### Склады

| gRPC | HTTP | Описание |
|---|---|---|
| `CreateWarehouse` | `POST /api/warehouses` | Создание склада |
| `GetWarehouse` | `GET /api/warehouses/{id}` | Склад по id |
| `ListWarehouses` | `GET /api/warehouses` | Список складов |
| `UpdateWarehouse` | `PUT /api/warehouses/{id}` | Переименование склада |
| `DeleteWarehouse` | `DELETE /api/warehouses/{id}` | Удаление склада |

## Сообщения

### DealerPoint (модель)
`id`, `name`, `address`, `created_at`, `updated_at`

### LegalEntity (модель)
`id`, `name`, `inn`, `address`, `created_at`, `updated_at`

### Warehouse (модель)
`id`, `dealer_point_id`, `legal_entity_id`, `type` (`cars`/`parts`), `name`, `created_at`, `updated_at`

### Create/Update DealerPoint
Request: `name`, `address` (+ optional `id` для update)
Response: `dealer_point`

### ListDealerPoints / ListLegalEntities
Request: `limit`, `offset`, `search`
Response: `[]`, `total`

### Create/Update LegalEntity
Request: `name`, `inn`, `address`
Response: `legal_entity`

### Link/Unlink
Request: `dealer_point_id`, `legal_entity_id`
Response: пусто

### ListLegalEntitiesByDealerPoint
Request: `dealer_point_id`, `limit`, `offset`
Response: `legal_entities[]`, `total`

### Create/Update Warehouse
Request: `dealer_point_id`, `legal_entity_id`, `type`, `name`
Response: `warehouse`

### ListWarehouses
Request: `limit`, `offset`, `dealer_point_id`, `legal_entity_id`, `type`
Response: `warehouses[]`, `total`
