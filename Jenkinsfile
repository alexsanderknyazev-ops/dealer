// =============================================================================
// Jenkinsfile — CI dealer (Jenkins + Sonar + registry + Minikube)
// =============================================================================
// Поток стадий:
//   1. Checkout — код из SCM.
//   2. Go test + coverage — go test, coverage.out.
//   3. SonarQube — sonar-scanner + Node, токен из credentials.
//   4. Docker build and push — вложенные stage (Docker: prepare, Docker: auth-service, …) для наглядности в UI; логика в scripts/ci/jenkins-docker.sh; тег = services/<svc>/VERSION; skip если тег уже в registry.
//   5. Deploy to Kubernetes (опционально) — apply только изменённых сервисов из services/*/k8s + новые migration-файлы.
//      Доступ к API: либо kubeconfig на агенте (любой драйвер minikube: qemu2, kvm, docker, …), либо ветка docker exec — только при --driver=docker (контейнер-нода на хосте).
//
// Инфраструктура поднимается отдельно, pipeline деплоит сервисы из services/*/k8s в namespace dealer (см. параметр K8S_NAMESPACE).
// =============================================================================

pipeline {
  agent any

  parameters {
    // --- SonarQube: доп. аргументы sonar-scanner (переопределение properties из UI) ---
    string(
      name: 'SONAR_EXTRA_OPTS',
      defaultValue: '',
      description: 'Доп. аргументы sonar-scanner (переопределяют properties), напр. -Dsonar.projectKey=КЛЮЧ_ИЗ_SONAR_UI'
    )
    // --- Docker registry (Skopeo push после сборки всех сервисов) ---
    string(
      name: 'DOCKER_REGISTRY',
      defaultValue: 'host.docker.internal:5050',
      description: 'Docker registry host:port. Jenkins в Docker → host.docker.internal:5050; агент на хосте → localhost:5050'
    )
    // --- Kubernetes / Minikube: включение деплоя, контейнер minikube, registry для pull-манифеста ---
    booleanParam(
      name: 'DEPLOY',
      defaultValue: false,
      description: 'После push: deploy изменённых сервисов из services/*/k8s + применение новых migration-файлов.'
    )
    string(
      name: 'KUBECONFIG_PATH',
      defaultValue: '',
      description: 'Необязательно: абсолютный путь к kubeconfig внутри агента (если не JENKINS_HOME/.kube/config). Удобно при volume, напр. /var/jenkins_home/secrets/kubeconfig.'
    )
    string(
      name: 'K8S_PULL_REGISTRY',
      defaultValue: 'host.minikube.internal:5050',
      description: 'Registry host:port для pull образов в Kubernetes'
    )
    string(
      name: 'K8S_NAMESPACE',
      defaultValue: 'dealer',
      description: 'Kubernetes namespace для сервисных манифестов'
    )
    string(
      name: 'POSTGRES_PASSWORD',
      defaultValue: '',
      description: 'Обязательный пароль БД dealer для k8s. На деплое Jenkins кладёт его в Secret dealer-db (POSTGRES_PASSWORD, POSTGRES_DSN).'
    )
    password(
      name: 'JWT_SECRET',
      defaultValue: '',
      description: 'Обязательный JWT secret для сервисов в k8s (Secret dealer-app-secrets).'
    )
  }

  // Переменные окружения для стадий Sonar/Go (версии инструментов; SONAR_HOST_URL — к Sonar в Docker на хосте).
  environment {
    SONAR_HOST_URL = 'http://host.docker.internal:9000'
    // Совпадает с `toolchain` в go.mod (корневой модуль dealer).
    GO_VERSION = '1.24.11'
    SONAR_SCANNER_VERSION = '8.0.1.6346'
    // Для сенсоров JS/TS/CSS Sonar нужен Node.js в PATH.
    NODE_JS_VERSION = '20.18.1'
  }

  stages {
    // --- Клонирование репозитория ---
    stage('Checkout') {
      steps {
        checkout scm
      }
    }

    stage('Detect changed modules') {
      steps {
        sh '''#!/bin/bash
set -euo pipefail
cd "${WORKSPACE}"
mkdir -p .ci

RANGE=""
if [ -n "${CHANGE_TARGET:-}" ]; then
  git fetch --no-tags origin "${CHANGE_TARGET}" || true
  BASE="$(git merge-base HEAD "origin/${CHANGE_TARGET}" || true)"
  if [ -n "${BASE}" ]; then RANGE="${BASE}..HEAD"; fi
fi
if [ -z "${RANGE}" ] && git rev-parse HEAD~1 >/dev/null 2>&1; then
  RANGE="HEAD~1..HEAD"
fi

if [ -n "${RANGE}" ]; then
  git diff --name-only "${RANGE}" > .ci/changed-files.txt
  git diff --name-status "${RANGE}" > .ci/changed-status.txt
else
  git ls-files > .ci/changed-files.txt
  : > .ci/changed-status.txt
fi

all_changed=0
if rg -n '^(pkg/|api/|proto/|go\\.mod|go\\.sum|go\\.work|go\\.work\\.sum|build/|scripts/ci/|Jenkinsfile)' .ci/changed-files.txt >/dev/null 2>&1; then
  all_changed=1
fi

services=()
if [ "${all_changed}" -eq 1 ]; then
  services=(auth-service customers-service vehicles-service deals-service parts-service brands-service dealer-points-service)
else
  rg -n '^services/auth/' .ci/changed-files.txt >/dev/null 2>&1 && services+=(auth-service)
  rg -n '^services/customers/' .ci/changed-files.txt >/dev/null 2>&1 && services+=(customers-service)
  rg -n '^services/vehicles/' .ci/changed-files.txt >/dev/null 2>&1 && services+=(vehicles-service)
  rg -n '^services/deals/' .ci/changed-files.txt >/dev/null 2>&1 && services+=(deals-service)
  rg -n '^services/parts/' .ci/changed-files.txt >/dev/null 2>&1 && services+=(parts-service)
  rg -n '^services/brands/' .ci/changed-files.txt >/dev/null 2>&1 && services+=(brands-service)
  rg -n '^services/dealerpoints/' .ci/changed-files.txt >/dev/null 2>&1 && services+=(dealer-points-service)
fi

uniq=()
for s in "${services[@]}"; do
  exists=0
  for u in "${uniq[@]}"; do [ "$u" = "$s" ] && exists=1 && break; done
  [ "$exists" -eq 0 ] && uniq+=("$s")
done

printf '%s\n' "${uniq[@]}" > .ci/changed-services.txt
new_migrations="$(awk '$1=="A" && $2 ~ /^migrations\\/.*\\.up\\.sql$/ {print $2}' .ci/changed-status.txt | sort -u | tr '\n' ' ' | sed 's/[[:space:]]*$//')"

{
  if [ "${#uniq[@]}" -gt 0 ]; then
    echo 'HAS_SERVICE_CHANGES=true'
    echo "CHANGED_SERVICES=\"${uniq[*]}\""
  else
    echo 'HAS_SERVICE_CHANGES=false'
    echo 'CHANGED_SERVICES=""'
  fi
  if [ -n "${new_migrations}" ]; then
    echo 'HAS_NEW_MIGRATIONS=true'
    echo "NEW_MIGRATIONS=\"${new_migrations}\""
  else
    echo 'HAS_NEW_MIGRATIONS=false'
    echo 'NEW_MIGRATIONS=""'
  fi
} > .ci/changed.env

echo "Changed services: ${uniq[*]:-(none)}"
echo "New migration files: ${new_migrations:-(none)}"
'''
      }
    }

    // --- Тесты Go и покрытие (артефакт coverage.out для Sonar) ---
    stage('Go test + coverage') {
      steps {
        // GOTOOLCHAIN=local до любого вызова go: иначе под auto может подтянуться другой toolchain, чем бинарь в /usr/local/go.
        // Проверяем только /usr/local/go/bin/go — не «go» из PATH с другим поведением.
        sh """#!/bin/bash
set -eux
GO_VER='${env.GO_VERSION}'
if [ -z "\${GO_VER}" ]; then GO_VER='1.24.11'; fi
export GOTOOLCHAIN=local
export PATH="/usr/local/go/bin:\${PATH}"

ARCH="\$(uname -m)"
case "\$ARCH" in
  aarch64|arm64) GOARCH=arm64 ;;
  x86_64) GOARCH=amd64 ;;
  *) echo "unsupported arch: \$ARCH"; exit 1 ;;
esac

if [ -x /usr/local/go/bin/go ] && /usr/local/go/bin/go version 2>/dev/null | grep -qF "go\${GO_VER}"; then
  echo "Go already at \${GO_VER} under /usr/local/go"
else
  GOURL="https://go.dev/dl/go\${GO_VER}.linux-\${GOARCH}.tar.gz"
  # Повреждённый .tar.gz (обрыв сети/прокси) даёт «gzip: invalid compressed data» — проверяем gzip до rm /usr/local/go.
  for attempt in 1 2 3; do
    echo "Downloading Go \${GO_VER} (\${GOARCH}), attempt \${attempt}"
    curl -fSL --connect-timeout 30 --max-time 600 --retry 5 --retry-delay 2 "\${GOURL}" -o /tmp/go.tgz
    if gzip -t /tmp/go.tgz 2>/dev/null; then break; fi
    echo "go.tgz is not valid gzip, retrying"
    rm -f /tmp/go.tgz
    if [ "\${attempt}" -eq 3 ]; then echo "Giving up after 3 attempts"; exit 1; fi
  done
  rm -rf /usr/local/go
  tar -C /usr/local -xzf /tmp/go.tgz
fi

export GOROOT=/usr/local/go
export GOTOOLCHAIN=local

go version
cd "\${WORKSPACE}"
# Не использовать общий /root/go/pkg/mod на агенте: при полном диске/обрыве скачивания там остаются пустые .go → «expected package, found EOF».
export GOMODCACHE="\${WORKSPACE}/.gomodcache"
export GOCACHE="\${WORKSPACE}/.gocache"
mkdir -p "\${GOMODCACHE}" "\${GOCACHE}"
# Корень + каждый модуль из go.work: один go test ./... из корня не включает services/* (отдельные go.mod).
# Склеиваем coverprofile (mode + строки блоков); без -coverpkg — в отчёт попадают все прогнанные пакеты.
rm -f coverage.out
first=1
for d in . services/auth services/customers services/vehicles services/deals services/parts services/brands services/dealerpoints; do
  (cd "\${d}" && go test ./... -coverprofile=cov_piece.out -covermode=atomic)
  if [ "\${first}" -eq 1 ]; then
    mv "\${d}/cov_piece.out" coverage.out
    first=0
  else
    tail -n +2 "\${d}/cov_piece.out" >> coverage.out
    rm -f "\${d}/cov_piece.out"
  fi
done
rm -f cov_piece.out
"""
      }
    }

    stage('Go lint (changed services)') {
      steps {
        sh '''#!/bin/bash
set -euo pipefail
cd "${WORKSPACE}"
. "${WORKSPACE}/.ci/changed.env"

export GOBIN="${WORKSPACE}/.ci/bin"
mkdir -p "${GOBIN}"
export PATH="${GOBIN}:${PATH}"

if ! command -v golangci-lint >/dev/null 2>&1; then
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
fi

targets=()
if [ "${HAS_SERVICE_CHANGES}" = "true" ]; then
  for svc in ${CHANGED_SERVICES}; do
    case "${svc}" in
      auth-service) targets+=(services/auth) ;;
      customers-service) targets+=(services/customers) ;;
      vehicles-service) targets+=(services/vehicles) ;;
      deals-service) targets+=(services/deals) ;;
      parts-service) targets+=(services/parts) ;;
      brands-service) targets+=(services/brands) ;;
      dealer-points-service) targets+=(services/dealerpoints) ;;
      *) echo "Unknown changed service: ${svc}" >&2; exit 1 ;;
    esac
  done
else
  echo "No changed services detected for lint."
  exit 0
fi

for d in "${targets[@]}"; do
  echo "Linting ${d}"
  (cd "${d}" && golangci-lint run ./...)
done
'''
      }
    }

    // --- Статический анализ SonarQube (sonar-project.properties в корне); при FAILED Quality Gate сканер падает с кодом 3 ---
    stage('SonarQube analysis') {
      environment {
        SONAR_TOKEN = credentials('dealer-app')
      }
      steps {
        sh """#!/bin/bash
set -eux
if ! command -v curl >/dev/null 2>&1 || ! command -v unzip >/dev/null 2>&1 || ! command -v xz >/dev/null 2>&1; then
  apt-get update -qq
  apt-get install -y -qq curl ca-certificates unzip xz-utils
fi

ARCH="\$(uname -m)"
case "\$ARCH" in
  aarch64|arm64) ZIP_ARCH=aarch64; NODE_DIST_ARCH=arm64 ;;
  x86_64) ZIP_ARCH=x64; NODE_DIST_ARCH=x64 ;;
  *) echo "unsupported arch: \$ARCH"; exit 1 ;;
esac

NODE_VER='${env.NODE_JS_VERSION}'
if [ -z "\${NODE_VER}" ]; then NODE_VER='20.18.1'; fi
NODE_BASE="node-v\${NODE_VER}-linux-\${NODE_DIST_ARCH}"
NODE_ROOT="/usr/local/\${NODE_BASE}"
if [ ! -x "\${NODE_ROOT}/bin/node" ]; then
  curl -fsSL "https://nodejs.org/dist/v\${NODE_VER}/\${NODE_BASE}.tar.xz" -o /tmp/node.txz
  tar -C /usr/local -xJf /tmp/node.txz
fi
export PATH="\${NODE_ROOT}/bin:\${PATH}"
node -v

ZIP="sonar-scanner-cli-${env.SONAR_SCANNER_VERSION}-linux-\${ZIP_ARCH}.zip"
URL="https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/\${ZIP}"
curl -fsSL "\$URL" -o "/tmp/\${ZIP}"
# Родитель нельзя называть sonar-scanner-* — find совпадёт с ним раньше, чем с каталогом из zip.
SCANNER_ROOT=/tmp/ss-unpack
rm -rf "\${SCANNER_ROOT}"
mkdir -p "\${SCANNER_ROOT}"
unzip -q -o "/tmp/\${ZIP}" -d "\${SCANNER_ROOT}"
SCANNER_HOME="\$(find "\${SCANNER_ROOT}" -maxdepth 1 -mindepth 1 -type d -name 'sonar-scanner-*' | head -1)"
test -x "\${SCANNER_HOME}/bin/sonar-scanner"

cd "\${WORKSPACE}"
SONAR_EXTRA_OPTS='${params.SONAR_EXTRA_OPTS}'
"\${SCANNER_HOME}/bin/sonar-scanner" \\
  -Dsonar.host.url="${env.SONAR_HOST_URL}" \\
  -Dsonar.token="\${SONAR_TOKEN}" \\
  -Dsonar.scm.revision="\$(git rev-parse HEAD)" \\
  \${SONAR_EXTRA_OPTS}
"""
      }
    }

    stage('Docker build and push (changed only)') {
      steps {
        sh """#!/bin/bash
set -eux
export DOCKER_REGISTRY='${params.DOCKER_REGISTRY}'
export BUILD_NUMBER='${env.BUILD_NUMBER}'
cd "\${WORKSPACE}"
bash scripts/ci/jenkins-docker.sh prepare
. "\${WORKSPACE}/.ci/changed.env"

if [ "\${HAS_SERVICE_CHANGES}" != "true" ]; then
  echo "Нет изменений в сервисных модулях — docker build/push пропускаем."
  exit 0
fi

for svc in \${CHANGED_SERVICES}; do
  case "\${svc}" in
    auth-service) df='build/auth-service.Dockerfile' ;;
    customers-service) df='build/customers-service.Dockerfile' ;;
    vehicles-service) df='build/vehicles-service.Dockerfile' ;;
    deals-service) df='build/deals-service.Dockerfile' ;;
    parts-service) df='build/parts-service.Dockerfile' ;;
    brands-service) df='build/brands-service.Dockerfile' ;;
    dealer-points-service) df='build/dealer-points-service.Dockerfile' ;;
    *) echo "Unknown service: \${svc}" >&2; exit 1 ;;
  esac
  bash scripts/ci/jenkins-docker.sh build "\${svc}" "\${df}"
done
"""
      }
    }

    // --- Деплой в Minikube: k8s/dealer-stack.yaml (все сервисы), только если DEPLOY=true ---
    stage('Deploy to Minikube') {
      when {
        expression { return params.DEPLOY }
      }
      steps {
        sh """#!/bin/bash
set -euo pipefail
cd "\${WORKSPACE}"
if [ -x "\${WORKSPACE}/.ci/docker-cli-bin/docker" ]; then
  export PATH="\${WORKSPACE}/.ci/docker-cli-bin:\${PATH}"
fi
command -v docker >/dev/null 2>&1 || true

KUBECTL=""
KP='${params.KUBECONFIG_PATH}'
if [ -n "\$KP" ]; then
  export KUBECONFIG="\$KP"
fi
if command -v kubectl >/dev/null 2>&1; then
  KUBECTL="kubectl"
else
  ARCH="\$(uname -m)"
  case "\$ARCH" in aarch64|arm64) KARCH=arm64 ;; x86_64) KARCH=amd64 ;; *) echo "unsupported arch: \$ARCH"; exit 1 ;; esac
  KVER="\$(curl -fsSL https://dl.k8s.io/release/stable.txt)"
  KUBECTL="/tmp/kubectl-\${KVER}"
  if [ ! -x "\$KUBECTL" ]; then
    curl -fSL "https://dl.k8s.io/release/\${KVER}/bin/linux/\${KARCH}/kubectl" -o "\$KUBECTL"
    chmod +x "\$KUBECTL"
  fi
fi

if [ ! -f "\${WORKSPACE}/.ci/image-versions.env" ]; then
  echo "Нет .ci/image-versions.env — сначала должна пройти стадия Docker build." >&2
  exit 1
fi

# shellcheck disable=SC1090
. "\${WORKSPACE}/.ci/image-versions.env"
. "\${WORKSPACE}/.ci/changed.env"

NS='${params.K8S_NAMESPACE}'
K8S_PULL_REG='${params.K8S_PULL_REGISTRY}'
POSTGRES_PASSWORD='${params.POSTGRES_PASSWORD}'
JWT_SECRET='${params.JWT_SECRET}'
if [ -z "\${POSTGRES_PASSWORD}" ]; then
  echo "POSTGRES_PASSWORD is required for deploy stage" >&2
  exit 1
fi
if [ -z "\${JWT_SECRET}" ]; then
  echo "JWT_SECRET is required for deploy stage" >&2
  exit 1
fi
POSTGRES_DSN="postgres://dealer:\${POSTGRES_PASSWORD}@postgres:5432/dealer?sslmode=disable"

if [ "\${HAS_SERVICE_CHANGES}" != "true" ] && [ "\${HAS_NEW_MIGRATIONS}" != "true" ]; then
  echo "Нет изменений для deploy."
  exit 0
fi

kctl() { "\$KUBECTL" "\$@"; }

set +x
kctl create namespace "\$NS" --dry-run=client -o yaml | kctl apply -f -
kctl -n "\$NS" create secret generic dealer-db \
  --from-literal=POSTGRES_PASSWORD="\$POSTGRES_PASSWORD" \
  --from-literal=POSTGRES_DSN="\$POSTGRES_DSN" \
  --dry-run=client -o yaml | kctl apply -f -
kctl -n "\$NS" create secret generic dealer-app-secrets \
  --from-literal=JWT_SECRET="\$JWT_SECRET" \
  --dry-run=client -o yaml | kctl apply -f -
set -x

apply_service() {
  local svc="$1" img="$2" dep="" svcf=""
  case "\$svc" in
    auth-service)
      dep="services/auth/k8s/auth-deployment.yaml"; svcf="services/auth/k8s/auth-service.yaml"
      sed -e "s|__IMG_AUTH__|\${img}|g" -e "s|__PULL_POLICY__|Always|g" "\$dep" | kctl apply -f -
      ;;
    customers-service)
      dep="services/customers/k8s/customer-deployment.yaml"; svcf="services/customers/k8s/customers-service.yaml"
      sed -e "s|__IMG_CUSTOMERS__|\${img}|g" -e "s|__PULL_POLICY__|Always|g" "\$dep" | kctl apply -f -
      ;;
    vehicles-service)
      dep="services/vehicles/k8s/vehicles-deployment.yaml"; svcf="services/vehicles/k8s/vehicles-service.yaml"
      sed -e "s|__IMG_VEHICLES__|\${img}|g" -e "s|__PULL_POLICY__|Always|g" "\$dep" | kctl apply -f -
      ;;
    deals-service)
      dep="services/deals/k8s/deals-deployment.yaml"; svcf="services/deals/k8s/deals-service.yaml"
      sed -e "s|__IMG_DEALS__|\${img}|g" -e "s|__PULL_POLICY__|Always|g" "\$dep" | kctl apply -f -
      ;;
    parts-service)
      dep="services/parts/k8s/parts-deployment.yaml"; svcf="services/parts/k8s/parts-service.yaml"
      sed -e "s|__IMG_PARTS__|\${img}|g" -e "s|__PULL_POLICY__|Always|g" "\$dep" | kctl apply -f -
      ;;
    brands-service)
      dep="services/brands/k8s/brand-deployment.yaml"; svcf="services/brands/k8s/brand-service.yaml"
      sed -e "s|__IMG_BRANDS__|\${img}|g" -e "s|__PULL_POLICY__|Always|g" "\$dep" | kctl apply -f -
      ;;
    dealer-points-service)
      dep="services/dealerpoints/k8s/dealerpoints-deployment.yaml"; svcf="services/dealerpoints/k8s/dealerpoints-service.yaml"
      sed -e "s|__IMG_DEALER_POINTS__|\${img}|g" -e "s|__PULL_POLICY__|Always|g" "\$dep" | kctl apply -f -
      ;;
    *) echo "Unknown service \${svc}" >&2; exit 1 ;;
  esac
  kctl apply -f "\$svcf"
  kctl -n "\$NS" rollout status "deployment/\$svc" --timeout=300s
}

if [ "\${HAS_SERVICE_CHANGES}" = "true" ]; then
  for svc in \${CHANGED_SERVICES}; do
    case "\$svc" in
      auth-service) IMG="\${K8S_PULL_REG}/auth-service:\${VER_AUTH_SERVICE}" ;;
      customers-service) IMG="\${K8S_PULL_REG}/customers-service:\${VER_CUSTOMERS_SERVICE}" ;;
      vehicles-service) IMG="\${K8S_PULL_REG}/vehicles-service:\${VER_VEHICLES_SERVICE}" ;;
      deals-service) IMG="\${K8S_PULL_REG}/deals-service:\${VER_DEALS_SERVICE}" ;;
      parts-service) IMG="\${K8S_PULL_REG}/parts-service:\${VER_PARTS_SERVICE}" ;;
      brands-service) IMG="\${K8S_PULL_REG}/brands-service:\${VER_BRANDS_SERVICE}" ;;
      dealer-points-service) IMG="\${K8S_PULL_REG}/dealer-points-service:\${VER_DEALER_POINTS_SERVICE}" ;;
      *) echo "Unknown changed service \${svc}" >&2; exit 1 ;;
    esac
    apply_service "\$svc" "\$IMG"
  done
fi

if [ "\${HAS_NEW_MIGRATIONS}" = "true" ]; then
  kctl -n "\$NS" wait --for=condition=available deployment/postgres --timeout=180s
  for f in \${NEW_MIGRATIONS}; do
    test -f "\${WORKSPACE}/\${f}"
    kctl -n "\$NS" exec -i deployment/postgres -- env PGPASSWORD="\$POSTGRES_PASSWORD" psql -U dealer -d dealer -v ON_ERROR_STOP=1 -f - < "\${WORKSPACE}/\${f}"
  done
fi
"""
      }
    }
  }
}
