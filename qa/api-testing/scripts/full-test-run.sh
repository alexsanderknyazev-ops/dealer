#!/usr/bin/env bash
# Full API test run — employee + client + E2E flows
set -uo pipefail

PATH="/usr/bin:/bin:/usr/local/bin:$PATH"
QA_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUN_ID="full-$(date -u +%Y%m%d-%H%M%S)"
RUN_DIR="${QA_ROOT}/results/runs/${RUN_ID}"
REPORT="${RUN_DIR}/full-report.md"
BUGS="${RUN_DIR}/bugs.md"
JSONL="${RUN_DIR}/results.jsonl"
mkdir -p "$RUN_DIR"

EMPLOYEE_API="${EMPLOYEE_API:-http://127.0.0.1:8090}"
CLIENT_PUBLIC="${CLIENT_PUBLIC:-http://127.0.0.1:8091}"
CLIENT_PROTECTED="${CLIENT_PROTECTED:-http://127.0.0.1:8093}"
POSTGRES_DSN="${POSTGRES_DSN:-postgres://dealer:changeme@127.0.0.1:5433/dealer?sslmode=disable}"
QA_PASSWORD="${QA_PASSWORD:-Test1234!}"

PASS=0
FAIL=0
SKIP=0
BUGS_LIST=()

log_result() {
  local id="$1" name="$2" status="$3" expected="$4" actual="$5" body="${6:-}"
  echo "{\"id\":\"$id\",\"name\":\"$name\",\"status\":\"$status\",\"expected\":\"$expected\",\"actual\":\"$actual\"}" >> "$JSONL"
  if [[ "$status" == "PASS" ]]; then PASS=$((PASS+1)); elif [[ "$status" == "SKIP" ]]; then SKIP=$((SKIP+1)); else FAIL=$((FAIL+1)); fi
}

