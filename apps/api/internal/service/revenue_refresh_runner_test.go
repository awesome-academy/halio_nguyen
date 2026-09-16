package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// stubRow returns a fixed pg_try_advisory_lock result.
type stubRow struct {
	acquired bool
	err      error
}

func (r stubRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) == 1 {
		if p, ok := dest[0].(*bool); ok {
			*p = r.acquired
		}
	}
	return nil
}

// stubConn records every statement it is asked to run so a test can assert
// what happened, and on which connection.
type stubConn struct {
	id       string
	acquired bool
	lockErr  error
	execErr  error
	panicOn  string

	mu         sync.Mutex
	statements []string
	released   int
	// done fires once the unlock statement has been issued, which is the
	// last database call the goroutine makes.
	done chan struct{}
	// gate, when non-nil, blocks the CALL until the test closes it.
	gate chan struct{}
}

func newStubConn(id string, acquired bool) *stubConn {
	return &stubConn{id: id, acquired: acquired, done: make(chan struct{})}
}

func (c *stubConn) QueryRow(context.Context, string, ...any) pgx.Row {
	c.record(tryLockSQL)
	return stubRow{acquired: c.acquired, err: c.lockErr}
}

func (c *stubConn) Exec(ctx context.Context, sql string, _ ...any) (pgconn.CommandTag, error) {
	if c.panicOn == sql {
		c.record(sql)
		panic("boom")
	}
	if sql == refreshSQL && c.gate != nil {
		<-c.gate
	}
	if sql == refreshSQL && ctx.Err() != nil {
		// A refresh wired to the request context would land here — record
		// the cancellation so the test can fail loudly rather than silently
		// "passing" on a refresh that never ran.
		c.record("CANCELLED")
		return pgconn.CommandTag{}, ctx.Err()
	}
	c.record(sql)
	if sql == unlockSQL {
		close(c.done)
	}
	if sql == refreshSQL {
		return pgconn.CommandTag{}, c.execErr
	}
	return pgconn.CommandTag{}, nil
}

func (c *stubConn) Release() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.released++
}

func (c *stubConn) record(sql string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.statements = append(c.statements, sql)
}

func (c *stubConn) ran(sql string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, s := range c.statements {
		if s == sql {
			return true
		}
	}
	return false
}

func (c *stubConn) releaseCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.released
}

// stubSource hands out the same connection every time, and counts how often
// it was asked — the runner must take exactly one.
type stubSource struct {
	conn       *stubConn
	acquireErr error

	mu       sync.Mutex
	acquires int
}

func (s *stubSource) Acquire(context.Context) (RefreshConn, error) {
	s.mu.Lock()
	s.acquires++
	s.mu.Unlock()
	if s.acquireErr != nil {
		return nil, s.acquireErr
	}
	return s.conn, nil
}

// waitDone fails the test if the background refresh has not finished its
// unlock within a short budget.
func waitDone(t *testing.T, c *stubConn) {
	t.Helper()
	select {
	case <-c.done:
	case <-time.After(2 * time.Second):
		t.Fatal("refresh goroutine did not issue the unlock")
	}
}

func TestTriggerRejectsWhenLockIsHeldAndRunsNothing(t *testing.T) {
	conn := newStubConn("c1", false)
	runner := NewRevenueRefreshRunner(&stubSource{conn: conn})

	started, err := runner.Trigger(context.Background(), "admin-1")
	if err != nil {
		t.Fatalf("Trigger returned %v, want nil", err)
	}
	if started {
		t.Error("Trigger reported a started refresh while the lock was held")
	}
	if conn.ran(refreshSQL) {
		t.Error("the CALL must not run when the advisory lock was not acquired (BR-001)")
	}
	if conn.releaseCount() != 1 {
		t.Errorf("connection released %d times, want 1 — a rejected trigger must not leak it", conn.releaseCount())
	}
	if runner.LastRefreshed() != nil {
		t.Error("a rejected trigger must not record a refresh time")
	}
}

