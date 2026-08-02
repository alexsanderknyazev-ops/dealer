#!/usr/bin/env bash
# Поднимает полный dealer-стек через docker compose (локально или в GitHub Codespaces):
# собирает образы, стартует контейнеры, применяет миграции и seed.
#
# Usage:
#   ./scripts/codespaces-up.sh [--build] [--skip-migrate] [--skip-seed]
#
#   --build   пересобрать образы (первый запуск; может занять 15-30 мин)
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

BUILD=0
SKIP_MIGRATE=0
SKIP_SEED=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --build) BUILD=1 ;;
    --skip-migrate) SKIP_MIGRATE=1 ;;
    --skip-seed) SKIP_SEED=1 ;;
    --help|-h) sed -n '2,9p' "$0"; exit 0 ;;
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

log "Waiting for docker daemon"
for _ in $(seq 1 30); do
  docker info >/dev/null 2>&1 && break
  sleep 2
done
docker info >/dev/null 2>&1 || {
  echo "ERROR: docker daemon is not ready" >&2
  exit 1
}

if [[ "$BUILD" -eq 1 ]]; then
  log "docker compose up -d --build"
  "${COMPOSE[@]}" up -d --build
else
  log "docker compose up -d"
  "${COMPOSE[@]}" up -d
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
  psql -U dealer -d dealer -tAc "SELECT count(*) FROM information_schema.tables WHERE table_schema='public'" | tr -d '[:space:]')"

if [[ "$SKIP_MIGRATE" -eq 0 && "${TABLE_COUNT:-0}" == "0" ]]; then
  log "Applying SQL migrations"
  for f in "$ROOT"/migrations/*.up.sql; do
    [[ -f "$f" ]] || continue
    log "migrate $(basename "$f")"
    "${COMPOSE[@]}" exec -T postgres env PGPASSWORD="${POSTGRES_PASSWORD:-changeme}" \
      psql -U dealer -d dealer -v ON_ERROR_STOP=1 -f - <"$f"
  done
else
  log "Skipping migrations (already applied or --skip-migrate)"
fi

if [[ "$SKIP_SEED" -eq 0 && "${TABLE_COUNT:-0}" == "0" ]]; then
  log "Applying seed_volume test data"
  POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-changeme}" "$ROOT/migrations/seed_volume/apply.sh" --compose
else
  log "Skipping seed_volume (already applied or --skip-seed)"
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

=== Dealer dev stand ===
Employee UI/API:   http://localhost:8080
Gateway API:       http://localhost:8090
Client public API: http://localhost:8091
Client protected:  http://localhost:8093
Client UI:         http://localhost:3001

Postgres: localhost:5433  Redis: 6379  Kafka: 9092  ClickHouse HTTP: 8123

В GitHub Codespaces порты пробрасываются автоматически:
  http://localhost:8080 -> https://8080-<имя>.github.dev
Сделать порт публичным: панель Ports -> правый клик -> Port Visibility -> Public.

Логины (seed):
  Employee: admin@dealer.local / admin123
            vol.employee1@test.dealer.local / Test1234!
  Client:   vol.client1@test.dealer.local / Test1234!
EOF
