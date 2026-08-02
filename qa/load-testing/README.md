# Load testing — employee gateway GET endpoints

Нагрузочное тестирование всех **GET** ручек employee gateway.

## Быстрый старт

```bash
# 1. Проба всех GET на удалённом хосте
BASE_URL=http://192.168.0.27:9080 \
LOGIN_EMAIL=qa.master@test.local \
LOGIN_PASSWORD='Test1234!' \
python3 qa/load-testing/scripts/probe_get.py

# 2. Нагрузка без k6 (curl, параллельно)
PROFILE=smoke BASE_URL=http://192.168.0.27:9080 \
  ./qa/load-testing/scripts/curl-load.sh

# 3. Полный цикл: проба + k6 (если установлен)
PROFILE=smoke BASE_URL=http://192.168.0.27:9080 \
  ./qa/load-testing/scripts/run-load-test.sh
```

## Переменные окружения

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `BASE_URL` | `http://192.168.0.27:9080` | Employee gateway |
| `LOGIN_EMAIL` | `qa.master@test.local` | Учётка сотрудника |
| `LOGIN_PASSWORD` | `Test1234!` | Пароль |
| `CLIENT_ID` | fixture UUID | Для `/api/clients/{id}/reviews` |
| `PROFILE` | `smoke` | Профиль k6: `smoke`, `load`, `stress` |

## Профили k6

| Профиль | VUs | Длительность | Назначение |
|---------|-----|--------------|------------|
| `smoke` | 5 | 30s | Быстрая проверка под нагрузкой |
| `load` | 50 | 5m | Рабочая нагрузка |
| `stress` | 100 | 10m | Стресс-тест |

```bash
PROFILE=load BASE_URL=http://192.168.0.27:9080 \
  k6 run -e PROFILE=load \
  -e MANIFEST_PATH=qa/load-testing/results/latest-endpoints.json \
  qa/load-testing/k6/get-endpoints.js
```

## Артефакты

| Файл | Описание |
|------|----------|
| `results/latest-probe-report.md` | Последний отчёт пробы (HTTP коды) |
| `results/latest-endpoints.json` | Манифест для k6 (paths + weights) |
| `results/probe-*/` | История проб с timestamp |

## Покрытие GET

Скрипт `probe_get.py` проверяет:

- `/healthz`, `/api/me`
- Списки: customers, vehicles, deals, parts, folders, brands, dealer-points, legal-entities, warehouses, work-orders, works, employees, reviews, movement-documents, suppliers, orders, repair-appointments
- Статистика: `/api/stats/employee/overview`, `/api/stats/client/overview`, `/api/reviews/stats`
- Resolve/slots: brand-labor-rates, repair-appointment-slots, dealer-point legal entities
- By-ID маршруты (ID берутся из list-ответов)
- `/api/parts/{id}/stock`
- `/api/clients/{client_id}/reviews`

## Результат пробы `192.168.0.27:9080` (2026-06-12)

| Метрика | Значение |
|---------|----------|
| PASS (JSON API) | 42 |
| WARN (200, но HTML SPA) | 4 |
| SKIP (нет ID) | 2 |
| FAIL | 0 |

**WARN — маршрут не проксируется на API**, отдаётся frontend HTML:

- `/api/suppliers`
- `/api/supplier-orders`
- `/api/customer-orders`
- `/api/repair-appointment-slots`

Эти ручки **исключены** из манифеста нагрузки (`latest-endpoints.json`).

Smoke load (50 req, 5 concurrent): **OK=50, p95≈125ms, ~116 rps**.

## Известные проблемы

- 4 GET возвращают HTML вместо JSON на `:9080` (см. WARN выше)
- By-ID тесты пропускаются (SKIP), если list пустой (`supplier-orders`, `customer-orders`)
- Отдельные сервисы могут отдавать **503**, если pod не поднят

## Установка k6

```bash
# macOS
brew install k6

# Linux (deb)
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg \
  --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | \
  sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update && sudo apt-get install k6
```

## Client contour (опционально)

Client GET (`/api/client/profile`, vehicles, reviews) идут через **другие** gateway (`:8091`/`:8093`). Для них нужен отдельный манифест и скрипт — пока не включены в этот пакет.
