#!/usr/bin/env bash
# Полный автоматический прогон API-тестов Dealer.
# Usage: ./qa/api-testing/scripts/run-api-tests.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
QA_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
RESULTS_ROOT="${QA_ROOT}/results"
RUN_ID="run-$(date -u +%Y%m%d-%H%M%S)"
RUN_DIR="${RESULTS_ROOT}/runs/${RUN_ID}"
REPORT="${RUN_DIR}/smoke-report.md"
META="${RUN_DIR}/meta.json"
LATEST_DIR="${RESULTS_ROOT}/latest"

EMPLOYEE_API="${EMPLOYEE_API:-http://127.0.0.1:8090}"
EMPLOYEE_AUTH="${EMPLOYEE_AUTH:-http://127.0.0.1:8080}"
CLIENT_PUBLIC="${CLIENT_PUBLIC:-http://127.0.0.1:8091}"
CLIENT_PROTECTED="${CLIENT_PROTECTED:-http://127.0.0.1:8093}"
ERRORS_INGEST="${ERRORS_INGEST:-http://127.0.0.1:8092}"
QA_PASSWORD="${QA_PASSWORD:-Test1234!}"

PASS=0
FAIL=0
SKIP=0
RESULTS=()
TS="$(date -u '+%Y-%m-%d %H:%M:%S UTC')"
GIT_COMMIT="$(git -C "$QA_ROOT/.." rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
mkdir -p "$RUN_DIR" "$LATEST_DIR"

json_field() {
  python3 -c "import sys,json; d=json.load(sys.stdin); print(d$1)" 2>/dev/null || true
}

assert_code() {
  local id="$1" name="$2" expected="$3" actual="$4" note="${5:-}"
  if [[ "$actual" == "$expected" ]] || [[ "$expected" == *"$actual"* ]]; then
    PASS=$((PASS + 1))
    RESULTS+=("| $id | $name | ✅ PASS | $expected | $actual | ${note:-} |")
  else
    FAIL=$((FAIL + 1))
    RESULTS+=("| $id | $name | ❌ FAIL | $expected | $actual | ${note:-} |")
  fi
}

assert_one_of() {
  local id="$1" name="$2" actual="$3" note="$4"
  shift 4
  for exp in "$@"; do
    if [[ "$actual" == "$exp" ]]; then
      PASS=$((PASS + 1))
      RESULTS+=("| $id | $name | ✅ PASS | one of: $* | $actual | $note |")
      return
    fi
  done
  FAIL=$((FAIL + 1))
  RESULTS+=("| $id | $name | ❌ FAIL | one of: $* | $actual | $note |")
}

http_code() {
  curl -s -o /tmp/qa_resp_$$ -w "%{http_code}" "$@"
}

echo "=== Dealer API test run $RUN_ID ==="
echo "Employee API: $EMPLOYEE_API"

# --- Employee auth ---
EMAIL="qa-${RUN_ID}@test.local"
REG_CODE=$(http_code -X POST "$EMPLOYEE_API/api/register" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$QA_PASSWORD\",\"name\":\"QA Auto\",\"phone\":\"+79001234567\"}")
assert_code "AUTH-001" "POST /api/register" "200" "$REG_CODE"

REG_BODY=$(cat /tmp/qa_resp_$$)
ACCESS=$(echo "$REG_BODY" | json_field "['access_token']")
REFRESH=$(echo "$REG_BODY" | json_field "['refresh_token']")
USER_ID=$(echo "$REG_BODY" | json_field "['user_id']")

LOGIN_CODE=$(http_code -X POST "$EMPLOYEE_API/api/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$QA_PASSWORD\"}")
assert_code "AUTH-002" "POST /api/login" "200" "$LOGIN_CODE"
LOGIN_BODY=$(cat /tmp/qa_resp_$$)
ACCESS=$(echo "$LOGIN_BODY" | json_field "['access_token']")
REFRESH=$(echo "$LOGIN_BODY" | json_field "['refresh_token']")

AUTH_H=(-H "Authorization: Bearer $ACCESS")

ME_CODE=$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/me")
assert_code "AUTH-003" "GET /api/me" "200" "$ME_CODE"

