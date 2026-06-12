#!/usr/bin/env bash
# Пересобрать и перезапустить только клиентский UI (полный стек: ./scripts/kube-up.sh).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
NS="${DEALER_NS:-dealer}"

eval "$(minikube docker-env)"
export DOCKER_BUILDKIT=0

docker rm -f dealer-client-ui 2>/dev/null || true
docker build -f "${ROOT}/build/client-frontend.Dockerfile" -t dealer-client-ui:latest "$ROOT"
kubectl apply -f "${ROOT}/k8s/client-frontend.yaml"
kubectl -n "$NS" rollout restart deployment/client-frontend
kubectl -n "$NS" rollout status deployment/client-frontend --timeout=120s
"${ROOT}/scripts/expose-lan.sh"
