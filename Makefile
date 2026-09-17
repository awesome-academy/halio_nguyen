.PHONY: up down db-shell migrate-up migrate-down migrate-seed run-api dev-web build-web test-api lint-web docker-api docker-web test-e2e e2e-report e2e-kill-web

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

# Run the Playwright admin E2E suite (needs `make up` + `make run-api` first)
test-e2e:
	cd apps/web && pnpm test:e2e

# Open the last HTML report
e2e-report:
	cd apps/web && pnpm test:e2e:report

# Kill whatever process is bound to :3000 (a stale `next dev` most often).
# Resolves the PID from the bound port rather than matching "next" in the
# command line, so it never reaches into an unrelated node.exe process
# (e.g. Nextcloud, or another project's dev server) that happens to be
# running at the same time.
e2e-kill-web:
	powershell -Command "Get-NetTCPConnection -LocalPort 3000 -State Listen -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess -Unique | ForEach-Object { Stop-Process -Id $$_ -Force -ErrorAction SilentlyContinue }"