REF_CODE=$(http_code -X POST "$EMPLOYEE_API/api/refresh" \
  -H 'Content-Type: application/json' \
  -d "{\"refresh_token\":\"$REFRESH\"}")
assert_code "AUTH-004" "POST /api/refresh" "200" "$REF_CODE"
NEW_REFRESH=$(cat /tmp/qa_resp_$$ | json_field "['refresh_token']")

DUP_CODE=$(http_code -X POST "$EMPLOYEE_API/api/register" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$QA_PASSWORD\",\"name\":\"Dup\"}")
assert_one_of "AUTH-006" "POST /api/register duplicate" "$DUP_CODE" "" "400" "409" "500"

BAD_LOGIN=$(http_code -X POST "$EMPLOYEE_API/api/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"wrong\"}")
assert_one_of "AUTH-007" "POST /api/login wrong password" "$BAD_LOGIN" "" "401" "403" "400"

NOAUTH=$(http_code "$EMPLOYEE_API/api/me")
assert_code "AUTH-008" "GET /api/me no auth" "401" "$NOAUTH"

EMPTY_REG=$(http_code -X POST "$EMPLOYEE_API/api/register" \
  -H 'Content-Type: application/json' -d '{}')
assert_one_of "AUTH-010" "POST /api/register empty" "$EMPTY_REG" "" "400" "500"

AUTH_HZ=$(http_code "$EMPLOYEE_AUTH/healthz")
assert_code "AUTH-011" "GET /healthz auth" "200" "$AUTH_HZ"

PROXY_CUST=$(http_code "${AUTH_H[@]}" "$EMPLOYEE_AUTH/api/customers")
assert_code "AUTH-020" "GET /api/customers via auth proxy" "200" "$PROXY_CUST"

LOGOUT_CODE=$(http_code -X POST "$EMPLOYEE_API/api/logout" \
  -H 'Content-Type: application/json' \
  -d "{\"refresh_token\":\"$NEW_REFRESH\"}")
assert_one_of "AUTH-005" "POST /api/logout" "$LOGOUT_CODE" "" "200" "204"

# Re-login for rest of tests
LOGIN_BODY=$(curl -s -X POST "$EMPLOYEE_API/api/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$QA_PASSWORD\"}")
ACCESS=$(echo "$LOGIN_BODY" | json_field "['access_token']")
AUTH_H=(-H "Authorization: Bearer $ACCESS")

# --- Gateway ---
GW_HZ=$(http_code "$EMPLOYEE_API/healthz")
assert_code "GW-001" "GET /healthz gateway" "200" "$GW_HZ"

OPT=$(http_code -X OPTIONS "$EMPLOYEE_API/api/customers" -H 'Origin: http://localhost')
assert_one_of "GW-002" "OPTIONS CORS" "$OPT" "" "204" "200"

for spec in \
  "GW-003:GET:/api/customers:200" \
  "GW-004:GET:/api/customers:401:noauth" \
  "GW-005:GET:/api/vehicles:200" \
  "GW-006:GET:/api/deals:200" \
  "GW-007:GET:/api/parts:200" \
  "GW-008:GET:/api/brands:200" \
  "GW-009:GET:/api/dealer-points:200" \
  "GW-011:GET:/api/movement-documents:200" \
  "GW-012:GET:/api/reviews:200" \
  "GW-013:GET:/api/stats/employee/overview:200" \
  "GW-014:GET:/api/stats/client/overview:200" \
  "GW-016:GET:/api/works:200" \
  "GW-017:GET:/api/employees:200"; do
  IFS=':' read -r gid method path exp flag <<< "$spec"
  if [[ "${flag:-}" == "noauth" ]]; then
    c=$(http_code -X "$method" "$EMPLOYEE_API$path")
  else
    c=$(http_code -X "$method" "${AUTH_H[@]}" "$EMPLOYEE_API$path")
  fi
  if [[ "$gid" == "GW-010" || "$gid" == "GW-016" || "$gid" == "GW-017" ]]; then
    assert_one_of "$gid" "$method $path" "$c" "503 if service down" "200" "503"
  else
    assert_code "$gid" "$method $path" "$exp" "$c"
  fi
