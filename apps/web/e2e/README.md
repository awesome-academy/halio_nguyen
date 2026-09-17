# Admin Portal E2E Suite (Playwright)

This suite drives the admin portal (`apps/web`) against the real local stack —
no mocks, no fixture SQL, no DB reset. Tests log in once, create the data they
need through the admin UI, and clean up after themselves.

## Prerequisites

Before running anything, the rest of the stack must already be up:

1. **Postgres** — `make up` (starts `sun_booking_postgres` on host port `5434`).
2. **Go API** — `make run-api` (serves `http://localhost:8080`).

Playwright's `webServer` starts the Next.js dev server on `:3000` for you (see
[Web server behavior](#web-server-behavior) below) — you do not need to run
`pnpm dev` yourself.

## First-time setup

```bash
cd apps/web
pnpm install
pnpm exec playwright install chromium
cp .env.e2e.example .env.e2e
# edit .env.e2e and fill in the real local admin credentials
```

`.env.e2e` is gitignored — it holds a real (dev-only) admin password and must
never be committed. See [Credentials](#credentials) below.

## How to run

```bash
# from the repo root
make test-e2e
# or, from apps/web
pnpm test:e2e

# just the login/session capture
pnpm test:e2e --project=setup

# UI mode / headed for debugging
pnpm test:e2e:ui
pnpm test:e2e:headed
```

The first run (or any run with `PW_FRESH_SERVER=1`) authenticates once via
`e2e/setup/auth.setup.ts` and writes `playwright/.auth/admin.json`. Every
`admin` project spec reuses that captured session — there is no second admin
login anywhere in the suite (see [Session expiry](#session-expiry) and the
plan's decisions §J6 for why: the API rate-limits login attempts).

## Where the report lands

- HTML report: `apps/web/e2e-report/` — open with `make e2e-report` or
  `pnpm test:e2e:report`.
- Machine-readable results: `apps/web/e2e-report/results.json`
  (`stats.expected` / `stats.unexpected` is the evidence of record — not an
  agent's or a person's prose summary of the run).
- Traces / videos / screenshots on failure: `apps/web/test-results/`.

All three directories are gitignored. **They are credential-bearing** — a
trace or video can capture the `Set-Cookie: sun_admin_token=...` header and
other request data from the live session. Never upload them as a CI artifact,
attach them to a ticket, or share them outside your machine.

## Credentials

Read from `apps/web/.env.e2e` (gitignored), with `.env.e2e.example`
(committed) as the documented placeholder template:

| Variable | Meaning |
|---|---|
| `E2E_BASE_URL` | Defaults to `http://localhost:3000`. **Never point this at a staging or production database** — the suite authenticates as a full administrator. |
| `E2E_ADMIN_EMAIL` / `E2E_ADMIN_PASSWORD` | The one admin login the whole run performs, done once in `auth.setup.ts`. |
| `E2E_USER_EMAIL` | Non-admin seed account, used only for deliberate wrong-password/bad-login specs. |
| `E2E_RATE_LIMIT_PROBE` | Opt-in (`1`), and run-alone: the 429 spec poisons the login rate-limit bucket for a few minutes. |

Never hardcode real credentials in any committed file — `.env.e2e.example`
holds placeholders only.

## Session expiry

The `sun_admin_token` cookie is `Max-Age=3600` (1 hour), and `middleware.ts`
only checks that the cookie is *present*, not that it's still valid — so a
stale captured session passes the middleware and only fails later, as a
scattered, hard-to-place 401 somewhere in the middle of an unrelated spec.

`auth.setup.ts` records the capture time next to the storageState file
(`playwright/.auth/admin.json.captured-at`). `e2e/config/session-guard.ts`
exports `assertSessionFresh()`, which throws a clear
`session expired — re-run pnpm test:e2e --project=setup` error instead of
letting expiry surface indirectly. Shared fixtures (phase 02) call this once
per authenticated spec file.

If a run is going to take longer than roughly 50 minutes, re-run
`pnpm test:e2e --project=setup` partway through rather than letting the
session run out mid-suite.

## Web server behavior

By default `webServer.command` is `pnpm dev` (**not** a production build) and
`reuseExistingServer` is true outside CI, so a server you already have running
on `:3000` is reused as-is. Two env switches change that:

- `PW_WEB_MODE=prod` — run `pnpm build && pnpm start` instead of `pnpm dev`.
- `PW_FRESH_SERVER=1` — force a brand-new server even outside CI.

## Build-machine hazards (read before touching any script here)

1. **Never call bare `next build`, bare `next start`, or `rm -rf .next`, in
   any script or config in this suite.** `next.config.mjs` sets
   `output: 'standalone'`, which on pnpm + Windows writes directory
   **symlinks** under `.next/standalone/node_modules/*`. A later build, or a
   git-bash `rm -rf .next`, follows those symlinks and empties the real
   `next` package in `node_modules`. Always go through the existing
   `pnpm dev` / `pnpm build` scripts — both route through
   `apps/web/scripts/clean-next-dir.mjs`, which unlinks `.next` with
   `fs.rmSync` (safe: it does not follow symlinks) before starting.
2. **A stale `next dev` on `:3000` silently serves old code.** `pkill` does
   not terminate Windows `node.exe` processes; a zombie dev server has
   previously sat on `:3000` returning HTTP `000` for an entire day. Before a
   confusing run, use `make e2e-kill-web` — it resolves the PID bound to port
   `3000` via `Get-NetTCPConnection` and stops exactly that process, rather
   than pattern-matching on command line text (which could otherwise kill an
   unrelated `node.exe`, e.g. Nextcloud, or another project's dev server).
3. **`auth.setup.ts` includes a liveness probe** (`GET /api/v1/admin/auth/me`
   right after login must return `200`) specifically because a stale or
   half-started server usually fails there first — it's the fastest signal
   that the dev server, the `/api/v1` rewrite, and the Go API are all really
   talking to each other.

## Scope of this suite

Selectors are role/aria/text only (`getByRole`, `getByLabel`, `getByText`,
`getByPlaceholder`) — no `data-testid` is ever added to application code
(`apps/web/src/**` and `apps/api/**` are read-only for this whole plan). If a
future spec genuinely cannot be written without a production change, that is
a blocker for the user to decide, not something to quietly work around.
