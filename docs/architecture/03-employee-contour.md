# Контур сотрудников

```mermaid
flowchart LR
    subgraph Employee [Контур сотрудников]
        FE[frontend/auth]
        AUTH_HTTP[auth-service :8080<br/>SPA + auth HTTP]
        GW[gateway-service :8090]

        AUTH[auth :50051]
        CUST[customers :50052]
        VEH[vehicles :50053]
        DEALS[deals :50054]
        PARTS[parts :50055]
        BRANDS[brands :50056]
        DP[dealer-points :50057]
        WO[workorders :50064]
        WORKS[works :50065]
        EMP[employees :50066]
        E_REV[employee-reviews :50063]
        E_STAT[employee-statistics :50061]
        C_STAT[client-statistics :50062]
        SCHED[scheduler :8100]

        KF{{Kafka}}
        PG[(PostgreSQL)]
        RD[(Redis)]
    end

    FE --> AUTH_HTTP
    AUTH_HTTP -->|/api/* proxy| GW
    AUTH_HTTP -->|/api/auth/*| AUTH

    GW -->|gRPC grpc-gateway| AUTH & CUST & VEH & DEALS & PARTS & BRANDS & DP & WO & WORKS & EMP & E_REV & E_STAT & C_STAT

    DEALS -->|gRPC| CUST & VEH
    DEALS -->|deal.completed.v1| KF

    VEH -->|gRPC| BRANDS & DP
    PARTS -->|gRPC| BRANDS & DP & WO & EMP
    WO -->|gRPC| CUST & VEH & DP & PARTS & WORKS & EMP

    E_REV -->|gRPC| VEH
    KF -->|review.published.v1| E_REV
    KF -->|deal.completed.v1| E_STAT

    AUTH --> RD
    AUTH -->|auth.events| KF
    SCHED -->|SQL join 6 schemas| PG

    AUTH & CUST & VEH & DEALS & PARTS & BRANDS & DP & WO & WORKS & EMP & E_REV & E_STAT --> PG
```

## Точка входа для фронтенда

`frontend/auth` обращается к **auth-service :8080**:

1. `/api/auth/*` — логин, регистрация, refresh (напрямую в auth gRPC)
2. `/api/customers`, `/api/vehicles`, `/api/deals`, … — reverse proxy на **gateway-service :8090**
3. `/` — статика SPA

## Маршрутизация gateway-service (:8090)

| Backend | Proto-сервис |
|---------|--------------|
| auth-service | `AuthService` |
| customers-service | `CustomersService` |
| vehicles-service | `VehiclesService` |
| deals-service | `DealsService` |
| parts-service | `PartsService` |
| brands-service | `BrandsService` |
| dealer-points-service | `DealerPointsService` |
| employee-statistics-service | `EmployeeStatisticsService` |
| client-statistics-service | `ClientStatisticsService` |
| employee-reviews-service | `EmployeeReviewsService` |
| workorders-service | `WorkOrdersService` |
| works-service | `WorksService` |
| employees-service | `EmployeesService` |

## Типовые сценарии

### Логин сотрудника

```
frontend/auth → auth-service HTTP → auth-service gRPC → PostgreSQL (auth) + Redis (refresh)
```

### CRUD доменной сущности

```
frontend/auth → auth-service (proxy) → gateway-service → <domain-service> gRPC → PostgreSQL
```

### Создание сделки

```mermaid
sequenceDiagram
    participant FE as frontend/auth
    participant GW as gateway-service
    participant DEALS as deals-service
    participant CUST as customers
    participant VEH as vehicles
    participant KF as Kafka
    participant STAT as employee-statistics

    FE->>GW: POST /api/deals
    GW->>DEALS: gRPC CreateDeal
    DEALS->>CUST: gRPC (проверка customer)
    DEALS->>VEH: gRPC (проверка vehicle)
    DEALS->>DEALS: сохранить в DB
    DEALS->>KF: deal.completed.v1
    KF->>STAT: consume
```

### Заказ-наряд (workorders)

`workorders-service` — наиболее связанный сервис. При создании/изменении ЗН синхронно валидирует ссылки через gRPC:

| Зависимость | Назначение |
|-------------|------------|
| customers | Проверка клиента |
| vehicles | Проверка автомобиля |
| dealer-points | Проверка точки дилера |
| parts | Проверка запчастей |
| works | Проверка работ |
| employees | Проверка исполнителей |

### Движение запчастей (parts → workorders)

```mermaid
flowchart LR
    PARTS[parts-service] -->|gRPC notify| WO[workorders-service]
    WO -->|gRPC validate| PARTS
```

Циклическая зависимость **parts ↔ workorders**: parts уведомляет workorders при подтверждении складского движения; workorders валидирует запчасти при создании ЗН.

### Приглашение на отзыв (scheduler)

`scheduler-service` периодически выполняет cross-schema SQL:

- читает `workorders`, `customers`, `vehicles`, `clients`
- пишет в `reviews.review_invitations`

Не использует gRPC — прямой доступ к нескольким схемам PostgreSQL.

## Доменные сервисы без исходящих gRPC

Сервисы, которые только принимают вызовы (leaf nodes):

| Сервис | Схема БД |
|--------|----------|
| customers-service | `customers` |
| brands-service | `brands` |
| dealer-points-service | `dealerpoints` |
| works-service | `works` |
| employees-service | `employees` |
| employee-statistics-service | `employee_statistics` |

## Связанные документы

- [Общая схема](./01-overview.md)
- [Kafka-события](./04-kafka-events.md)
- [Матрица сервисов](./05-services-interactions.md)
