package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// advisoryLockKey is technical-spec §4.6's ADV_LOCK_KEY. It is a
// compile-time constant and is never taken from request input.
const advisoryLockKey = 4271001

// refreshTimeout caps a single refresh so a pathological run cannot hold
// the advisory lock indefinitely (BR-002 / R4).
const refreshTimeout = 15 * time.Minute

// unlockTimeout bounds the unlock issued on the unwind path. It is derived
// from context.Background(), not from the refresh's own context, so the
// unlock still runs when the refresh was killed by refreshTimeout.
const unlockTimeout = 30 * time.Second

const (
	tryLockSQL = "SELECT pg_try_advisory_lock($1)"
	unlockSQL  = "SELECT pg_advisory_unlock($1)"
	refreshSQL = "CALL refresh_revenue_reports()"
)

// RefreshConn is the subset of *pgxpool.Conn the runner uses. The whole
// operation must run on ONE session: pg_try_advisory_lock is session-scoped,
// so acquiring through the pool and unlocking through the pool would unlock
// a different backend and wedge every future refresh at 409.
type RefreshConn interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Release()
}

// ConnSource hands out one dedicated session. *pgxpool.Pool satisfies it
// through PoolConnSource; tests substitute a stub.
type ConnSource interface {
	Acquire(ctx context.Context) (RefreshConn, error)
}

// PoolConnSource adapts *pgxpool.Pool to ConnSource. It exists because the
// shared repository.DB interface deliberately does not expose Acquire.
type PoolConnSource struct {
	Pool *pgxpool.Pool
}

// Acquire takes one connection out of the pool for the caller to own.
func (s PoolConnSource) Acquire(ctx context.Context) (RefreshConn, error) {
	return s.Pool.Acquire(ctx)
}

// RevenueRefreshRunner owns A3: it starts at most one concurrent refresh of
// the two materialized views, detached from the triggering request, and
// remembers when the last one finished.
//
// The dedupe arbiter is the database's advisory lock, never a Go mutex: a
// mutex only dedupes inside one replica, the lock dedupes across all of
// them. The mutex here guards the displayed timestamp only.
type RevenueRefreshRunner struct {
	conns   ConnSource
	timeout time.Duration

	mu            sync.RWMutex
	lastRefreshed *time.Time
}

// NewRevenueRefreshRunner builds the runner over a connection source.
func NewRevenueRefreshRunner(conns ConnSource) *RevenueRefreshRunner {
	return &RevenueRefreshRunner{conns: conns, timeout: refreshTimeout}
}

// Trigger attempts to start a refresh. It returns true when this call won
// the advisory lock and a refresh is now running in the background, and
// false when another refresh already holds it (BR-001 — rejected, never
// queued). It returns before the refresh completes, always.
//
// ctx bounds only the acquire and the try-lock. The refresh itself runs on
// a context derived from context.Background(): echo cancels the request
// context the instant the 202 is written, and a refresh tied to it would be
// killed before it did any work (R3).
func (r *RevenueRefreshRunner) Trigger(ctx context.Context, actor string) (bool, error) {
	conn, err := r.conns.Acquire(ctx)
	if err != nil {
		return false, fmt.Errorf("service: acquiring refresh connection: %w", err)
	}

	var acquired bool
	if err := conn.QueryRow(ctx, tryLockSQL, advisoryLockKey).Scan(&acquired); err != nil {
		conn.Release()
		return false, fmt.Errorf("service: acquiring revenue refresh lock: %w", err)
	}
	if !acquired {
		conn.Release()
		return false, nil
	}

	// Ownership of conn transfers to the goroutine, which releases it.
	go r.run(conn, actor)
	return true, nil
}

// LastRefreshed returns when this process last completed a refresh, or nil.
func (r *RevenueRefreshRunner) LastRefreshed() *time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lastRefreshed
}

// run performs the refresh on the caller's already-locked connection and
// always gives the lock and the connection back.
//
// Defer order is load-bearing. Declared recover -> cancel -> release ->
// unlock, so unwinding runs unlock, then release, then cancel, then
// recover. The recover is outermost because a panic in ANY goroutine takes
// the whole process down; containing it degrades a failed refresh to one
// logged error instead of an outage.
func (r *RevenueRefreshRunner) run(conn RefreshConn, actor string) {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("revenue refresh panicked", "panic", rec, "actor_id", actor)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	defer conn.Release()
	defer r.unlock(conn, actor)

	slog.Info("revenue refresh started", "actor_id", actor)
	started := time.Now()

	// REFRESH MATERIALIZED VIEW CONCURRENTLY cannot run inside a
	// transaction block, and pgx's default extended protocol wraps each
	// statement in an implicit one. The simple protocol is what makes this
	// CALL legal; without it the procedure fails, and the only trace is
	// this goroutine's log.
	if _, err := conn.Exec(ctx, refreshSQL, pgx.QueryExecModeSimpleProtocol); err != nil {
		slog.Error("revenue refresh failed", "error", err, "actor_id", actor,
			"duration_ms", time.Since(started).Milliseconds())
		return
	}

	r.setLastRefreshed(time.Now())
	slog.Info("revenue refresh completed", "actor_id", actor,
		"duration_ms", time.Since(started).Milliseconds())
}

// unlock releases the advisory lock on the same session that took it. Its
// context is fresh rather than the refresh's, so a refresh killed by the
// timeout still gives the lock back.
func (r *RevenueRefreshRunner) unlock(conn RefreshConn, actor string) {
	ctx, cancel := context.WithTimeout(context.Background(), unlockTimeout)
	defer cancel()

	if _, err := conn.Exec(ctx, unlockSQL, advisoryLockKey); err != nil {
		slog.Error("releasing revenue refresh lock failed", "error", err, "actor_id", actor)
	}
}

// setLastRefreshed records a successful run. It is never called on the
// failure or panic path — a stale success time is worse than "unknown".
func (r *RevenueRefreshRunner) setLastRefreshed(at time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastRefreshed = &at
}
