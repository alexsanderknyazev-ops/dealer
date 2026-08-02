# НТ-стенд на домашней Ubuntu-машине (локальный Wi-Fi)

Постоянный тест-стенд на отдельном компьютере с Ubuntu в твоей домашней сети.
Доступ — только из локального Wi-Fi (не из интернета).

Почему так: стенд всегда включён, не тратит деньги/квоты облака, а данные и образы
живут локально. Для дев-стенда по-прежнему GitHub Codespaces (см. `docs/deploy/codespaces.md`).

Что используется из репозитория:

- `scripts/kube-up.sh` — поднимает **minikube**, собирает образы, применяет
  `k8s/dealer-stack.yaml` + `k8s/client-frontend.yaml`, создаёт секреты, прогоняет
  миграции и seed, затем запускает `expose-lan.sh`;
- `scripts/expose-lan.sh` — `kubectl port-forward` на `0.0.0.0`: 9080/8090/8091/8093/3001;
- `scripts/diagnose-lan.sh` — диагностика сети, если что-то не открывается.

## 1. Подготовка машины (Ubuntu 22.04/24.04)

```bash
sudo apt-get update && sudo apt-get install -y curl git
```

Установи Docker:

```bash
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER && newgrp docker
```

Установи kubectl (latest):

```bash
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl && rm kubectl
```

Установи minikube:

```bash
curl -LO https://github.com/kubernetes/minikube/releases/latest/download/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube && rm minikube-linux-amd64
```

## 2. Развернуть стенд

```bash
git clone https://github.com/alexsanderknyazev-ops/dealer.git
cd dealer
./scripts/kube-up.sh
```

Первый запуск — 15–30 мин (сборка 21 образа + старт minikube на 4 CPU / 16 ГБ RAM).
Дальше — минуты. В конце скрипт напечатает адреса и логины.

## 3. Доступ с других устройств (Mac/телефон в том же Wi-Fi)

| Сервис | URL |
|---|---|
| Employee UI | http://`<IP>`:9080 |
| Employee API gateway | http://`<IP>`:8090 |
| Client UI | http://`<IP>`:3001/login |
| Client public gateway | http://`<IP>`:8091 |
| Client protected GW | http://`<IP>`:8093 |

`<IP>` — адрес Ubuntu-машины в Wi-Fi (скрипт печатает свой LAN IP).

Логины (seed-данные):

| Зона | Логин | Пароль |
|---|---|---|
| Employee | `admin@dealer.local` | `admin123` |
| Employee | `vol.employee1@test.dealer.local` | `Test1234!` |
| Client | `vol.client1@test.dealer.local` | `Test1234!` |

> Стенд по http — норма для домашней сети. В интернет он не публикуется.

## 4. Чтобы IP не менялся

После переподключения к Wi-Fi IP может поменяться (и ссылки устареют).

1. На роутере зарезервируй адрес за MAC-адресом Wi-Fi/eth интерфейса Ubuntu-машины
   (DHCP reservation). Либо настрой статический IP в Netplan.
2. После смены IP просто перезапусти проброс портов:

   ```bash
   cd ~/dealer && ./scripts/expose-lan.sh
   ```

## 5. Обновление стенда

```bash
cd ~/dealer
git pull --ff-only
./scripts/kube-up.sh --skip-expose   # пересоберёт изменившееся и переприменит
./scripts/expose-lan.sh
```

`kube-up.sh` идемпотентен: повторный запуск не ломает данные (volumes в minikube).

## 6. Диагностика

Если с другого устройства не открывается:

```bash
./scripts/diagnose-lan.sh
```

Типичные причины (по убыванию частоты):

1. Устройства в **разных сетях** (гостевая/основная Wi-Fi) — включи одно.
2. **AP isolation** на роутере — отключи (устройства не видят друг друга).
3. Файрвол на Ubuntu:
   ```bash
   sudo ufw allow 9080,8090,8091,8093,3001/tcp
   ```
4. Пробросы упали после перезагрузки — снова `./scripts/expose-lan.sh`.

## 7. Чистка Docker (диск на Ubuntu-машине)

Образы стенда собираются **внутри minikube** (`minikube docker-env`), поэтому и чистить надо там:

```bash
./scripts/docker-cleanup.sh --minikube   # build cache + dangling, образы стенда сохраняются
```

Для полного сброса (если minikube разросся):

```bash
minikube stop && minikube delete   # снесёт весь кластер и образы (данные volumes тоже)
./scripts/kube-up.sh               # заново
```

## 8. Ограничения подхода

- Доступ только из локальной сети. Удалённый доступ — только через проброс порта на роутере
  (+ белый IP/dyndns), это уже за рамками стенда.
- CI (`deploy.yml`) сюда деплоить не сможет: машина за домашним NAT и недоступна из GitHub
  Actions. Обновления — вручную через `kube-up.sh`.
- Требования к машине: ~4+ CPU, 16 ГБ+ RAM, 20+ ГБ диска.
