# Project Changelog

Running record of significant changes. Newest first. Plan of record:
`plans/260908-0912-admin-portal-full-stack/plan.md` (local working notes; `plans/` is git-ignored).

## 2026-09-16 — Revenue analytics, and plan closeout (Phases 9 & 10, F007)

**Added**
- 3 admin routes: `GET /api/v1/admin/revenue/daily` (per tour per day, paginated, allowlisted
  sort, ISO date range defaulting to the current month), `GET .../monthly` (per category per
  month, whole range in one response), and `POST .../refresh` (202 / 409).
- Admin UI at `/admin/revenue`: daily and monthly tables over one shared date range, VND
  formatting, a refresh trigger, and a last-refresh indicator. A shared
  `components/admin/date-range-filter.tsx` drives both tables from the URL.
- Both revenue reports and the dashboard tile read the materialized views directly; no aggregation
  is recomputed in Go.

**Changed**
- `/admin/dashboard` and `/admin/revenue` no longer render mock data. Both pages were rewritten:
  the dashboard is now **revenue-only** (decisions.md B1) — one tile with this month's total and
  top category — and the fabricated KPI cards and "Recent Booking Requests" table are gone, not
  rewired. The revenue page's sample row arrays and its non-functional "Export CSV" button are
  removed (§7 out of scope). No mock, sample or hardcoded business data remains anywhere under
  `apps/web/src/app/(admin)/`.

**The substance of this feature — the refresh, not the reads**
- `REFRESH MATERIALIZED VIEW CONCURRENTLY` cannot run inside a transaction, and pgx's default
  extended protocol wraps every statement in an implicit one. The `CALL` is issued with
  `pgx.QueryExecModeSimpleProtocol` on a dedicated connection; without that it fails silently in a
  background goroutine with no HTTP response to fail into.
- `pg_try_advisory_lock` is **session**-scoped. The acquire, the `CALL` and the unlock all run on
  one `*pgxpool.Conn` owned by the goroutine, because taking the lock through the pool would leave
  it held on whichever pooled connection happened to serve the call — wedging every later refresh
  at 409 permanently. Verified live: `pg_locks` is empty afterwards and three consecutive triggers
  all return 202.
- The refresh runs on a `context.Background()`-derived context with a 15-minute ceiling, so it
  survives the request that started it (a client aborted mid-refresh still saw the views update).
  An outermost `recover()` contains a panic that would otherwise take the whole process down.
- Last-refresh time lives in process memory only (L3): it is `null` after a restart and the UI
  says "unknown" rather than inventing a time. No table was added; the schema stays frozen.

**Verified (Phase 10)**
- All 7 CI-equivalent gate commands exit 0, run before and after cleanup.
- 50-call RBAC sweep across every admin route: 401 without a cookie, 403 with `role:"user"`, zero
  failures. Exactly two ungated routes (`auth/login`, `auth/logout`), as designed.
- Full walkthrough passed with database-level proofs: booking cancel restores schedule slots and
  writes its `activity_logs` row; deleting a user with a live booking is blocked 409 and succeeds
  204 once cancelled; the last-admin guard closes and re-opens; a deleted parent comment's replies
  survive as moderable rows under a redacted tombstone.
- Limitations L1–L7 each carry a written disposition. Nothing was handed back to an earlier phase.
- Report: `plans/260908-0912-admin-portal-full-stack/reports/integration-walkthrough.md`.

**Fixed**
- `apps/api/internal/service/category_service_test.go` was 256 lines, over the project's 200-line
  limit; split mechanically into two files with no test changed or removed.

**Known limitations carried forward** — L1 no admin audit trail (`slog` is the record), L2 logout
cannot revoke a token (1h TTL verified live as the mitigation), L3 no persisted last-refresh
timestamp, L4 login throttle is per-replica, L5 repository SQL still untested in CI (7.4% package
coverage; the manual `psql`/`curl` pass is the only integration evidence), L6 monthly revenue has
no per-tour breakdown. Follow-ups recorded, none started: Postgres in CI, shared-store rate
limiting, session revocation, a scheduler for `refresh_revenue_reports()` (nothing schedules it
today — the views only move when an admin presses the button).

**Local gate note** — `go test -race` could not run on the build host (no C toolchain); the suite
was run without `-race`. The race detector remains covered by CI only.

## 2026-09-16 — Review and comment moderation (Phase 8, F006)

