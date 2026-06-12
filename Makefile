.PHONY: proto docker-up docker-down run-auth seed-admin frontend-dev frontend-build frontend-client-dev frontend-client-build

proto:
	@which protoc >/dev/null || (echo "install protoc (brew install protobuf)" && exit 1)
	@which protoc-gen-go >/dev/null || (echo "go install google.golang.org/protobuf/cmd/protoc-gen-go@latest" && exit 1)
	@which protoc-gen-go-grpc >/dev/null || (echo "go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest" && exit 1)
	@which protoc-gen-grpc-gateway >/dev/null || (echo "go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest" && exit 1)
	mkdir -p pkg/pb/auth/v1 pkg/pb/customers/v1 pkg/pb/vehicles/v1 pkg/pb/deals/v1 pkg/pb/parts/v1 pkg/pb/brands/v1 pkg/pb/dealerpoints/v1 pkg/pb/clients/v1 pkg/pb/clientauth/v1 pkg/pb/reviews/v1 pkg/pb/statistics/employee/v1 pkg/pb/statistics/client/v1 pkg/pb/workorders/v1 pkg/pb/works/v1 pkg/pb/employees/v1
	protoc -I api/proto --go_out=module=github.com/dealer/dealer:. \
		--go-grpc_out=module=github.com/dealer/dealer:. \
		--grpc-gateway_out=module=github.com/dealer/dealer:. \
		api/proto/auth/v1/auth.proto api/proto/customers/v1/customers.proto api/proto/vehicles/v1/vehicles.proto api/proto/deals/v1/deals.proto api/proto/parts/v1/parts.proto api/proto/brands/v1/brands.proto api/proto/dealerpoints/v1/dealerpoints.proto \
		api/proto/clients/v1/common.proto api/proto/clients/v1/registration_public.proto api/proto/clients/v1/account.proto \
		api/proto/clientauth/v1/clientauth.proto api/proto/clientauth/v1/clientauth_public.proto api/proto/clientauth/v1/clientauth_session.proto \
		api/proto/reviews/v1/reviews.proto api/proto/reviews/v1/employee_reviews.proto \
		api/proto/statistics/employee/v1/employee_stats.proto api/proto/statistics/client/v1/client_stats.proto \
		api/proto/workorders/v1/work_orders.proto api/proto/works/v1/works.proto api/proto/employees/v1/employees.proto

docker-up:
	docker compose up -d

docker-down:
	docker compose down

# Применить миграции к БД (нужен запущенный Postgres, порт 5433 при Docker)
migrate:
	@: $${POSTGRES_DSN:?Set POSTGRES_DSN (see .env.example; copy .env from .env.example)}
	@for f in migrations/000_schemas.up.sql migrations/001_users.up.sql migrations/002_roles.up.sql migrations/003_customers.up.sql migrations/004_vehicles.up.sql migrations/005_deals.up.sql migrations/006_parts.up.sql migrations/007_part_folders.up.sql migrations/008_brands.up.sql migrations/009_dealer_points.up.sql migrations/010_part_stock.up.sql migrations/011_clients.up.sql migrations/012_client_role.up.sql migrations/013_clientauth.up.sql migrations/014_reviews.up.sql migrations/015_employee_statistics.up.sql migrations/016_client_statistics.up.sql migrations/017_employee_reviews.up.sql migrations/018_work_orders.up.sql migrations/019_stock_movements.up.sql migrations/020_movement_documents.up.sql migrations/021_work_order_movement_doc.up.sql migrations/022_works.up.sql migrations/023_work_order_labor_work_id.up.sql migrations/024_employees.up.sql migrations/025_review_invitations.up.sql migrations/026_movement_document_statuses.up.sql migrations/027_work_folders.up.sql migrations/028_brand_labor_rates.up.sql migrations/029_production_extraction.up.sql migrations/030_movement_destination_warehouse.up.sql; do \
		echo "Applying $$f..."; psql "$$POSTGRES_DSN" -f "$$f" || exit 1; \
	done
	@echo "Migrations done."

run-auth:
	go run ./services/employee/auth

run-customers:
	go run ./services/employee/customers

run-vehicles:
	go run ./services/employee/vehicles

run-deals:
	go run ./services/employee/deals

run-parts:
	go run ./services/employee/parts

run-brands:
	go run ./services/employee/brands

run-works:
	go run ./services/employee/works

run-employees:
	go run ./services/employee/employees

run-dealer-points:
	go run ./services/employee/dealerpoints

run-gateway:
	go run ./services/gateway/employee

run-client-registration:
	go run ./services/client/registration

run-client-public-gateway:
	go run ./services/gateway/client-public

run-client-protected-gateway:
	go run ./services/gateway/client-protected

run-client-reviews:
	go run ./services/client/reviews

run-client-auth:
	go run ./services/client/auth

run-employee-statistics:
	go run ./services/statistics/employee

run-client-statistics:
	go run ./services/statistics/client

run-employee-reviews:
	go run ./services/employee/reviews

run-workorders:
	go run ./services/employee/workorders

run-scheduler:
	go run ./services/scheduler

# Тестовые клиенты, автомобили, запчасти (нужны миграции 001–006 и POSTGRES_DSN)
seed-data:
	@: $${POSTGRES_DSN:?Set POSTGRES_DSN}
	psql "$$POSTGRES_DSN" -f migrations/seed_test_data.sql

# Дилерские точки, юр. лица, склады, бренды, папки запчастей и привязка авто/запчастей (нужны миграции 008–009)
seed-dealer-brands:
	@: $${POSTGRES_DSN:?Set POSTGRES_DSN}
	psql "$$POSTGRES_DSN" -f migrations/seed_dealer_brands.sql

# Тестовые запчасти (15 шт) + папки + привязка к складам. Сначала выполните seed-dealer-brands.
seed-parts:
	@: $${POSTGRES_DSN:?Set POSTGRES_DSN}
	psql "$$POSTGRES_DSN" -f migrations/seed_parts.sql

# Папки и работы СТО (нужна миграция 022 и 027)
seed-works:
	@: $${POSTGRES_DSN:?Set POSTGRES_DSN}
	psql "$$POSTGRES_DSN" -f migrations/seed_works.sql

# Все тестовые данные: клиенты/авто/запчасти + дилерские точки/юрлица/склады/бренды/папки + работы
full-seed: seed-data seed-dealer-brands seed-parts seed-works

# Сборка образов и запуск всех сервисов (Postgres, auth, vehicles, parts, …). Перед первым запуском: make migrate && make full-seed
deploy:
	docker compose up -d --build

# Создаёт пользователя admin (email и пароль из ADMIN_EMAIL, ADMIN_PASSWORD; по умолчанию admin@dealer.local / admin123)
seed-admin:
	cd services/employee/auth && go run ./cmd/seed-admin

frontend-dev:
	cd frontend/auth && npm install && npm run dev

frontend-build:
	cd frontend/auth && npm install && npm run build

frontend-client-dev:
	cd frontend/client && npm install && npm run dev

frontend-client-build:
	cd frontend/client && npm install && npm run build
