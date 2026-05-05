#!/usr/bin/env bash
# Вызывается из Jenkins: prepare (файлы версий + skopeo) | build <имя> <dockerfile>.
set -euo pipefail

WS="${WORKSPACE:?}"
cd "$WS"

REG="${DOCKER_REGISTRY:?}"
TAG="${BUILD_NUMBER:?}"
LOCAL_TAG="jenkins-${TAG}"
SKOPEO_IMG='quay.io/skopeo/stable:latest'

# Jenkins-агент иногда даёт урезанный PATH без /usr/bin — docker при этом установлен.
ensure_docker_on_path() {
	if command -v docker >/dev/null 2>&1; then
		return 0
	fi
	local d
	for d in /usr/bin/docker /usr/local/bin/docker; do
		if [ -x "$d" ]; then
			export PATH="$(dirname "$d"):$PATH"
			return 0
		fi
	done
	return 1
}

# Статический клиент с download.docker.com (только docker), кэш в WORKSPACE/.ci/docker-cli-bin/
# Версию можно переопределить: DOCKER_STATIC_CLI_VERSION=27.4.1
bootstrap_docker_cli() {
	if command -v curl >/dev/null 2>&1; then
		:
	else
		echo "ERROR: нет curl — не могу скачать Docker CLI. Установите curl в образе агента." >&2
		return 1
	fi
	local arch darch ver dest_dir bin tmpd tgz url
	arch="$(uname -m)"
	case "$arch" in
	aarch64 | arm64) darch=aarch64 ;;
	x86_64 | amd64) darch=x86_64 ;;
	*)
		echo "ERROR: архитектура $arch не поддерживается для статического Docker CLI." >&2
		return 1
		;;
	esac
	ver="${DOCKER_STATIC_CLI_VERSION:-27.4.1}"
	dest_dir="$WS/.ci/docker-cli-bin"
	bin="$dest_dir/docker"
	mkdir -p "$dest_dir"
	if [ -x "$bin" ] && "$bin" version --format '{{.Client.Version}}' >/dev/null 2>&1; then
		export PATH="$dest_dir:$PATH"
		echo "Using cached Docker CLI at $bin ($("$bin" version --format '{{.Client.Version}}'))"
		return 0
	fi
	rm -f "$bin"
	url="https://download.docker.com/linux/static/stable/${darch}/docker-${ver}.tgz"
	tgz="/tmp/docker-static-${ver}-${darch}-$$.tgz"
	echo "Downloading Docker CLI ${ver} (${darch})…"
	curl -fSL --connect-timeout 30 --max-time 300 "$url" -o "$tgz"
	tmpd="$(mktemp -d)"
	tar -xzf "$tgz" -C "$tmpd"
	mv "$tmpd/docker/docker" "$bin"
	rm -rf "$tmpd" "$tgz"
	chmod +x "$bin"
	export PATH="$dest_dir:$PATH"
	"$bin" version --format 'Docker CLI {{.Client.Version}} (bootstrapped)'
}

docker_ok() {
	if ! ensure_docker_on_path; then
		echo "Docker CLI не найден в образе — ставлю статический клиент…"
		bootstrap_docker_cli || return 1
	fi
	if ! command -v docker >/dev/null 2>&1; then
		echo "ERROR: после bootstrap docker всё ещё не в PATH." >&2
		return 1
	fi
	if ! docker info >/dev/null 2>&1; then
		echo "ERROR: Docker daemon not reachable (docker info failed). For Jenkins in Docker mount the host socket, e.g. -v /var/run/docker.sock:/var/run/docker.sock" >&2
		return 1
	fi
	return 0
}

setup_skopeo() {
	docker_ok || exit 1
	if docker image inspect "$SKOPEO_IMG" >/dev/null 2>&1; then
		echo "Using local image $SKOPEO_IMG (skip pull)"
	else
		local attempt
		for attempt in 1 2 3; do
			echo "Pulling $SKOPEO_IMG (attempt $attempt/3)…"
			if docker pull -q "$SKOPEO_IMG" || docker pull "$SKOPEO_IMG"; then
				break
			fi
			if [ "$attempt" -eq 3 ]; then
				echo "ERROR: could not pull $SKOPEO_IMG after 3 attempts (network or registry)." >&2
				exit 1
			fi
			sleep 5
		done
	fi
	HOST_ARGS=()
	if docker run --help 2>&1 | grep -qF 'host-gateway'; then
		HOST_ARGS=(--add-host=host.docker.internal:host-gateway)
	fi
}

run_skopeo_copy_daemon() {
	local local_ref="$1"
	local dest="$2"
	docker run --rm \
		"${HOST_ARGS[@]}" \
		-v /var/run/docker.sock:/var/run/docker.sock \
		"$SKOPEO_IMG" \
		copy --dest-tls-verify=false \
		"docker-daemon:${local_ref}" \
		"docker://${dest}"
}