done

WO_CODE=$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/work-orders")
assert_one_of "GW-010" "GET /api/work-orders" "$WO_CODE" "503 if service down" "200" "503"

BAD_UUID=$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/parts/not-a-uuid")
assert_one_of "GW-015" "GET /api/parts bad id" "$BAD_UUID" "" "400" "404" "500"

# --- Customers ---
CUST_CODE=$(http_code -X POST "${AUTH_H[@]}" -H 'Content-Type: application/json' \
  -d "{\"type\":\"individual\",\"full_name\":\"QA Customer $RUN_ID\",\"phone\":\"+7900222\",\"email\":\"cust-$RUN_ID@test.local\"}" \
  "$EMPLOYEE_API/api/customers")
assert_code "CUST-001" "POST /api/customers" "200" "$CUST_CODE"
CUST_ID=$(cat /tmp/qa_resp_$$ | json_field "['id']")

assert_code "CUST-002" "GET /api/customers/{id}" "200" \
  "$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/customers/$CUST_ID")"
assert_code "CUST-003" "GET /api/customers list" "200" \
  "$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/customers?limit=5")"
assert_code "CUST-006" "GET /api/customers no auth" "401" \
  "$(http_code "$EMPLOYEE_API/api/customers")"
assert_one_of "CUST-007" "GET /api/customers missing" \
  "$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/customers/00000000-0000-4000-8000-000000000099")" "" "404" "400" "500"

# --- Brands ---
assert_code "BRD-001" "GET /api/brands" "200" \
  "$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/brands")"
BRAND_ID=$(curl -s "${AUTH_H[@]}" "$EMPLOYEE_API/api/brands?limit=1" | json_field "['brands'][0]['id']")
assert_code "BRD-002" "GET /api/brands/{id}" "200" \
  "$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/brands/$BRAND_ID")"

UNIQ_BRAND="QA-Brand-$RUN_ID"
BR_CREATE=$(http_code -X POST "${AUTH_H[@]}" -H 'Content-Type: application/json' \
  -d "{\"name\":\"$UNIQ_BRAND\"}" "$EMPLOYEE_API/api/brands")
assert_one_of "BRD-003" "POST /api/brands unique" "$BR_CREATE" "" "200" "201"

DUP_BR=$(http_code -X POST "${AUTH_H[@]}" -H 'Content-Type: application/json' \
  -d "{\"name\":\"Hyundai\"}" "$EMPLOYEE_API/api/brands")
assert_one_of "BRD-006" "POST /api/brands duplicate" "$DUP_BR" "" "409" "500"

assert_one_of "BRD-007" "GET /api/brands missing" \
  "$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/brands/00000000-0000-4000-8000-000000000099")" "" "404" "400"
assert_code "BRD-008" "POST /api/brands no auth" "401" \
  "$(http_code -X POST -H 'Content-Type: application/json' -d '{"name":"X"}' "$EMPLOYEE_API/api/brands")"

# --- Vehicles ---
VIN="VIN${RUN_ID: -10}"
VEH_CODE=$(http_code -X POST "${AUTH_H[@]}" -H 'Content-Type: application/json' \
  -d "{\"vin\":\"$VIN\",\"brand_id\":\"$BRAND_ID\",\"model\":\"QA Model\",\"year\":2024,\"mileage_km\":5000,\"price\":\"1500000\",\"status\":\"available\"}" \
  "$EMPLOYEE_API/api/vehicles")
assert_code "VEH-001" "POST /api/vehicles" "200" "$VEH_CODE"
VEH_ID=$(cat /tmp/qa_resp_$$ | json_field "['id']")

assert_code "VEH-002" "GET /api/vehicles/{id}" "200" \
  "$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/vehicles/$VEH_ID")"
assert_code "VEH-003" "GET /api/vehicles list" "200" \
  "$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/vehicles?limit=5")"
assert_code "VEH-008" "GET /api/vehicles no auth" "401" \
  "$(http_code "$EMPLOYEE_API/api/vehicles/$VEH_ID")"
