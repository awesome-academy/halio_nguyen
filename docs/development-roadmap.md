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
| 8 | Review and comment moderation (F006) | Pending | — |
| 9 | Revenue analytics and dashboard (F007) | Pending | — |
| 10 | Integration and hardening | Pending | — |

**Progress:** 7 / 10 phases (≈ 57h of 74h estimated).

## Open items carried between phases

- Apply `apps/api/scripts/fix_admin_seed_hash.sql` on any database seeded before 2026-09-10.
  (Applied to the local dev database on 2026-09-16; still outstanding for every other environment.)
- R3: confirm `c.RealIP()` differs per client through the Next proxy (Phase 10).
- CI has no Postgres (L5); repository SQL is mock-tested only. Phases 6 and 7 were additionally
  verified by hand against a local Postgres — Phase 6 caught a NULL-scan defect the mocks could
  not, and Phase 7 proved BR-002's `FOR UPDATE` race fix, which no mock can exercise. Worth
  repeating for Phases 8–9 rather than trusting mocks alone.
- `go test -race` needs a C toolchain — unavailable on the current build machine.
- L8: soft-deleted categories keep `name`/`slug` reserved (table-level UNIQUE, schema frozen).
- Phase 6 / L1: `domain.UserBankAccount.AccountNumber` has no `json:"-"` tag, so any future query
  that populates `Payment.UserBankAccount` serialises a full bank account number. No current path
  does, and Phase 6 asserts it on raw bytes — but the tag belongs on the struct (Phase 10).
- Phase 6 / M1: booking search (`FR-203`) runs `ILIKE` over unindexed `contact_name`/`contact_email`.
  Acceptable now; consider `pg_trgm` before volume grows.
- SC-002/SC-003 (delete persistence, reorder contiguity) proven at service level only until a
  Postgres is available.
- Phase 7 / L1: deactivate / demote / soft-delete write no `activity_logs` row —
  `activity_logs.action`'s CHECK enum has no value for them and the schema is frozen. Structured
  `slog` (actor id + target id + changed-field flags, never PII) is the trail. Phase 10 to revisit.
- Phase 7 / L2a: a demotion or deactivation does not revoke the target's already-issued JWT — the
  RBAC gate reads the `role` claim, never the live row, so a demoted admin keeps access for up to
  the 1h token TTL. Accepted consequence of the stateless-JWT decision (A1).
- Phase 7 / M1: `guardLastAdmin` row-locks the whole active-admin set on every A3/A4 mutation, not
  only when the target is an admin — as the plan specifies. Harmless at single-digit admin scale;
  if admin-mutation volume ever grows, pre-check the target row first and escalate to the full
  lock only when it is currently an active admin.
- **Operational:** promote a second admin through `PATCH /api/v1/admin/users/{id} {"role":"admin"}`
  as the first real action after Phase 7 ships. Until then BR-002 has exactly one account to
  protect and RISK-001's manual-DB-operation scenario stays live.