**Added**
- 7 admin routes: `GET /api/v1/admin/reviews` (list, title search on the existing trigram index,
  status + review-category filters, allowlisted sort), `GET .../:id` (review plus its full nested
  comment tree), `PATCH .../:id/status`, `DELETE .../:id` (soft-delete), plus the review-scoped
  `PATCH .../:id/comments/:commentId/visibility` and `DELETE .../:id/comments/:commentId`.
- `GET /api/v1/admin/review-categories` — a read-only lookup added beyond the written spec (A7).
  The spec called for a category filter but exposed no way to populate it; deliberately not CRUD,
  since managing that taxonomy stays out of scope. `review_categories` is a different table from
  Phase 3's *tour* categories.
- Admin UI at `/admin/reviews`: list with status/category filters and row-level moderation, and a
  detail screen carrying the review body and the nested thread. The sidebar "Reviews & Comments"
  link, dead since Phase 3, now resolves.
- `comment_thread_builder.go` — a pure, separately tested function that assembles the tree
  iteratively: attach, promote orphans to root, prune deleted leaves bottom-up, redact the rest.

**The substance of this feature — reading, not writing**
- Every write is a single-column flip on one row; there is no transaction anywhere in the phase.
  The complexity is entirely in assembling a thread that stays complete when parts of it have been
  deleted out from under it.
- `comments.parent_id`'s `ON DELETE CASCADE` fires only on a hard `DELETE`, which this app never
  issues. So A2 fetches **all** comments regardless of `deleted_at` and redacts server-side: a
  soft-deleted parent with live replies ships as `{id, parent_id, created_at, is_deleted}` carrying
  no content and no author, with its replies intact and independently moderable beneath it
  (BR-003). A deleted comment with no surviving descendant is omitted entirely, so the admin never
  scrolls phantom placeholders.
- `like_count`/`comment_count` are trigger-maintained; no repository, service or handler in this
  feature has a write path to either (BR-002/BR-004), and the service tests assert the absence
  rather than assuming it.

**Fixed (found by live verification, invisible to the mock suite)**
- `GET /api/v1/admin/reviews/:id` returned 500 on **every** call: the shared column list was
  unqualified and interpolated into a three-table join where `reviews`, `users` and
  `review_categories` each carry `id`/`created_at`/`updated_at`/`deleted_at`, so Postgres rejected
  the statement with `column reference "id" is ambiguous` (SQLSTATE 42702). Severity: blocking —
  the detail screen could never have loaded. Every mock-based test passed it.
- The tree builder silently dropped comments caught in a `parent_id` cycle or self-reference: with
  no member reachable from a root, the walk never entered the ring and those comments vanished from
  the response — invisible and therefore unmoderatable, violating the builder's own invariant.
  Rescued in `comment_thread_cycle_rescue.go`, with tests proven RED before the fix.
- `ListCommentsForReview` had no `ORDER BY` tiebreaker, leaving sibling order non-deterministic
  under equal timestamps; now `c.created_at ASC, c.id`, matching how the list query appends `r.id`.

**Security**
- A2 no longer ships `author_email`. It was joined, carried through the DTO and declared in the TS
  types but rendered nowhere — dead PII on the wire. Author identity is name and avatar,
  display-only, as the phase specifies.
- Deleted comment content is redacted in the builder, server-side. The placeholder carries
  structure only; content is never sent and hidden in the client.
- Review bodies and comment text are user-authored and render as plain text in a high-privilege
  admin session — no `dangerouslySetInnerHTML`, no markdown-to-HTML, no auto-linking.
- Both comment writes are scoped `id = @comment_id AND review_id = @review_id`, so a comment cannot
  be moderated through a review it does not belong to (404, nothing written).
- `draft` is rejected as a moderation target by a two-value allowlist in the service, before any
  query runs (422) — not left to the database's three-value CHECK.

**Verification**
- Backend gates green (`go build` / `go vet` / `go test ./...`; no `-race`, no C toolchain).
  Frontend `pnpm lint` and `pnpm build` green.
- Exercised live against local Postgres, and again through the Next rewrites proxy with the real
  session cookie: hiding a comment moved `comment_count` 5→4 and unhiding restored it, by trigger
  alone; a childless deleted comment vanished; a deleted parent returned as a content-free,
  author-free placeholder with both replies intact; a soft-deleted review 404'd on A2, left A1, and
  kept its comments in storage (BR-001). Six concurrent hides of one comment all returned 200 with
  a final count matching a ground-truth recount.