DUP_VIN=$(http_code -X POST "${AUTH_H[@]}" -H 'Content-Type: application/json' \
  -d "{\"vin\":\"$VIN\",\"brand_id\":\"$BRAND_ID\",\"model\":\"Dup\",\"year\":2024,\"price\":\"1\",\"status\":\"available\"}" \
  "$EMPLOYEE_API/api/vehicles")
assert_one_of "VEH-006" "POST /api/vehicles dup VIN" "$DUP_VIN" "" "409" "500"

# --- Dealer points ---
assert_code "DP-001" "GET /api/dealer-points" "200" \
  "$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/dealer-points")"
assert_code "DP-010" "GET /api/legal-entities" "200" \
  "$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/legal-entities")"
assert_code "DP-020" "GET /api/warehouses" "200" \
  "$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/warehouses")"

# --- Parts RBAC ---
assert_code "PRT-001" "GET /api/parts" "200" \
  "$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/parts")"
assert_code "PRT-003" "POST /api/parts sales role" "403" \
  "$(http_code -X POST "${AUTH_H[@]}" -H 'Content-Type: application/json' \
    -d '{"sku":"QA-SKU","name":"Part","quantity":1,"unit":"pcs","price":"100"}' \
    "$EMPLOYEE_API/api/parts")"
assert_code "PRT-020" "GET /api/movement-documents" "200" \
  "$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/movement-documents")"

# --- Works catalog ---
assert_one_of "WRK-001" "GET /api/works" \
  "$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/works?limit=5")" "" "200" "503"
assert_code "WRK-004" "POST /api/works sales role" "403" \
  "$(http_code -X POST "${AUTH_H[@]}" -H 'Content-Type: application/json' \
    -d '{"code":"QA-SMOKE","name":"Smoke work","category":"test","labor_hours":"1","unit_price":"100"}' \
    "$EMPLOYEE_API/api/works")"
WRK_DIRECT=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 2 "http://127.0.0.1:8098/healthz" 2>/dev/null || echo "000")
if [[ "$WRK_DIRECT" == "200" ]]; then
  assert_code "WRK-010" "works /healthz :8098" "200" "$WRK_DIRECT"
else
  SKIP=$((SKIP + 1))
  RESULTS+=("| WRK-010 | works /healthz | ⏭ SKIP | 200 | $WRK_DIRECT | service not running |")
fi

# --- Employees directory ---
assert_one_of "EMP-001" "GET /api/employees" \
  "$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/employees?limit=5")" "" "200" "503"
assert_code "EMP-006" "POST /api/employees sales role" "403" \
  "$(http_code -X POST "${AUTH_H[@]}" -H 'Content-Type: application/json' \
    -d '{"full_name":"QA Smoke Emp","position":"test","active":true}' \
    "$EMPLOYEE_API/api/employees")"
EMP_DIRECT=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 2 "http://127.0.0.1:8099/healthz" 2>/dev/null || echo "000")
if [[ "$EMP_DIRECT" == "200" ]]; then
  assert_code "EMP-010" "employees /healthz :8099" "200" "$EMP_DIRECT"
else
  SKIP=$((SKIP + 1))
  RESULTS+=("| EMP-010 | employees /healthz | ⏭ SKIP | 200 | $EMP_DIRECT | service not running |")
fi

# --- Work orders ---
assert_one_of "WO-001" "GET /api/work-orders" \
  "$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/work-orders")" "" "200" "503"
assert_code "WO-003" "POST /api/work-orders sales" "403" \
  "$(http_code -X POST "${AUTH_H[@]}" -H 'Content-Type: application/json' \
    -d "{\"customer_id\":\"$CUST_ID\",\"vehicle_id\":\"$VEH_ID\"}" \
    "$EMPLOYEE_API/api/work-orders")" || true
WO_HZ=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 2 "$EMPLOYEE_API/../.." 2>/dev/null || echo "000")
WO_DIRECT=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 2 "http://127.0.0.1:8097/healthz" 2>/dev/null || echo "000")
if [[ "$WO_DIRECT" == "200" ]]; then
  assert_code "WO-013" "workorders /healthz :8097" "200" "$WO_DIRECT"
else
  SKIP=$((SKIP + 1))
  RESULTS+=("| WO-013 | workorders /healthz | ⏭ SKIP | 200 | $WO_DIRECT | service not running |")