record_bug() {
  local id="$1" title="$2" repro="$3" expected="$4" actual="$5"
  BUGS_LIST+=("### BUG-${id}: ${title}

**Expected:** ${expected}
**Actual:** ${actual}

**Reproduction:**
\`\`\`bash
${repro}
\`\`\`
")
}

login() {
  local base="$1" email="$2" pass="$3"
  curl -s -X POST "$base/api/login" -H 'Content-Type: application/json' \
    -d "{\"email\":\"$email\",\"password\":\"$pass\"}"
}

api() {
  local method="$1" url="$2" token="$3" data="${4:-}"
  if [[ -n "$data" ]]; then
    curl -s -w "\n__HTTP__:%{http_code}" -X "$method" "$url" \
      -H "Authorization: Bearer $token" -H 'Content-Type: application/json' -d "$data"
  else
    curl -s -w "\n__HTTP__:%{http_code}" -X "$method" "$url" \
      -H "Authorization: Bearer $token"
  fi
}

split_http() {
  HTTP_CODE="${1##*__HTTP__:}"
  BODY="${1%$'\n'__HTTP__:*}"
}

assert_http() {
  local id="$1" name="$2" exp="$3" resp="$4" repro="${5:-}"
  split_http "$resp"
  if [[ "$HTTP_CODE" == "$exp" ]] || [[ "$exp" == *"|"* && "$exp" == *"$HTTP_CODE"* ]]; then
    log_result "$id" "$name" "PASS" "$exp" "$HTTP_CODE" "$BODY"
    echo "PASS $id $name ($HTTP_CODE)"
  else
    log_result "$id" "$name" "FAIL" "$exp" "$HTTP_CODE" "$BODY"
    echo "FAIL $id $name expected=$exp got=$HTTP_CODE body=${BODY:0:200}"
    if [[ -n "$repro" ]]; then
      record_bug "$id" "$name" "$repro" "HTTP $exp" "HTTP $HTTP_CODE — ${BODY:0:300}"
    fi
  fi
}

echo "=== Full test run $RUN_ID ==="

# --- Tokens ---
MASTER=$(login "$EMPLOYEE_API" "qa.master@test.local" "$QA_PASSWORD" | python3 -c "import sys,json; print(json.load(sys.stdin).get('access_token',''))")
ADMIN=$(login "$EMPLOYEE_API" "qa.admin@test.local" "$QA_PASSWORD" | python3 -c "import sys,json; print(json.load(sys.stdin).get('access_token',''))")
SALES=$(login "$EMPLOYEE_API" "qa.sales@test.local" "$QA_PASSWORD" | python3 -c "import sys,json; print(json.load(sys.stdin).get('access_token',''))")
PARTS=$(login "$EMPLOYEE_API" "qa.parts@test.local" "$QA_PASSWORD" | python3 -c "import sys,json; print(json.load(sys.stdin).get('access_token',''))")
CLIENT=$(login "$CLIENT_PUBLIC" "qa.client@test.local" "$QA_PASSWORD" | python3 -c "import sys,json; print(json.load(sys.stdin).get('access_token',''))")

# Manifest refs
CUST_ID="a2200001-0000-4000-8000-000000000001"
VEH_ID="a3300001-0000-4000-8000-000000000001"
DP_ID="10000000-0000-4000-8000-000000000001"
WH_ID="30000000-0000-4000-8000-000000000002"
PART_ID="a4400001-0000-4000-8000-000000000001"
WORK_ID="a8800001-0000-4000-8000-000000000002"
EMP_MASTER="a1100001-0000-4000-8000-000000000003"
WO_FIXTURE="a6600001-0000-4000-8000-000000000001"
BRAND_ID="40000000-0000-4000-8000-000000000003"

# --- Health ---
for spec in "gateway:$EMPLOYEE_API/healthz" "auth:http://127.0.0.1:8080/healthz" "client-pub:$CLIENT_PUBLIC/healthz" "client-prot:$CLIENT_PROTECTED/healthz"; do
  name="${spec%%:*}"
  url="${spec#*:}"
  c=$(curl -s -o /dev/null -w "%{http_code}" "$url")
  assert_http "HL-$(echo $name|tr -d '-')" "health $name" "200" "$(echo ok; echo __HTTP__:$c)"
done

# --- Auth RBAC ---
assert_http "RBAC-001" "sales POST work-orders 403" "403" \
  "$(api POST "$EMPLOYEE_API/api/work-orders" "$SALES" "{\"customer_id\":\"$CUST_ID\",\"vehicle_id\":\"$VEH_ID\"}")" \
  "curl -X POST $EMPLOYEE_API/api/work-orders -H 'Authorization: Bearer \$SALES_TOKEN' -d '{\"customer_id\":\"$CUST_ID\",\"vehicle_id\":\"$VEH_ID\"}'"

assert_http "RBAC-002" "sales POST works 403" "403" \
  "$(api POST "$EMPLOYEE_API/api/works" "$SALES" '{"code":"X","name":"X"}')" \
  "curl -X POST $EMPLOYEE_API/api/works -H 'Authorization: Bearer \$SALES_TOKEN' ..."

assert_http "RBAC-003" "master POST works 200" "200" \
  "$(api POST "$EMPLOYEE_API/api/works" "$MASTER" "{\"code\":\"LAB-QA-RUN-$RUN_ID\",\"name\":\"Run test\",\"category\":\"test\",\"labor_hours\":\"1\",\"unit_price\":\"100\"}")"

# --- Works CRUD ---
assert_http "WRK-001" "GET works list" "200" "$(api GET "$EMPLOYEE_API/api/works?limit=5" "$MASTER")"
assert_http "WRK-002" "GET work by id" "200" "$(api GET "$EMPLOYEE_API/api/works/$WORK_ID" "$MASTER")"

# --- Employees ---
assert_http "EMP-001" "GET employees list" "200" "$(api GET "$EMPLOYEE_API/api/employees?limit=5" "$MASTER")"

# --- Customers / Vehicles ---
assert_http "CUST-001" "GET customers" "200" "$(api GET "$EMPLOYEE_API/api/customers?limit=5" "$MASTER")"
assert_http "VEH-001" "GET vehicles" "200" "$(api GET "$EMPLOYEE_API/api/vehicles?limit=5" "$MASTER")"

# --- Create WO with work_id ---
WO_CREATE=$(api POST "$EMPLOYEE_API/api/work-orders" "$MASTER" "{
  \"customer_id\":\"$CUST_ID\",
  \"vehicle_id\":\"$VEH_ID\",
  \"dealer_point_id\":\"$DP_ID\",
  \"warehouse_id\":\"$WH_ID\",
  \"service_advisor_id\":\"$EMP_MASTER\",
  \"complaint\":\"Full test run\",
  \"labor\":[{\"work_id\":\"$WORK_ID\",\"executor_id\":\"$EMP_MASTER\",\"sort_order\":1}],
  \"parts\":[{\"part_id\":\"$PART_ID\",\"warehouse_id\":\"$WH_ID\",\"quantity\":\"2\",\"unit_price\":\"900\",\"sort_order\":1}]
}")
assert_http "WO-001" "POST create work-order" "200" "$WO_CREATE" \
  "curl -X POST $EMPLOYEE_API/api/work-orders -H 'Authorization: Bearer \$MASTER_TOKEN' -d '{...work_id, parts...}'"

WO_ID=$(echo "${WO_CREATE%$'\n'__HTTP__:*}" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('id', d.get('work_order',{}).get('id','')))" 2>/dev/null || echo "")
if [[ -z "$WO_ID" ]]; then
  WO_ID=$(echo "${WO_CREATE%$'\n'__HTTP__:*}" | python3 -c "import sys,json; print(json.load(sys.stdin).get('work_order',{}).get('id',''))" 2>/dev/null || echo "")
fi
echo "Created WO_ID=$WO_ID"

# --- Get WO fixture ---
assert_http "WO-002" "GET fixture WO" "200" "$(api GET "$EMPLOYEE_API/api/work-orders/$WO_FIXTURE" "$MASTER")"

# --- move-parts-to-work on fixture ---
MOVE=$(api POST "$EMPLOYEE_API/api/work-orders/$WO_FIXTURE/move-parts-to-work" "$MASTER" "{\"issued_by\":\"$EMP_MASTER\"}")
assert_http "WO-003" "POST move-parts-to-work" "200" "$MOVE" \
  "curl -X POST $EMPLOYEE_API/api/work-orders/$WO_FIXTURE/move-parts-to-work -H 'Authorization: Bearer \$MASTER_TOKEN' -d '{\"issued_by\":\"$EMP_MASTER\"}'"

DOC_ID=$(echo "${MOVE%$'\n'__HTTP__:*}" | python3 -c "
import sys,json
d=json.load(sys.stdin)
wo=d.get('work_order',d)
print(wo.get('movement_document_id',''))
" 2>/dev/null || echo "")
echo "DOC_ID=$DOC_ID"

# --- Confirm movement ---
if [[ -n "$DOC_ID" && "$DOC_ID" != "None" ]]; then
  CONF=$(api POST "$EMPLOYEE_API/api/movement-documents/$DOC_ID/confirm" "$PARTS" '{}')
  assert_http "PRT-001" "POST confirm movement" "200|204" "$CONF" \
    "curl -X POST $EMPLOYEE_API/api/movement-documents/$DOC_ID/confirm -H 'Authorization: Bearer \$PARTS_TOKEN'"
  
  WO_AFTER=$(api GET "$EMPLOYEE_API/api/work-orders/$WO_FIXTURE" "$MASTER")
  split_http "$WO_AFTER"
  PI=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); wo=d.get('work_order',d); print(wo.get('parts_issued', False))" 2>/dev/null)
  if [[ "$PI" == "True" || "$PI" == "true" ]]; then
    log_result "WO-004" "parts_issued after confirm" "PASS" "true" "$PI"
    echo "PASS WO-004 parts_issued=true"
  else
    log_result "WO-004" "parts_issued after confirm" "FAIL" "true" "$PI" "$BODY"
    record_bug "WO-004" "parts_issued not set after confirm" \
      "curl GET $EMPLOYEE_API/api/work-orders/$WO_FIXTURE" "parts_issued=true" "parts_issued=$PI"
    echo "FAIL WO-004 parts_issued=$PI"
  fi
  
  STOCK=$(psql "$POSTGRES_DSN" -tAc "SELECT quantity FROM parts.part_stock WHERE part_id='$PART_ID' AND warehouse_id='$WH_ID'")
  echo "Stock after confirm: $STOCK (was 50, expect 48)"
  if [[ "$STOCK" == "48" ]]; then
    log_result "PRT-002" "stock decreased by 2" "PASS" "48" "$STOCK"
  else
    log_result "PRT-002" "stock decreased by 2" "FAIL" "48" "$STOCK"
    record_bug "PRT-002" "Stock not decreased after confirm" \
      "psql -c \"SELECT quantity FROM parts.part_stock WHERE part_id='$PART_ID'\"" "48" "$STOCK"
  fi
else
  log_result "PRT-001" "confirm movement" "SKIP" "doc_id" "empty"
  log_result "WO-004" "parts_issued" "SKIP" "" "no doc"
fi

# --- Deals ---
assert_http "DEL-001" "GET deals" "200" "$(api GET "$EMPLOYEE_API/api/deals?limit=5" "$SALES")"
DEAL=$(api POST "$EMPLOYEE_API/api/deals" "$SALES" "{\"customer_id\":\"$CUST_ID\",\"vehicle_id\":\"a3300001-0000-4000-8000-000000000003\",\"amount\":\"1000000\",\"stage\":\"draft\"}")
assert_http "DEL-002" "POST deal sales" "200" "$DEAL"

# --- Brands ---
assert_http "BRD-001" "GET brands" "200" "$(api GET "$EMPLOYEE_API/api/brands?limit=3" "$SALES")"

# --- Dealer points ---
assert_http "DP-001" "GET dealer-points" "200" "$(api GET "$EMPLOYEE_API/api/dealer-points" "$SALES")"
assert_http "DP-002" "GET warehouses" "200" "$(api GET "$EMPLOYEE_API/api/warehouses" "$SALES")"

# --- Reviews employee ---
assert_http "REV-001" "GET employee reviews" "200" "$(api GET "$EMPLOYEE_API/api/reviews?limit=5" "$SALES")"

# --- Stats ---
assert_http "ST-001" "employee stats" "200" "$(api GET "$EMPLOYEE_API/api/stats/employee/overview" "$SALES")"
assert_http "ST-002" "client stats" "200" "$(api GET "$EMPLOYEE_API/api/stats/client/overview" "$SALES")"

# --- Client contour ---
assert_http "CL-001" "GET /api/me client" "200" "$(api GET "$CLIENT_PROTECTED/api/me" "$CLIENT")"
assert_http "CL-002" "GET profile" "200" "$(api GET "$CLIENT_PROTECTED/api/client/profile" "$CLIENT")"
assert_http "CL-003" "GET vehicles" "200" "$(api GET "$CLIENT_PROTECTED/api/client/vehicles" "$CLIENT")"
assert_http "CL-004" "GET reviews" "200" "$(api GET "$CLIENT_PROTECTED/api/client/reviews" "$CLIENT")"

REV=$(api POST "$CLIENT_PROTECTED/api/client/reviews" "$CLIENT" "{\"vehicle_id\":\"$VEH_ID\",\"rating\":5,\"text\":\"Full test review $RUN_ID\"}")
assert_http "CL-005" "POST client review" "200|201" "$REV"

# --- Cross-token isolation ---
assert_http "ISO-001" "employee token on client profile" "401|403" "$(api GET "$CLIENT_PROTECTED/api/client/profile" "$MASTER")"
assert_http "ISO-002" "client token on employee customers" "401|403" "$(api GET "$EMPLOYEE_API/api/customers" "$CLIENT")"

# --- Invalid refs WO ---
assert_http "WO-005" "bad work_id" "400|404|500" \
  "$(api POST "$EMPLOYEE_API/api/work-orders" "$MASTER" "{\"customer_id\":\"$CUST_ID\",\"vehicle_id\":\"$VEH_ID\",\"labor\":[{\"work_id\":\"00000000-0000-4000-8000-000000000099\"}]}")"

# --- service_advisor_name ---
if [[ -n "$WO_ID" ]]; then
  GWO=$(api GET "$EMPLOYEE_API/api/work-orders/$WO_ID" "$MASTER")
  split_http "$GWO"
  SAN=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); wo=d.get('work_order',d); print(wo.get('service_advisor_name',''))" 2>/dev/null)
  if [[ -n "$SAN" && "$SAN" != "None" ]]; then
    log_result "WO-006" "service_advisor_name populated" "PASS" "non-empty" "$SAN"
  else
    log_result "WO-006" "service_advisor_name populated" "FAIL" "non-empty" "empty"
    record_bug "WO-006" "service_advisor_name empty in WO response" \
      "curl GET $EMPLOYEE_API/api/work-orders/$WO_ID" "QA Master" "empty"
  fi
fi

# --- Write report ---
TOTAL=$((PASS+FAIL))
{
  echo "# Full API Test Report — $RUN_ID"
  echo ""
  echo "| Metric | Value |"
  echo "|--------|-------|"
  echo "| Timestamp | $(date -u '+%Y-%m-%d %H:%M:%S UTC') |"
  echo "| Passed | $PASS |"
  echo "| Failed | $FAIL |"
  echo "| Skipped | $SKIP |"
  echo "| Pass rate | $(( PASS * 100 / (TOTAL+SKIP>0?TOTAL:1) ))% |"
  echo ""
  echo "## Results (see results.jsonl)"
  echo ""
  if [[ ${#BUGS_LIST[@]} -gt 0 ]]; then
    echo "## Bugs found: ${#BUGS_LIST[@]}"
    echo ""
    for b in "${BUGS_LIST[@]}"; do echo "$b"; echo ""; done
  else
    echo "## Bugs found: 0"
  fi
} > "$REPORT"

if [[ ${#BUGS_LIST[@]} -gt 0 ]]; then
  printf '%s\n' "${BUGS_LIST[@]}" > "$BUGS"
fi

cp "$REPORT" "${QA_ROOT}/results/latest/full-report.md" 2>/dev/null || true
cp "$BUGS" "${QA_ROOT}/results/latest/bugs.md" 2>/dev/null || true

echo ""
echo "Report: $REPORT"
echo "PASS=$PASS FAIL=$FAIL SKIP=$SKIP BUGS=${#BUGS_LIST[@]}"
exit $([[ $FAIL -eq 0 ]] && echo 0 || echo 1)