**Known limitations carried forward** — L1 no moderation audit trail (`activity_logs.action` has no
value for publish/hide/delete and the schema is frozen; structured `slog` with actor and target ids
is the trail). No restore path for a soft-deleted review or comment — an accepted default, stated
plainly in the confirm dialogs; recovery is a manual `UPDATE ... SET deleted_at = NULL`. A1's title
search rides the existing trigram index, but the comment-tree read is not paginated: a review with
a very large thread returns it whole (acceptable at current volume, revisit in Phase 10).

## 2026-09-16 — Platform user management (Phase 7, F005)

**Added**
- 4 admin routes: `GET /api/v1/admin/users` (list, search over email + full name, role and
  active-status filters, allowlisted sort), `GET .../:id` (profile + lifetime booking/review
  counts), one combined `PATCH .../:id` for `is_active` and/or `role`, and `DELETE .../:id`
  (soft-delete). No create-user endpoint by design — see D1 below.
- Admin UI at `/admin/users`: list with row-level status toggle and delete, and a detail screen
  carrying the profile card, the two history-count badges, the role dialog, the status toggle and
  delete. Role change is detail-only (FR-005 / SCR002). The sidebar "User Management" link, dead
  since Phase 3, now resolves.
- `BookingRepository.CountActiveByUser` added as a third sibling of `CountActiveByTour` /
  `CountActiveBySchedule` on Phase 4/6's file — no second repository, no fourth copy of the count.

**Guards — the substance of this feature**
- **BR-001 self-lockout.** Any self-target on `PATCH` or `DELETE` is `403`, without inspecting
  which fields the body carries. Enforced server-side against the verified JWT subject (B2); the
  hidden buttons in the UI are usability only and are not the control.
- **BR-002 last-admin, refined.** The spec's literal pseudocode ("target is admin and active
  admins <= 1 → 409") would block *reactivating* a deactivated sole admin and block a no-op
  self-restatement, leaving no way out of a locked-out state. The implemented guard fires on the
  *effective change* instead: only a mutation that would **remove** an active admin — demote,
  deactivate, delete — trips it. Promotions, reactivations and no-ops succeed.
- **BR-002 concurrency.** The guard runs inside the mutation's own transaction and takes the count
  via `SELECT id ... WHERE role='admin' AND is_active AND deleted_at IS NULL FOR UPDATE`. A bare
  `COUNT(*)` outside a transaction would let two simultaneous demotions of the last two admins
  both pass and lock the portal out permanently, with no recovery path.
- **BR-003 active-booking delete guard.** `bookings.user_id` is `ON DELETE RESTRICT`, which fires
  only on a hard `DELETE` — this app soft-deletes with `UPDATE ... SET deleted_at`, so the FK is
  inert and this service check is the *only* thing preventing an orphaned active booking. `409`
  naming the blocking count.
- Guard order is load-bearing and observable: BR-001 → BR-002 → BR-003. A sole admin deleting
  themselves gets `403`, not `409`.

**Decisions carried through**
- **D1 — role-promotion IS the second-admin creation path.** `PATCH {"role":"admin"}` on a
  non-self account is the sanctioned mechanism; no create-admin endpoint was added, and BR-002's
  recovery path depends on this existing. The F005 spec's "out of scope" framing is superseded.
- **L1 — no audit row for deactivate / demote / delete.** `activity_logs.action`'s CHECK enum has
  no value for any of them and the schema is frozen. Structured `slog` carries actor id, target id
  and changed-field flags — never email, phone or name.
- **D6 — an unrecognised `sort_by` falls back silently** to `created_at DESC` through the shared
  `ResolveSort`, rather than F005 §9's `400`. One sort path across the portal beats a per-feature
  status code; both are safe, only the observable code differs. A malformed path UUID likewise
  returns `400` through the shared `pathUUID` rather than the phase todo's `422`, for the same
  one-parser reason. Both deviations are recorded in code comments.

**Security**
- `password_hash` never leaves the API. `json:"-"` on `domain.User` is the mechanism; `UserDetail`
  embeds `*User` rather than re-declaring its fields or building a map, so the guarantee is
  structural. Asserted against the raw marshalled bytes of every response, not a typed struct.
