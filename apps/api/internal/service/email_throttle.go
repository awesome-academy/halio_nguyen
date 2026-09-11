package service

import (
	"strings"
	"sync"
	"time"
)

// bucket is a fixed-window counter for one email key.
type bucket struct {
	count     int
	windowEnd time.Time
}

// EmailThrottle is the per-email limb of D3's login throttle (the per-IP
// limb is middleware.NewLoginRateLimiter — Echo's store keys off one
// extracted identifier and runs before the body is readable, so per-email
// throttling cannot live there). It is a fixed window, not a sliding one:
// simpler, and the plan's D3 accepts an in-memory, per-replica limiter.
type EmailThrottle struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	limit    int
	window   time.Duration
	maxKeys  int
	stopOnce sync.Once
	stop     chan struct{}
}

// maxTrackedEmails caps the bucket map so a dictionary attack across many
// distinct emails cannot grow it unboundedly between janitor sweeps (R5).
// Once the cap is hit, a brand-new email is allowed through (fails open on
// the email limb only — the per-IP limb still applies).
const maxTrackedEmails = 10000

// NewEmailThrottle builds a throttle allowing limit attempts per window for
// each lower-cased email, and starts a janitor goroutine that prunes
// expired buckets every window so memory does not grow unbounded.
func NewEmailThrottle(limit int, window time.Duration) *EmailThrottle {
	t := &EmailThrottle{
		buckets: make(map[string]*bucket),
		limit:   limit,
		window:  window,
		maxKeys: maxTrackedEmails,
		stop:    make(chan struct{}),
	}
	go t.runJanitor()
	return t
}

// Allow reports whether another attempt for email is permitted right now,
// consuming one attempt from the window if so.
func (t *EmailThrottle) Allow(email string) bool {
	key := strings.ToLower(email)
	now := time.Now()

	t.mu.Lock()
	defer t.mu.Unlock()

	b, ok := t.buckets[key]
	if !ok || now.After(b.windowEnd) {
		if !ok && len(t.buckets) >= t.maxKeys {
			return true
		}
		t.buckets[key] = &bucket{count: 1, windowEnd: now.Add(t.window)}
		return true
	}

	if b.count >= t.limit {
		return false
	}
	b.count++
	return true
}

// Close stops the janitor goroutine. Tests call this to avoid leaking it
// across cases; production callers may leave it running for process life.
func (t *EmailThrottle) Close() {
	t.stopOnce.Do(func() { close(t.stop) })
}

func (t *EmailThrottle) runJanitor() {
	ticker := time.NewTicker(t.window)
	defer ticker.Stop()

	for {
		select {
		case <-t.stop:
			return
		case now := <-ticker.C:
			t.prune(now)
		}
	}
}

func (t *EmailThrottle) prune(now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for key, b := range t.buckets {
		if now.After(b.windowEnd) {
			delete(t.buckets, key)
		}
	}
}