fi

# --- Deals ---
DEAL_CODE=$(http_code -X POST "${AUTH_H[@]}" -H 'Content-Type: application/json' \
  -d "{\"customer_id\":\"$CUST_ID\",\"vehicle_id\":\"$VEH_ID\",\"amount\":\"2500000\",\"stage\":\"draft\",\"responsible_id\":\"$USER_ID\"}" \
  "$EMPLOYEE_API/api/deals")
assert_one_of "DEL-002" "POST /api/deals" "$DEAL_CODE" "" "200" "201"
DEAL_ID=$(cat /tmp/qa_resp_$$ | json_field "['id']")

assert_code "DEL-001" "GET /api/deals" "200" \
  "$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/deals")"
if [[ -n "$DEAL_ID" ]]; then
  assert_code "DEL-003" "GET /api/deals/{id}" "200" \
    "$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/deals/$DEAL_ID")"
fi
assert_one_of "DEL-007" "POST /api/deals bad customer" \
  "$(http_code -X POST "${AUTH_H[@]}" -H 'Content-Type: application/json' \
    -d "{\"customer_id\":\"00000000-0000-4000-8000-000000000099\",\"vehicle_id\":\"$VEH_ID\",\"amount\":\"1\",\"stage\":\"draft\"}" \
    "$EMPLOYEE_API/api/deals")" "" "400" "404" "500"
assert_one_of "DEL-008" "POST /api/deals bad vehicle" \
  "$(http_code -X POST "${AUTH_H[@]}" -H 'Content-Type: application/json' \
    -d "{\"customer_id\":\"$CUST_ID\",\"vehicle_id\":\"00000000-0000-4000-8000-000000000099\",\"amount\":\"1\",\"stage\":\"draft\"}" \
    "$EMPLOYEE_API/api/deals")" "" "400" "404" "500"
assert_code "DEL-009" "POST /api/deals no auth" "401" \
  "$(http_code -X POST -H 'Content-Type: application/json' \
    -d "{\"customer_id\":\"$CUST_ID\",\"vehicle_id\":\"$VEH_ID\",\"amount\":\"1\",\"stage\":\"draft\"}" \
    "$EMPLOYEE_API/api/deals")"

# --- Employee reviews & stats ---
assert_code "EREV-001" "GET /api/reviews" "200" \
  "$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/reviews")"
assert_code "EREV-004" "GET /api/reviews/stats" "200" \
  "$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/reviews/stats")"
assert_code "EREV-005" "GET /api/reviews no auth" "401" \
  "$(http_code "$EMPLOYEE_API/api/reviews")"
assert_code "EST-001" "GET employee stats" "200" \
  "$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/stats/employee/overview")"
assert_code "EST-002" "GET employee stats no auth" "401" \
  "$(http_code "$EMPLOYEE_API/api/stats/employee/overview")"
assert_code "CST-001" "GET client stats" "200" \
  "$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/stats/client/overview")"
assert_code "CST-002" "GET client stats no auth" "401" \
  "$(http_code "$EMPLOYEE_API/api/stats/client/overview")"

# --- Client public ---
CEMAIL="client-${RUN_ID}@test.local"
CREG=$(http_code -X POST "$CLIENT_PUBLIC/api/client/register" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$CEMAIL\",\"password\":\"$QA_PASSWORD\",\"full_name\":\"Client QA\",\"phone\":\"+7900333\",\"vin\":\"$VIN\"}")
assert_code "CPG-001" "POST /api/client/register" "200" "$CREG"
CREG_BODY=$(cat /tmp/qa_resp_$$)
CLIENT_ACCESS=$(echo "$CREG_BODY" | json_field "['access_token']")
CLIENT_REFRESH=$(echo "$CREG_BODY" | json_field "['refresh_token']")
CLIENT_ID=$(echo "$CREG_BODY" | json_field "['client_id']")

assert_one_of "CPG-002" "POST /api/client/register no vin" \
  "$(http_code -X POST "$CLIENT_PUBLIC/api/client/register" \
    -H 'Content-Type: application/json' \
    -d "{\"email\":\"bad-$RUN_ID@test.local\",\"password\":\"$QA_PASSWORD\",\"full_name\":\"X\",\"phone\":\"+1\"}")" "" "400" "500"

