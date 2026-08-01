#!/usr/bin/env bash
# Деплой сервисов dealer в Kubernetes.
#
# Заменяет плейсхолдеры __IMG_*__, __PULL_POLICY__, __K8S_DB_HOST__, __K8S_DB_PORT__,
# __NIP_HOST__, __CLIENT_NIP_HOST__ в манифестах services/*/k8s/*.yaml и k8s/*.yaml,
# затем применяет их через kubectl.
#
# Использование:
#   ./scripts/k8s-deploy.sh [services...]     # по умолчанию: all (все сервисы)
#   ./scripts/k8s-deploy.sh auth-service gateway-service
#
# Переменные окружения:
#   REGISTRY            префикс registry (по умолчанию ghcr.io/ВАШ_USER, или host.docker.internal:5050 локально)
#   K8S_DB_HOST         хост Postgres в кластере (по умолчанию postgres)
#   K8S_DB_PORT         порт Postgres (по умолчанию 5432)
#   NIP_HOST            внешний host для ingress (по умолчанию пусто — пропускает подстановку)
#   PULL_POLICY         imagePullPolicy (по умолчанию IfNotPresent)
set -euo pipefail

cd "$(dirname "$0")/.."

REPO_ROOT="$(pwd)"
REGISTRY="${REGISTRY:-ghcr.io/${GITHUB_REPOSITORY_OWNER:-local}}"
K8S_DB_HOST="${K8S_DB_HOST:-postgres}"
K8S_DB_PORT="${K8S_DB_PORT:-5432}"
PULL_POLICY="${PULL_POLICY:-IfNotPresent}"

# image-имя сервиса -> плейсхолдер в манифестах
declare -A IMG=(
  [auth-service]="__IMG_AUTH__"
  [customers-service]="__IMG_CUSTOMERS__"
  [vehicles-service]="__IMG_VEHICLES__"
  [deals-service]="__IMG_DEALS__"
  [parts-service]="__IMG_PARTS__"
  [brands-service]="__IMG_BRANDS__"
  [dealer-points-service]="__IMG_DEALER_POINTS__"
  [workorders-service]="__IMG_WORKORDERS__"
  [works-service]="__IMG_WORKS__"
  [employees-service]="__IMG_EMPLOYEES__"
  [employee-reviews-service]="__IMG_EMPLOYEE_REVIEWS__"
  [appointments-service]="__IMG_APPOINTMENTS__"
  [client-auth-service]="__IMG_CLIENT_AUTH__"
  [client-registration-service]="__IMG_CLIENT_REGISTRATION__"
  [client-reviews-service]="__IMG_CLIENT_REVIEWS__"
  [gateway-service]="__IMG_GATEWAY__"
  [client-public-gateway-service]="__IMG_CLIENT_PUBLIC_GATEWAY__"
  [client-protected-gateway-service]="__IMG_CLIENT_PROTECTED_GATEWAY__"
  [employee-statistics-service]="__IMG_EMPLOYEE_STATISTICS__"
  [client-statistics-service]="__IMG_CLIENT_STATISTICS__"
  [errors-ingest-service]="__IMG_ERRORS_INGEST__"
  [scheduler-service]="__IMG_SCHEDULER__"
)

# сервис -> директория модуля (для VERSION)
declare -A MOD=(
  [auth-service]="services/employee/auth"
  [customers-service]="services/employee/customers"
  [vehicles-service]="services/employee/vehicles"
  [deals-service]="services/employee/deals"
  [parts-service]="services/employee/parts"
  [brands-service]="services/employee/brands"
  [dealer-points-service]="services/employee/dealerpoints"
  [workorders-service]="services/employee/workorders"
  [works-service]="services/employee/works"
  [employees-service]="services/employee/employees"
  [employee-reviews-service]="services/employee/reviews"
  [appointments-service]="services/employee/appointments"
  [client-auth-service]="services/client/auth"
  [client-registration-service]="services/client/registration"
  [client-reviews-service]="services/client/reviews"
  [gateway-service]="services/gateway/employee"
  [client-public-gateway-service]="services/gateway/client-public"
  [client-protected-gateway-service]="services/gateway/client-protected"
  [employee-statistics-service]="services/statistics/employee"
  [client-statistics-service]="services/statistics/client"
  [errors-ingest-service]="services/errors-ingest"
  [scheduler-service]="services/scheduler"
)

if [ "$#" -gt 0 ]; then
  SERVICES=("$@")
else
  SERVICES=("${!IMG[@]}")
fi

echo "Registry:  $REGISTRY"
echo "Services:  ${SERVICES[*]}"
echo "PullPolicy: $PULL_POLICY"

WORKDIR="$(mktemp -d)"
trap 'rm -rf "$WORKDIR"' EXIT

# 1. Заполняем плейсхолдеры __IMG_*__ из VERSION каждого сервиса
SUBST=()
for svc in "${SERVICES[@]}"; do
  placeholder="${IMG[$svc]:-}"
  mod="${MOD[$svc]:-}"
  [ -n "$placeholder" ] || { echo "skip $svc: no placeholder mapping"; continue; }
  if [ -n "$mod" ] && [ -f "$REPO_ROOT/$mod/VERSION" ]; then
    ver="$(tr -d '[:space:]' < "$REPO_ROOT/$mod/VERSION")"
    SUBST+=("-e" "s|${placeholder}|${REGISTRY}/${svc}:${ver}|g")
  fi
done

SUBST+=( "-e" "s|__PULL_POLICY__|${PULL_POLICY}|g" )
SUBST+=( "-e" "s|__K8S_DB_HOST__|${K8S_DB_HOST}|g" )
SUBST+=( "-e" "s|__K8S_DB_PORT__|${K8S_DB_PORT}|g" )
if [ -n "${NIP_HOST:-}" ]; then
  SUBST+=( "-e" "s|__NIP_HOST__|${NIP_HOST}|g" )
  SUBST+=( "-e" "s|__CLIENT_NIP_HOST__|${NIP_HOST}|g" )
fi

# 2. Копируем манифесты с подстановкой
apply_files=()
for svc in "${SERVICES[@]}"; do
  mod="${MOD[$svc]:-}"
  [ -n "$mod" ] || continue
  for src in "$REPO_ROOT/$mod"/k8s/*.yaml; do
    [ -e "$src" ] || continue
    dst="$WORKDIR/$(basename "$(dirname "$src")")-$(basename "$src")"
    sed "${SUBST[@]}" "$src" > "$dst"
    apply_files+=("$dst")
  done
done

# dealer-stack.yaml (инфраструктура + все сервисы) всегда применяем
if [ -f "$REPO_ROOT/k8s/dealer-stack.yaml" ]; then
  dst="$WORKDIR/dealer-stack.yaml"
  sed "${SUBST[@]}" "$REPO_ROOT/k8s/dealer-stack.yaml" > "$dst"
  apply_files+=("$dst")
fi

# 3. Применяем
for f in "${apply_files[@]}"; do
  echo ">>> kubectl apply -f $f"
  kubectl apply -f "$f"
done

echo "Done."
