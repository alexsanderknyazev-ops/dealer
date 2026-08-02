#!/usr/bin/env bash
# Deploy the full dealer stack to GKE (Artifact Registry + dealer-stack.yaml).
#
# Prerequisites:
#   gcloud auth login
#   gcloud auth application-default login
#   gcloud auth configure-docker ${GCP_REGION}-docker.pkg.dev
#   export USE_GKE_GCLOUD_AUTH_PLUGIN=True   # add to ~/.zshrc for kubectl
#
# Usage:
#   ./scripts/gcp-up.sh
#
# Options:
#   --skip-build     do not build/push Docker images
#   --skip-migrate   do not apply SQL migrations
#   --skip-seed      do not apply seed_volume
#   --help
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
NS="${DEALER_NS:-dealer}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-changeme}"
JWT_SECRET="${JWT_SECRET:-change-me-in-production}"
POSTGRES_DSN="postgres://dealer:${POSTGRES_PASSWORD}@postgres:5432/dealer?sslmode=disable"

GCP_PROJECT="${GCP_PROJECT:-$(gcloud config get-value project 2>/dev/null)}"
GCP_REGION="${GCP_REGION:-europe-west1}"
GCP_ZONE="${GCP_ZONE:-europe-west1-b}"
AR_REPO="${AR_REPO:-dealer}"
GKE_CLUSTER="${GKE_CLUSTER:-dealer-cluster}"
CLIENT_UI_TAG="${CLIENT_UI_TAG:-latest}"
DOCKER_PLATFORM="${DOCKER_PLATFORM:-linux/amd64}"

SKIP_BUILD=0
SKIP_MIGRATE=0
SKIP_SEED=0

usage() {
  sed -n '2,18p' "$0"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --skip-build) SKIP_BUILD=1 ;;
    --skip-migrate) SKIP_MIGRATE=1 ;;
    --skip-seed) SKIP_SEED=1 ;;
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

registry_base() {
  echo "${GCP_REGION}-docker.pkg.dev/${GCP_PROJECT}/${AR_REPO}"
}

build_and_push() {
  local name="$1" dockerfile="$2" version_file="$3"
  local ver reg img
  ver="$(read_version "$version_file")"
  reg="$(registry_base)"
  img="${reg}/${name}:${ver}"
  log "BUILD+PUSH ${img} (platform: ${DOCKER_PLATFORM})"
  docker build --platform "$DOCKER_PLATFORM" -f "$dockerfile" --build-arg "SERVICE_VERSION=${ver}" -t "$img" "$ROOT"
  docker push "$img"
}

apply_seed_admin() {
  log "Creating admin user (seed-admin)"
  kubectl -n "$NS" wait --for=condition=available deployment/auth-service --timeout=300s
  kubectl -n "$NS" exec deployment/auth-service -- /seed-admin
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
  kubectl -n "$NS" wait --for=condition=available deployment/postgres --timeout=600s
  kubectl -n "$NS" wait --for=condition=available deployment/redis --timeout=300s
}

wait_apps() {
  log "Waiting for application deployments"
  local dep
  for dep in $(kubectl -n "$NS" get deploy -o name | sed 's|deployment.apps/||'); do
    case "$dep" in
      postgres|redis|zookeeper|kafka) continue ;;
    esac
    kubectl -n "$NS" rollout status "deployment/${dep}" --timeout=900s || {
      echo "WARN: rollout timeout for ${dep}" >&2
      kubectl -n "$NS" logs "deployment/${dep}" --tail=30 >&2 || true
    }
  done
}

render_stack() {
  local reg ver_auth ver_customers ver_vehicles ver_deals ver_parts ver_brands ver_dealer_points
  local ver_works ver_employees ver_workorders ver_appointments ver_gateway ver_employee_reviews
  local ver_client_auth ver_client_registration ver_client_reviews
  local ver_client_public ver_client_protected ver_employee_stats ver_client_stats

  reg="$(registry_base)"
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
    -e "s|__IMG_AUTH__|${reg}/auth-service:${ver_auth}|g" \
    -e "s|__IMG_CUSTOMERS__|${reg}/customers-service:${ver_customers}|g" \
    -e "s|__IMG_VEHICLES__|${reg}/vehicles-service:${ver_vehicles}|g" \
    -e "s|__IMG_DEALS__|${reg}/deals-service:${ver_deals}|g" \
    -e "s|__IMG_PARTS__|${reg}/parts-service:${ver_parts}|g" \
    -e "s|__IMG_BRANDS__|${reg}/brands-service:${ver_brands}|g" \
    -e "s|__IMG_DEALER_POINTS__|${reg}/dealer-points-service:${ver_dealer_points}|g" \
    -e "s|__IMG_WORKS__|${reg}/works-service:${ver_works}|g" \
    -e "s|__IMG_EMPLOYEES__|${reg}/employees-service:${ver_employees}|g" \
    -e "s|__IMG_WORKORDERS__|${reg}/workorders-service:${ver_workorders}|g" \
    -e "s|__IMG_APPOINTMENTS__|${reg}/appointments-service:${ver_appointments}|g" \
    -e "s|__IMG_GATEWAY__|${reg}/gateway-service:${ver_gateway}|g" \
    -e "s|__IMG_EMPLOYEE_REVIEWS__|${reg}/employee-reviews-service:${ver_employee_reviews}|g" \
    -e "s|__IMG_CLIENT_AUTH__|${reg}/client-auth-service:${ver_client_auth}|g" \
    -e "s|__IMG_CLIENT_REGISTRATION__|${reg}/client-registration-service:${ver_client_registration}|g" \
    -e "s|__IMG_CLIENT_REVIEWS__|${reg}/client-reviews-service:${ver_client_reviews}|g" \
    -e "s|__IMG_CLIENT_PUBLIC_GATEWAY__|${reg}/client-public-gateway-service:${ver_client_public}|g" \
    -e "s|__IMG_CLIENT_PROTECTED_GATEWAY__|${reg}/client-protected-gateway-service:${ver_client_protected}|g" \
    -e "s|__IMG_EMPLOYEE_STATISTICS__|${reg}/employee-statistics-service:${ver_employee_stats}|g" \
    -e "s|__IMG_CLIENT_STATISTICS__|${reg}/client-statistics-service:${ver_client_stats}|g" \
    -e "s|__PULL_POLICY__|Always|g" \
    "$ROOT/k8s/dealer-stack.yaml"
}

