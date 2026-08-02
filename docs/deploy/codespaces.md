# Дев-стенд в GitHub Codespaces

Cloud-контейнер со всем стеком dealer, поднимаемый в один клик. Для разработки и быстрого
ручного тестирования. Не является 24/7-стендом: останавливается по бездействию и удаляется
после 30 дней без запуска.

Полный стек (тот же, что `docker compose up`):

- Postgres, Redis, Zookeeper, Kafka, ClickHouse;
- все сервисы dealer (включая gateway, errors-ingest, scheduler);
- клиентский UI (Vite-dev-сервер в контейнере `client-ui`);
- employee-UI встроен в `auth-service` (порт 8080).

## Как создать

1. Открой репозиторий `alexsanderknyazev-ops/dealer` на GitHub.
2. **Code → Codespaces → Create codespace on main**.
3. Дождись окончания первого запуска (сборка 22 образов — 15–30 мин, идёт в фоне
   с прогрессом в панели «postCreateCommand»).
4. Когда сборка завершится, стенд доступен:

| Сервис | Локальный адрес | Публичный (в Codespaces) |
|---|---|---|
| Employee UI/API | http://localhost:8080 | https://8080-`<имя>`.github.dev |
| Gateway API | http://localhost:8090 | https://8090-`<имя>`.github.dev |
| Client public API | http://localhost:8091 | https://8091-`<имя>`.github.dev |
| Client protected | http://localhost:8093 | https://8093-`<имя>`.github.dev |
| Client UI | http://localhost:3001 | https://3001-`<имя>`.github.dev |

Порты 8080/8090/8091/8093/3001 настроены как **public** — ссылками можно делиться
(у кого есть URL, тот получит доступ; это dev-стенд).

## Клиентский UI

В контейнере `client-ui` уже крутится Vite (порт 3001). Дополнительно можно запустить
dev-сервер на хосте codespace, если нужно быстро пересобирать фронт без пересборки образа:

```bash
./scripts/codespaces-client-ui.sh   # npm install + npm run dev (порт 3001)
```

## Логины (seed-данные)

| Зона | Логин | Пароль |
|---|---|---|
| Employee | `admin@dealer.local` | `admin123` |
| Employee | `vol.employee1@test.dealer.local` | `Test1234!` |
| Client | `vol.client1@test.dealer.local` | `Test1234!` |

## Обновление

Стенд в codespace живёт по коммиту `main`, на котором создан. Чтобы обновить:

```bash
git pull --ff-only
./scripts/codespaces-up.sh --build   # пересборка изменившихся образов
```

## Лимиты и стоимость (важно)

- Бесплатно **120 core-часов/мес** (личный аккаунт).
  - 4-core машина ≈ **30 часов** работы в месяц;
  - 8-core ≈ 15 часов; 2-core ≈ 60 часов (но сборка на 2-core очень медленная).
- Codespace **останавливается сам** через 30 мин бездействия (настраивается: Machine → Settings).
- Остановленный codespace **не тратит** часы. Данные (volumes БД) сохраняются.
- Codespace **автоудаляется** после 30 дней без запуска — образы и данные пропадут.
- Сборка образов считается в лимит часов. Поэтому не удаляй codespace без нужды —
  пересоздание = новая дорогая сборка.

Советы:
- Останавливай codespace, когда закончил работу.
- Для «всегда-включённого» тест-стенда используй Oracle VM (см. `docs/deploy/oracle-always-free.md`).

## Диагностика

| Проблема | Решение |
|---|---|
| Порты не открываются | Проверь панель **Ports**; если не видно — `./scripts/codespaces-up.sh` |
| `ImagePullBackOff` / падение контейнера | `docker compose ps`, `docker compose logs <service>` |
| Нужен чистый стенд | `docker compose down -v` и заново `./scripts/codespaces-up.sh --build` |
| Мало места | `./scripts/docker-cleanup.sh` (build cache + неиспользуемые образы; volumes не трогает) |