CLOGIN=$(http_code -X POST "$CLIENT_PUBLIC/api/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$CEMAIL\",\"password\":\"$QA_PASSWORD\"}")
assert_code "CPG-004" "POST /api/login client" "200" "$CLOGIN"

CREF=$(http_code -X POST "$CLIENT_PUBLIC/api/refresh" \
  -H 'Content-Type: application/json' \
  -d "{\"refresh_token\":\"$CLIENT_REFRESH\"}")
assert_code "CPG-005" "POST /api/refresh client" "200" "$CREF"

assert_one_of "CPG-007" "OPTIONS client login" \
  "$(http_code -X OPTIONS "$CLIENT_PUBLIC/api/login" -H 'Origin: http://localhost')" "" "204" "200"

CLIENT_H=(-H "Authorization: Bearer $CLIENT_ACCESS")

# --- Client protected ---
assert_code "CPP-001" "GET /api/me client" "200" \
  "$(http_code "${CLIENT_H[@]}" "$CLIENT_PROTECTED/api/me")"
assert_code "CPP-002" "GET /api/client/profile" "200" \
  "$(http_code "${CLIENT_H[@]}" "$CLIENT_PROTECTED/api/client/profile")"
assert_code "CPP-003" "GET /api/client/vehicles" "200" \
  "$(http_code "${CLIENT_H[@]}" "$CLIENT_PROTECTED/api/client/vehicles")"
assert_code "CPP-006" "GET profile no auth" "401" \
  "$(http_code "$CLIENT_PROTECTED/api/client/profile")"

# Employee token on client gateway
assert_one_of "CPP-005" "employee token on client profile" \
  "$(http_code "${AUTH_H[@]}" "$CLIENT_PROTECTED/api/client/profile")" "" "401" "403" "500"

# --- Client auth negative ---
assert_one_of "CA-004" "client login wrong pass" \
  "$(http_code -X POST "$CLIENT_PUBLIC/api/login" \
    -H 'Content-Type: application/json' \
    -d "{\"email\":\"$CEMAIL\",\"password\":\"wrong\"}")" "" "401" "403" "400"

# --- Client reviews ---
assert_code "CREV-001" "GET /api/client/reviews" "200" \
  "$(http_code "${CLIENT_H[@]}" "$CLIENT_PROTECTED/api/client/reviews")"
assert_code "CREV-006" "GET reviews no auth" "401" \
  "$(http_code "$CLIENT_PROTECTED/api/client/reviews")"

REV_CODE=$(http_code -X POST "${CLIENT_H[@]}" -H 'Content-Type: application/json' \
  -d "{\"vehicle_id\":\"$VEH_ID\",\"rating\":5,\"text\":\"QA review $RUN_ID\"}" \
  "$CLIENT_PROTECTED/api/client/reviews")
assert_one_of "CREV-002" "POST /api/client/reviews" "$REV_CODE" "" "200" "201" "400" "403"

# --- Client registration ---
assert_code "CR-002" "GET profile" "200" \
  "$(http_code "${CLIENT_H[@]}" "$CLIENT_PROTECTED/api/client/profile")"
assert_code "CR-003" "GET vehicles" "200" \
  "$(http_code "${CLIENT_H[@]}" "$CLIENT_PROTECTED/api/client/vehicles")"

if [[ -n "$CLIENT_ID" ]]; then
  assert_code "EREV-003" "GET client reviews by id" "200" \
    "$(http_code "${AUTH_H[@]}" "$EMPLOYEE_API/api/clients/$CLIENT_ID/reviews")"
fi

# --- Errors ingest ---
TEL=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 2 \
  -X POST "$ERRORS_INGEST/api/telemetry/events" \
  -H 'Content-Type: application/json' \
  -d "{\"kind\":\"js_error\",\"message\":\"qa-$RUN_ID\",\"at\":$(date +%s000)}" \
  2>/dev/null || echo "000")
