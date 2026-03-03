package session

import (
	"sync"
	"testing"
	"time"
)

func TestBroker_SubscribeAndInvalidate(t *testing.T) {
	b := NewBroker()
	ch, unsub := b.Subscribe("sess1")
	defer unsub()

	go func() {
		time.Sleep(10 * time.Millisecond)
		b.Invalidate("sess1")
	}()

	select {
	case <-ch:
		// success
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for invalidation")
	}
}

func TestBroker_InvalidateUnknownSession(t *testing.T) {
	b := NewBroker()
	// Should not panic
	b.Invalidate("nonexistent")
}

func TestBroker_MultipleSubscribers(t *testing.T) {
	b := NewBroker()
	ch1, unsub1 := b.Subscribe("sess1")
	defer unsub1()
	ch2, unsub2 := b.Subscribe("sess1")
	defer unsub2()

	b.Invalidate("sess1")

	for i, ch := range []<-chan struct{}{ch1, ch2} {
		select {
		case <-ch:
			// success
		case <-time.After(time.Second):
			t.Fatalf("subscriber %d: timed out waiting for invalidation", i)
		}
	}
}

func TestBroker_Unsubscribe(t *testing.T) {
	b := NewBroker()
	ch, unsub := b.Subscribe("sess1")
	unsub()

	b.Invalidate("sess1")

	select {
	case <-ch:
		t.Fatal("should not receive after unsubscribe")
	case <-time.After(50 * time.Millisecond):
		// success: no notification
	}
}

func TestBroker_UnsubscribeCleansUp(t *testing.T) {
	b := NewBroker()
	_, unsub := b.Subscribe("sess1")
	unsub()

	b.mu.RLock()
	_, exists := b.subscribers["sess1"]
	b.mu.RUnlock()
	if exists {
		t.Error("subscriber map entry should be removed after last unsubscribe")
	}
}

func TestBroker_DifferentSessions(t *testing.T) {
	b := NewBroker()
	ch1, unsub1 := b.Subscribe("sess1")
	defer unsub1()
	ch2, unsub2 := b.Subscribe("sess2")
	defer unsub2()

	b.Invalidate("sess1")

	select {
	case <-ch1:
		// success
	case <-time.After(time.Second):
		t.Fatal("sess1 subscriber should have been notified")
	}

	select {
	case <-ch2:
		t.Fatal("sess2 subscriber should not have been notified")
	case <-time.After(50 * time.Millisecond):
		// success
	}
}

func TestBroker_BufferedChannel(t *testing.T) {
	b := NewBroker()
	ch, unsub := b.Subscribe("sess1")
	defer unsub()

	// Multiple invalidations should not block the sender
	b.Invalidate("sess1")
	b.Invalidate("sess1")
	b.Invalidate("sess1")

	select {
	case <-ch:
		// success: received at least one
	case <-time.After(time.Second):
		t.Fatal("should have received notification")
	}
}

func TestBroker_ConcurrentAccess(t *testing.T) {
	b := NewBroker()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, unsub := b.Subscribe("sess1")
			defer unsub()
			b.Invalidate("sess1")
		}(i)
	}
	wg.Wait()
}
