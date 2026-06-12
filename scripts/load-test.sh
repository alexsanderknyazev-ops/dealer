#!/usr/bin/env bash
# Нагрузочное тестирование Dealer stack (HTTP).
#
# Usage:
#   ./scripts/load-test.sh                     # интерактивно: RPS, длительность, сценарий
#   ./scripts/load-test.sh --rps 150 --duration 600
#   ./scripts/load-test.sh --smoke
#   LOAD_RPS=200 LOAD_DURATION=300 ./scripts/load-test.sh --yes
#
# Scenarios:
#   health | employee-read | employee-login | client-read | mixed (default)
#
# Reports: ${TMPDIR}/dealer-load-test/results/<run-id>/ (outside repo)
#   report.txt      — текстовый отчёт
#   summary.json    — сводка + всё вместе
#   services.json   — детально по сервисам
#   endpoints.json  — детально по endpoint
#   timeline.json   — RPS/latency/errors по 30s
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
RESULTS_DIR="${LOAD_TEST_RESULTS_DIR:-${TMPDIR:-/tmp}/dealer-load-test/results}"
RUN_ID="run-$(date -u +%Y%m%d-%H%M%S)"
RUN_DIR="${RESULTS_DIR}/${RUN_ID}"

LAN_IP="$(ip -4 route get 1.1.1.1 2>/dev/null | awk '{for(i=1;i<=NF;i++) if($i=="src") print $(i+1)}')"
LAN_IP="${LAN_IP:-127.0.0.1}"

export EMPLOYEE_API="${EMPLOYEE_API:-http://${LAN_IP}:8090}"
export EMPLOYEE_AUTH="${EMPLOYEE_AUTH:-http://${LAN_IP}:9080}"
export CLIENT_PUBLIC="${CLIENT_PUBLIC:-http://${LAN_IP}:8091}"
export CLIENT_PROTECTED="${CLIENT_PROTECTED:-http://${LAN_IP}:8093}"

DEFAULT_RPS=150
DEFAULT_DURATION=600
DEFAULT_SCENARIO=mixed

SKIP_PROMPT=0
PY_ARGS=()

display_arg() {
  local flag=$1 default=$2
  local i
  for i in "${!PY_ARGS[@]}"; do
    if [[ "${PY_ARGS[$i]}" == "$flag" && -n "${PY_ARGS[$i+1]:-}" ]]; then
      echo "${PY_ARGS[$i+1]}"
      return
    fi
  done
  echo "$default"
}

usage() {
  sed -n '2,20p' "$0" | sed 's/^# \?//'
  exit 0
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help) usage ;;
    -y|--yes) SKIP_PROMPT=1; shift ;;
    --smoke) PY_ARGS+=("$1"); shift ;;
    --rps|--duration|--scenario|--concurrency|--bucket|--progress)
      PY_ARGS+=("$1" "$2"); shift 2 ;;
    --*) PY_ARGS+=("$1"); shift ;;
    *) PY_ARGS+=("$1"); shift ;;
  esac
done

has_arg() {
  local name=$1
  local i
  for i in "${!PY_ARGS[@]}"; do
    [[ "${PY_ARGS[$i]}" == "$name" ]] && return 0
  done
  return 1
}

if [[ "$SKIP_PROMPT" -eq 0 ]] && ! has_arg --smoke; then
  if ! has_arg --rps; then
    read -rp "RPS [${DEFAULT_RPS}]: " input_rps
    export LOAD_RPS="${input_rps:-$DEFAULT_RPS}"
  fi
  if ! has_arg --duration; then
    read -rp "Duration (seconds) [${DEFAULT_DURATION} = 10 min]: " input_dur
    export LOAD_DURATION="${input_dur:-$DEFAULT_DURATION}"
  fi
  if ! has_arg --scenario; then
    echo "Scenarios: health | employee-read | employee-login | client-read | mixed"
    read -rp "Scenario [${DEFAULT_SCENARIO}]: " input_scenario
    export LOAD_SCENARIO="${input_scenario:-$DEFAULT_SCENARIO}"
  fi
fi

mkdir -p "$RUN_DIR"

DISPLAY_RPS="$(display_arg --rps "${LOAD_RPS:-$DEFAULT_RPS}")"
DISPLAY_DURATION="$(display_arg --duration "${LOAD_DURATION:-$DEFAULT_DURATION}")"
DISPLAY_SCENARIO="$(display_arg --scenario "${LOAD_SCENARIO:-$DEFAULT_SCENARIO}")"

echo "=== Dealer load test ==="
echo "Target:        ${DISPLAY_RPS} RPS × ${DISPLAY_DURATION}s (~$(( ${DISPLAY_RPS%.*} * DISPLAY_DURATION )) requests)"
echo "Scenario:      ${DISPLAY_SCENARIO}"
echo "Employee API:  $EMPLOYEE_API"
echo "Employee auth: $EMPLOYEE_AUTH"
echo "Client public: $CLIENT_PUBLIC"
echo "Results:       $RUN_DIR"
echo ""

python3 "$ROOT/scripts/load-test.py" \
  "${PY_ARGS[@]}" \
  --report-dir "$RUN_DIR" \
  | tee "$RUN_DIR/console.log"

rc=${PIPESTATUS[0]}
ln -sfn "$RUN_DIR" "${RESULTS_DIR}/latest"
exit "$rc"
