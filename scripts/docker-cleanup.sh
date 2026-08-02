#!/usr/bin/env bash
# Безопасная чистка Docker: build cache + неиспользуемые образы. Volumes не трогаются.
#
# Usage:
#   ./scripts/docker-cleanup.sh              # безопасный режим (build cache + образы)
#   ./scripts/docker-cleanup.sh --volumes    # + удалить неиспользуемые volumes (⚠️ локальная БД!)
#   ./scripts/docker-cleanup.sh --minikube   # чистка build cache внутри minikube (НТ-стенд)
#   ./scripts/docker-cleanup.sh --help
set -euo pipefail

VOLUMES=0
MINIKUBE=0

for arg in "$@"; do
  case "$arg" in
    --volumes) VOLUMES=1 ;;
    --minikube) MINIKUBE=1 ;;
    --help|-h) sed -n '2,9p' "$0"; exit 0 ;;
    *) echo "Unknown option: $arg" >&2; exit 1 ;;
  esac
done

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "ERROR: required command not found: $1" >&2
    exit 1
  }
}

if [[ "$MINIKUBE" -eq 1 ]]; then
  need_cmd minikube
  echo "==> Cleaning minikube docker (build cache + dangling, keep stand images)"
  eval "$(minikube docker-env)"
  docker builder prune -af
  docker image prune -f
  echo "==> Done"
  exit 0
fi

need_cmd docker
docker info >/dev/null 2>&1 || {
  echo "ERROR: docker daemon is not running" >&2
  exit 1
}

echo "==> Before:"
docker system df
echo

echo "==> 1. Build cache (главный источник мусора)"
docker builder prune -af

echo "==> 2. Неиспользуемые образы"
docker image prune -af

if [[ "$VOLUMES" -eq 1 ]]; then
  echo
  echo "!! --volumes: будут удалены ВСЕ volumes, не привязанные к запущенным контейнерам."
  echo "!! Это включает postgres_data / clickhouse_data локального compose-стенда (потеря БД и данных)."
  read -r -p "Продолжить? [y/N] " ans
  [[ "${ans,,}" == "y" ]] || { echo "Отменено."; exit 0; }
  docker volume prune -af
fi

echo
echo "==> After:"
docker system df

cat <<EOF

Подсказка:
  - Образы остановленных стендов (docker compose / minikube) могут быть удалены шагом 2 —
    следующий запуск пересоберёт их. Данные (volumes) при этом сохраняются.
  - Чистка внутри minikube (НТ-стенд): ./scripts/docker-cleanup.sh --minikube
EOF
