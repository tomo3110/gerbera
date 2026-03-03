package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware_SetsContext(t *testing.T) {
	store := NewMemoryStore([]byte("secret"))
	defer store.Close()

	var gotSession *Session
	var wasNew bool
	handler := Middleware(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSession = FromContext(r.Context())
		wasNew = gotSession != nil && gotSession.IsNew
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if gotSession == nil {
		t.Fatal("session should be set in context")
	}
	if !wasNew {
		t.Error("first request should have new session")
	}
}

func TestMiddleware_PersistsSession(t *testing.T) {
	store := NewMemoryStore([]byte("secret"))
	defer store.Close()

	handler := Middleware(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess := FromContext(r.Context())
		if sess.IsNew {
			sess.Set("count", 1)
		} else {
			count := sess.Get("count").(int)
			sess.Set("count", count+1)
		}
		w.WriteHeader(http.StatusOK)
	}))

	// First request
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	cookies := w.Result().Cookies()

	// Second request with cookie
	r2 := httptest.NewRequest("GET", "/", nil)
	for _, c := range cookies {
		r2.AddCookie(c)
	}
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, r2)

	// Verify via direct store access
	r3 := httptest.NewRequest("GET", "/", nil)
	for _, c := range cookies {
		r3.AddCookie(c)
	}
	sess, _ := store.Get(r3)
	if got := sess.Get("count"); got != 2 {
		t.Errorf("count = %v, want 2", got)
	}
}

func TestFromContext_NoSession(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	sess := FromContext(r.Context())
	if sess != nil {
		t.Error("FromContext should return nil when no session middleware")
	}
}

func TestRequireKey_Redirect(t *testing.T) {
	store := NewMemoryStore([]byte("secret"))
	defer store.Close()

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := Middleware(store)(RequireKey("user", "/login")(inner))

	r := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusFound)
	}
	if loc := w.Header().Get("Location"); loc != "/login" {
		t.Errorf("Location = %q, want /login", loc)
	}
}

func TestRequireKey_Allowed(t *testing.T) {
	store := NewMemoryStore([]byte("secret"))
	defer store.Close()

	// Set up a session with a "user" key
	handler := Middleware(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess := FromContext(r.Context())
		sess.Set("user", "alice")
		w.WriteHeader(http.StatusOK)
	}))
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	cookies := w.Result().Cookies()

	// Now test RequireKey with that session
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	protectedHandler := Middleware(store)(RequireKey("user", "/login")(inner))

	r2 := httptest.NewRequest("GET", "/protected", nil)
	for _, c := range cookies {
		r2.AddCookie(c)
	}
	w2 := httptest.NewRecorder()
	protectedHandler.ServeHTTP(w2, r2)

	if w2.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w2.Code, http.StatusOK)
	}
}
