#!/usr/bin/env bash
# Поднимает весь локальный dev-стенд одной командой:
# docker compose up, миграции, seed и admin-пользователь (если БД пустая).
#
# Usage:
#   ./scripts/local-up.sh [--build] [--full] [--volumes]
#
#   (без флагов)  поднять стек из уже собранных образов, без Docker Hub.
#   --build       пересобрать изменившиеся сервисы (--pull never: без registry).
#                 Если сборка не удалась — стартует существующие образы.
#   --full        пересоздать контейнеры и пересобрать все сервисы.
#   --volumes     удалить data-тома (postgres_data, clickhouse_data) —
#                 БД с нуля, миграции/seed заново.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

BUILD=0
FULL=0
VOLUMES=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --build) BUILD=1 ;;
    --full) FULL=1 ;;
    --volumes) VOLUMES=1 ;;
    --help|-h) sed -n '2,16p' "$0"; exit 0 ;;
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

# Не ходим в registry, если образы уже есть локально.
PULL_NEVER=()
if "${COMPOSE[@]}" up --help 2>/dev/null | grep -q -- '--pull'; then
  PULL_NEVER=(--pull never)
fi

compose_up() {
  "${COMPOSE[@]}" up -d --remove-orphans "${PULL_NEVER[@]}" "$@"
}

if [[ "$VOLUMES" -eq 1 ]]; then
  log "Stopping stack and removing volumes"
  "${COMPOSE[@]}" down -v --remove-orphans
fi

DID_BUILD=0
if [[ "$FULL" -eq 1 ]]; then
  log "docker compose up -d --build --force-recreate"
  if compose_up --build --force-recreate; then
    DID_BUILD=1
  else
    log "Build failed (registry unavailable?). Starting existing images"
    compose_up --force-recreate
  fi
elif [[ "$BUILD" -eq 1 ]]; then
  log "docker compose up -d --build"
  if compose_up --build; then
    DID_BUILD=1
  else
    log "Build failed (registry unavailable?). Starting existing images"
    compose_up
  fi
else
  log "docker compose up -d (existing images)"
  if ! compose_up; then
    log "Images missing, trying --build"
    compose_up --build
    DID_BUILD=1
  fi
fi

if [[ "$DID_BUILD" -eq 1 ]]; then
  PROJECT="$("${COMPOSE[@]}" config --format json 2>/dev/null | python3 -c 'import json,sys; print(json.load(sys.stdin)["name"])' 2>/dev/null || echo dealer)"
  log "Removing old versions of rebuilt service images (dangling, only ${PROJECT}-*)"
  docker image prune -f --filter "label=com.docker.compose.project=${PROJECT}"
fi

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