render_client_frontend() {
  local reg img
  reg="$(registry_base)"
  img="${reg}/dealer-client-ui:${CLIENT_UI_TAG}"
  sed \
    -e "s|image: dealer-client-ui:latest|image: ${img}|g" \
    -e "s|imagePullPolicy: Never|imagePullPolicy: Always|g" \
    "$ROOT/k8s/client-frontend.yaml"
}

print_status() {
  cat <<EOF

=== Dealer stack deployed to GKE (namespace: ${NS}) ===

Project:  ${GCP_PROJECT}
Cluster:  ${GKE_CLUSTER} (${GCP_ZONE})
Registry: $(registry_base)

Check pods:
  kubectl -n ${NS} get pods

Port-forward examples (local only):
  kubectl -n ${NS} port-forward svc/gateway-service 8090:8090
  kubectl -n ${NS} port-forward svc/client-frontend 3001:3001

Public internet (share with friends, no domain):
  ./scripts/gcp-expose.sh

Logins (after seed):
  Employee: admin@dealer.local / admin123
  Client:   vol.client1@test.dealer.local / Test1234!

EOF
}

main() {
  need_cmd gcloud
  need_cmd docker
  need_cmd kubectl

  if [[ -z "$GCP_PROJECT" || "$GCP_PROJECT" == "(unset)" ]]; then
    echo "ERROR: GCP project not set. Run: gcloud config set project PROJECT_ID" >&2
    exit 1
  fi

  export USE_GKE_GCLOUD_AUTH_PLUGIN="${USE_GKE_GCLOUD_AUTH_PLUGIN:-True}"

  log "Connecting kubectl to ${GKE_CLUSTER} (${GCP_ZONE})"
  gcloud container clusters get-credentials "$GKE_CLUSTER" --zone="$GCP_ZONE" --project="$GCP_PROJECT"

  log "Configuring Docker for Artifact Registry"
  gcloud auth configure-docker "${GCP_REGION}-docker.pkg.dev" --quiet

  if [[ "$SKIP_BUILD" -eq 0 ]]; then
    log "Building and pushing service images"
    build_and_push auth-service build/auth-service.Dockerfile services/employee/auth/VERSION
    build_and_push customers-service build/customers-service.Dockerfile services/employee/customers/VERSION
    build_and_push vehicles-service build/vehicles-service.Dockerfile services/employee/vehicles/VERSION
    build_and_push deals-service build/deals-service.Dockerfile services/employee/deals/VERSION
    build_and_push parts-service build/parts-service.Dockerfile services/employee/parts/VERSION
    build_and_push brands-service build/brands-service.Dockerfile services/employee/brands/VERSION
    build_and_push dealer-points-service build/dealer-points-service.Dockerfile services/employee/dealerpoints/VERSION
    build_and_push works-service build/works-service.Dockerfile services/employee/works/VERSION
    build_and_push employees-service build/employees-service.Dockerfile services/employee/employees/VERSION
    build_and_push workorders-service build/workorders-service.Dockerfile services/employee/workorders/VERSION
    build_and_push appointments-service build/appointments-service.Dockerfile services/employee/appointments/VERSION
    build_and_push gateway-service build/gateway-service.Dockerfile services/gateway/employee/VERSION
    build_and_push employee-reviews-service build/employee-reviews-service.Dockerfile services/employee/reviews/VERSION
    build_and_push client-auth-service build/client-auth-service.Dockerfile services/client/auth/VERSION
    build_and_push client-registration-service build/client-registration-service.Dockerfile services/client/registration/VERSION
    build_and_push client-reviews-service build/client-reviews-service.Dockerfile services/client/reviews/VERSION
    build_and_push client-public-gateway-service build/client-public-gateway-service.Dockerfile services/gateway/client-public/VERSION
    build_and_push client-protected-gateway-service build/client-protected-gateway-service.Dockerfile services/gateway/client-protected/VERSION
    build_and_push employee-statistics-service build/employee-statistics-service.Dockerfile services/statistics/employee/VERSION
    build_and_push client-statistics-service build/client-statistics-service.Dockerfile services/statistics/client/VERSION
    reg="$(registry_base)"
    client_img="${reg}/dealer-client-ui:${CLIENT_UI_TAG}"
    log "BUILD+PUSH ${client_img} (platform: ${DOCKER_PLATFORM})"
    docker build --platform "$DOCKER_PLATFORM" -f "$ROOT/build/client-frontend-dev.Dockerfile" -t "$client_img" "$ROOT"
    docker push "$client_img"
  else
    log "Skipping image build/push (--skip-build)"
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
  render_client_frontend | kubectl apply -f -

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

  wait_apps
  apply_seed_admin
  print_status
}

main "$@"
