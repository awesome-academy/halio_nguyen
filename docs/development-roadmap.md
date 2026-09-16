# Development Roadmap

The detailed, authoritative roadmap is the admin-portal plan at
`plans/260908-0912-admin-portal-full-stack/plan.md` (phase files, dependency graph, decisions log —
local working notes; `plans/` is git-ignored). This page is the at-a-glance status.

## Admin Portal — full-stack build (branch `dev`)

| Phase | Scope | Status | Done |
|-------|-------|--------|------|
| 1 | Shared foundation (Go skeleton, error contract, shadcn, proxy, query client) | Completed | 2026-09-09 |
| 2 | Admin authentication + RBAC gate (F001) | Completed | 2026-09-10 |
| 3 | Tour category management + shared FE data-table library (F002) | Completed | 2026-09-11 |
| 4 | Tour package backend (F003) | Completed | 2026-09-14 |
| 5 | Tour package frontend (F003) | Completed | 2026-09-15 |
| 6 | Booking request management (F004) | Completed | 2026-09-16 |
| 7 | Platform user management (F005) | Completed | 2026-09-16 |
| 8 | Review and comment moderation (F006) | Completed | 2026-09-16 |
| 9 | Revenue analytics and dashboard (F007) | Completed | 2026-09-16 |
| 10 | Integration and hardening | Completed | 2026-09-16 |

**Progress:** 10 / 10 phases — **admin portal build complete**.

Phase 10's closing walkthrough passed every row with nothing handed back to an earlier phase.
Evidence (gate exit codes, a 50-call RBAC sweep, database-level proofs, and the L1–L7
dispositions) is in `plans/260908-0912-admin-portal-full-stack/reports/integration-walkthrough.md`.

## Open items carried between phases

- Apply `apps/api/scripts/fix_admin_seed_hash.sql` on any database seeded before 2026-09-10.
  (Applied to the local dev database on 2026-09-16; still outstanding for every other environment.)
- ~~R3: confirm `c.RealIP()` differs per client through the Next proxy.~~ **Closed 2026-09-16.**
  Distinct `X-Forwarded-For` clients get distinct throttle buckets; a fresh IP is not throttled by
  another IP's exhausted budget. Note for anyone re-testing: D3's *per-email* limb answers 429 with
  the same body as the per-IP limb, so vary the email too or the result reads as a false failure.
- CI has no Postgres (L5); repository SQL is mock-tested only. Phases 6, 7 and 8 were additionally
  verified by hand against a local Postgres — Phase 6 caught a NULL-scan defect the mocks could
  not, Phase 7 proved BR-002's `FOR UPDATE` race fix, which no mock can exercise, and Phase 8 found
  an ambiguous-column bug that made its detail endpoint 500 on **every** call while the whole mock
  suite stayed green. Three phases, three defects only a real database surfaced: treat the live
  pass as mandatory for Phase 9, not optional. **Phase 9 followed that rule**: its live pass proved
  the advisory-lock release and that the refresh survives a disconnected client — neither of which
  any mock can show. Phase 10 re-ran every phase's `psql`/`curl` proof together against one
  database; that manual pass remains the only integration evidence for the repository SQL
  (7.4% package coverage).
- `go test -race` needs a C toolchain — unavailable on the current build machine (confirmed again
  2026-09-16: no `gcc` on PATH). The suite runs locally without `-race`; the race detector is
  covered by CI only.
- L8: soft-deleted categories keep `name`/`slug` reserved (table-level UNIQUE, schema frozen).
- ~~Phase 6 / L1: `domain.UserBankAccount.AccountNumber` has no `json:"-"` tag.~~ **Closed
  2026-09-16 (Phase 10).** The field is now `json:"-"` with a regression test in
  `domain/user_test.go`; the struct is safe regardless of which query populates it.
- Phase 6 / M1: booking search (`FR-203`) runs `ILIKE` over unindexed `contact_name`/`contact_email`.
  Acceptable now; consider `pg_trgm` before volume grows.
- SC-002/SC-003 (delete persistence, reorder contiguity) proven at service level only until a
  Postgres is available.
- Phase 7 / L1: deactivate / demote / soft-delete write no `activity_logs` row —
  `activity_logs.action`'s CHECK enum has no value for them and the schema is frozen. Structured
  `slog` (actor id + target id + changed-field flags, never PII) is the trail. **Phase 10 revisited
  and accepted it**: the fallback lines were confirmed present in all four mutating services and
  observed firing during the walkthrough. A real audit trail needs the schema change A2 forbids.
- Phase 7 / L2a: a demotion or deactivation does not revoke the target's already-issued JWT — the
  RBAC gate reads the `role` claim, never the live row, so a demoted admin keeps access for up to
  the 1h token TTL. Accepted consequence of the stateless-JWT decision (A1).
- Phase 7 / M1: `guardLastAdmin` row-locks the whole active-admin set on every A3/A4 mutation, not
  only when the target is an admin — as the plan specifies. Harmless at single-digit admin scale;
  if admin-mutation volume ever grows, pre-check the target row first and escalate to the full
  lock only when it is currently an active admin.
- Phase 8 / L1: review and comment publish/hide/delete write no `activity_logs` row — same frozen
  CHECK enum as Phase 7. Structured `slog` (actor id, action, review id, comment id) is the only
  record that a moderation decision was made, and it is what an incident review would work from.
- Phase 8 / L2: no restore for a soft-deleted review or comment. Accepted default, stated in the
  confirm dialogs; recovery needs a manual `UPDATE ... SET deleted_at = NULL`.
- Phase 8 / M1: A2 returns the whole comment tree unpaginated. Fine at current volume; a review
  with a very large thread would ship it all in one response. **Phase 10: still accepted** — no
  volume problem at current data; paginating a nested tree is its own design, not a closing task.
- Phase 8 / M2: A3 accepts `draft → published`, which the spec's state diagram frames as the
  authoring flow ("out of scope"). Permitted by design — publishing is the only sensible action on
  a draft row — but it is a product call worth confirming rather than an accident.
- Phase 8 / note: browser-automation agents proved unreliable in this environment — two consecutive
  runs reported results for actions that never reached the server (verified against the API access
  log). Cross-check any browser-driven E2E claim against the access log or the database.
- **Operational:** promote a second admin through `PATCH /api/v1/admin/users/{id} {"role":"admin"}`
  as the first real action in any new environment. Phase 10 proved the path end to end (D1) on the
  local database and then restored the seeded roles, so **this is still outstanding for every
  deployed environment**: until a second admin exists there, BR-002 has one account to protect and
  RISK-001's manual-DB-operation scenario stays live.

## Follow-ups after the admin portal (each needs its own design)

- Postgres service in CI (L5) — repository SQL has no automated coverage today.
- Shared-store login rate limiting (L4) — the per-replica budget multiplies with replica count.
- Admin session revocation or short-lived refresh (L2) — a leaked token lives up to an hour.
- A scheduler for `refresh_revenue_reports()` — nothing schedules it; the revenue views only move
  when an admin presses the button.
- An audit trail (L1) — requires the schema change A2 currently forbids.
