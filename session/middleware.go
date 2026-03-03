package session

import (
	"context"
	"net/http"
)

type contextKey struct{}

// Middleware loads the session from the store, sets it in the request context,
// and automatically saves it after the handler returns.
func Middleware(store Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, err := store.Get(r)
			if err != nil {
				http.Error(w, "session error", http.StatusInternalServerError)
				return
			}
			ctx := context.WithValue(r.Context(), contextKey{}, sess)
			r = r.WithContext(ctx)
			// Save session before handler to set cookie in response
			// headers before they are committed by the handler.
			store.Save(w, r, sess)
			next.ServeHTTP(w, r)
			// Save again if handler modified the session to persist changes.
			if sess.Modified() {
				store.Save(w, r, sess)
			}
		})
	}
}

// FromContext returns the Session from the context, or nil if none.
func FromContext(ctx context.Context) *Session {
	sess, _ := ctx.Value(contextKey{}).(*Session)
	return sess
}

// RequireKey returns middleware that redirects to redirectTo if the session
// does not contain the given key.
func RequireKey(key string, redirectTo string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess := FromContext(r.Context())
			if sess == nil || sess.Get(key) == nil {
				http.Redirect(w, r, redirectTo, http.StatusFound)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
