# Порты и владение схемами PostgreSQL

## Порты сервисов (docker-compose)

### Gateways и инфраструктура

| Сервис | HTTP | gRPC |
|--------|------|------|
| auth-service | 8080 | 50051 |
| gateway-service | 8090 | — |
| client-public-gateway | 8091 | — |
| client-protected-gateway | 8093 | — |
| errors-ingest | 8092 | — |
| scheduler-service | 8100 | — |

### Контур клиентов

| Сервис | HTTP | gRPC |
|--------|------|------|
| client-registration-service | 8087 | 50058 |
| client-auth-service | 8088 | 50059 |
| client-reviews-service | 8089 | 50060 |
| client-statistics-service | 8095 | 50062 |

### Контур сотрудников

| Сервис | HTTP | gRPC |
|--------|------|------|
| customers-service | 8081 | 50052 |
| vehicles-service | 8082 | 50053 |
| deals-service | 8083 | 50054 |
| parts-service | 8084 | 50055 |
| brands-service | 8085 | 50056 |
| dealer-points-service | 8086 | 50057 |
| employee-statistics-service | 8094 | 50061 |
| employee-reviews-service | 8096 | 50063 |
| workorders-service | 8097 | 50064 |
| works-service | 8098 | 50065 |
| employees-service | 8099 | 50066 |

### Инфраструктура данных

| Компонент | Порт |
|-----------|------|
| PostgreSQL | 5433 (host) → 5432 (container) |
| Redis | 6379 |
| Kafka | 9092 (host), 29092 (internal) |
| ClickHouse HTTP | 8123 |
| ClickHouse native | 9000 |

---

## Владение схемами PostgreSQL

Константы схем определены в `pkg/dbschema/schemas.go`.

| Сервис | Схема(ы) | Владелец данных |
|--------|----------|-----------------|
| auth-service | `auth` | Учётные записи сотрудников |
| customers-service | `customers` | Клиенты дилера (CRM) |
| vehicles-service | `vehicles` | Автомобили на складе / в парке |
| deals-service | `deals` | Сделки продаж |
| parts-service | `parts` | Запчасти, склад, движения |
| brands-service | `brands` | Марки, модели, нормо-часы |
| dealer-points-service | `dealerpoints` | Точки дилера, юрлица, склады |
| workorders-service | `workorders` | Заказ-наряды |
| works-service | `works` | Справочник работ |
| employees-service | `employees` | Сотрудники |
| employee-reviews-service | `employee_reviews` | Отзывы (вид сотрудника) |
| employee-statistics-service | `employee_statistics` | Агрегаты продаж |
| client-registration-service | `clients` | Профили клиентов (B2C) |
| client-auth-service | `clientauth` | Учётные записи клиентов |
| client-reviews-service | `reviews`, read `clients` | Отзывы клиентов |
| client-statistics-service | `client_statistics` | Агрегаты клиентской активности |
| scheduler-service | read: `reviews`, `workorders`, `deals`, `customers`, `clients`, `vehicles` | Фоновые задачи (не владеет данными) |
| errors-ingest | — | ClickHouse `analytics` |

### Общая схема `public`

Все сервисы подключаются также к схеме `public` (миграции, общие объекты).

---

## Расположение кода сервисов

Канонический путь исходников (собирается Docker):

| Docker-образ | Путь в репозитории |
|--------------|-------------------|
| auth-service | `services/employee/auth/` |
| customers-service | `services/employee/customers/` |
| vehicles-service | `services/employee/vehicles/` |
| deals-service | `services/employee/deals/` |
| parts-service | `services/employee/parts/` |
| brands-service | `services/employee/brands/` |
| dealer-points-service | `services/employee/dealerpoints/` |
| workorders-service | `services/employee/workorders/` |
| works-service | `services/employee/works/` |
| employees-service | `services/employee/employees/` |
| employee-reviews-service | `services/employee/reviews/` |
| gateway-service | `services/gateway/employee/` |
| client-* | `services/client/*` |
| client-public-gateway | `services/gateway/client-public/` |
| client-protected-gateway | `services/gateway/client-protected/` |
| *-statistics | `services/statistics/*` |
| errors-ingest | `services/errors-ingest/` |
| scheduler-service | `services/scheduler/` |

---

## Proto-контракты

| Домен | Файл |
|-------|------|
| auth | `api/proto/auth/v1/auth.proto` |
| customers | `api/proto/customers/v1/customers.proto` |
| vehicles | `api/proto/vehicles/v1/vehicles.proto` |
| deals | `api/proto/deals/v1/deals.proto` |
| parts | `api/proto/parts/v1/parts.proto` |
| brands | `api/proto/brands/v1/brands.proto` |
| dealer-points | `api/proto/dealerpoints/v1/dealerpoints.proto` |
| workorders | `api/proto/workorders/v1/work_orders.proto` |
| works | `api/proto/works/v1/works.proto` |
| employees | `api/proto/employees/v1/employees.proto` |
| reviews (client) | `api/proto/reviews/v1/reviews.proto` |
| reviews (employee) | `api/proto/reviews/v1/employee_reviews.proto` |
| client-auth | `api/proto/clientauth/v1/*.proto` |
| client-registration | `api/proto/clients/v1/*.proto` |
| employee statistics | `api/proto/statistics/employee/v1/employee_stats.proto` |
| client statistics | `api/proto/statistics/client/v1/client_stats.proto` |

Сгенерированный код: `pkg/pb/`.

## Связанные документы

- [Общая схема](./01-overview.md)
- [Матрица сервисов](./05-services-interactions.md)
