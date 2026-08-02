#!/usr/bin/env bash
# Применяет объёмные SQL-фикстуры (идемпотентно).
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ "${1:-}" == "--k8s" ]]; then
  POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-changeme}"
  exec_sql() {
    kubectl -n dealer exec -i deployment/postgres -- \
      env PGPASSWORD="$POSTGRES_PASSWORD" psql -U dealer -d dealer -v ON_ERROR_STOP=1 -f -
  }
elif [[ "${1:-}" == "--compose" ]]; then
  POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-changeme}"
  exec_sql() {
    docker compose exec -T postgres env PGPASSWORD="$POSTGRES_PASSWORD" psql -U dealer -d dealer -v ON_ERROR_STOP=1 -f -
  }
elif [[ -n "${POSTGRES_DSN:-}" ]]; then
  exec_sql() {
    psql "$POSTGRES_DSN" -v ON_ERROR_STOP=1 -f -
  }
else
  echo "Set POSTGRES_DSN or run: $0 --k8s | --compose" >&2
  exit 1
fi

echo "seed_volume apply — $(date -u '+%Y-%m-%d %H:%M:%S UTC')"

for f in \
  "$DIR/01_auth.sql" \
  "$DIR/02_employees.sql" \
  "$DIR/03_dealerpoints.sql" \
  "$DIR/04_brands.sql" \
  "$DIR/05_customers.sql" \
  "$DIR/06_vehicles.sql" \
  "$DIR/07_parts.sql" \
  "$DIR/08_works.sql" \
  "$DIR/09_deals.sql" \
  "$DIR/10_workorders.sql" \
  "$DIR/11_clientauth.sql" \
  "$DIR/12_clients.sql" \
  "$DIR/13_reviews.sql" \
  "$DIR/14_employee_reviews.sql" \
  "$DIR/15_employee_statistics.sql" \
  "$DIR/16_client_statistics.sql" \
  "$DIR/17_appointments.sql" \
  "$DIR/18_part_orders.sql"; do
  echo "==> $(basename "$f")"
  exec_sql < "$f"
done

echo "Done. See migrations/seed_volume/README.md"
