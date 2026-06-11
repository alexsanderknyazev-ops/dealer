#!/usr/bin/env bash
# Проверка записи в PostgreSQL после API-теста.
# Usage:
#   export POSTGRES_DSN='postgres://dealer:pass@127.0.0.1:5433/dealer?sslmode=disable'
#   ./qa/api-testing/scripts/db-check.sh customer <UUID>
#   ./qa/api-testing/scripts/db-check.sh deal <UUID>
#   ./qa/api-testing/scripts/db-check.sh wo-movement <WO_UUID>
set -euo pipefail

if [[ -z "${POSTGRES_DSN:-}" ]]; then
  echo "Set POSTGRES_DSN" >&2
  exit 1
fi

ENTITY="${1:-}"
ID="${2:-}"
if [[ -z "$ENTITY" || -z "$ID" ]]; then
  echo "Usage: $0 <entity> <uuid>" >&2
  echo "Entities: user, customer, vehicle, deal, brand, part, movement, wo, wo-movement, work, employee, client, clientauth, review, employee-review, deal-event, reg-event" >&2
  exit 1
fi

case "$ENTITY" in
  user)
    psql "$POSTGRES_DSN" -c "SELECT id, email, role, created_at FROM auth.users WHERE id = '$ID';"
    ;;
  customer)
    psql "$POSTGRES_DSN" -c "SELECT * FROM customers.customers WHERE id = '$ID';"
    ;;
  vehicle)
    psql "$POSTGRES_DSN" -c "SELECT id, vin, model, year, status, brand_id FROM vehicles.vehicles WHERE id = '$ID';"
    ;;
  deal)
    psql "$POSTGRES_DSN" -c "SELECT id, customer_id, vehicle_id, amount, stage FROM deals.deals WHERE id = '$ID';"
    ;;
  brand)
    psql "$POSTGRES_DSN" -c "SELECT * FROM brands.brands WHERE id = '$ID';"
    ;;
  part)
    psql "$POSTGRES_DSN" -c "SELECT p.id, p.sku, p.name, p.quantity, ps.warehouse_id, ps.quantity AS wh_qty
      FROM parts.parts p
      LEFT JOIN parts.part_stock ps ON ps.part_id = p.id
      WHERE p.id = '$ID';"
    ;;
  movement)
    psql "$POSTGRES_DSN" -c "SELECT md.*, (SELECT COUNT(*) FROM parts.movement_document_lines l WHERE l.document_id = md.id) AS lines
      FROM parts.movement_documents md WHERE md.id = '$ID';"
    ;;
  wo)
    psql "$POSTGRES_DSN" -c "SELECT id, order_number, status, parts_issued, movement_document_id, movement_document_status
      FROM workorders.work_orders WHERE id = '$ID';"
    ;;
  wo-movement)
    psql "$POSTGRES_DSN" -c "
      SELECT wo.id, wo.order_number, wo.parts_issued, wo.movement_document_status,
             md.id AS doc_id, md.status AS doc_status,
             (SELECT quantity FROM parts.part_stock ps
              JOIN workorders.work_order_parts wop ON wop.part_id = ps.part_id::text
              WHERE wop.work_order_id = wo.id LIMIT 1) AS stock_qty
      FROM workorders.work_orders wo
      LEFT JOIN parts.movement_documents md ON md.id = wo.movement_document_id
      WHERE wo.id = '$ID';"
    ;;
  work)
    psql "$POSTGRES_DSN" -c "SELECT id, code, name, labor_hours, unit_price, category FROM works.works WHERE id = '$ID';"
    ;;
  employee)
    psql "$POSTGRES_DSN" -c "SELECT id, user_id, full_name, position, department, active FROM employees.employees WHERE id = '$ID';"
    ;;
  client)
    psql "$POSTGRES_DSN" -c "SELECT c.*, (SELECT COUNT(*) FROM clients.client_vehicles cv WHERE cv.client_id = c.id) AS vehicles
      FROM clients.clients c WHERE c.id = '$ID';"
    ;;
  clientauth)
    psql "$POSTGRES_DSN" -c "SELECT id, email, full_name, LEFT(password_hash, 20) AS hash_prefix FROM clientauth.users WHERE id = '$ID';"
    ;;
  review)
    psql "$POSTGRES_DSN" -c "SELECT * FROM reviews.reviews WHERE id = '$ID';"
    ;;
  employee-review)
    psql "$POSTGRES_DSN" -c "SELECT * FROM employee_reviews.reviews WHERE review_id = '$ID';"
    ;;
  deal-event)
    psql "$POSTGRES_DSN" -c "SELECT * FROM employee_statistics.deal_events WHERE deal_id = '$ID';"
    ;;
  reg-event)
    psql "$POSTGRES_DSN" -c "SELECT * FROM client_statistics.client_registration_events WHERE user_id = '$ID';"
    ;;
  *)
    echo "Unknown entity: $ENTITY" >&2
    exit 1
    ;;
esac
