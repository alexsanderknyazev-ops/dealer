#!/usr/bin/env bash
# Expose dealer stack on the public internet (no domain needed — uses nip.io).
#
# Run AFTER ./scripts/gcp-up.sh completes and pods are healthy.
#
# Usage:
#   ./scripts/gcp-expose.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
NS="${DEALER_NS:-dealer}"
INGRESS_NS="${INGRESS_NS:-ingress-nginx}"
INGRESS_MANIFEST="${INGRESS_MANIFEST:-https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.11.3/deploy/static/provider/cloud/deploy.yaml}"

export USE_GKE_GCLOUD_AUTH_PLUGIN="${USE_GKE_GCLOUD_AUTH_PLUGIN:-True}"

log() { echo "==> $*" >&2; }

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "ERROR: required command not found: $1" >&2
    exit 1
  }
}

wait_for_external_ip() {
  local ip=""
  log "Waiting for Load Balancer external IP (up to 5 min)"
  for _ in $(seq 1 60); do
    ip="$(kubectl get svc -n "$INGRESS_NS" ingress-nginx-controller \
      -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || true)"
    [[ -n "$ip" ]] && { echo "$ip"; return 0; }
    sleep 5
  done
  echo "ERROR: external IP not assigned. Check: kubectl get svc -n $INGRESS_NS" >&2
  exit 1
}

ip_to_nip_host() {
  local ip="$1"
  echo "$(echo "$ip" | tr '.' '-').nip.io"
}

main() {
  need_cmd kubectl

  if ! kubectl get ns "$NS" >/dev/null 2>&1; then
    echo "ERROR: namespace '$NS' not found. Run ./scripts/gcp-up.sh first." >&2
    exit 1
  fi

  if ! kubectl -n "$NS" get deploy auth-service >/dev/null 2>&1; then
    echo "ERROR: auth-service not deployed. Run ./scripts/gcp-up.sh first." >&2
    exit 1
  fi

  if ! kubectl get ns "$INGRESS_NS" >/dev/null 2>&1; then
    log "Installing nginx Ingress Controller"
    kubectl apply -f "$INGRESS_MANIFEST"
  elif ! kubectl -n "$INGRESS_NS" get deploy ingress-nginx-controller >/dev/null 2>&1; then
    log "Installing nginx Ingress Controller"
    kubectl apply -f "$INGRESS_MANIFEST"
  else
    log "nginx Ingress Controller already installed"
  fi

  local ip nip_host client_host
  ip="$(wait_for_external_ip)"
  nip_host="$(ip_to_nip_host "$ip")"
  client_host="client.${nip_host}"

  log "Public IP: ${ip}"
  log "Employee host: ${nip_host}"
  log "Client host: ${client_host}"

  sed \
    -e "s|__NIP_HOST__|${nip_host}|g" \
    -e "s|__CLIENT_NIP_HOST__|${client_host}|g" \
    "$ROOT/k8s/gcp-public-ingress.yaml" | kubectl apply -f -

  log "Waiting for Ingress to receive an address"
  for _ in $(seq 1 30); do
    local addr
    addr="$(kubectl -n "$NS" get ingress dealer-http -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || true)"
    [[ -n "$addr" ]] && break
    sleep 5
  done

  cat <<EOF

=== Dealer is on the public internet ===

Share with your friend:

  Employee app:  http://${nip_host}/
  Client app:    http://${client_host}/login

Logins:
  Employee: admin@dealer.local / admin123
  Client:   vol.client1@test.dealer.local / Test1234!

Notes:
  - HTTP only (no HTTPS) — fine for a quick demo.
  - First load may take 1–2 min while pods start on a single node.
  - External IP: ${ip}  (changes if you recreate the ingress controller)

Check status:
  kubectl -n ${NS} get pods
  kubectl -n ${NS} get ingress

EOF
}

main "$@"
