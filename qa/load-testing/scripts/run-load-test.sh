#!/usr/bin/env bash
# Run GET probe then k6 load test.
# Usage:
#   PROFILE=smoke BASE_URL=http://192.168.0.27:9080 ./qa/load-testing/scripts/run-load-test.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BASE_URL="${BASE_URL:-http://192.168.0.27:9080}"
PROFILE="${PROFILE:-smoke}"
LOGIN_EMAIL="${LOGIN_EMAIL:-qa.master@test.local}"
LOGIN_PASSWORD="${LOGIN_PASSWORD:-Test1234!}"

export BASE_URL LOGIN_EMAIL LOGIN_PASSWORD

echo "=== Step 1: probe GET endpoints ==="
python3 "${ROOT}/scripts/probe_get.py"

if ! command -v k6 >/dev/null 2>&1; then
  echo ""
  echo "k6 not installed. Install: https://grafana.com/docs/k6/latest/set-up/install-k6/"
  echo "Manifest ready at: ${ROOT}/results/latest-endpoints.json"
  exit 0
fi

echo ""
echo "=== Step 2: k6 load test (profile=${PROFILE}) ==="
k6 run \
  -e "BASE_URL=${BASE_URL}" \
  -e "LOGIN_EMAIL=${LOGIN_EMAIL}" \
  -e "LOGIN_PASSWORD=${LOGIN_PASSWORD}" \
  -e "PROFILE=${PROFILE}" \
  -e "MANIFEST_PATH=${ROOT}/results/latest-endpoints.json" \
  "${ROOT}/k6/get-endpoints.js"
