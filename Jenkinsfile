// =============================================================================
// Jenkinsfile — CI dealer (Jenkins + Sonar + registry + Minikube)
// =============================================================================
// Поток стадий:
//   1. Checkout — код из SCM.
//   2. Go test + coverage — go test, coverage.out.
//   3. SonarQube — sonar-scanner + Node, токен из credentials.
//   4. Docker build and push — вложенные stage (Docker: prepare, Docker: auth-service, …) для наглядности в UI; логика в scripts/ci/jenkins-docker.sh; тег = services/<svc>/VERSION; skip если тег уже в registry.
//   5. Deploy to Minikube (опционально) — при K8S_DELETE_NS_BEFORE_DEPLOY удаление namespace, apply, rollout; при K8S_BOOTSTRAP_DEV_DATA — миграции, сиды, seed-admin.
//      Доступ к API: либо kubeconfig на агенте (любой драйвер minikube: qemu2, kvm, docker, …), либо ветка docker exec — только при --driver=docker (контейнер-нода на хосте).
//
// Деплой k8s/dealer-stack.yaml поднимает Postgres, Redis, Zookeeper, Kafka и 7 сервисов в namespace dealer (см. параметр K8S_NAMESPACE).
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
      description: 'После push: kubectl apply k8s/dealer-stack.yaml. Нужен kubeconfig: JENKINS_HOME/.kube/config или параметр KUBECONFIG_PATH (qemu2/kvm); либо minikube --driver=docker + docker.sock. При qemu2 контейнера minikube на хосте нет.'
    )
    booleanParam(
      name: 'K8S_BOOTSTRAP_DEV_DATA',
      defaultValue: false,
      description: 'После деплоя: миграции (если нет users), seed_test_data + seed_dealer_brands + seed_parts, /seed-admin (admin@dealer.local / admin123). Только лаборатория; на прод выставьте false. Не видно в UI — первый прогон / Scan Multibranch.'
    )
    booleanParam(
      name: 'K8S_DELETE_NS_BEFORE_DEPLOY',
      defaultValue: false,
      description: 'Перед kubectl apply: удалить namespace K8S_NAMESPACE (чистый деплой, PVC сбрасывается). На общем кластере — false.'
    )
    string(
      name: 'KUBECONFIG_PATH',
      defaultValue: '',
      description: 'Необязательно: абсолютный путь к kubeconfig внутри агента (если не JENKINS_HOME/.kube/config). Удобно при volume, напр. /var/jenkins_home/secrets/kubeconfig.'
    )
    string(
      name: 'MINIKUBE_CONTAINER',
      defaultValue: 'minikube',
      description: 'Только для minikube --driver=docker: имя контейнера-ноды (docker ps). При qemu2/kvm игнорируйте; настройте .kube/config с API VM.'
    )
    string(
      name: 'K8S_PULL_REGISTRY',
      defaultValue: 'host.minikube.internal:5050',
      description: 'Registry host:port для образа в манифесте, если K8S_CTR_IMPORT_IMAGE=false'
    )
    // Локальная загрузка образа в dockerd ноды minikube (см. комментарий в stage Deploy).
    booleanParam(
      name: 'K8S_CTR_IMPORT_IMAGE',
      defaultValue: true,
      description: 'Только при ветке docker exec (minikube --driver=docker): save|load в dockerd ноды; imagePullPolicy Never. Для qemu2/kvm — false: образы из K8S_PULL_REGISTRY (VM тянет по сети).'
    )
    string(
      name: 'K8S_NAMESPACE',
      defaultValue: 'dealer',
      description: 'Kubernetes namespace для k8s/dealer-stack.yaml'
    )
    // --- БД для подов (плейсхолдеры __K8S_DB_HOST__ / __K8S_DB_PORT__ в k8s/dealer-stack.yaml) ---
    string(
      name: 'K8S_DB_HOST',
      defaultValue: 'postgres',
      description: 'Postgres host из подов (например Service postgres в namespace dealer) или host.minikube.internal'
    )
    string(
      name: 'K8S_DB_PORT',
      defaultValue: '5432',
      description: 'Порт Postgres: 5432 in-cluster; 5433 если БД на хосте (property.yaml)'
    )
    string(
      name: 'POSTGRES_PASSWORD',
      defaultValue: 'changeme',
      description: 'Пароль БД dealer для k8s (плейсхолдер __POSTGRES_PASSWORD__ в dealer-stack.yaml). Смените вне лаборатории; для sed избегайте символов | \\ & в значении.'
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

    // --- Тесты Go и покрытие (артефакт coverage.out для Sonar) ---
    stage('Go test + coverage') {
      steps {
        // GOTOOLCHAIN=local до любого вызова go: иначе под auto может подтянуться другой toolchain, чем бинарь в /usr/local/go.
        // Проверяем только /usr/local/go/bin/go — не «go» из PATH с другим поведением.
        sh """#!/bin/bash
set -eux
GO_VER='${env.GO_VERSION ?: '1.24.11'}'
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

NODE_VER='${env.NODE_JS_VERSION ?: '20.18.1'}'
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
"\${SCANNER_HOME}/bin/sonar-scanner" \\
  -Dsonar.host.url="${env.SONAR_HOST_URL}" \\
  -Dsonar.token="\${SONAR_TOKEN}" \\
  -Dsonar.scm.revision="\$(git rev-parse HEAD)" ${params.SONAR_EXTRA_OPTS?.trim() ?: ''}
"""
      }
    }

    // --- Сборка микросервисов: вложенные stage — в UI видно «Docker: prepare», «Docker: auth-service», … ---
    stage('Docker build and push') {
      stages {
        stage('Docker: prepare') {
          steps {
            sh """#!/bin/bash
set -eux
export DOCKER_REGISTRY='${params.DOCKER_REGISTRY}'
export BUILD_NUMBER='${env.BUILD_NUMBER}'
cd "\${WORKSPACE}"
bash scripts/ci/jenkins-docker.sh prepare
"""
          }
        }
        stage('Docker: auth-service') {
          steps {
            sh """#!/bin/bash
set -eux
export DOCKER_REGISTRY='${params.DOCKER_REGISTRY}'
export BUILD_NUMBER='${env.BUILD_NUMBER}'
cd "\${WORKSPACE}"
bash scripts/ci/jenkins-docker.sh build auth-service build/auth-service.Dockerfile
"""
          }
        }
        stage('Docker: customers-service') {
          steps {
            sh """#!/bin/bash
set -eux
export DOCKER_REGISTRY='${params.DOCKER_REGISTRY}'
export BUILD_NUMBER='${env.BUILD_NUMBER}'
cd "\${WORKSPACE}"
bash scripts/ci/jenkins-docker.sh build customers-service build/customers-service.Dockerfile
"""
          }
        }
        stage('Docker: vehicles-service') {
          steps {
            sh """#!/bin/bash
set -eux
export DOCKER_REGISTRY='${params.DOCKER_REGISTRY}'
export BUILD_NUMBER='${env.BUILD_NUMBER}'
cd "\${WORKSPACE}"
bash scripts/ci/jenkins-docker.sh build vehicles-service build/vehicles-service.Dockerfile
"""
          }
        }
        stage('Docker: deals-service') {
          steps {
            sh """#!/bin/bash
set -eux
export DOCKER_REGISTRY='${params.DOCKER_REGISTRY}'
export BUILD_NUMBER='${env.BUILD_NUMBER}'
cd "\${WORKSPACE}"
bash scripts/ci/jenkins-docker.sh build deals-service build/deals-service.Dockerfile
"""
          }
        }
        stage('Docker: parts-service') {
          steps {
            sh """#!/bin/bash
set -eux
export DOCKER_REGISTRY='${params.DOCKER_REGISTRY}'
export BUILD_NUMBER='${env.BUILD_NUMBER}'
cd "\${WORKSPACE}"
bash scripts/ci/jenkins-docker.sh build parts-service build/parts-service.Dockerfile
"""
          }
        }
        stage('Docker: brands-service') {
          steps {
            sh """#!/bin/bash
set -eux
export DOCKER_REGISTRY='${params.DOCKER_REGISTRY}'
export BUILD_NUMBER='${env.BUILD_NUMBER}'
cd "\${WORKSPACE}"
bash scripts/ci/jenkins-docker.sh build brands-service build/brands-service.Dockerfile
"""
          }
        }
        stage('Docker: dealer-points-service') {
          steps {
            sh """#!/bin/bash
set -eux
export DOCKER_REGISTRY='${params.DOCKER_REGISTRY}'
export BUILD_NUMBER='${env.BUILD_NUMBER}'
cd "\${WORKSPACE}"
bash scripts/ci/jenkins-docker.sh build dealer-points-service build/dealer-points-service.Dockerfile
"""
          }
        }
      }
    }

    // --- Деплой в Minikube: k8s/dealer-stack.yaml (все сервисы), только если DEPLOY=true ---
    stage('Deploy to Minikube') {
      when {
        expression { return params.DEPLOY }
      }
      steps {
        sh """#!/bin/bash
set -eux
cd "\${WORKSPACE}"
# Новый shell не видит PATH из jenkins-docker.sh — тот же статический клиент, что кладётся в .ci/docker-cli-bin/
if [ -x "\${WORKSPACE}/.ci/docker-cli-bin/docker" ]; then
  export PATH="\${WORKSPACE}/.ci/docker-cli-bin:\${PATH}"
fi
command -v docker

# --- Контекст: kubeconfig в Jenkins (любой драйвер) или docker exec в контейнер-ноду (только minikube --driver=docker) ---
JHOME="\${JENKINS_HOME:-\$HOME}"
MK='${params.MINIKUBE_CONTAINER}'
KCFG_SRC="\$JHOME/.kube/config"
KP='${params.KUBECONFIG_PATH}'
if [ -n "\$KP" ]; then
  if [ ! -f "\$KP" ]; then echo "KUBECONFIG_PATH задан, но файл не найден: \$KP" >&2; exit 1; fi
  KCFG_SRC="\$KP"
fi
USE_DOCKER_EXEC=0
KUBECTL=""

# Ветка A: есть kubeconfig (JENKINS_HOME/.kube/config или KUBECONFIG_PATH) — kubectl в /tmp, 127.0.0.1 → host.docker.internal для API из контейнера Jenkins.
if [ -f "\$KCFG_SRC" ]; then
  ARCH="\$(uname -m)"
  case "\$ARCH" in aarch64|arm64) KARCH=arm64 ;; x86_64) KARCH=amd64 ;; *) echo "unsupported arch: \$ARCH"; exit 1 ;; esac
  KVER="\$(curl -fsSL https://dl.k8s.io/release/stable.txt)"
  KUBECTL="/tmp/kubectl-\${KVER}"
  if [ ! -x "\$KUBECTL" ]; then
    curl -fSL "https://dl.k8s.io/release/\${KVER}/bin/linux/\${KARCH}/kubectl" -o "\$KUBECTL"
    chmod +x "\$KUBECTL"
  fi
  KCFG="/tmp/kubeconfig-jenkins-\${BUILD_NUMBER}"
  cp "\$KCFG_SRC" "\$KCFG"
  export KUBECONFIG="\$KCFG"
  sed -i.bak 's|127.0.0.1|host.docker.internal|g' "\$KCFG"
  CLUSTER="\$( "\$KUBECTL" config view --minify -o jsonpath='{.clusters[0].name}' 2>/dev/null || echo minikube )"
  if [ -f "\$JHOME/.minikube/ca.crt" ]; then
    SERVER="\$( "\$KUBECTL" config view --minify -o jsonpath='{.clusters[0].cluster.server}' )"
    "\$KUBECTL" config set-cluster "\$CLUSTER" --server="\$SERVER" --certificate-authority="\$JHOME/.minikube/ca.crt" --embed-certs=false --kubeconfig="\$KCFG"
  else
    echo "WARN: нет \$JHOME/.minikube/ca.crt — TLS verify off (лаборатория)."
    "\$KUBECTL" config set-cluster "\$CLUSTER" --insecure-skip-tls-verify=true --kubeconfig="\$KCFG"
  fi
  "\$KUBECTL" cluster-info
else
  # Ветка B: kubeconfig нет — kubectl через docker exec в контейнер-ноду (только драйвер docker у minikube).
  if docker inspect "\$MK" >/dev/null 2>&1; then
    USE_DOCKER_EXEC=1
    echo "kubeconfig в Jenkins нет — kubectl через docker exec \$MK"
  else
    MK_AUTO="\$(docker ps --format '{{.Names}}' 2>/dev/null | grep -E 'minikube' | head -n1 || true)"
    if [ -n "\$MK_AUTO" ] && docker inspect "\$MK_AUTO" >/dev/null 2>&1; then
      echo "Контейнер '\$MK' не найден — использую running-контейнер '\$MK_AUTO' (проверьте MINIKUBE_CONTAINER при следующем запуске)."
      MK="\$MK_AUTO"
      USE_DOCKER_EXEC=1
    else
      echo "=== Деплой: нет Kubernetes-доступа ==="
      echo "Нет kubeconfig (\$JHOME/.kube/config или KUBECONFIG_PATH='\${KP}') и не найден Docker-контейнер-нода minikube (MINIKUBE_CONTAINER='\${MK}')."
      echo ""
      echo "Что сделать:"
      echo "  A) Смонтируйте kubeconfig в контейнер Jenkins: volume хост ~/.kube/config -> /var/jenkins_home/.kube/config, либо другой путь + параметр KUBECONFIG_PATH. Для qemu2/kvm отдельная VM — контейнера minikube на Docker хоста нет."
      echo "  B) Только если minikube --driver=docker: Jenkins должен видеть тот же docker.sock; docker ps показывает контейнер ноды (часто minikube)."
      echo "  C) Для docker driver: MINIKUBE_CONTAINER = имя из «docker ps»."
      echo "  D) Отключите DEPLOY, если деплой не нужен."
      echo "  E) При qemu2: K8S_CTR_IMPORT_IMAGE=false, образы из registry (K8S_PULL_REGISTRY), доступного из VM."
      echo ""
      echo "Запущенные контейнеры (имена):"
      docker ps --format '{{.Names}}' 2>/dev/null | head -30 || echo "(docker ps недоступен)"
      exit 1
    fi
  fi
fi

# --- Подготовка kubectl для ветки B: бинарь в образе minikube или скачанный в /tmp внутри контейнера ---
# KUBECONFIG на ноде: /etc/kubernetes/admin.conf (или запасной путь minikube).
MK_KUBECTL=""
MK_KUBECONFIG="/etc/kubernetes/admin.conf"
if [ "\$USE_DOCKER_EXEC" = 1 ]; then
  MK_KUBECTL="\$(docker exec "\$MK" sh -c 'command -v kubectl 2>/dev/null || find /var/lib/minikube -type f -name kubectl 2>/dev/null | head -n1')"
  if [ -z "\$MK_KUBECTL" ] || ! docker exec "\$MK" test -x "\$MK_KUBECTL" 2>/dev/null; then
    echo "kubectl в образе \$MK не найден — качаем бинарь и копируем в контейнер"
    ARCH="\$(docker exec "\$MK" uname -m)"
    case "\$ARCH" in aarch64|arm64) KARCH=arm64 ;; x86_64) KARCH=amd64 ;; *) echo "unsupported minikube arch: \$ARCH"; exit 1 ;; esac
    KVER="\$(curl -fsSL https://dl.k8s.io/release/stable.txt)"
    TMPK="/tmp/kubectl-mk-\${BUILD_NUMBER}-\${KVER}"
    curl -fSL "https://dl.k8s.io/release/\${KVER}/bin/linux/\${KARCH}/kubectl" -o "\$TMPK"
    chmod +x "\$TMPK"
    docker cp "\$TMPK" "\$MK:/tmp/kubectl-from-jenkins"
    docker exec "\$MK" chmod +x /tmp/kubectl-from-jenkins
    MK_KUBECTL=/tmp/kubectl-from-jenkins
  fi
  echo "kubectl in \$MK: \$MK_KUBECTL"
  if ! docker exec "\$MK" test -f "\$MK_KUBECONFIG" 2>/dev/null; then
    MK_KUBECONFIG=/var/lib/minikube/kubeconfig
  fi
  docker exec -e KUBECONFIG="\$MK_KUBECONFIG" "\$MK" "\$MK_KUBECTL" cluster-info
fi

# Единый вызов kubectl (ветка A: \$KUBECTL; ветка B: docker exec в minikube-ноду).
kctl() {
  if [ "\$USE_DOCKER_EXEC" = 1 ]; then
    # -i обязателен: иначе stdin Jenkins (pipe в kubectl exec -i … psql -f -) не доходит в minikube.
    docker exec -i -e KUBECONFIG="\$MK_KUBECONFIG" "\$MK" "\$MK_KUBECTL" "\$@"
  else
    "\$KUBECTL" "\$@"
  fi
}

# --- Подготовка образов и манифеста k8s/dealer-stack.yaml (теги из services/*/VERSION, см. Docker stage) ---
TAG='${env.BUILD_NUMBER}'
LOCAL_TAG="jenkins-\${TAG}"
if [ ! -f "\${WORKSPACE}/.ci/image-versions.env" ]; then
  echo "Нет \${WORKSPACE}/.ci/image-versions.env — сначала должна пройти стадия Docker build and push." >&2
  exit 1
fi
# shellcheck disable=SC1090
. "\${WORKSPACE}/.ci/image-versions.env"
CTR_IMP='${params.K8S_CTR_IMPORT_IMAGE}'
NS='${params.K8S_NAMESPACE}'
K8S_PULL_REG='${params.K8S_PULL_REGISTRY}'
K8S_DB_HOST='${params.K8S_DB_HOST}'
K8S_DB_PORT='${params.K8S_DB_PORT}'
POSTGRES_PASSWORD='${params.POSTGRES_PASSWORD ?: 'changeme'}'
BOOTSTRAP_DEV='${params.K8S_BOOTSTRAP_DEV_DATA ? "true" : "false"}'
DELETE_NS_BEFORE='${params.K8S_DELETE_NS_BEFORE_DEPLOY ? "true" : "false"}'
INFRA_DPL=(postgres redis zookeeper kafka)
SVC_LIST=(auth-service customers-service vehicles-service deals-service parts-service brands-service dealer-points-service)

test -f k8s/dealer-stack.yaml

PULL_POLICY=Always
if [ "\$USE_DOCKER_EXEC" = 1 ] && [ "\$CTR_IMP" = "true" ]; then
  if ! docker exec "\$MK" sh -c 'command -v docker >/dev/null 2>&1'; then
    echo "В \$MK нет docker — отключите K8S_CTR_IMPORT_IMAGE" >&2
    exit 1
  fi
  for NAME in "\${SVC_LIST[@]}"; do
    LOCAL_REF="\${NAME}:\${LOCAL_TAG}"
    docker image inspect "\$LOCAL_REF" >/dev/null
    echo "docker load \${LOCAL_REF} → minikube…"
    docker save "\$LOCAL_REF" | docker exec -i "\$MK" docker load
  done
  PULL_POLICY=Never
  IMG_AUTH="auth-service:\${LOCAL_TAG}"
  IMG_CUSTOMERS="customers-service:\${LOCAL_TAG}"
  IMG_VEHICLES="vehicles-service:\${LOCAL_TAG}"
  IMG_DEALS="deals-service:\${LOCAL_TAG}"
  IMG_PARTS="parts-service:\${LOCAL_TAG}"
  IMG_BRANDS="brands-service:\${LOCAL_TAG}"
  IMG_DEALER_POINTS="dealer-points-service:\${LOCAL_TAG}"
else
  IMG_AUTH="\${K8S_PULL_REG}/auth-service:\${VER_AUTH_SERVICE}"
  IMG_CUSTOMERS="\${K8S_PULL_REG}/customers-service:\${VER_CUSTOMERS_SERVICE}"
  IMG_VEHICLES="\${K8S_PULL_REG}/vehicles-service:\${VER_VEHICLES_SERVICE}"
  IMG_DEALS="\${K8S_PULL_REG}/deals-service:\${VER_DEALS_SERVICE}"
  IMG_PARTS="\${K8S_PULL_REG}/parts-service:\${VER_PARTS_SERVICE}"
  IMG_BRANDS="\${K8S_PULL_REG}/brands-service:\${VER_BRANDS_SERVICE}"
  IMG_DEALER_POINTS="\${K8S_PULL_REG}/dealer-points-service:\${VER_DEALER_POINTS_SERVICE}"
fi

render_stack() {
  sed \\
    -e "s|__K8S_DB_HOST__|\${K8S_DB_HOST}|g" \\
    -e "s|__K8S_DB_PORT__|\${K8S_DB_PORT}|g" \\
    -e "s|__POSTGRES_PASSWORD__|\${POSTGRES_PASSWORD}|g" \\
    -e "s|__IMG_AUTH__|\${IMG_AUTH}|g" \\
    -e "s|__IMG_CUSTOMERS__|\${IMG_CUSTOMERS}|g" \\
    -e "s|__IMG_VEHICLES__|\${IMG_VEHICLES}|g" \\
    -e "s|__IMG_DEALS__|\${IMG_DEALS}|g" \\
    -e "s|__IMG_PARTS__|\${IMG_PARTS}|g" \\
    -e "s|__IMG_BRANDS__|\${IMG_BRANDS}|g" \\
    -e "s|__IMG_DEALER_POINTS__|\${IMG_DEALER_POINTS}|g" \\
    -e "s|__PULL_POLICY__|\${PULL_POLICY}|g" \\
    k8s/dealer-stack.yaml
}

REG_PULL='${params.K8S_PULL_REGISTRY}'
registry_probe() {
  if [ "\$USE_DOCKER_EXEC" != 1 ] || [ "\$CTR_IMP" = "true" ]; then return 0; fi
  if ! docker exec "\$MK" sh -c 'command -v curl >/dev/null 2>&1'; then
    echo "WARN: в minikube нет curl — проверку http://\${REG_PULL}/v2/ пропускаем"
    return 0
  fi
  if docker exec "\$MK" sh -ec "curl -sf --connect-timeout 5 http://\${REG_PULL}/v2/ >/dev/null"; then
    echo "OK: с ноды minikube отвечает http://\${REG_PULL}/v2/"
    return 0
  fi
  echo "=== ОШИБКА: с ноды minikube НЕ открывается http://\${REG_PULL}/v2/ ===" >&2
  return 1
}

if [ "\$USE_DOCKER_EXEC" = 1 ]; then
  registry_probe
fi

if [ "\$DELETE_NS_BEFORE" = "true" ]; then
  echo "=== K8S_DELETE_NS_BEFORE_DEPLOY: удаляю namespace \$NS ==="
  kctl delete namespace "\$NS" --ignore-not-found --wait=true
fi

if [ "\$USE_DOCKER_EXEC" = 1 ]; then
  render_stack | docker exec -i -e KUBECONFIG="\$MK_KUBECONFIG" "\$MK" "\$MK_KUBECTL" apply -f -
else
  render_stack | "\$KUBECTL" apply -f -
fi

for d in "\${INFRA_DPL[@]}"; do
  kctl -n "\$NS" rollout status "deployment/\$d" --timeout=300s
done

# Миграции + dev-сидеры в Postgres; admin — через /seed-admin в auth (образ с build/auth-service.Dockerfile).
if [ "\$BOOTSTRAP_DEV" = "true" ]; then
  echo "=== K8S_BOOTSTRAP_DEV_DATA: миграции и тестовые данные ==="
  kctl -n "\$NS" wait --for=condition=available "deployment/postgres" --timeout=180s
  HAS_USERS=\$(kctl -n "\$NS" exec "deployment/postgres" -- env PGPASSWORD="\$POSTGRES_PASSWORD" psql -U dealer -d dealer -qtAc "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='users');")
  if [ "\$HAS_USERS" != "t" ]; then
    echo "Применяю миграции (пустая БД)…"
    for f in \\
      001_users.up.sql 002_roles.up.sql 003_customers.up.sql 004_vehicles.up.sql \\
      005_deals.up.sql 006_parts.up.sql 007_part_folders.up.sql 008_brands.up.sql \\
      009_dealer_points.up.sql 010_part_stock.up.sql; do
      test -f "\${WORKSPACE}/migrations/\$f"
      kctl -n "\$NS" exec -i "deployment/postgres" -- env PGPASSWORD="\$POSTGRES_PASSWORD" psql -U dealer -d dealer -v ON_ERROR_STOP=1 -f - < "\${WORKSPACE}/migrations/\$f"
    done
  else
    echo "Таблица users уже есть — миграции пропускаем."
  fi
  for f in seed_test_data.sql seed_dealer_brands.sql seed_parts.sql; do
    test -f "\${WORKSPACE}/migrations/\$f"
    echo "Сид: \$f"
    kctl -n "\$NS" exec -i "deployment/postgres" -- env PGPASSWORD="\$POSTGRES_PASSWORD" psql -U dealer -d dealer -v ON_ERROR_STOP=1 -f - < "\${WORKSPACE}/migrations/\$f"
  done
  echo "Bootstrap SQL готов."
fi

for d in "\${SVC_LIST[@]}"; do
  kctl -n "\$NS" rollout status "deployment/\$d" --timeout=300s
done

if [ "\$BOOTSTRAP_DEV" = "true" ]; then
  echo "=== seed-admin (admin@dealer.local / admin123 по умолчанию) ==="
  SEED_DSN="postgres://dealer:\${POSTGRES_PASSWORD}@\${K8S_DB_HOST}:\${K8S_DB_PORT}/dealer?sslmode=disable"
  kctl -n "\$NS" exec "deployment/auth-service" -- env POSTGRES_DSN="\$SEED_DSN" /seed-admin
fi

echo "=== Как открыть приложение после деплоя ==="
ING_ADDR="\$(kctl -n "\$NS" get ingress dealer-http -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || true)"
if [ -z "\$ING_ADDR" ]; then
  ING_ADDR="\$(kctl -n "\$NS" get ingress dealer-http -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null || true)"
fi
if [ -n "\$ING_ADDR" ]; then
  echo "Ingress address: \$ING_ADDR"
  echo "На вашей машине добавьте в /etc/hosts:"
  echo "  \$ING_ADDR dealer.local"
else
  echo "Ingress address пока пустой (контроллер ещё обновляет status)."
  echo "Проверьте позже: kubectl -n \$NS get ingress dealer-http -o wide"
  echo "Если используете minikube + docker driver, можно взять: minikube ip"
fi
echo "URL: http://dealer.local/"
"""
      }
    }
  }
}
