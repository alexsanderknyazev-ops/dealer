# Архитектура dealer-1

Документация по взаимодействию микросервисов платформы дилерского центра.

## Содержание

| Файл | Описание |
|------|----------|
| [01-overview.md](./01-overview.md) | Общая схема платформы |
| [02-client-contour.md](./02-client-contour.md) | Контур клиентов |
| [03-employee-contour.md](./03-employee-contour.md) | Контур сотрудников |
| [04-kafka-events.md](./04-kafka-events.md) | Kafka-топики и события |
| [05-services-interactions.md](./05-services-interactions.md) | Взаимодействия по каждому сервису |
| [06-ports-and-schemas.md](./06-ports-and-schemas.md) | Порты и владение схемами PostgreSQL |
| [grpc-gateway.md](./grpc-gateway.md) | Поток HTTP → gRPC через grpc-gateway |

## Условные обозначения

| Обозначение | Значение |
|-------------|----------|
| `──gRPC──►` | Синхронный вызов по gRPC |
| `──Kafka──►` | Публикация события в Kafka |
| `◄──Kafka──` | Потребление события из Kafka |
| `[(schema)]` | Схема PostgreSQL |
| `──HTTP──►` | HTTP-прокси / REST через grpc-gateway |

Все сервисы при наличии `KAFKA_BROKERS` публикуют ошибки в топик `platform.errors.v1` через `pkg/errorreport`.

## Два контура доступа

- **Сотрудники** — `frontend/auth` → `auth-service` (:8080) → `gateway-service` (:8090) → доменные gRPC-сервисы
- **Клиенты** — `frontend/client` → `client-public-gateway` (:8091) / `client-protected-gateway` (:8093) → client gRPC-сервисы