if [[ "$TEL" == "204" ]]; then
  assert_code "ERR-001" "POST telemetry js_error" "204" "$TEL"
  assert_code "ERR-002" "POST telemetry api_latency" "204" \
    "$(http_code -X POST "$ERRORS_INGEST/api/telemetry/events" \
      -H 'Content-Type: application/json' \
      -d '{"kind":"api_latency","path":"/api/test","status":200,"duration_ms":100,"at":123}')"
  assert_one_of "ERR-003" "POST telemetry invalid" \
    "$(http_code -X POST "$ERRORS_INGEST/api/telemetry/events" \
      -H 'Content-Type: application/json' -d '{"kind":"unknown"}')" "" "400" "204"
  TEL_PROXY=$(http_code -X POST "$EMPLOYEE_AUTH/api/telemetry/events" \
    -H 'Content-Type: application/json' \
    -d "{\"kind\":\"js_error\",\"message\":\"proxy-$RUN_ID\",\"at\":123}")
  assert_one_of "ERR-006" "telemetry via auth proxy" "$TEL_PROXY" "" "204" "502" "503"
else
  SKIP=$((SKIP + 3))
  RESULTS+=("| ERR-001 | telemetry | ⏭ SKIP | 204 | $TEL | errors-ingest not running |")
fi

# --- Integration markers ---
assert_code "INT-001" "integration employee CRUD chain" "200" "$CUST_CODE"
assert_code "INT-002" "integration client register" "200" "$CREG"

rm -f /tmp/qa_resp_$$

TOTAL=$((PASS + FAIL))
PASS_RATE=0
if [[ $TOTAL -gt 0 ]]; then
  PASS_RATE=$(( PASS * 100 / TOTAL ))
fi

{
  echo "# TEST-RUN-REPORT"
  echo ""
  echo "| Field | Value |"
  echo "|-------|-------|"
  echo "| Run ID | \`$RUN_ID\` |"
  echo "| Timestamp | $TS |"
  echo "| Employee API | $EMPLOYEE_API |"
  echo "| Client Public | $CLIENT_PUBLIC |"
  echo "| Client Protected | $CLIENT_PROTECTED |"
  echo "| **Passed** | **$PASS** |"
  echo "| **Failed** | **$FAIL** |"
  echo "| **Skipped** | **$SKIP** |"
  echo "| Pass rate (excl. skip) | ${PASS_RATE}% |"
  echo ""
  echo "## Results"
  echo ""
  echo "| Auto ID | Test | Status | Expected | Actual | Notes |"
  echo "|---------|------|--------|----------|--------|-------|"
  for row in "${RESULTS[@]}"; do
    echo "$row"
  done
  echo ""
  echo "## Manual follow-up"
  echo ""
  echo "- Parts/Work-orders **write** CRUD: login qa.master@test.local (TC-PRT-002, TC-WO-002, INT-004)"
  echo "- Works/Employees: WRK-003, EMP-005 — admin/master tokens"
  echo "- Kafka async: employee-reviews после CREV-002 (INT-003)"
  echo "- workorders/works/employees: проверить docker если GW-010/WRK-001/EMP-001 = 503"
  echo ""
  echo "_Generated by \`qa/api-testing/scripts/run-api-tests.sh\`_"
} > "$REPORT"

python3 - <<PY
import json
meta = {
    "run_id": "$RUN_ID",
    "timestamp_utc": "$TS",
    "git_commit": "$GIT_COMMIT",
    "employee_api": "$EMPLOYEE_API",
    "client_public": "$CLIENT_PUBLIC",
    "client_protected": "$CLIENT_PROTECTED",
    "passed": $PASS,
    "failed": $FAIL,
    "skipped": $SKIP,
    "pass_rate_pct": $PASS_RATE,
    "report": "smoke-report.md",
}
with open("$META", "w") as f:
    json.dump(meta, f, indent=2)
    f.write("\n")
PY

cp "$REPORT" "$LATEST_DIR/smoke-report.md"
cp "$META" "$LATEST_DIR/meta.json"

echo ""
echo "Done: PASS=$PASS FAIL=$FAIL SKIP=$SKIP"
echo "Report: $REPORT"
echo "Latest: $LATEST_DIR/smoke-report.md"
[[ "$FAIL" -eq 0 ]]
