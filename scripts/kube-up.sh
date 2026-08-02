#!/usr/bin/env bash
# Поднять весь стек dealer в minikube одной командой:
#   ./scripts/kube-up.sh
#
# Опции:
#   --skip-build     не пересобирать Docker-образы
#   --skip-migrate   не применять SQL-миграции
#   --skip-seed      не применять seed_volume
#   --skip-expose    не запускать port-forward для LAN
#   --help
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
NS="${DEALER_NS:-dealer}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-changeme}"
JWT_SECRET="${JWT_SECRET:-change-me-in-production}"
POSTGRES_DSN="postgres://dealer:${POSTGRES_PASSWORD}@postgres:5432/dealer?sslmode=disable"
MINIKUBE_CPUS="${MINIKUBE_CPUS:-4}"
MINIKUBE_MEMORY="${MINIKUBE_MEMORY:-16384}"

SKIP_BUILD=0
SKIP_MIGRATE=0
SKIP_SEED=0
SKIP_EXPOSE=0

usage() {
  sed -n '2,12p' "$0"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --skip-build) SKIP_BUILD=1 ;;
    --skip-migrate) SKIP_MIGRATE=1 ;;
    --skip-seed) SKIP_SEED=1 ;;
    --skip-expose) SKIP_EXPOSE=1 ;;
    --help|-h) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage; exit 1 ;;
  esac
  shift
done

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "ERROR: required command not found: $1" >&2
    exit 1
  }
}

log() { echo "==> $*"; }

read_version() {
  local vf="$1"
  tr -d '[:space:]' <"$vf"
}

build_service() {
  local name="$1" dockerfile="$2" version_file="$3"
  local ver
  ver="$(read_version "$version_file")"
  log "BUILD ${name}:${ver}"
  docker build -f "$dockerfile" --build-arg "SERVICE_VERSION=${ver}" -t "${name}:${ver}" "$ROOT"
}

