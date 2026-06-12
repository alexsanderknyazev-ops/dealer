# Контур клиентов

```mermaid
flowchart LR
    subgraph Client [Клиентский контур]
        FE[frontend/client]
        PUB[client-public-gateway<br/>:8091]
        PR[client-protected-gateway<br/>:8093]

        REG[client-registration<br/>:50058]
        AUTH[client-auth<br/>:50059]
        REV[client-reviews<br/>:50060]
        STAT[client-statistics<br/>:50062]

        VEH[vehicles-service<br/>:50053]
        KF{{Kafka}}
        PG[(PostgreSQL)]
        RD[(Redis)]
    end

    FE -->|публичные API<br/>login, register| PUB
    FE -->|защищённые API<br/>профиль, отзывы| PR

    PUB -->|gRPC| AUTH
    PUB -->|gRPC| REG

    PR -->|gRPC| AUTH
    PR -->|gRPC| REG
    PR -->|gRPC| REV

    REG -->|gRPC sync| VEH
    REG -->|gRPC sync<br/>IssueTokens| AUTH
    REG -->|produce<br/>client.registration.v1| KF

    KF -->|consume| AUTH
    KF -->|consume| STAT

    REV -->|gRPC sync| VEH
    REV -->|produce<br/>review.published.v1| KF

    REG -->|schema clients| PG
    AUTH -->|schema clientauth| PG
    REV -->|schemas reviews, clients| PG
    STAT -->|schema client_statistics| PG
    AUTH --> RD
```

## Сервисы клиентского контура

| Сервис | gRPC | HTTP | Схема БД |
|--------|------|------|----------|
| client-registration-service | 50058 | 8087 | `clients` |
| client-auth-service | 50059 | 8088 | `clientauth` |
| client-reviews-service | 50060 | 8089 | `reviews`, read `clients` |
| client-statistics-service | 50062 | 8095 | `client_statistics` |
| client-public-gateway | — | 8091 | — |
| client-protected-gateway | — | 8093 | — |

## Маршрутизация gateway

### client-public-gateway (:8091)

| Backend | Proto-сервис |
|---------|--------------|
| client-auth | `ClientAuthPublicService` |
| client-registration | `ClientRegistrationPublicService` |

### client-protected-gateway (:8093)

| Backend | Proto-сервис |
|---------|--------------|
| client-auth | `ClientAuthSessionService` |
| client-registration | `ClientAccountService` |
| client-reviews | `ReviewsService` |

## Типовые сценарии

### Регистрация клиента

```mermaid
sequenceDiagram
    participant FE as frontend/client
    participant GW as client-public-gateway
    participant REG as client-registration
    participant VEH as vehicles
    participant KF as Kafka
    participant AUTH as client-auth
    participant PG as PostgreSQL

    FE->>GW: POST /register
    GW->>REG: gRPC Register
    REG->>VEH: gRPC GetVehicleByVIN
    VEH-->>REG: vehicle
    REG->>PG: INSERT clients + client_vehicles
    REG->>KF: client.registration.v1
    KF->>AUTH: consume (создать user)
    AUTH->>PG: INSERT clientauth.users
    REG->>AUTH: gRPC IssueTokens (retry)
    AUTH-->>REG: tokens
    REG-->>GW: RegisterResult
    GW-->>FE: access + refresh token
```

### Логин

```
frontend/client → client-public-gateway → client-auth (gRPC)
```

### Профиль и автомобили

```
frontend/client → client-protected-gateway → client-registration (ClientAccountService)
```

### Отзыв

```mermaid
sequenceDiagram
    participant FE as frontend/client
    participant GW as client-protected-gateway
    participant REV as client-reviews
    participant VEH as vehicles
    participant KF as Kafka
    participant E_REV as employee-reviews
    participant STAT as client-statistics

    FE->>GW: POST /reviews
    GW->>REV: gRPC CreateReview
    REV->>VEH: gRPC (проверка авто)
    REV->>REV: сохранить в DB
    REV->>KF: review.published.v1
    KF->>E_REV: consume
    KF->>STAT: consume
```

## Зависимости от контура сотрудников

Клиентский контур использует **vehicles-service** из employee-контура:

- `client-registration` — поиск авто по VIN при регистрации
- `client-reviews` — привязка отзыва к автомобилю

Публичный gRPC-метод `GetVehicleByVIN` в vehicles-service доступен без employee JWT.

## Связанные документы

- [Общая схема](./01-overview.md)
- [Kafka-события](./04-kafka-events.md)
- [Матрица сервисов](./05-services-interactions.md)
