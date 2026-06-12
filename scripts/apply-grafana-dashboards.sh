#!/usr/bin/env bash
# Обновить ConfigMap с дашбордами Grafana из k8s/grafana/dashboards/*.json
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
NS="${MONITORING_NS:-monitoring}"
DIR="${ROOT}/k8s/grafana/dashboards"

if [[ ! -d "$DIR" ]]; then
  echo "Dashboard dir not found: $DIR" >&2
  exit 1
fi

count=$(find "$DIR" -maxdepth 1 -name '*.json' | wc -l)
if [[ "$count" -eq 0 ]]; then
  echo "No JSON dashboards in $DIR" >&2
  exit 1
fi

kubectl create namespace "$NS" --dry-run=client -o yaml | kubectl apply -f -

kubectl create configmap grafana-dashboards -n "$NS" \
  --from-file="$DIR" \
  --dry-run=client -o yaml | kubectl apply -f -

echo "Applied grafana-dashboards ($count files) to namespace $NS"

if kubectl get deployment/grafana -n "$NS" &>/dev/null; then
  kubectl -n "$NS" rollout restart deployment/grafana
  kubectl -n "$NS" rollout status deployment/grafana --timeout=120s
  echo "Grafana restarted — dashboards will reload in ~30s"
fi