apply_migrations() {
  log "Applying SQL migrations"
  local f
  for f in "$ROOT"/migrations/*.up.sql; do
    [[ -f "$f" ]] || continue
    log "migrate $(basename "$f")"
    kubectl -n "$NS" exec -i deployment/postgres -- \
      env PGPASSWORD="$POSTGRES_PASSWORD" psql -U dealer -d dealer -v ON_ERROR_STOP=1 -f - <"$f"
  done
}

wait_infra() {
  log "Waiting for Postgres and Redis"
  kubectl -n "$NS" wait --for=condition=available deployment/postgres --timeout=300s
  kubectl -n "$NS" wait --for=condition=available deployment/redis --timeout=180s
}

wait_apps() {
  log "Waiting for application deployments"
  local dep
  for dep in $(kubectl -n "$NS" get deploy -o name | sed 's|deployment.apps/||'); do
    case "$dep" in
      postgres|redis|zookeeper|kafka) continue ;;
    esac
    kubectl -n "$NS" rollout status "deployment/${dep}" --timeout=600s || {
      echo "WARN: rollout timeout for ${dep}, showing last logs:" >&2
      kubectl -n "$NS" logs "deployment/${dep}" --tail=20 >&2 || true
      kubectl -n "$NS" rollout status "deployment/${dep}" --timeout=120s || true
    }
  done
}

render_stack() {
  local ver_auth ver_customers ver_vehicles ver_deals ver_parts ver_brands ver_dealer_points
  local ver_works ver_employees ver_workorders ver_appointments ver_gateway ver_employee_reviews
  local ver_client_auth ver_client_registration ver_client_reviews
  local ver_client_public ver_client_protected ver_employee_stats ver_client_stats

  ver_auth="$(read_version "$ROOT/services/employee/auth/VERSION")"
  ver_customers="$(read_version "$ROOT/services/employee/customers/VERSION")"
  ver_vehicles="$(read_version "$ROOT/services/employee/vehicles/VERSION")"
  ver_deals="$(read_version "$ROOT/services/employee/deals/VERSION")"
  ver_parts="$(read_version "$ROOT/services/employee/parts/VERSION")"
  ver_brands="$(read_version "$ROOT/services/employee/brands/VERSION")"
  ver_dealer_points="$(read_version "$ROOT/services/employee/dealerpoints/VERSION")"
  ver_works="$(read_version "$ROOT/services/employee/works/VERSION")"
  ver_employees="$(read_version "$ROOT/services/employee/employees/VERSION")"
  ver_workorders="$(read_version "$ROOT/services/employee/workorders/VERSION")"
  ver_appointments="$(read_version "$ROOT/services/employee/appointments/VERSION")"
  ver_gateway="$(read_version "$ROOT/services/gateway/employee/VERSION")"
  ver_employee_reviews="$(read_version "$ROOT/services/employee/reviews/VERSION")"
  ver_client_auth="$(read_version "$ROOT/services/client/auth/VERSION")"
  ver_client_registration="$(read_version "$ROOT/services/client/registration/VERSION")"
  ver_client_reviews="$(read_version "$ROOT/services/client/reviews/VERSION")"
  ver_client_public="$(read_version "$ROOT/services/gateway/client-public/VERSION")"
  ver_client_protected="$(read_version "$ROOT/services/gateway/client-protected/VERSION")"
  ver_employee_stats="$(read_version "$ROOT/services/statistics/employee/VERSION")"
  ver_client_stats="$(read_version "$ROOT/services/statistics/client/VERSION")"

  sed \
    -e "s|__IMG_AUTH__|auth-service:${ver_auth}|g" \
    -e "s|__IMG_CUSTOMERS__|customers-service:${ver_customers}|g" \
    -e "s|__IMG_VEHICLES__|vehicles-service:${ver_vehicles}|g" \
    -e "s|__IMG_DEALS__|deals-service:${ver_deals}|g" \
    -e "s|__IMG_PARTS__|parts-service:${ver_parts}|g" \
    -e "s|__IMG_BRANDS__|brands-service:${ver_brands}|g" \
    -e "s|__IMG_DEALER_POINTS__|dealer-points-service:${ver_dealer_points}|g" \
    -e "s|__IMG_WORKS__|works-service:${ver_works}|g" \
    -e "s|__IMG_EMPLOYEES__|employees-service:${ver_employees}|g" \
    -e "s|__IMG_WORKORDERS__|workorders-service:${ver_workorders}|g" \
    -e "s|__IMG_APPOINTMENTS__|appointments-service:${ver_appointments}|g" \
    -e "s|__IMG_GATEWAY__|gateway-service:${ver_gateway}|g" \
    -e "s|__IMG_EMPLOYEE_REVIEWS__|employee-reviews-service:${ver_employee_reviews}|g" \
    -e "s|__IMG_CLIENT_AUTH__|client-auth-service:${ver_client_auth}|g" \
    -e "s|__IMG_CLIENT_REGISTRATION__|client-registration-service:${ver_client_registration}|g" \
    -e "s|__IMG_CLIENT_REVIEWS__|client-reviews-service:${ver_client_reviews}|g" \
    -e "s|__IMG_CLIENT_PUBLIC_GATEWAY__|client-public-gateway-service:${ver_client_public}|g" \
    -e "s|__IMG_CLIENT_PROTECTED_GATEWAY__|client-protected-gateway-service:${ver_client_protected}|g" \
    -e "s|__IMG_EMPLOYEE_STATISTICS__|employee-statistics-service:${ver_employee_stats}|g" \
    -e "s|__IMG_CLIENT_STATISTICS__|client-statistics-service:${ver_client_stats}|g" \
    -e "s|__PULL_POLICY__|Never|g" \
    "$ROOT/k8s/dealer-stack.yaml"
}

print_urls() {
  local lan_ip
  lan_ip="$(ip -4 route get 1.1.1.1 2>/dev/null | awk '{for(i=1;i<=NF;i++) if($i=="src") print $(i+1)}')"
  lan_ip="${lan_ip:-$(hostname -I | awk '{print $1}')}"

  cat <<EOF

=== Dealer stack is up (namespace: ${NS}) ===

Employee UI:        http://${lan_ip}:9080
Employee API:       http://${lan_ip}:8090
Client UI (login):  http://${lan_ip}:3001/login
Client public API:  http://${lan_ip}:8091
Client protected:   http://${lan_ip}:8093

Logins:
  Employee: admin@dealer.local / admin123
            vol.employee1@test.dealer.local / Test1234!
  Client:   vol.client1@test.dealer.local / Test1234!

Re-expose LAN ports after reboot:
  ./scripts/expose-lan.sh

Monitoring:
  Prometheus: http://${lan_ip}:9090
  Grafana:    http://${lan_ip}:3030  (admin / admin)

EOF
}

main() {
  need_cmd docker
  need_cmd kubectl
  need_cmd minikube

  if ! minikube status >/dev/null 2>&1; then
    log "Starting minikube (${MINIKUBE_CPUS} CPU, ${MINIKUBE_MEMORY}MB RAM)"
    minikube start --cpus="$MINIKUBE_CPUS" --memory="$MINIKUBE_MEMORY"
  fi

  if ! minikube addons list 2>/dev/null | grep -q 'metrics-server.*enabled'; then
    log "Enabling metrics-server addon"
    minikube addons enable metrics-server
    kubectl -n kube-system rollout status deployment/metrics-server --timeout=120s || true
  fi

  eval "$(minikube docker-env)"
  export DOCKER_BUILDKIT=0

  if [[ "$SKIP_BUILD" -eq 0 ]]; then
    log "Building service images in minikube docker"
    build_service auth-service build/auth-service.Dockerfile services/employee/auth/VERSION
    build_service customers-service build/customers-service.Dockerfile services/employee/customers/VERSION
    build_service vehicles-service build/vehicles-service.Dockerfile services/employee/vehicles/VERSION
    build_service deals-service build/deals-service.Dockerfile services/employee/deals/VERSION
    build_service parts-service build/parts-service.Dockerfile services/employee/parts/VERSION
    build_service brands-service build/brands-service.Dockerfile services/employee/brands/VERSION
    build_service dealer-points-service build/dealer-points-service.Dockerfile services/employee/dealerpoints/VERSION
    build_service works-service build/works-service.Dockerfile services/employee/works/VERSION
    build_service employees-service build/employees-service.Dockerfile services/employee/employees/VERSION
    build_service workorders-service build/workorders-service.Dockerfile services/employee/workorders/VERSION
    build_service appointments-service build/appointments-service.Dockerfile services/employee/appointments/VERSION
    build_service gateway-service build/gateway-service.Dockerfile services/gateway/employee/VERSION
    build_service employee-reviews-service build/employee-reviews-service.Dockerfile services/employee/reviews/VERSION
    build_service client-auth-service build/client-auth-service.Dockerfile services/client/auth/VERSION
    build_service client-registration-service build/client-registration-service.Dockerfile services/client/registration/VERSION
    build_service client-reviews-service build/client-reviews-service.Dockerfile services/client/reviews/VERSION
    build_service client-public-gateway-service build/client-public-gateway-service.Dockerfile services/gateway/client-public/VERSION
    build_service client-protected-gateway-service build/client-protected-gateway-service.Dockerfile services/gateway/client-protected/VERSION
    build_service employee-statistics-service build/employee-statistics-service.Dockerfile services/statistics/employee/VERSION
    build_service client-statistics-service build/client-statistics-service.Dockerfile services/statistics/client/VERSION
    log "BUILD dealer-client-ui:latest"
    docker build -f "$ROOT/build/client-frontend.Dockerfile" -t dealer-client-ui:latest "$ROOT"
  else
    log "Skipping image build (--skip-build)"
  fi

  log "Creating namespace and secrets"
  kubectl create namespace "$NS" --dry-run=client -o yaml | kubectl apply -f -
  kubectl -n "$NS" create secret generic dealer-db \
    --from-literal=POSTGRES_PASSWORD="$POSTGRES_PASSWORD" \
    --from-literal=POSTGRES_DSN="$POSTGRES_DSN" \
    --dry-run=client -o yaml | kubectl apply -f -
  kubectl -n "$NS" create secret generic dealer-app-secrets \
    --from-literal=JWT_SECRET="$JWT_SECRET" \
    --dry-run=client -o yaml | kubectl apply -f -

  log "Applying k8s/dealer-stack.yaml"
  render_stack | kubectl apply -f -

  log "Applying k8s/client-frontend.yaml"
  kubectl apply -f "$ROOT/k8s/client-frontend.yaml"

  log "Applying k8s/monitoring-stack.yaml (Prometheus + Grafana)"
  kubectl apply -f "$ROOT/k8s/monitoring-stack.yaml"
  "$ROOT/scripts/apply-grafana-dashboards.sh"
  kubectl -n monitoring rollout status deployment/prometheus --timeout=120s || true
  kubectl -n monitoring rollout status deployment/grafana --timeout=120s || true

  wait_infra

  if [[ "$SKIP_MIGRATE" -eq 0 ]]; then
    apply_migrations
  else
    log "Skipping migrations (--skip-migrate)"
  fi

  if [[ "$SKIP_SEED" -eq 0 ]]; then
    log "Applying seed_volume test data"
    POSTGRES_PASSWORD="$POSTGRES_PASSWORD" "$ROOT/migrations/seed_volume/apply.sh" --k8s
  else
    log "Skipping seed_volume (--skip-seed)"
  fi

  log "Creating admin user (seed-admin)"
  AUTH_VER="$(read_version "$ROOT/services/employee/auth/VERSION")"
  kubectl -n "$NS" run seed-admin --rm -i --restart=Never \
    --image="auth-service:${AUTH_VER}" --image-pull-policy=Never \
    --env="POSTGRES_DSN=postgres://dealer:${POSTGRES_PASSWORD}@postgres:5432/dealer?sslmode=disable" \
    --command -- /seed-admin

  wait_apps

  if [[ "$SKIP_EXPOSE" -eq 0 ]]; then
    "$ROOT/scripts/expose-lan.sh"
  else
    log "Skipping LAN port-forward (--skip-expose)"
  fi

  print_urls
}

main "$@"
