# Project Changelog

Running record of significant changes. Newest first. Plan of record:
`plans/260908-0912-admin-portal-full-stack/plan.md` (local working notes; `plans/` is git-ignored).

## 2026-09-11 — Tour category management + shared admin DataTable (Phase 3, F002)

**Added**
- `GET/POST /api/v1/admin/categories`, `PUT /…/:id`, `PATCH /…/:id/sort-order`, `DELETE /…/:id`
  — list (search, `is_active` filter, allowlisted sort, `tour_count` per row, `COUNT(*) OVER()`),
  create/update with server-side slug derivation (`service.Slugify`, Vietnamese-aware), reorder,
  soft-delete.
- ALG-001 reorder as a `SELECT … FOR UPDATE` rank-rewrite in one transaction; it is the **only**
  write path for `sort_order` (create appends, update preserves, a requested position moves) so the
  non-deleted set stays contiguous `0..n-1`.
- BR-001 / D4 delete guard: row lock + live-tour count + soft-delete in one transaction; blocked
  deletes return `409` naming the count and write nothing.
- Unique violations (`23505`) mapped by constraint name to field-scoped `409`s
  (`repository.UniqueViolationField`) — no pre-check `SELECT`, no constraint names in bodies.
- Web: shared admin library for every later screen — `DataTable` family (react-table v9,
  server-driven), `useListQueryState` (URL-backed page/sort/search/filters), `ConfirmDialog`,
  `PageHeader`; `/admin/categories` screen with create/edit dialog (slug auto-fill until edited,
  field-level `409` errors), reorder arrows (boundary-disabled), guarded delete dialog.
- `router.Build(cfg, pool)` constructs every handler chain; `cmd/server/main.go` is now frozen.
- `apps/web/scripts/clean-next-dir.mjs` — symlink-safe `.next` cleanup before `dev`/`build`.

**Fixed**
- Recurring "Cannot find module next/dist/…" build failures on Windows + pnpm: `output: 'standalone'`
  symlinks under `.next/standalone/node_modules` were followed by the `.next` clear-out, emptying the
  real `next` package. Severity: high (blocked builds); environment-only, no runtime impact.

**Known limitations** — L8 (new): `categories.name`/`slug` are table-level `UNIQUE`, so a
soft-deleted category keeps both reserved forever (the 409 message says so). Repository SQL still
mock-tested only (L5).

## 2026-09-10 — Admin authentication and RBAC gate (Phase 2, F001)

**Added**
- `POST /api/v1/admin/auth/login` · `POST /api/v1/admin/auth/logout` · `GET /api/v1/admin/auth/me`.
- A0 gate on `/api/v1/admin/*`: `echo-jwt` reading the HttpOnly `sun_admin_token` cookie (HS256
  pinned) then `RequireAdminRole`; a wildcard fallback makes unbuilt admin routes 401, not 404.
- Login throttling (D3): per-IP Echo rate limiter on the login route + per-email fixed-window
  throttle in the service. In-memory, per-replica (L4).
- Anti-enumeration login (BR-001): one message/status for wrong password, non-admin, inactive and
  soft-deleted accounts; bcrypt always runs (dummy hash on miss) for timing parity.
- `ValidateRedirect` (DEC-001): only `/admin/*` same-origin paths, no `..` segments, no `\`.
- Web: `/admin/login` (RHF + zod), `AdminShell` with real identity and working Sign Out,
  global 401 → `/admin/login?from=…` redirect (loop-guarded).
- `apps/api/scripts/fix_admin_seed_hash.sql` — idempotent hash fix for already-seeded databases.

**Fixed**
- Seeded admin `password_hash` did not verify against its documented password (decisions.md D5);
  replaced with a verified bcrypt cost-10 hash of `Admin@123456`. Severity: blocking — no login
  was possible.

**Security**
- `JWT_SECRET` is now fail-closed: outside development, unset or < 32 chars aborts boot.
- `TRUSTED_PROXY_CIDRS` now actually drives the X-Forwarded-For extractor; echo's default trust of
  every RFC1918 address is off, so only the listed proxy hops can set the client IP.
- Redirect containment hardened against backslash traversal (`/admin/..\..\x`).

**Known limitations carried forward** — L1 no admin audit trail, L2 logout cannot revoke a token
(1h TTL is the mitigation), L2a demotion takes effect only at token expiry, L4 per-replica limits,
L5 repository SQL untested in CI (no Postgres). R3 (real client IP through the Next proxy) not yet
empirically verified — Phase 10.

## 2026-09-09 — Shared foundation (Phase 1)

- Go package skeleton (`apperror`, `repository` params/DB interface, `router`), single JSON error
  contract, list pagination/sort-allowlist primitives, hardening config (`COOKIE_SECURE`,
  login rate-limit knobs, `TRUSTED_PROXY_CIDRS`, `JWT_ACCESS_EXPIRY_HOURS=1`).
- Web: shadcn/ui primitives, TanStack Query provider, `rewrites()` proxy to the Go API, base
  admin API client with a pluggable 401 handler.
