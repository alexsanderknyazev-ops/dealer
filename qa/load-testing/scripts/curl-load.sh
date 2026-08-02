#!/usr/bin/env bash
# Simple parallel GET load test without k6. Uses manifest from probe.
# Usage:
#   CONCURRENCY=10 REQUESTS=200 PROFILE=smoke ./qa/load-testing/scripts/curl-load.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BASE_URL="${BASE_URL:-http://192.168.0.27:9080}"
LOGIN_EMAIL="${LOGIN_EMAIL:-qa.master@test.local}"
LOGIN_PASSWORD="${LOGIN_PASSWORD:-Test1234!}"
MANIFEST="${MANIFEST:-${ROOT}/results/latest-endpoints.json}"
CONCURRENCY="${CONCURRENCY:-10}"
REQUESTS="${REQUESTS:-100}"

case "${PROFILE:-smoke}" in
  smoke) CONCURRENCY=5; REQUESTS=50 ;;
  load)  CONCURRENCY=30; REQUESTS=1500 ;;
  stress) CONCURRENCY=50; REQUESTS=5000 ;;
esac

TOKEN=$(curl -sf -X POST "${BASE_URL}/api/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"${LOGIN_EMAIL}\",\"password\":\"${LOGIN_PASSWORD}\"}" \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])")

PATHS=$(python3 -c "
import json
with open('${MANIFEST}') as f:
    eps = [e['path'] for e in json.load(f)['endpoints'] if e.get('auth')]
print('\n'.join(eps))
")

echo "Load: ${REQUESTS} requests, concurrency=${CONCURRENCY}, base=${BASE_URL}"
START=$(python3 -c "import time; print(time.time())")

export BASE_URL TOKEN
printf '%s\n' $PATHS | python3 -c "
import os, random, subprocess, sys, time
paths = [l.strip() for l in sys.stdin if l.strip()]
base = os.environ['BASE_URL']
token = os.environ['TOKEN']
requests = int('${REQUESTS}')
concurrency = int('${CONCURRENCY}')
ok = err = 0
latencies = []

def one():
    global ok, err
    p = random.choice(paths)
    t0 = time.time()
    r = subprocess.run(
        ['curl','-s','-o','/dev/null','-w','%{http_code}',
         '-H', f'Authorization: Bearer {token}',
         '-H', 'Accept: application/json',
         f'{base}{p}'],
        capture_output=True, text=True, timeout=30,
    )
    code = int(r.stdout.strip() or '0')
    latencies.append(time.time() - t0)
    if code == 200:
        ok += 1
    else:
        err += 1

from concurrent.futures import ThreadPoolExecutor, as_completed
with ThreadPoolExecutor(max_workers=concurrency) as ex:
    futs = [ex.submit(one) for _ in range(requests)]
    for _ in as_completed(futs):
        pass

latencies.sort()
p95 = latencies[int(len(latencies)*0.95)-1] if latencies else 0
elapsed = time.time() - float('${START}')
print(f'OK={ok} ERR={err} elapsed={elapsed:.1f}s rps={requests/elapsed:.1f} p95={p95*1000:.0f}ms')
"
