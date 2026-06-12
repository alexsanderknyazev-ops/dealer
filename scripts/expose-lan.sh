#!/usr/bin/env bash
# Expose dealer stack on all interfaces for LAN access (kubectl port-forward).
set -euo pipefail

NS="${DEALER_NS:-dealer}"
MON_NS="${MONITORING_NS:-monitoring}"

LAN_IP=$(ip -4 route get 1.1.1.1 2>/dev/null | awk '{for(i=1;i<=NF;i++) if($i=="src") print $(i+1)}')
LAN_IP="${LAN_IP:-$(hostname -I | awk '{print $1}')}"

stop_one() {
  local port=$1
  local pid_file="/tmp/pf-watch-${port}.pid"
  if [[ -f "$pid_file" ]]; then
    kill "$(cat "$pid_file")" 2>/dev/null || true
    rm -f "$pid_file"
  fi
  pkill -f "kubectl port-forward --address 0.0.0.0 -n ${NS}.* ${port}:" 2>/dev/null || true
  pkill -f "kubectl port-forward --address 0.0.0.0 -n ${MON_NS}.* ${port}:" 2>/dev/null || true
}

start_pf() {
  local ns=$1 svc=$2 local_port=$3 remote_port=$4
  local log="/tmp/pf-${ns}-${svc}-${local_port}.log"
  stop_one "$local_port"
  # Restart port-forward when backend pods roll (kubectl PF exits on pod delete).
  (
    while true; do
      kubectl port-forward --address 0.0.0.0 -n "$ns" "svc/${svc}" "${local_port}:${remote_port}" \
        >>"$log" 2>&1 || true
      sleep 2
    done
  ) &
  echo $! >"/tmp/pf-watch-${local_port}.pid"
  disown 2>/dev/null || true
}

echo "LAN IP: ${LAN_IP}"
echo "Stopping old port-forwards..."
for port in 9080 8090 8091 8093 3001 9090 3030; do
  stop_one "$port"
done
sleep 1

echo "Starting port-forwards on 0.0.0.0..."
start_pf "$NS" auth-service 9080 8080
start_pf "$NS" gateway-service 8090 8090
start_pf "$NS" client-public-gateway-service 8091 8091
start_pf "$NS" client-protected-gateway-service 8093 8093
start_pf "$NS" client-frontend 3001 3001
if kubectl get svc -n "$MON_NS" prometheus &>/dev/null; then
  start_pf "$MON_NS" prometheus 9090 9090
  start_pf "$MON_NS" grafana 3030 3000
fi

sleep 2

echo ""
echo "=== URLs for other devices on the same Wi‑Fi (http, not https) ==="
echo "  Employee UI (auth):     http://${LAN_IP}:9080"
echo "  Employee API gateway:   http://${LAN_IP}:8090"
echo "  Client UI (login):      http://${LAN_IP}:3001/login"
echo "  Client public gateway:  http://${LAN_IP}:8091"
echo "  Client protected GW:    http://${LAN_IP}:8093"
echo "  Prometheus:             http://${LAN_IP}:9090"
echo "  Grafana:                http://${LAN_IP}:3030  (admin / admin)"
echo ""
echo "CI (Docker, same host):"
echo "  Jenkins:    http://${LAN_IP}:8080"
echo "  SonarQube:  http://${LAN_IP}:9000"
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
  echo "Some dealer checks failed. Logs: /tmp/pf-*.log"
  exit 1
fi

mon_ok=0
for spec in "9090:/-/ready" "3030:/login"; do
  port=${spec%%:*}
  path=${spec#*:}
  if ! ss -tln | grep -q ":${port} "; then
    continue
  fi
  code=$(curl -s -o /dev/null -w '%{http_code}' --connect-timeout 3 "http://${LAN_IP}:${port}${path}" || echo fail)
  if [[ "$code" == "200" ]]; then
    echo "OK  ${LAN_IP}:${port}${path}"
    mon_ok=$((mon_ok + 1))
  else
    echo "WARN ${LAN_IP}:${port}${path} (http ${code})"
  fi
done

if [[ "$mon_ok" -lt 2 ]]; then
  echo ""
  echo "Monitoring optional: apply k8s/monitoring-stack.yaml if Prometheus/Grafana needed."
fi

echo ""
echo "Port-forwards auto-restart when pods roll. Re-run this script only if all ports are dead."
echo ""
echo "Host is ready. If another laptop still cannot connect:"
echo "  1. Same Wi‑Fi as this PC (not guest network)"
echo "  2. Ping: ping ${LAN_IP}"
echo "  3. Router: disable AP/client isolation"
echo "  4. Firewall on this PC: sudo ufw allow 3001,3030,9090,9080,8090,8091,8093/tcp"
