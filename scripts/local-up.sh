#!/usr/bin/env bash
# Поднимает локальный dev-стенд через docker compose инкрементально:
# останавливает стек, пересобирает только те образы, где изменился контекст сборки
# (BuildKit через кэш пропускает неизменённые), стартует, сносит старые версии
# пересобранных образов (dangling) и применяет миграции, seed, admin-пользователя.
#
# Usage:
#   ./scripts/local-up.sh [--full] [--volumes]
#
#   --full     снести ВСЕ образы проекта (--rmi all) и пересобрать всё с нуля.
#              По умолчанию пересобираются только сервисы с изменениями.
#   --volumes  дополнительно удалить data-тома (postgres_data, clickhouse_data) —
#              БД начнётся с нуля и миграции/seed выполнятся заново.
#              По умолчанию данные сохраняются (миграции/seed пропускаются, если БД заполнена).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

FULL=0
VOLUMES=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --full) FULL=1 ;;
    --volumes) VOLUMES=1 ;;
    --help|-h) sed -n '2,15p' "$0"; exit 0 ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
  shift
done

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "ERROR: required command not found: $1" >&2
    exit 1
  }
}

log() { echo "==> $*"; }

need_cmd docker

COMPOSE=(docker compose)
if ! docker compose version >/dev/null 2>&1; then
  need_cmd docker-compose
  COMPOSE=(docker-compose)
fi

# Ограничиваем число параллельных сборок: 22 Go-компилятора одновременно
# съедают всю RAM (OOM: compile: signal: killed). 4 хватает, чтобы не упереться в память.
export COMPOSE_PARALLEL_LIMIT="${COMPOSE_PARALLEL_LIMIT:-4}"

log "Waiting for docker daemon"
for _ in $(seq 1 30); do
  docker info >/dev/null 2>&1 && break
  sleep 2
done
docker info >/dev/null 2>&1 || {
  echo "ERROR: docker daemon is not ready" >&2
  exit 1
}

log "Stopping stack"
if [[ "$VOLUMES" -eq 1 ]]; then
  "${COMPOSE[@]}" down -v --remove-orphans
else
  "${COMPOSE[@]}" down --remove-orphans
fi

if [[ "$FULL" -eq 1 ]]; then
  log "docker compose up -d --build --force-recreate (full rebuild: removing all project images)"
  "${COMPOSE[@]}" up -d --build --force-recreate
else
  log "docker compose up -d --build (incremental: only changed services rebuild)"
  "${COMPOSE[@]}" up -d --build
fi

PROJECT="$("${COMPOSE[@]}" config --format json 2>/dev/null | python3 -c 'import json,sys; print(json.load(sys.stdin)["name"])' 2>/dev/null || echo dealer)"

log "Removing old versions of rebuilt service images (dangling, only ${PROJECT}-*)"
docker image prune -f --filter "label=com.docker.compose.project=${PROJECT}"

log "Waiting for Postgres"
for _ in $(seq 1 60); do
  "${COMPOSE[@]}" exec -T postgres pg_isready -U dealer >/dev/null 2>&1 && break
  sleep 5
done
"${COMPOSE[@]}" exec -T postgres pg_isready -U dealer >/dev/null 2>&1 || {
  echo "ERROR: Postgres is not ready" >&2
  exit 1
}

TABLE_COUNT="$("${COMPOSE[@]}" exec -T postgres env PGPASSWORD="${POSTGRES_PASSWORD:-changeme}" \
  psql -U dealer -d dealer -tAc "SELECT count(*) FROM information_schema.tables WHERE table_schema NOT IN ('pg_catalog','information_schema')" | tr -d '[:space:]')"

if [[ "${TABLE_COUNT:-0}" == "0" ]]; then
  log "Applying SQL migrations"
  for f in "$ROOT"/migrations/*.up.sql; do
    [[ -f "$f" ]] || continue
    log "migrate $(basename "$f")"
    "${COMPOSE[@]}" exec -T postgres env PGPASSWORD="${POSTGRES_PASSWORD:-changeme}" \
      psql -U dealer -d dealer -v ON_ERROR_STOP=1 -f - <"$f"
  done
else
  log "Skipping migrations (DB already initialized)"
fi

if [[ "${TABLE_COUNT:-0}" == "0" ]]; then
  log "Applying seed_volume test data"
  POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-changeme}" "$ROOT/migrations/seed_volume/apply.sh" --compose
else
  log "Skipping seed_volume (DB already initialized)"
fi

if [[ "${TABLE_COUNT:-0}" == "0" ]]; then
  log "Creating admin user (seed-admin)"
  "${COMPOSE[@]}" run --rm --no-deps --entrypoint /seed-admin \
    -e POSTGRES_DSN="postgres://dealer:${POSTGRES_PASSWORD:-changeme}@postgres:5432/dealer?sslmode=disable" \
    auth-service
fi

log "Stack status"
"${COMPOSE[@]}" ps

cat <<EOF

=== Dealer local dev stand ===
Employee UI/API:   http://localhost:8080
Gateway API:       http://localhost:8090
Client public API: http://localhost:8091
Client protected:  http://localhost:8093
Client UI:         http://localhost:3001

Postgres: localhost:5433  Redis: 6379  Kafka: 9092  ClickHouse HTTP: 8123

Логины (seed):
  Employee: admin@dealer.local / admin123
            vol.employee1@test.dealer.local / Test1234!
  Client:   vol.client1@test.dealer.local / Test1234!
EOF
