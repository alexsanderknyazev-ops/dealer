# Проверка данных в PostgreSQL

Одна БД `dealer`, схемы по сервисам (см. `pkg/dbschema/schemas.go`).

## Подключение

```bash
export POSTGRES_DSN="postgres://dealer:PASSWORD@127.0.0.1:5433/dealer?sslmode=disable"

# интерактивно
psql "$POSTGRES_DSN"

# один запрос
psql "$POSTGRES_DSN" -c "SELECT COUNT(*) FROM customers.customers;"
```

## Карта схем → таблицы

| Схема | Таблицы (основные) | Сервис |
|-------|-------------------|--------|
| `auth` | `users` | employee-auth |
| `customers` | `customers` | customers |
| `vehicles` | `vehicles` | vehicles |
| `deals` | `deals` | deals |
| `brands` | `brands` | brands |
| `dealerpoints` | `dealer_points`, `legal_entities`, `dealer_point_legal_entities`, `warehouses` | dealer-points |
| `parts` | `parts`, `part_folders`, `part_stock`, `movement_documents`, `movement_document_lines`, `stock_movements` | parts |
| `workorders` | `work_orders`, `work_order_labor`, `work_order_parts` | work-orders |
| `clients` | `clients`, `client_vehicles` | client-registration |
| `clientauth` | `users` | client-auth |
| `reviews` | `reviews` | client-reviews |
| `employee_reviews` | `reviews` | employee-reviews |
| `employee_statistics` | `deal_events` | employee-statistics |
| `client_statistics` | `client_registration_events`, `review_events` | client-statistics |

## Паттерн проверки (каждый детальный кейс)

1. **Snapshot до** — `COUNT(*)` или `SELECT ... FOR` ключевой таблицы
2. **API call** — сохранить id из ответа
3. **Snapshot после** — row exists, поля совпадают с request/response
4. **Delta** — count +1 (create) / 0 (failed) / -1 (delete)
5. **Связанные таблицы** — FK-строки, дочерние записи

## Согласованность API ↔ БД

| API поле | Колонка БД | Заметка |
|----------|------------|---------|
| `customer_type` | `customers.customer_type` | individual / legal |
| `full_name` (client) | `clients.full_name` | не `name` |
| `assigned_to` / `responsible_id` | `deals.assigned_to` | UUID без FK |
| `movement_document_id` | `workorders.work_orders.movement_document_id` | зеркало parts |
| `review_id` (employee) | `employee_reviews.reviews.review_id` | id из reviews.reviews |

## Готовые запросы

См. [`db-queries.sql`](./db-queries.sql) — шаблоны с плейсхолдерами `<UUID>`.

## ClickHouse (telemetry / errors)

```bash
curl 'http://127.0.0.1:8123/?query=SELECT+count()+FROM+analytics.events+WHERE+route+LIKE+%27%25telemetry%25%27'
```

Только если подняты `clickhouse` + `errors-ingest-service`.
