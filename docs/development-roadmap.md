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
| 4 | Tour package backend (F003) | Pending | — |
| 5 | Tour package frontend (F003) | Pending | — |
| 6 | Booking request management (F004) | Pending | — |
| 7 | Platform user management (F005) | Pending | — |
| 8 | Review and comment moderation (F006) | Pending | — |
| 9 | Revenue analytics and dashboard (F007) | Pending | — |
| 10 | Integration and hardening | Pending | — |

**Progress:** 3 / 10 phases (≈ 22h of 74h estimated).

## Open items carried between phases

- Apply `apps/api/scripts/fix_admin_seed_hash.sql` on any database seeded before 2026-09-10.
- R3: confirm `c.RealIP()` differs per client through the Next proxy (Phase 10).
- CI has no Postgres (L5); repository SQL is mock-tested only.
- `go test -race` needs a C toolchain — unavailable on the current build machine.
- L8: soft-deleted categories keep `name`/`slug` reserved (table-level UNIQUE, schema frozen).
- SC-002/SC-003 (delete persistence, reorder contiguity) proven at service level only until a
  Postgres is available.
