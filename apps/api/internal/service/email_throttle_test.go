package service

import (
	"sync"
	"testing"
	"time"
)

func TestEmailThrottleTripsAtLimit(t *testing.T) {
	th := NewEmailThrottle(5, time.Minute)
	defer th.Close()

	for i := 0; i < 5; i++ {
		if !th.Allow("admin@sunbooking.com") {
			t.Fatalf("attempt %d should be allowed", i+1)
		}
	}
	if th.Allow("admin@sunbooking.com") {
		t.Fatal("6th attempt within the window should be denied")
	}
}

func TestEmailThrottleIsCaseInsensitive(t *testing.T) {
	th := NewEmailThrottle(1, time.Minute)
	defer th.Close()

	if !th.Allow("Admin@SunBooking.com") {
		t.Fatal("first attempt should be allowed")
	}
	if th.Allow("admin@sunbooking.com") {
		t.Fatal("same email in a different case should share the bucket")
	}
}

func TestEmailThrottleRecoversAfterWindow(t *testing.T) {
	th := NewEmailThrottle(1, 20*time.Millisecond)
	defer th.Close()

	if !th.Allow("a@b.com") {
		t.Fatal("first attempt should be allowed")
	}
	if th.Allow("a@b.com") {
		t.Fatal("second attempt within the window should be denied")
	}

	time.Sleep(30 * time.Millisecond)
	if !th.Allow("a@b.com") {
		t.Fatal("attempt after the window elapsed should be allowed again")
	}
}

func TestEmailThrottleDifferentEmailsAreIndependent(t *testing.T) {
	th := NewEmailThrottle(1, time.Minute)
	defer th.Close()

	if !th.Allow("a@b.com") {
		t.Fatal("first email's first attempt should be allowed")
	}
	if !th.Allow("c@d.com") {
		t.Fatal("a different email must not be affected by another email's throttle")
	}
}

func TestEmailThrottleConcurrentSafe(t *testing.T) {
	th := NewEmailThrottle(1000000, time.Minute)
	defer th.Close()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			th.Allow("concurrent@test.com")
		}(i)
	}
	wg.Wait()
}