func TestUnlockRunsOnTheSameConnectionEvenWhenTheCallFails(t *testing.T) {
	conn := newStubConn("c1", true)
	conn.execErr = errors.New("refresh exploded")
	source := &stubSource{conn: conn}
	runner := NewRevenueRefreshRunner(source)

	started, err := runner.Trigger(context.Background(), "admin-1")
	if err != nil || !started {
		t.Fatalf("Trigger = (%v, %v), want (true, nil)", started, err)
	}
	waitDone(t, conn)

	if !conn.ran(unlockSQL) {
		t.Error("the advisory lock must be released even when the CALL errors (R1)")
	}
	if source.acquires != 1 {
		t.Errorf("acquired %d connections, want exactly 1 — lock and unlock must share one session", source.acquires)
	}
	if conn.releaseCount() != 1 {
		t.Errorf("connection released %d times, want 1", conn.releaseCount())
	}
	if runner.LastRefreshed() != nil {
		t.Error("a failed refresh must not record a success time")
	}
}

func TestRefreshOutlivesTheCallersCancelledContext(t *testing.T) {
	conn := newStubConn("c1", true)
	conn.gate = make(chan struct{})
	runner := NewRevenueRefreshRunner(&stubSource{conn: conn})

	callerCtx, cancelCaller := context.WithCancel(context.Background())
	started, err := runner.Trigger(callerCtx, "admin-1")
	if err != nil || !started {
		t.Fatalf("Trigger = (%v, %v), want (true, nil)", started, err)
	}

	// echo cancels the request context the moment the 202 is written. The
	// refresh is only correct if it survives that.
	cancelCaller()
	close(conn.gate)
	waitDone(t, conn)

	if conn.ran("CANCELLED") {
		t.Fatal("the refresh saw a cancelled context — it is wired to the request, not context.Background (R3)")
	}
	if !conn.ran(refreshSQL) {
		t.Error("the CALL never ran")
	}
	if runner.LastRefreshed() == nil {
		t.Error("a successful refresh must record its completion time")
	}
}

func TestSuccessfulRefreshReleasesLockAndRecordsTime(t *testing.T) {
	conn := newStubConn("c1", true)
	runner := NewRevenueRefreshRunner(&stubSource{conn: conn})

	if runner.LastRefreshed() != nil {
		t.Fatal("a fresh runner must report nil, not a fabricated time (L3)")
	}

	started, _ := runner.Trigger(context.Background(), "admin-1")
	if !started {
		t.Fatal("Trigger did not start a refresh")
	}
	waitDone(t, conn)

	if !conn.ran(refreshSQL) || !conn.ran(unlockSQL) {
		t.Errorf("statements = %v, want the CALL followed by the unlock", conn.statements)
	}
	if runner.LastRefreshed() == nil {
		t.Error("LastRefreshed is nil after a successful refresh")
	}
}

func TestPanicDuringRefreshIsContainedAndStillUnlocks(t *testing.T) {
	conn := newStubConn("c1", true)
	conn.panicOn = refreshSQL
	runner := NewRevenueRefreshRunner(&stubSource{conn: conn})

	started, _ := runner.Trigger(context.Background(), "admin-1")
	if !started {
		t.Fatal("Trigger did not start a refresh")
	}
	waitDone(t, conn)

	// Reaching here at all is the assertion: an unrecovered panic in the
	// goroutine would take the test binary down with it.
	if !conn.ran(unlockSQL) {
		t.Error("a panicking refresh must still release the advisory lock")
	}
	if conn.releaseCount() != 1 {
		t.Errorf("connection released %d times, want 1", conn.releaseCount())
	}
	if runner.LastRefreshed() != nil {
		t.Error("a panicking refresh must not record a success time")
	}
}

func TestTriggerReleasesConnectionWhenTheLockQueryFails(t *testing.T) {
	conn := newStubConn("c1", false)
	conn.lockErr = errors.New("connection reset")
	runner := NewRevenueRefreshRunner(&stubSource{conn: conn})

	started, err := runner.Trigger(context.Background(), "admin-1")
	if started {
		t.Error("Trigger reported a start after the lock query failed")
	}
	if err == nil {
		t.Error("a failed lock query must surface as an error, not a silent 409")
	}
	if conn.releaseCount() != 1 {
		t.Errorf("connection released %d times, want 1", conn.releaseCount())
	}
}

func TestTriggerSurfacesAcquireFailure(t *testing.T) {
	runner := NewRevenueRefreshRunner(&stubSource{acquireErr: errors.New("pool exhausted")})

	if started, err := runner.Trigger(context.Background(), "admin-1"); started || err == nil {
		t.Errorf("Trigger = (%v, %v), want (false, error)", started, err)
	}
}
