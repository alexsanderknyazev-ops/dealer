# Dealer — система для автодилеров

Многомодульный проект: микросервисы на Go (gRPC), PostgreSQL, Kafka, Redis. JS-клиенты подключаются по gRPC-Web или через gateway.

## Стек

- **Backend:** Go 1.22+, gRPC, pgx (PostgreSQL), go-redis, Kafka (segmentio/kafka-go)
- **Инфраструктура:** Docker Compose (PostgreSQL, Redis, Zookeeper, Kafka, auth-service)

## Микросервисы

- **auth-service** (`services/auth`) — авторизация и аутентификация (регистрация, логин, JWT, сессии в Redis).
- **customers-service** (`services/customers`) — клиенты (физ./юр. лица): CRUD, поиск. HTTP API на 8081, gRPC на 50052. Требует JWT (тот же `JWT_SECRET`, что и auth).
- **vehicles-service** (`services/vehicles`) — автомобили на складе: VIN, марка, модель, год, пробег, цена, статус (available/sold/reserved). HTTP API на 8082, gRPC на 50053. Требует JWT.
- **deals-service** (`services/deals`) — сделки: клиент, автомобиль, сумма, этап (draft → in_progress → paid → completed), ответственный. HTTP API на 8083, gRPC на 50054. Требует JWT.
- **parts-service** (`services/parts`) — запасные части: артикул (SKU), название, категория, количество, единица, цена, расположение. HTTP API на 8084, gRPC на 50055. Требует JWT.

## Запуск инфраструктуры

```bash
cp .env.example .env   # при необходимости поправьте POSTGRES_PASSWORD и POSTGRES_DSN
docker compose up -d
```

Пароль БД в репозитории не хранится: в `docker-compose.yml` используется `POSTGRES_PASSWORD` из `.env` (или значение по умолчанию `changeme` только для локального стенда — смените его).

**После первого запуска (или новой БД)** примените миграции и при необходимости создайте админа и тестовые данные:

```bash
export POSTGRES_DSN="postgres://dealer:ВАШ_ПАРОЛЬ@127.0.0.1:5433/dealer?sslmode=disable"  # пароль как в .env / POSTGRES_PASSWORD
make migrate                                    # создать таблицы users, customers, vehicles, deals, parts
make seed-admin                                 # админ admin@dealer.local / admin123 (нужен POSTGRES_DSN)
make seed-data                                  # тестовые клиенты и автомобили (если таблицы пусты)
```

Без `make migrate` сервисы customers, vehicles и deals будут падать с ошибкой «relation does not exist».

## Миграции и запуск auth-service

```bash
# Зависимости (workspace: корень + services/auth)
go mod download
cd services/auth && go mod download && cd ../..

# Генерация gRPC (опционально, если меняли .proto)
make proto

# Поднять инфраструктуру (Postgres, Redis, Kafka + auth-service)
make docker-up

# Либо только инфраструктура, сервис локально:
docker compose up -d postgres redis zookeeper kafka
# Миграции (порт 5433 при доступе с хоста к Docker); задайте DSN с паролем из .env
export POSTGRES_DSN="postgres://dealer:ВАШ_ПАРОЛЬ@127.0.0.1:5433/dealer?sslmode=disable"
psql "$POSTGRES_DSN" -f migrations/001_users.up.sql
psql "$POSTGRES_DSN" -f migrations/002_roles.up.sql
psql "$POSTGRES_DSN" -f migrations/003_customers.up.sql
psql "$POSTGRES_DSN" -f migrations/004_vehicles.up.sql
# Создать пользователя admin (по умолчанию admin@dealer.local / admin123)
make seed-admin
# Тестовые клиенты и автомобили (только если таблицы пусты)
make seed-data
# Запуск сервисов (несколько терминалов)
make run-auth
make run-customers
make run-vehicles
make run-deals
make run-parts
```

auth-service: gRPC 50051, HTTP 8080. customers: 50052/8081, vehicles: 50053/8082, deals: 50054/8083, parts: 50055/8084.

## Фронт авторизации

React (Vite) SPA в `frontend/auth`: экраны входа, регистрации и личный кабинет после входа.

**Локальная разработка (фронт отдельно):**
```bash
# Терминалы 1–3: бэкенды
make run-auth
make run-customers
make run-vehicles

# Терминал 4: фронт (прокси /api на 8080, /api/customers, /api/vehicles, /api/deals, /api/parts на соответствующие порты)
make frontend-dev
```
Откройте http://127.0.0.1:3000 — вход, регистрация, разделы «Клиенты», «Автомобили», «Сделки», «Запчасти» (список, создание, просмотр, редактирование, удаление).

**Продакшен (фронт из auth-service):** при `docker compose up` фронт собирается в образ и раздаётся с http://127.0.0.1:8080 (тот же хост, что и API).

## Конфигурация

Скопируйте `.env.example` в `.env` и задайте переменные (в т.ч. `JWT_SECRET`, DSN, порты).
