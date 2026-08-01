#!/usr/bin/env bash
# Expose dealer stack on all interfaces for LAN access (kubectl port-forward).
set -euo pipefail

NS="${DEALER_NS:-dealer}"

LAN_IP=$(ip -4 route get 1.1.1.1 2>/dev/null | awk '{for(i=1;i<=NF;i++) if($i=="src") print $(i+1)}')
LAN_IP="${LAN_IP:-$(hostname -I | awk '{print $1}')}"

stop_one() {
  local port=$1
  pkill -f "kubectl port-forward --address 0.0.0.0 -n ${NS}.* ${port}:" 2>/dev/null || true
}

start_pf() {
  local svc=$1 local_port=$2 remote_port=$3
  stop_one "$local_port"
  kubectl port-forward --address 0.0.0.0 -n "$NS" "svc/${svc}" "${local_port}:${remote_port}" \
    >/tmp/pf-"${svc}"-"${local_port}".log 2>&1 &
  disown 2>/dev/null || true
}

echo "LAN IP: ${LAN_IP}"
echo "Stopping old port-forwards..."
pkill -f "kubectl port-forward --address 0.0.0.0 -n ${NS}" 2>/dev/null || true
sleep 1

echo "Starting port-forwards on 0.0.0.0..."
start_pf auth-service 9080 8080
start_pf gateway-service 8090 8090
start_pf client-public-gateway-service 8091 8091
start_pf client-protected-gateway-service 8093 8093
start_pf client-frontend 3001 3001

sleep 2

echo ""
echo "=== URLs for other devices on the same Wi‑Fi (http, not https) ==="
echo "  Employee UI (auth):     http://${LAN_IP}:9080"
echo "  Employee API gateway:   http://${LAN_IP}:8090"
echo "  Client UI (login):      http://${LAN_IP}:3001/login"
echo "  Client public gateway:  http://${LAN_IP}:8091"
echo "  Client protected GW:    http://${LAN_IP}:8093"
echo ""
echo "CI (Docker, same host):"
echo "  Registry:   http://${LAN_IP}:5050"
echo ""

ok=0
for spec in "9080:/healthz" "8090:/healthz" "8091:/healthz" "8093:/healthz" "3001:/login"; do
  port=${spec%%:*}
  path=${spec#*:}
  code=$(curl -s -o /dev/null -w '%{http_code}' --connect-timeout 3 "http://${LAN_IP}:${port}${path}" || echo fail)
  if [[ "$code" == "200" ]]; then
    echo "OK  ${LAN_IP}:${port}${path}"
    ok=$((ok + 1))
  else
    echo "FAIL ${LAN_IP}:${port}${path} (http ${code})"
  fi
done

if [[ "$ok" -lt 5 ]]; then
  echo ""
  echo "Some checks failed. Logs: /tmp/pf-*.log"
  exit 1
fi

echo ""
echo "Host is ready. If another laptop still cannot connect:"
echo "  1. Same Wi‑Fi as this PC (not guest network)"
echo "  2. Ping: ping ${LAN_IP}"
echo "  3. Router: disable AP/client isolation"
echo "  4. Firewall on this PC: sudo ufw allow 3001,9080,8090,8091,8093/tcp"
