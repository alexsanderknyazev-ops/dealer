# Общая схема платформы

```mermaid
flowchart TB
    subgraph FE [Фронтенды]
        FE_EMP[frontend/auth<br/>сотрудники]
        FE_CLI[frontend/client<br/>клиенты]
    end

    subgraph GW [API Gateways — HTTP]
        AUTH_HTTP[auth-service :8080<br/>SPA + proxy]
        GW_EMP[gateway-service :8090]
        GW_PUB[client-public-gateway :8091]
        GW_PR[client-protected-gateway :8093]
    end

    subgraph EMP [Контур сотрудников — gRPC]
        AUTH[auth-service :50051]
        CUST[customers :50052]
        VEH[vehicles :50053]
        DEALS[deals :50054]
        PARTS[parts :50055]
        BRANDS[brands :50056]
        DP[dealer-points :50057]
        WO[workorders :50064]
        WORKS[works :50065]
        EMP_SVC[employees :50066]
        E_REV[employee-reviews :50063]
        E_STAT[employee-statistics :50061]
    end

    subgraph CLI [Контур клиентов — gRPC]
        C_REG[client-registration :50058]
        C_AUTH[client-auth :50059]
        C_REV[client-reviews :50060]
        C_STAT[client-statistics :50062]
    end

    subgraph INFRA [Инфраструктура]
        PG[(PostgreSQL)]
        RD[(Redis)]
        KF{{Kafka}}
        CH[(ClickHouse)]
        E_ING[errors-ingest :8092]
        SCHED[scheduler :8100]
    end

    FE_EMP --> AUTH_HTTP
    AUTH_HTTP -->|proxy API| GW_EMP
    AUTH_HTTP --> AUTH
    GW_EMP --> AUTH & CUST & VEH & DEALS & PARTS & BRANDS & DP & WO & WORKS & EMP_SVC & E_REV & E_STAT & C_STAT

    FE_CLI --> GW_PUB & GW_PR
    GW_PUB --> C_AUTH & C_REG
    GW_PR --> C_AUTH & C_REG & C_REV

    EMP & CLI --> PG
    AUTH & C_AUTH --> RD

    DEALS -->|deal.completed.v1| KF
    C_REG -->|client.registration.v1| KF
    C_REV -->|review.published.v1| KF
    AUTH -->|auth.events| KF

    KF -->|consume| C_AUTH & E_STAT & C_STAT & E_REV & E_ING
    E_ING --> CH

    SCHED -->|cross-schema SQL| PG

    VEH & DEALS & WO & PARTS & C_REG & C_REV & E_REV -.->|gRPC| EMP
```

## Краткое описание слоёв

### Фронтенды

- `frontend/auth` — SPA для сотрудников дилера (продажи, сервис, склад, HR)
- `frontend/client` — SPA для конечных клиентов (регистрация, профиль, отзывы)

### API Gateways

Все gateway используют **grpc-gateway**: HTTP/JSON снаружи, gRPC внутри.

| Gateway | Порт | Назначение |
|---------|------|------------|
| auth-service HTTP | 8080 | SPA + `/api/auth/*` + reverse proxy на gateway-service |
| gateway-service | 8090 | Единая точка входа для доменных API сотрудников |
| client-public-gateway | 8091 | Публичные API клиентов (login, register) |
| client-protected-gateway | 8093 | Защищённые API клиентов (профиль, отзывы) |

### Инфраструктура

| Компонент | Роль |
|-----------|------|
| PostgreSQL | Основное хранилище, schema-per-service |
| Redis | Refresh-токены (auth-service, client-auth) |
| Kafka | Асинхронные события между контурами |
| ClickHouse | Аналитика ошибок (errors-ingest) |
| scheduler-service | Фоновые задачи, cross-schema SQL |

## Связанные документы

- [Контур клиентов](./02-client-contour.md)
- [Контур сотрудников](./03-employee-contour.md)
- [Kafka-события](./04-kafka-events.md)