- Privilege escalation is this feature's headline risk — `PATCH` can mint an admin. The actor id is
  read only from the verified JWT subject, `role` is validated against the two domain constants
  before it reaches SQL (the DB CHECK is the backstop, not the check), and a `role:"user"` token
  gets `403` on every route.
- `sort_by` / `sort_dir` go through the `ResolveSort` allowlist; `search`, `role`, `is_active` and
  the id are all bound through `pgx.NamedArgs`, and the `ILIKE` pattern is built by parameter
  binding, never concatenation.
- **Soft-delete is not erasure.** A deleted user's row, bookings and reviews all persist by design
  (history integrity) — this feature is not a GDPR-erasure mechanism and must not be described as
  one.
- **L2a, accepted:** a demotion or deactivation does not revoke the target's already-issued JWT.
  The RBAC gate reads the `role` claim, never the live row, so a demoted admin keeps admin access
  for up to the 1h token TTL.

**Verified against a live database**
Full walkthrough on real Postgres with the server up. BR-001: all three self-target mutations
`403`, admin row unchanged. BR-002: all three mutations against the sole active admin `409`, row
unchanged; reactivating a deactivated sole admin and promoting a user both `200` (the R1
refinement); after promoting a second admin the previously-`409` demotion returns `200` (D1's
recovery path). **BR-002 concurrency:** two simultaneous demotions of the last two admins returned
exactly `{200, 409}` with exactly one active admin left — the `FOR UPDATE` serialization proven,
not asserted. BR-003: deleting a user with one `pending` booking gave `409` naming the count with
`deleted_at` still NULL; after flipping that booking to `completed`, `204` with the booking row
intact, the user gone from the list and `404` on detail. Also confirmed: no `password_hash` in any
raw response body, `401` on every route without a cookie, silent sort fallback, case-insensitive
partial search across both email and full name, and `422` / `422` / `400` / `404` for empty patch /
invalid role / malformed uuid / unknown id. All test data was removed and the database restored to
its pre-test state afterwards.

**Known limits**
- `guardLastAdmin` row-locks the entire active-admin set on every mutation, not only when the
  target is an admin — as the plan specifies. Two admins editing two unrelated regular users
  serialize against each other. A non-issue at single-digit admin scale; noted for Phase 10.
- No optimistic locking on user rows (RISK-002, accepted for v1) — concurrent edits are
  last-write-wins.

## 2026-09-16 — Booking request management (Phase 6, F004)

**Added**
- 5 admin routes: `GET /api/v1/admin/bookings` (list), `GET .../:id` (detail),
  and three guarded transitions `PATCH .../:id/{confirm,cancel,complete}`. Bookings are never
  created or edited here — the customer site owns creation, and nothing in this feature writes
  the `payments` table (refunds are out of scope).
- SM-001 enforced by guarded `UPDATE ... WHERE status = ANY(expected) ... RETURNING`, not by a
  pre-read. A zero-row result *is* the conflict signal; 404 vs 409 is resolved by a follow-up
  existence read that runs only after the guard already rejected the write, so it cannot
  reintroduce the race.
- Cancel is the one multi-table write: booking flip + `tour_schedules.available_slots` restore
  (BR-005) + the single `activity_logs('cancel_tour')` row this portal may write outside
  login/logout, all in one transaction. A failure in any of the three rolls back all three.
- Admin UI at `/admin/bookings`: list (status/tour/schedule/created-date filters, search over
  booking code + contact name + email, allowlisted sort) and detail (tour/departure panel,
  customer panel, read-only payment card, SM-001-driven actions).
- Shared `BookingStatusBadge` lifted out of the dashboard mock into `components/admin/`, so it
  survives that mock's removal in Phase 9. Adds the `cancelled` variant the mock lacked.

**Decisions carried through**
- **B3 — confirm warns but does not block on an incomplete payment.** The server deliberately
  performs no payment check: cash and offline `bank_transfer` settlement are real, so a 409 here
  would be a defect, not extra safety. The warning is client-side only, and the confirm button
  stays enabled. A service test asserts the unpaid confirm *succeeds*, to stop a future
  "hardening" from silently breaking offline settlement.
- **L1 — confirm and complete are not audit-logged.** `activity_logs.action` has a CHECK enum
  with no value for either, and the schema is frozen. Structured `slog` (ids only, never contact
  details) carries the trail instead; the gap is Phase 10's to revisit.
- List responses are fixed to FR-201's columns — `contact_phone`, `contact_email` and
  `special_requests` are detail-only and never widened into the list.

