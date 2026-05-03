# Лабораторный Jenkins + Registry

Стек для CI из `Jenkinsfile`: сборка образов, Skopeo push в registry на **5050**, деплой в Minikube по **kubeconfig** (в т.ч. **qemu2**).

## Предусловия

1. На хосте: `minikube start` (или другой кластер), рабочий `kubectl get nodes`.
2. Файл kubeconfig и каталог `.minikube` на хосте (для `ca.crt` в пайплайне).

## Запуск

```bash
cd infra/lab
cp env.example .env
# В .env — абсолютные пути к ~/.kube/config и ~/.minikube
docker compose up -d --build
```

- Jenkins UI: **http://localhost:9080**
- Registry: **host.docker.internal:5050** (из контейнеров) / **localhost:5050** (с хоста)

Пароль администратора Jenkins:

```bash
docker exec lab-jenkins cat /var/jenkins_home/secrets/initialAdminPassword
```

## Job-параметры (Minikube qemu2)

| Параметр | Значение |
|----------|----------|
| `DOCKER_REGISTRY` | `host.docker.internal:5050` |
| `DEPLOY_MINIKUBE` | `true` |
| `K8S_CTR_IMPORT_IMAGE` | `false` (образы из registry в VM) |
| `K8S_PULL_REGISTRY` | `host.minikube.internal:5050` (если registry на хосте; иначе IP, доступный из minikube VM) |
| `KUBECONFIG_PATH` | пусто, если смонтировано в `/var/jenkins_home/.kube/config` |

Если API в kubeconfig не `127.0.0.1`, пайплайн не подменяет адрес — сеть до API из контейнера Jenkins должна быть доступна (часто так и есть для IP minikube).

## Конфликт имён

Если у вас уже есть контейнеры `lab-jenkins` / `lab-registry`, остановите их или измените `container_name` и порты в `docker-compose.yml`.

## SonarQube / Gitea

В этом compose только Jenkins и registry; Sonar и Gitea подключайте отдельно (как у вас в lab), адреса уже заданы в `Jenkinsfile` / `sonar-project.properties`.
