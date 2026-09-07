# 🌞 SUN Booking Tours

An enterprise-grade online travel and tour booking platform built with a **Monorepo** architecture, featuring a high-throughput **Golang (Echo Framework)** backend, an SEO-optimized **Next.js 14 App Router (TypeScript + Tailwind CSS + shadcn/ui)** frontend, **PostgreSQL 16** running via Docker, and an automated **GitHub Actions CI/CD pipeline**.

---

## 📑 Table of Contents
- [1. Monorepo Architecture](#1-monorepo-architecture)
- [2. Tech Stack](#2-tech-stack)
- [3. Database Design (PostgreSQL 16)](#3-database-design-postgresql-16)
  - [3.1. Entity Relationship Diagram (ERD)](#31-entity-relationship-diagram-erd)
  - [3.2. Detailed Overview of 15 Tables](#32-detailed-overview-of-15-tables)
  - [3.3. Business Rules & Performance Optimizations](#33-business-rules--performance-optimizations)
- [4. Frontend Architecture (Next.js App Router)](#4-frontend-architecture-nextjs-app-router)
  - [4.1. Customer Subsystem (Guest & User) — SEO Optimized](#41-customer-subsystem-guest--user--seo-optimized)
  - [4.2. Admin Subsystem (Back-office Management)](#42-admin-subsystem-back-office-management)
  - [4.3. RBAC Route Guard Middleware](#43-rbac-route-guard-middleware)
- [5. CI/CD Flow & Branching Strategy (GitHub Actions)](#5-cicd-flow--branching-strategy-github-actions)
  - [5.1. Git Branching Model (`dev` & `main`)](#51-git-branching-model-dev--main)
  - [5.2. Automated Quality Gate on `dev` (Dev CI)](#52-automated-quality-gate-on-dev-dev-ci)
  - [5.3. Production Release & Containerization on `main` (Production CD)](#53-production-release--containerization-on-main-production-cd)
- [6. Environment Configuration (.env.example)](#6-environment-configuration-envexample)
- [7. Quick Start Guide](#7-quick-start-guide)
- [8. Makefile Cheatsheet](#8-makefile-cheatsheet)

---

## 1. Monorepo Architecture

```
sun-booking-tours/
├── .github/
│   └── workflows/
│       ├── dev-ci.yml                 # Automated Quality Gate on push/PR to dev
│       └── production-cd.yml          # Production build & GHCR container release on merge to main
├── apps/
│   ├── api/                           # Backend API (Golang + Echo Framework)
│   │   ├── cmd/server/main.go         # HTTP Server Entry Point & Graceful Shutdown
│   │   ├── internal/
│   │   │   ├── config/                # Environment configuration & unit tests
│   │   │   │   ├── config.go
│   │   │   │   └── config_test.go
│   │   │   ├── db/postgres.go         # PostgreSQL connection pool (pgx/v5)
│   │   │   └── domain/                # Entity structs & unit tests
│   │   │       ├── user.go
│   │   │       ├── user_test.go
│   │   │       ├── tour.go
│   │   │       ├── booking.go
│   │   │       ├── payment.go
│   │   │       ├── review.go
│   │   │       ├── activity.go
│   │   │       └── revenue.go
│   │   ├── migrations/                # Database Migrations (SQL)
│   │   │   ├── 000001_init_schema.up.sql
│   │   │   ├── 000001_init_schema.down.sql
│   │   │   ├── 000002_seed_data.up.sql
│   │   │   └── 000002_seed_data.down.sql
│   │   ├── Dockerfile                 # Multi-stage Go 1.25+ Alpine production image
│   │   ├── .dockerignore
│   │   ├── .env.example
│   │   ├── go.mod
│   │   └── go.sum
│   │
│   └── web/                           # Frontend (Next.js App Router, TypeScript, Tailwind)
│       ├── src/
│       │   ├── app/
│       │   │   ├── (customer)/        # === CUSTOMER SUBSYSTEM ===
│       │   │   │   ├── layout.tsx     # Public Layout (Navbar, Footer)
│       │   │   │   ├── page.tsx       # Landing Page (Hero, Categories, Featured Tours)
│       │   │   │   ├── tours/         # Catalog & Discovery
│       │   │   │   │   └── [slug]/    # Dynamic Tour Detail (JSON-LD, generateMetadata)
│       │   │   │   ├── reviews/       # Community Reviews (Place, Food, News)
│       │   │   │   ├── robots.ts      # SEO robots.txt generator
│       │   │   │   └── sitemap.ts     # SEO dynamic XML sitemap
│       │   │   │
│       │   │   ├── (admin)/           # === ADMIN BACK-OFFICE SUBSYSTEM ===
│       │   │   │   └── admin/
│       │   │   │       ├── layout.tsx # Admin Shell (Sidebar, Topbar)
│       │   │   │       ├── dashboard/ # Executive Dashboard (KPIs, Recent Bookings)
│       │   │   │       └── revenue/   # Batch Revenue Visualizer (Materialized Views)
│       │   │   ├── globals.css        # Sun Orange theme variables
│       │   │   └── layout.tsx         # Root HTML/Body layout with base metadata
│       │   ├── components/            # UI primitives, Customer & Admin components
│       │   ├── lib/                   # Utility helpers, formatters (VND currency, dates)
│       │   ├── middleware.ts          # RBAC Route Guard protecting /admin/* and /user/*
│       │   └── types/                 # TypeScript interfaces mirroring Go backend models
│       ├── Dockerfile                 # Multi-stage Next.js Standalone production image
│       ├── .dockerignore
│       ├── .eslintrc.json             # ESLint configuration
│       ├── package.json
│       ├── tsconfig.json
│       ├── tailwind.config.ts
│       └── next.config.mjs
│
├── docker/
│   ├── docker-compose.yml             # PostgreSQL 16 + pgAdmin 4 services
│   └── postgres/init.sql              # Database initialization (uuid, pg_trgm, unaccent)
├── .env.example                       # Unified root environment variables template
├── .gitignore                         # Excludes binaries, .env, and local agent tooling
├── Makefile                           # Unified developer commands
└── README.md
```

---

## 2. Tech Stack

| Component | Technology | Rationale & Suitability |
|---|---|---|
| **Backend** | **Golang 1.25+** & **Echo v4** | Sub-millisecond routing, minimal memory footprint, clean domain design |
| **Database** | **PostgreSQL 16** (Docker) | Relational DB with JSONB, Trigram search, Unaccent dictionaries, and Materialized Views |
| **DB Driver** | **`pgx/v5` (pgxpool)** | Fastest native PostgreSQL driver in Go with connection pooling and type safety |
| **Frontend** | **Next.js 14 App Router + TypeScript** | React ecosystem with Server-Side Rendering (SSR) for SEO-optimized tour & review pages |
| **UI Kit** | **Tailwind CSS + shadcn/ui** | Modern, accessible, customizable component library with Sun Orange theme |
| **Authentication** | **JWT + OAuth 2.0** | Supports Google, Facebook, Twitter (X) and standard email/password registration |
| **Payment** | **Internet Banking (VNPay Sandbox / Bank Transfer)** | Secure online transactions with AES-256 encrypted bank account storage |
| **CI/CD** | **GitHub Actions** | Automated testing on `dev` and Docker container publishing to GHCR on `main` |
| **DB Management** | **pgAdmin 4** | Web-based database management GUI exposed on port `5050` |

---

## 3. Database Design (PostgreSQL 16)

### 3.1. Entity Relationship Diagram (ERD)

```mermaid
erDiagram
    users ||--o{ user_oauth_accounts : "links social account"
    users ||--o{ user_bank_accounts : "manages bank account"
    users ||--o{ bookings : "creates booking"
    users ||--o{ reviews : "writes review"
    users ||--o{ review_likes : "likes review"
    users ||--o{ comments : "posts comment"
    users ||--o{ tour_ratings : "rates completed tour"
    users ||--o{ activity_logs : "records activity"

    categories ||--o{ tours : "categorizes (1:N)"
    tours ||--o{ tour_images : "contains gallery images"
    tours ||--o{ tour_schedules : "has departure dates"
    tours ||--o{ bookings : "receives bookings"
    tours ||--o{ tour_ratings : "receives ratings"

    tour_schedules ||--o{ bookings : "selected departure"

    bookings ||--o{ payments : "paid via"
    bookings ||--o{ tour_ratings : "verifies tour completion"

    review_categories ||--o{ reviews : "classifies review topic"
    reviews ||--o{ comments : "has discussion comments"
    reviews ||--o{ review_likes : "receives likes"

    comments ||--o{ comments : "nested replies (parent_id)"
```

---

### 3.2. Detailed Overview of 15 Tables

1. **`users`**: User and Administrator accounts with email/password or OAuth login, soft delete (`deleted_at`), and status flag (`is_active`).
2. **`user_oauth_accounts`**: Social authentication providers (Google, Facebook, Twitter) with encrypted provider tokens and expiration tracking.
3. **`user_bank_accounts`**: User bank accounts for Internet Banking transactions, including bank code, default selection flag, and soft delete.
4. **`categories`**: Tour categories (e.g., Island & Coastal, Cultural Heritage, Mountain Treks). 1 category contains multiple tours (1:N).
5. **`tours`**: Travel tour packages with destinations, itineraries, duration (days/nights), pricing, max participant capacity, cached `avg_rating`, cached `total_ratings`, and soft delete.
6. **`tour_images`**: Photo galleries for tours ordered by `sort_order`.
7. **`tour_schedules`**: Specific departure dates, return dates, slot capacity, price overrides, and availability status.
8. **`bookings`**: Tour reservation records with unique `booking_code`, contact information, participant count, total price, and lifecycle status (`pending`, `confirmed`, `completed`, `cancelled`).
9. **`payments`**: Internet Banking payment transactions (`amount`, `payment_method`, `transaction_ref`, `bank_name`, gateway JSON response, timestamps).
10. **`review_categories`**: Independent review categories per requirement: **Places & Attractions**, **Food & Dining**, **Travel News**.
11. **`reviews`**: Standalone community articles on places, food, or news (not bound to specific tours), with cached `like_count`, cached `comment_count`, and soft delete.
12. **`review_likes`**: Article likes with composite primary key `(user_id, review_id)` ensuring one like per user.
13. **`comments`**: Community discussion comments supporting unlimited hierarchical nested replies via self-referencing `parent_id`.
14. **`tour_ratings`**: 1-to-5 star tour ratings. **Mandatory business rule**: Strictly restricted to users with a valid booking (`booking_id NOT NULL REFERENCES bookings(id)`), ensuring only verified travelers can rate.
15. **`activity_logs`**: System audit trail tracking core actions (`booking_tour`, `cancel_tour`, `create_review`, `pay_tour`, `rate_tour`, IP address, User-Agent, JSON metadata).

---

### 3.3. Business Rules & Performance Optimizations

#### A. Soft Delete & Partial Indexing
All primary operational tables feature a `deleted_at TIMESTAMPTZ NULL` column:
- Record deletion is performed logically: `UPDATE <table> SET deleted_at = NOW() WHERE id = $1`
- All lookup and filtering indexes are defined as **Partial Indexes** (`WHERE deleted_at IS NULL`), minimizing index footprint and speeding up active record lookups:
  ```sql
  CREATE INDEX idx_tours_status ON tours(status) WHERE deleted_at IS NULL;
  CREATE INDEX idx_bookings_user_id ON bookings(user_id) WHERE deleted_at IS NULL;
  ```

#### B. Accent-Insensitive Full-Text & Trigram Search
An immutable `f_unaccent` function is paired with PostgreSQL GIN Trigram indexes to enable lightning-fast fuzzy and accent-insensitive searches (e.g., searching "da nang" matches "Đà Nẵng"):
```sql
CREATE INDEX idx_tours_trgm_title ON tours USING gin (title gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX idx_tours_fts ON tours USING gin (
    to_tsvector('simple', f_unaccent(title || ' ' || destination || ' ' || COALESCE(description, '')))
) WHERE deleted_at IS NULL;
```

#### C. Real-Time Counter Synchronization via Database Triggers
- **Tour Rating Aggregation**: The `trg_tour_ratings_stat` trigger re-computes `avg_rating` and `total_ratings` on the `tours` table upon any rating change.
- **Review Likes**: The `trg_review_likes_stat` trigger updates `like_count` on `reviews`.
- **Discussion Comments**: The `trg_comments_stat` trigger keeps `comment_count` in sync on `reviews`.
- **Automatic Timestamps**: The `trigger_set_timestamp()` trigger updates `updated_at = NOW()` on every row modification.

#### D. Batch Revenue Reporting with Materialized Views
To allow administrators to query comprehensive revenue analytics without straining online transaction processing (OLTP):
- **`mv_daily_revenue_report`**: Aggregates completed booking counts, travelers, and net revenue daily per tour.
- **`mv_monthly_revenue_report`**: Aggregates revenue on a monthly cadence per category.
- **Periodic Batch Refresh**:
  ```sql
  CALL refresh_revenue_reports();
  ```

---

## 4. Frontend Architecture (Next.js App Router)

The frontend is organized into two distinct subsystems using Next.js **Route Groups**:

### 4.1. Customer Subsystem (Guest & User) — SEO Optimized
Located at `apps/web/src/app/(customer)`:
- **Landing Page (`/`)**: High-converting hero search widget, categories grid, featured tour cards, verified traveler testimonials, and trust badges.
- **Tour Detail Page (`/tours/[slug]`)**:
  - **Dynamic Metadata (`generateMetadata`)**: Automatically generates custom title, meta description, canonical URL, and OpenGraph/Twitter card images.
  - **Structured Data (JSON-LD)**: Injects Schema.org `Product`, `TouristTrip`, and `AggregateRating` schemas for Google Rich Snippets.
  - **Interactive Features**: Itinerary breakdown, inclusion/exclusion checklist, verified review rating badges, and upcoming departure schedule selection.
- **Community Reviews Feed (`/reviews`)**: Independent articles covering scenic places, local gastronomy, and travel news with like counters and nested comments.
- **Dynamic SEO Assets**:
  - `sitemap.ts`: Generates dynamic `/sitemap.xml` listing all tour and review URLs with priority scores.
  - `robots.ts`: Generates `/robots.txt` allowing public routes while keeping `/admin/` and `/user/` private.

### 4.2. Admin Subsystem (Back-office Management)
Located at `apps/web/src/app/(admin)/admin`:
- **Admin Shell Layout (`admin/layout.tsx`)**: Collapsible navigation sidebar (Dashboard, Tours, Categories, Bookings, Reviews, Users, Revenue), notifications indicator, and direct preview link to customer website.
- **Executive Dashboard (`admin/dashboard/page.tsx`)**: Key performance metrics (Net Revenue MTD, Total Bookings, Active Tours, Review Ratings) and recent booking requests queue.
- **Revenue Analytics Visualizer (`admin/revenue/page.tsx`)**: Daily and monthly breakdowns powered by PostgreSQL Materialized Views, with CSV export and manual **Trigger Batch Refresh** button.

### 4.3. RBAC Route Guard Middleware
`apps/web/src/middleware.ts` enforces Role-Based Access Control:
- Intercepts `/admin/*` routes and checks for valid admin credentials (redirects unauthenticated visitors to `/admin/login`).
- Intercepts `/user/*` routes and redirects guests to `/login`.

---

## 5. CI/CD Flow & Branching Strategy (GitHub Actions)

### 5.1. Git Branching Model (`dev` & `main`)

```mermaid
gitGraph
   commit id: "Initial Commit"
   branch dev
   checkout dev
   commit id: "feat: Add tour filter"
   commit id: "fix: Booking race condition"
   checkout main
   merge dev tag: "v1.0.0 (Trigger Production CD)"
```

- **`main` (Production Branch)**:
  - Contains strictly production-ready code.
  - Protected branch requiring pull request approvals and passing status checks.
  - Merging code into `main` automatically triggers `.github/workflows/production-cd.yml`.
- **`dev` (Development Branch)**:
  - Active integration branch for feature development and testing.
  - Every commit or pull request to `dev` automatically triggers `.github/workflows/dev-ci.yml`.

---

### 5.2. Automated Quality Gate on `dev` (Dev CI)

Workflow file: [`.github/workflows/dev-ci.yml`](file:///.github/workflows/dev-ci.yml)

Triggered on:
- `push` to branch `dev`
- `pull_request` targeting `dev` or `main`

Automated pipeline jobs run in parallel:
1. **`backend-ci`**:
   - Sets up Go `1.25+` with module caching.
   - Runs `go vet ./...` for static code analysis.
   - Executes Go unit tests with race detection: `go test -v -race -coverprofile=coverage.out ./...`.
   - Verifies clean binary compilation: `go build -v ./cmd/server`.
2. **`frontend-ci`**:
   - Sets up Node.js `20.x` and `pnpm 10.x` with store caching.
   - Runs `pnpm install --frozen-lockfile`.
   - Executes ESLint checking: `pnpm run lint`.
   - Validates TypeScript typing and builds production bundle: `pnpm run build`.

---

### 5.3. Production Release & Containerization on `main` (Production CD)

Workflow file: [`.github/workflows/production-cd.yml`](file:///.github/workflows/production-cd.yml)

Triggered on:
- `push` to branch `main` (merge commits)

Deployment pipeline stages:
1. **`quality-gate`**: Re-verifies all backend unit tests and frontend linting/compilation before container builds.
2. **`build-and-push-api`**:
   - Uses `docker/build-push-action@v5` with Buildx and GitHub Actions cache.
   - Builds multi-stage production image from `apps/api/Dockerfile`.
   - Automatically tags and publishes to GitHub Container Registry:
     - `ghcr.io/<org>/sun-booking-tours/api:<git-sha>`
     - `ghcr.io/<org>/sun-booking-tours/api:latest`
3. **`build-and-push-web`**:
   - Builds standalone production Next.js image from `apps/web/Dockerfile`.
   - Automatically tags and publishes to GHCR:
     - `ghcr.io/<org>/sun-booking-tours/web:<git-sha>`
     - `ghcr.io/<org>/sun-booking-tours/web:latest`
4. **`deploy-production`**:
   - Triggers production deployment environment via SSH, Kubernetes manifests, or webhook.

---

## 6. Environment Configuration (.env.example)

Copy `.env.example` to `.env` in the repository root:

```bash
cp .env.example .env
```

| Variable | Recommended Default | Description |
|---|---|---|
| `POSTGRES_USER` | `sun_booking` | Database username |
| `POSTGRES_PASSWORD` | `sun_booking_secret` | Database user password |
| `POSTGRES_DB` | `sun_booking_tours` | PostgreSQL database name |
| `POSTGRES_PORT` | `5434` | Host port mapping (prevents conflicts with host 5432) |
| `DATABASE_URL` | `postgres://sun_booking:sun_booking_secret@localhost:5434/sun_booking_tours?sslmode=disable` | Connection DSN string for API server |
| `PGADMIN_PORT` | `5050` | pgAdmin Web UI host port |
| `PGADMIN_EMAIL` | `admin@sunbooking.com` | pgAdmin login email |
| `JWT_SECRET` | `super-secret-jwt-key-...` | Secret key used for signing authentication tokens |
| `BANK_ENCRYPTION_KEY` | `my-super-secret-key-32b` | AES-256 secret key for encrypting user bank account numbers |
| `GOOGLE_CLIENT_ID` / `_SECRET` | `...` | Google OAuth 2.0 credentials |
| `FACEBOOK_CLIENT_ID` / `_SECRET` | `...` | Facebook OAuth credentials |
| `TWITTER_CLIENT_ID` / `_SECRET` | `...` | Twitter / X OAuth credentials |
| `VNPAY_TMN_CODE` / `_SECRET` | `...` | Sandbox Internet Banking gateway credentials |
| `NEXT_PUBLIC_SITE_URL` | `http://localhost:3000` | Frontend public domain |
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080/api/v1` | Backend API base URL |

---

## 7. Quick Start Guide

### Step 1: Start PostgreSQL & pgAdmin
```bash
# Launch PostgreSQL 16 and pgAdmin 4 in Docker
make up
```
- **PostgreSQL**: `localhost:5434` (User: `sun_booking`, DB: `sun_booking_tours`)
- **pgAdmin 4**: [http://localhost:5050](http://localhost:5050) (Login: `admin@sunbooking.com` / `admin123`)

### Step 2: Apply Database Schema & Seed Data
```bash
# Apply schema migration (15 tables, soft delete, triggers, materialized views)
make migrate-up

# Insert initial seed data (categories, sample tours, departure schedules, admin user)
make migrate-seed
```

### Step 3: Run Backend Tests & Start API Server
```bash
# Run backend unit tests
make test-api

# Start Go backend server
make run-api
```
- API Base: `http://localhost:8080`
- Health Check: `curl http://localhost:8080/health`

### Step 4: Run Frontend Lint & Start Next.js App
```bash
# Check frontend lint
make lint-web

# Run Next.js in development mode
make dev-web
```
- **Customer Portal**: [http://localhost:3000](http://localhost:3000)
- **Tour Details & SEO Preview**: [http://localhost:3000/tours/phu-quoc-tropical-island-discovery-3d2n](http://localhost:3000/tours/phu-quoc-tropical-island-discovery-3d2n)
- **Community Travel Guides**: [http://localhost:3000/reviews](http://localhost:3000/reviews)
- **Admin Dashboard**: [http://localhost:3000/admin/dashboard](http://localhost:3000/admin/dashboard)
- **Batch Revenue Reports**: [http://localhost:3000/admin/revenue](http://localhost:3000/admin/revenue)

---

## 8. Makefile Cheatsheet

| Command | Action |
|---|---|
| `make up` | Start PostgreSQL 16 and pgAdmin containers in Docker |
| `make down` | Stop and remove the project containers |
| `make db-shell` | Open an interactive `psql` shell inside the database container |
| `make migrate-up` | Apply schema migration creating all 15 tables, indexes, and triggers |
| `make migrate-down` | Rollback the database schema to an empty state |
| `make migrate-seed` | Populate initial seed data (categories, admin account, tours, schedules) |
| `make run-api` | Run the Go Echo API server locally |
| `make test-api` | Run backend Go unit tests with test output |
| `make dev-web` | Run the Next.js frontend in development mode (`localhost:3000`) |
| `make lint-web` | Run Next.js frontend ESLint inspection |
| `make build-web` | Build the Next.js frontend production standalone bundle |
| `make docker-api` | Build backend Docker container image locally |
| `make docker-web` | Build frontend Next.js Docker container image locally |