#!/usr/bin/env bash
# Применяет QA SQL-фикстуры (идемпотентно).
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ -z "${POSTGRES_DSN:-}" ]]; then
  echo "Set POSTGRES_DSN (see qa/api-testing/_shared/prerequisites.md)" >&2
  exit 1
fi

run() {
  local f="$1"
  echo "==> $f"
  psql "$POSTGRES_DSN" -v ON_ERROR_STOP=1 -f "$f"
}

echo "QA fixtures apply — $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
echo "DSN: ${POSTGRES_DSN%%@*}@..." 

# Preflight: dealer point from full-seed exists
if ! psql "$POSTGRES_DSN" -tAc "SELECT 1 FROM dealerpoints.dealer_points WHERE id = '10000000-0000-4000-8000-000000000001'" | grep -q 1; then
  echo "WARN: seed_dealer_brands not found. Run: make full-seed" >&2
  if [[ "${QA_FIXTURES_SKIP_SEED_CHECK:-}" != "1" ]]; then
    echo "Set QA_FIXTURES_SKIP_SEED_CHECK=1 to apply anyway." >&2
    exit 1
  fi
fi

run "$DIR/01_employee_users.sql"
run "$DIR/02_customers_vehicles.sql"
run "$DIR/03_parts_stock.sql"
run "$DIR/04_deals.sql"
run "$DIR/08_works_employees.sql"
run "$DIR/05_work_orders.sql"
run "$DIR/06_client.sql"
run "$DIR/07_reviews_stats.sql"
run "$DIR/08_client_review_invitation.sql"

echo ""
echo "Done. Credentials: qa/api-testing/fixtures/README.md"
echo "IDs: qa/api-testing/fixtures/manifest.json"
