#!/usr/bin/env bash
# LAN connectivity diagnostic for dealer stack (run on the SERVER host).
set -uo pipefail

LAN_IP=$(ip -4 route get 1.1.1.1 2>/dev/null | awk '{for(i=1;i<=NF;i++) if($i=="src") print $(i+1)}')
LAN_IP="${LAN_IP:-unknown}"
WIFI_IF=$(ip -4 route get 1.1.1.1 2>/dev/null | awk '{for(i=1;i<=NF;i++) if($i=="dev") print $(i+1)}')
SSID=$(iw dev 2>/dev/null | awk '/ssid/{print $2}' | head -1)

pass=0
fail=0
warn=0

ok()   { echo "  [OK]   $*"; pass=$((pass+1)); }
bad()  { echo "  [FAIL] $*"; fail=$((fail+1)); }
note() { echo "  [??]   $*"; warn=$((warn+1)); }

echo "=============================================="
echo "  Dealer LAN diagnostic — $(date '+%Y-%m-%d %H:%M')"
echo "=============================================="
echo ""
echo "Server LAN IP:  $LAN_IP"
echo "WiFi interface: ${WIFI_IF:-?}"
echo "WiFi SSID:      ${SSID:-?}"
echo ""

echo "--- A. Port-forward listeners (must be 0.0.0.0) ---"
for port in 9080 8090 8091 8093; do
  line=$(ss -tln | awk -v p=":$port" '$4 ~ p"$" {print}')
  if echo "$line" | grep -q '0.0.0.0:'; then
    ok "Port $port bound to 0.0.0.0"
  elif [[ -n "$line" ]]; then
    bad "Port $port bound to localhost only: $line"
  else
    bad "Port $port not listening — run: ./scripts/expose-lan.sh"
  fi
done
echo ""

echo "--- B. HTTP health via LAN IP ---"
for spec in "9080:/healthz:Employee UI" "8090:/healthz:API gateway" "8091:/healthz:Client public" "8093:/healthz:Client protected"; do
  IFS=: read -r port path label <<< "$spec"
  code=$(curl -s -o /dev/null -w '%{http_code}' --connect-timeout 3 "http://${LAN_IP}:${port}${path}" 2>/dev/null || echo 000)
  if [[ "$code" == "200" ]]; then
    ok "$label — http://${LAN_IP}:${port}"
  else
    bad "$label — http://${LAN_IP}:${port} (HTTP $code)"
  fi
done
echo ""

echo "--- C. Kubernetes ---"
if kubectl get ns dealer &>/dev/null; then
  not_ready=$(kubectl get pods -n dealer --no-headers 2>/dev/null | awk '!/(Running|Completed)/{c++} END{print c+0}')
  if [[ "$not_ready" -eq 0 ]]; then
    ok "All dealer pods Running"
  else
    bad "$not_ready pod(s) not Running — kubectl get pods -n dealer"
  fi
else
  bad "Cannot reach cluster — minikube running?"
fi
echo ""

echo "--- D. Firewall ---"
if command -v ufw &>/dev/null; then
  ufw_out=$(ufw status 2>/dev/null || true)
  if echo "$ufw_out" | grep -qi inactive; then
    ok "ufw inactive"
  elif echo "$ufw_out" | grep -qi active; then
    bad "ufw ACTIVE — run: sudo ufw allow 9080,8090,8091,8093/tcp"
  else
    note "ufw status unknown (need sudo?)"
  fi
else
  ok "ufw not installed"
fi
if systemctl is-active firewalld &>/dev/null; then
  bad "firewalld active — open ports 9080,8090,8091,8093"
else
  ok "firewalld inactive"
fi
echo ""

echo "--- E. Other devices on WiFi (ARP) ---"
neigh=$(ip neigh show dev "${WIFI_IF}" 2>/dev/null | grep -v FAILED | grep -v '^fe80' || true)
if echo "$neigh" | grep -qE '192\.168\.'; then
  others=$(echo "$neigh" | grep -v '192.168.0.1 ' | grep '192.168' || true)
  if [[ -n "$others" ]]; then
    ok "Devices seen on LAN:"
    echo "$others" | sed 's/^/         /'
  else
    note "Only router in ARP — other laptop may be on guest WiFi or AP isolation enabled"
  fi
else
  note "No LAN neighbors in ARP table"
fi
echo ""

echo "--- F. Wrong IPs (do NOT use from other laptop) ---"
note "192.168.49.x — minikube internal network"
note "172.17.x / 172.18.x — Docker bridges"
echo ""

echo "=============================================="
echo "  Summary: $pass passed, $fail failed, $warn warnings"
echo "=============================================="
echo ""
if [[ "$fail" -eq 0 ]]; then
  echo "SERVER SIDE IS OK. Problem is likely on the CLIENT or ROUTER."
  echo ""
  echo "On the OTHER laptop run:"
  echo "  ping $LAN_IP"
  echo "  curl http://$LAN_IP:9080/healthz"
  echo ""
  echo "Expected: ping replies, curl prints 'ok'"
  echo ""
  echo "If ping fails:"
  echo "  • Same WiFi: $SSID (not guest network)"
  echo "  • Router: disable AP/client isolation"
  echo "  • Check client IP is 192.168.0.x (same subnet)"
  echo ""
  echo "If ping OK but browser fails:"
  echo "  • Use http:// (not https)"
  echo "  • Port 9080 for UI (8080 = Jenkins)"
else
  echo "Fix server issues first, then run: ./scripts/expose-lan.sh"
fi