**Fixed**
- `GET /bookings/:id` returned 500 for any booking with no `payments` row
  (`cannot scan NULL into *float64`). `payments` declares `amount`/`payment_method`/`status`/
  `created_at`/`updated_at` NOT NULL, but a LEFT JOIN that matches nothing yields NULL in *every*
  column. Now scanned through a `nullablePayment` struct whose fields are all pointers, with a
  pure `toDomain()`; regression-guarded in CI by `booking_payment_scan_test.go` (no Postgres
  needed). Severity: high — it broke the detail page for every unpaid booking, which is the
  common case for cash/offline settlement.

**Security**
- `user_bank_accounts` is never joined in any booking response. `domain.Payment.UserBankAccount`
  is a nested pointer whose `AccountNumber` carries no `json:"-"` tag, so a populated pointer
  would serialise the full account number verbatim — that nil is the only thing preventing the
  leak. Asserted against the raw response bytes for both A1 and A2.
- The cancellation reason is length-capped server-side (1000 chars) and rendered as text, never
  HTML.
- Hidden/disabled action buttons mirror SM-001 for usability only; the guarded UPDATE is the
  control, verified by calling each transition out of order.

**Verified against a live database**
Docker/Postgres turned out to be available on this machine (the earlier "no Docker" note was
stale), so Phase 6 ran the live steps Phases 3–5 had to skip. On real Postgres: SC-001–SC-004 all
pass; cancel moved slots 3 → 5 for a 2-passenger booking with exactly one audit row attributed to
the acting admin; an empty reason returned 422 with zero writes; `date_from > date_to` returned
422 before the query ran; and all 5 routes returned 401 without a cookie.

Concurrency was proven, not asserted: two simultaneous cancels returned `{200, 409}` across 5 runs
with slots restored exactly once each time (R4 closed), and two simultaneous confirms returned
`{200, 409}` across 3 runs (R3 closed). Simultaneous confirm+cancel returned `{200, 200}` — the
plan predicted `{200, 409}`, but that prediction was wrong rather than the code: cancel's guard
legitimately accepts `confirmed` as well as `pending`, so confirm-then-cancel is a valid
serialisation and the at-most-once slot restore still held.

**Known limitations**
- `go test -race` still cannot run here (no C toolchain); plain `go test ./...` is the gate.
- Repository SQL has no automated integration test — the live verification above was manual and
  its fixtures were removed afterwards. The scan-assembly logic is unit-tested in CI, but the
  queries themselves are only covered by that manual pass.

## 2026-09-14 — Tour package backend (Phase 4, F003)

**Added**
- 13 admin routes across tours/images/schedules: `GET/POST /api/v1/admin/tours`,
  `GET/PUT/PATCH .../:id[/status]`, `DELETE .../:id`; nested `POST/PUT/DELETE .../:id/images[/:imageId]`
  and `.../:id/schedules[/:scheduleId][/status]`.
- A3 create is one transaction across `tours` → `tour_images` → `tour_schedules`; any failure
  (validation, BR-007 category check, a `23505` on either unique constraint) rolls back all three —
  proven by a test that injects a failure on the last insert.
- SM-001 (tour status: draft↔published↔archived) and SM-002 (schedule status: open↔closed,
  →cancelled terminal) as explicit `map[string][]string` transition tables, each with a 422 on an
  illegal transition.
- D4's BR-012 delete guard on both tours (A6) and schedules (A13): row lock + active-booking count +
  soft-delete inside one transaction, so a concurrent booking-create cannot slip between the count
  and the write. Blocks with `409` naming the count; the functional spec's superseded "warn and
  allow" draft was not implemented.
- ALG-001 effective price (`schedule.price_override` → `tour.discount_price` → `tour.price`) as a
  pure function, unit-tested independent of the database.
- `BookingRepository` created here, minimal (`CountActiveByTour`/`CountActiveBySchedule` only) —
  Phase 6 will extend the same file rather than duplicate the count logic.
- Tour slug is always server-derived from the title (`FR-002`); unlike categories, F003 gives the
  admin no client-side override.

**Known limitations**
- No live-database verification was possible on this build machine (no Docker/Postgres); every test
  mocks the `repository.DB`/tx interfaces (D-A7). The phase's "13-action curl walkthrough" step was
  skipped for the same reason.
- `go test -race` cannot run here (no C toolchain) — plain `go test ./...` is the verified gate.

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
