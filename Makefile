.PHONY: proto docker-up docker-down run-auth seed-admin frontend-dev frontend-build

proto:
	@which protoc >/dev/null || (echo "install: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest; go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest" && exit 1)
	mkdir -p pkg/pb/auth/v1 pkg/pb/customers/v1 pkg/pb/vehicles/v1 pkg/pb/deals/v1 pkg/pb/parts/v1 pkg/pb/brands/v1 pkg/pb/dealerpoints/v1
	protoc -I api/proto --go_out=module=github.com/dealer/dealer:. \
		--go-grpc_out=module=github.com/dealer/dealer:. \
		api/proto/auth/v1/auth.proto api/proto/customers/v1/customers.proto api/proto/vehicles/v1/vehicles.proto api/proto/deals/v1/deals.proto api/proto/parts/v1/parts.proto api/proto/brands/v1/brands.proto api/proto/dealerpoints/v1/dealerpoints.proto

docker-up:
	docker compose up -d

docker-down:
	docker compose down

# Применить миграции к БД (нужен запущенный Postgres, порт 5433 при Docker)
migrate:
	@DSN="$${POSTGRES_DSN:-postgres://dealer:dealer_secret@127.0.0.1:5433/dealer?sslmode=disable}"; \
	for f in migrations/001_users.up.sql migrations/002_roles.up.sql migrations/003_customers.up.sql migrations/004_vehicles.up.sql migrations/005_deals.up.sql migrations/006_parts.up.sql migrations/007_part_folders.up.sql migrations/008_brands.up.sql migrations/009_dealer_points.up.sql migrations/010_part_stock.up.sql; do \
		echo "Applying $$f..."; psql "$$DSN" -f "$$f" || exit 1; \
	done
	@echo "Migrations done."

run-auth:
	go run ./services/auth

run-customers:
	go run ./services/customers

run-vehicles:
	go run ./services/vehicles

run-deals:
	go run ./services/deals

run-parts:
	go run ./services/parts

run-brands:
	go run ./services/brands

run-dealer-points:
	go run ./services/dealerpoints

# Тестовые клиенты, автомобили, запчасти (нужны миграции 001–006 и POSTGRES_DSN)
seed-data:
	psql "$${POSTGRES_DSN:-postgres://dealer:dealer_secret@127.0.0.1:5433/dealer?sslmode=disable}" -f migrations/seed_test_data.sql

# Дилерские точки, юр. лица, склады, бренды, папки запчастей и привязка авто/запчастей (нужны миграции 008–009)
seed-dealer-brands:
	psql "$${POSTGRES_DSN:-postgres://dealer:dealer_secret@127.0.0.1:5433/dealer?sslmode=disable}" -f migrations/seed_dealer_brands.sql

# Тестовые запчасти (15 шт) + папки + привязка к складам. Сначала выполните seed-dealer-brands.
seed-parts:
	psql "$${POSTGRES_DSN:-postgres://dealer:dealer_secret@127.0.0.1:5433/dealer?sslmode=disable}" -f migrations/seed_parts.sql

# Все тестовые данные: клиенты/авто/запчасти + дилерские точки/юрлица/склады/бренды/папки
full-seed: seed-data seed-dealer-brands seed-parts

# Сборка образов и запуск всех сервисов (Postgres, auth, vehicles, parts, …). Перед первым запуском: make migrate && make full-seed
deploy:
	docker compose up -d --build

# Создаёт пользователя admin (email и пароль из ADMIN_EMAIL, ADMIN_PASSWORD; по умолчанию admin@dealer.local / admin123)
seed-admin:
	cd services/auth && go run ./cmd/seed-admin

frontend-dev:
	cd frontend/auth && npm install && npm run dev

frontend-build:
	cd frontend/auth && npm install && npm run build
