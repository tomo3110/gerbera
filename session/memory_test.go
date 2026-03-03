package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMemoryStore_CreateAndRetrieve(t *testing.T) {
	store := NewMemoryStore([]byte("secret"))
	defer store.Close()

	// First request creates a new session
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	sess, err := store.Get(r)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !sess.IsNew {
		t.Error("first session should be new")
	}
	sess.Set("user", "alice")
	store.Save(w, r, sess)

	// Extract cookie from response
	resp := w.Result()
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie in response")
	}

	// Second request with cookie retrieves existing session
	r2 := httptest.NewRequest("GET", "/", nil)
	r2.AddCookie(cookies[0])
	sess2, err := store.Get(r2)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if sess2.IsNew {
		t.Error("second session should not be new")
	}
	if got := sess2.GetString("user"); got != "alice" {
		t.Errorf("user = %q, want alice", got)
	}
}

func TestMemoryStore_Destroy(t *testing.T) {
	store := NewMemoryStore([]byte("secret"))
	defer store.Close()

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	sess, _ := store.Get(r)
	sess.Set("user", "alice")
	store.Save(w, r, sess)

	resp := w.Result()
	cookies := resp.Cookies()

	// Destroy session
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("GET", "/", nil)
	r2.AddCookie(cookies[0])
	store.Destroy(w2, r2, sess)

	// After destroy, should get a new session
	r3 := httptest.NewRequest("GET", "/", nil)
	r3.AddCookie(cookies[0])
	sess3, _ := store.Get(r3)
	if !sess3.IsNew {
		t.Error("session should be new after destroy")
	}
}

func TestMemoryStore_Expiry(t *testing.T) {
	store := NewMemoryStore([]byte("secret"), WithMaxAge(50*time.Millisecond))
	defer store.Close()

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	sess, _ := store.Get(r)
	sess.Set("key", "value")
	store.Save(w, r, sess)

	resp := w.Result()
	cookies := resp.Cookies()

	time.Sleep(100 * time.Millisecond)

	r2 := httptest.NewRequest("GET", "/", nil)
	r2.AddCookie(cookies[0])
	sess2, _ := store.Get(r2)
	if !sess2.IsNew {
		t.Error("session should be new after expiry")
	}
}

func TestMemoryStore_InvalidCookie(t *testing.T) {
	store := NewMemoryStore([]byte("secret"))
	defer store.Close()

	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "gerbera_session", Value: "invalid"})
	sess, err := store.Get(r)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !sess.IsNew {
		t.Error("should return new session for invalid cookie")
	}
}

func TestMemoryStore_GC(t *testing.T) {
	store := NewMemoryStore([]byte("secret"),
		WithMaxAge(50*time.Millisecond),
		WithGCInterval(50*time.Millisecond),
	)
	defer store.Close()

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	sess, _ := store.Get(r)
	store.Save(w, r, sess)

	store.mu.RLock()
	count := len(store.sessions)
	store.mu.RUnlock()
	if count != 1 {
		t.Fatalf("expected 1 session, got %d", count)
	}

	time.Sleep(200 * time.Millisecond)

	store.mu.RLock()
	count = len(store.sessions)
	store.mu.RUnlock()
	if count != 0 {
		t.Errorf("expected 0 sessions after GC, got %d", count)
	}
}

func TestMemoryStore_ImplementsBrokerStore(t *testing.T) {
	store := NewMemoryStore([]byte("secret"))
	defer store.Close()

	var _ BrokerStore = store
}

func TestMemoryStore_Broker(t *testing.T) {
	store := NewMemoryStore([]byte("secret"))
	defer store.Close()

	if store.Broker() == nil {
		t.Fatal("Broker() should not return nil")
	}
}

func TestMemoryStore_DestroyNotifiesBroker(t *testing.T) {
	store := NewMemoryStore([]byte("secret"))
	defer store.Close()

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	sess, _ := store.Get(r)
	store.Save(w, r, sess)

	// Subscribe before destroy
	ch, unsub := store.Broker().Subscribe(sess.ID)
	defer unsub()

	// Destroy should trigger notification
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("GET", "/", nil)
	store.Destroy(w2, r2, sess)

	select {
	case <-ch:
		// success
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for broker notification on Destroy")
	}
}

func TestMemoryStore_GCNotifiesBroker(t *testing.T) {
	store := NewMemoryStore([]byte("secret"),
		WithMaxAge(50*time.Millisecond),
		WithGCInterval(50*time.Millisecond),
	)
	defer store.Close()

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	sess, _ := store.Get(r)
	store.Save(w, r, sess)

	ch, unsub := store.Broker().Subscribe(sess.ID)
	defer unsub()

	select {
	case <-ch:
		// success: GC triggered broker notification
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for broker notification on GC expiry")
	}
}
