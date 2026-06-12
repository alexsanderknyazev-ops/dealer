# Объёмные тестовые данные (50–150 записей на схему)

SQL-файлы для наполнения БД демо-данными с фиксированным UUID-пространством `9000000*`.

**Пароль всех тестовых пользователей:** `Test1234!`

## Порядок применения

```bash
export POSTGRES_DSN='postgres://dealer:PASSWORD@HOST:PORT/dealer?sslmode=disable'

# После миграций и базового seed (make full-seed рекомендуется)
./migrations/seed_volume/apply.sh
```

Или в Kubernetes:

```bash
./migrations/seed_volume/apply.sh --k8s
```

## Файлы

| Файл | Схема | Записей (основная таблица) |
|------|-------|----------------------------|
| `01_auth.sql` | auth | 100 users |
| `02_employees.sql` | employees | 100 |
| `03_dealerpoints.sql` | dealerpoints | 60 точек + 60 юрлиц + 120 складов |
| `04_brands.sql` | brands | 80 брендов |
| `05_customers.sql` | customers | 120 |
| `06_vehicles.sql` | vehicles | 120 |
| `07_parts.sql` | parts | папки запчастей + 120 запчастей + 60 поставщиков |
| `08_works.sql` | works | папки работ + 100 работ |
| `09_deals.sql` | deals | 100 |
| `10_workorders.sql` | workorders | 80 заказ-нарядов |
| `11_clientauth.sql` | clientauth | 100 |
| `12_clients.sql` | clients | 100 профилей + 100 авто |
| `13_reviews.sql` | reviews | 80 |
| `14_employee_reviews.sql` | employee_reviews | 80 |
| `15_employee_statistics.sql` | employee_statistics | 80 |
| `16_client_statistics.sql` | client_statistics | 80 + 80 событий |
| `17_appointments.sql` | appointments | 80 записей + работы и запчасти |
| `18_part_orders.sql` | parts | 80 заказов поставщику + 80 заказов покупателя |

## Очистка

```bash
psql "$POSTGRES_DSN" -f migrations/seed_volume/99_cleanup.sql
```

Удаляет только записи с UUID `9000000*`.

## Примеры логина

```bash
# Сотрудник (volume)
curl -X POST http://HOST:8090/api/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"vol.employee1@test.dealer.local","password":"Test1234!"}'

# B2C клиент (volume)
curl -X POST http://HOST:8091/api/login \
  -d '{"email":"vol.client1@test.dealer.local","password":"Test1234!"}'
```
