# Локальный дев-стенд (docker compose на своей машине)

Быстрый дев-стенд для Mac/Linux: весь стек dealer поднимается через `docker compose`
прямо на твоей машине. Удобен, когда нужно разрабатывать без облака и без
установки minikube/Kubernetes.

Что поднимается (тот же стек, что и в Codespaces):

- Postgres, Redis, Zookeeper, Kafka, ClickHouse;
- все сервисы dealer (включая gateway, errors-ingest, scheduler);
- клиентский UI (Vite в контейнере `client-ui`);
- employee-UI встроен в `auth-service` (порт 8080).

## Быстрый старт

```bash
./scripts/local-up.sh
```

или

```bash
make local-up
```

Скрипт:

1. поднимает весь стек из уже собранных образов (`docker compose up -d`, без Docker Hub);
2. применяет миграции и seed_volume (пропускает, если БД уже заполнена);
3. создаёт admin-пользователя;
4. печатает адреса и логины.

Пересборка образов — отдельным флагом (`BUILD=1` / `--build`): BuildKit-кэш
пропускает неизменённые сервисы. Если registry недоступен, скрипт стартует
существующие образы, а не оставляет стенд выключенным.

### Оптимизация сборки

Все Go-сервисы собираются с общим кэшем модулей и компиляции (BuildKit cache mounts:
`/go/pkg/mod` и `/root/.cache/go-build`) — зависимости скачиваются один раз на машину,
а не по 22 раза, и пересборка изменённого сервиса идёт секунды. Образы собраны с
`-trimpath -ldflags="-w -s"` (меньше размер, без путей сборки в бинарнике).

## Пересборка образов

```bash
make local-up BUILD=1
```

или `./scripts/local-up.sh --build` — пересобрать изменившиеся сервисы.

```bash
make local-up FULL=1
```

или `./scripts/local-up.sh --full` — пересоздать контейнеры и пересобрать все сервисы
(долго — 15–30 мин). Обычно не нужно.

## Сброс данных БД

```bash
./scripts/local-up.sh --volumes
```

или

```bash
make local-up VOLUMES=1
```

Дополнительно удаляет data-тома (`postgres_data`, `clickhouse_data`) — БД начнётся
с нуля, миграции и seed выполнятся заново.

Комбинация: `make local-up FULL=1 VOLUMES=1` — полностью чистый стенд с нуля.

## Доступ

| Сервис | URL |
|---|---|
| Employee UI/API | http://localhost:8080 |
| Gateway API | http://localhost:8090 |
| Client public API | http://localhost:8091 |
| Client protected | http://localhost:8093 |
| Client UI | http://localhost:3001 |

Инфраструктура: Postgres `localhost:5433`, Redis `6379`, Kafka `9092`,
ClickHouse HTTP `8123`.

Логины (seed-данные):

| Зона | Логин | Пароль |
|---|---|---|
| Employee | `admin@dealer.local` | `admin123` |
| Employee | `vol.employee1@test.dealer.local` | `Test1234!` |
| Client | `vol.client1@test.dealer.local` | `Test1234!` |

## Остановить стенд

```bash
docker compose down
```

С данными: `docker compose down -v` (удалит и тома БД).

## Диагностика

| Проблема | Решение |
|---|---|
| Сервис не поднялся | `docker compose ps`, `docker compose logs <service>` |
| Не хватает места | `./scripts/docker-cleanup.sh` |
| Ошибка миграции на старой БД | `make local-up VOLUMES=1` (сброс данных) |

## Связь с другими стендами

- GitHub Codespaces — облачный дев-стенд в один клик (см. `docs/deploy/codespaces.md`);
- НТ-стенд на Ubuntu в локальной сети — через minikube (см. `docs/deploy/nt-lan.md`).
