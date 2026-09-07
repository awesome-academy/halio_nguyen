.PHONY: up down db-shell migrate-up migrate-down migrate-seed run-api dev-web build-web test-api lint-web docker-api docker-web

# Start Postgres and pgAdmin containers
up:
	docker compose -f docker/docker-compose.yml up -d

# Stop docker containers
down:
	docker compose -f docker/docker-compose.yml down

# Connect to Postgres database CLI
db-shell:
	docker exec -it sun_booking_postgres psql -U sun_booking -d sun_booking_tours

# Apply initial schema migration
migrate-up:
	powershell -Command "Get-Content apps/api/migrations/000001_init_schema.up.sql -Raw | docker exec -i sun_booking_postgres psql -U sun_booking -d sun_booking_tours"

# Rollback schema migration
migrate-down:
	powershell -Command "Get-Content apps/api/migrations/000001_init_schema.down.sql -Raw | docker exec -i sun_booking_postgres psql -U sun_booking -d sun_booking_tours"

# Seed initial data (categories, tours, schedules, admin user)
migrate-seed:
	powershell -Command "Get-Content apps/api/migrations/000002_seed_data.up.sql -Raw | docker exec -i sun_booking_postgres psql -U sun_booking -d sun_booking_tours"

# Run Go backend API server locally
run-api:
	cd apps/api && go run ./cmd/server

# Run backend unit tests
test-api:
	cd apps/api && go test -v ./...

# Run Next.js frontend dev server
dev-web:
	cd apps/web && pnpm dev

# Check Next.js frontend lint
lint-web:
	cd apps/web && .\node_modules\.bin\next.cmd lint

# Build Next.js frontend production bundle
build-web:
	cd apps/web && .\node_modules\.bin\next.cmd build

# Build backend Docker container image locally
docker-api:
	docker build -t sun-booking-api:latest -f apps/api/Dockerfile apps/api

# Build frontend Docker container image locally
docker-web:
	docker build -t sun-booking-web:latest -f apps/web/Dockerfile apps/web