# docker pull к plain-HTTP registry даёт «http: server gave HTTP response to HTTPS client»;
# skopeo с --src-tls-verify=false совпадает с inspect/push в этом скрипте.
run_skopeo_copy_registry_to_daemon() {
	local remote="$1"
	local local_ref="$2"
	docker run --rm \
		"${HOST_ARGS[@]}" \
		-v /var/run/docker.sock:/var/run/docker.sock \
		"$SKOPEO_IMG" \
		copy --src-tls-verify=false \
		"docker://${remote}" \
		"docker-daemon:${local_ref}"
}

registry_has_tag() {
	docker run --rm \
		"${HOST_ARGS[@]}" \
		"$SKOPEO_IMG" \
		inspect --tls-verify=false "docker://$1" >/dev/null 2>&1
}

version_file_for_service() {
	case "$1" in
	auth-service) echo services/auth/VERSION ;;
	customers-service) echo services/customers/VERSION ;;
	vehicles-service) echo services/vehicles/VERSION ;;
	deals-service) echo services/deals/VERSION ;;
	parts-service) echo services/parts/VERSION ;;
	brands-service) echo services/brands/VERSION ;;
	dealer-points-service) echo services/dealerpoints/VERSION ;;
	*) echo ""; return 1 ;;
	esac
}

append_version_export() {
	case "$1" in
	auth-service) echo "export VER_AUTH_SERVICE=$2" ;;
	customers-service) echo "export VER_CUSTOMERS_SERVICE=$2" ;;
	vehicles-service) echo "export VER_VEHICLES_SERVICE=$2" ;;
	deals-service) echo "export VER_DEALS_SERVICE=$2" ;;
	parts-service) echo "export VER_PARTS_SERVICE=$2" ;;
	brands-service) echo "export VER_BRANDS_SERVICE=$2" ;;
	dealer-points-service) echo "export VER_DEALER_POINTS_SERVICE=$2" ;;
	*) return 1 ;;
	esac
}

cmd_prepare() {
	export DOCKER_BUILDKIT=0
	export BUILDKIT_PROGRESS=plain
	echo "=== jenkins-docker.sh prepare (WORKSPACE=$WS) ==="
	setup_skopeo
	mkdir -p "$WS/.ci"
	: >"$WS/.ci/image-versions.env"
	local entry name vf ver
	for entry in \
		'auth-service:build/auth-service.Dockerfile' \
		'customers-service:build/customers-service.Dockerfile' \
		'vehicles-service:build/vehicles-service.Dockerfile' \
		'deals-service:build/deals-service.Dockerfile' \
		'parts-service:build/parts-service.Dockerfile' \
		'brands-service:build/brands-service.Dockerfile' \
		'dealer-points-service:build/dealer-points-service.Dockerfile'; do
		name="${entry%%:*}"
		vf="$(version_file_for_service "$name")"
		test -n "$vf" && test -f "$vf" || {
			echo "VERSION not found for $name (expected $vf)"
			exit 1
		}
		ver="$(tr -d '[:space:]' <"$vf")"
		test -n "$ver" || {
			echo "Empty VERSION in $vf"
			exit 1
		}
		append_version_export "$name" "$ver" >>"$WS/.ci/image-versions.env"
	done
}

cmd_build() {
	local name="$1"
	local dockerfile="$2"
	export DOCKER_BUILDKIT=0
	export BUILDKIT_PROGRESS=plain
	echo "=== jenkins-docker.sh build $name ==="
	setup_skopeo
	local vf ver remote
	vf="$(version_file_for_service "$name")"
	ver="$(tr -d '[:space:]' <"$vf")"
	test -f "$dockerfile" || {
		echo "Dockerfile not found: $dockerfile"
		exit 1
	}
	remote="${REG}/${name}:${ver}"
	if registry_has_tag "$remote"; then
		echo "=== SKIP $name: registry already has $remote ==="
		run_skopeo_copy_registry_to_daemon "$remote" "${name}:${LOCAL_TAG}"
		return 0
	fi
	echo "=== docker build $name ($dockerfile) version=$ver ==="
	docker build -f "$dockerfile" --build-arg "SERVICE_VERSION=${ver}" -t "${name}:${LOCAL_TAG}" .
	run_skopeo_copy_daemon "${name}:${LOCAL_TAG}" "$remote"
	run_skopeo_copy_daemon "${name}:${LOCAL_TAG}" "${REG}/${name}:latest"
}

case "${1:-}" in
prepare) cmd_prepare ;;
build)
	test "${2:-}" && test "${3:-}" || {
		echo "usage: $0 build <service-name> <path-to-Dockerfile>" >&2
		exit 1
	}
	cmd_build "$2" "$3"
	;;
*)
	echo "usage: $0 prepare | $0 build <service> <dockerfile>" >&2
	exit 1
	;;
esac
