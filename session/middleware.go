package session

import (
	"context"
	"net/http"
)

type contextKey struct{}

// Middleware loads the session from the store, sets it in the request context,
// and automatically saves it before the response is written.
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
			sw := &saveWriter{
				ResponseWriter: w,
				r:              r,
				store:          store,
				sess:           sess,
			}
			next.ServeHTTP(sw, r)
			sw.save()
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

// saveWriter wraps ResponseWriter to auto-save the session before writing.
type saveWriter struct {
	http.ResponseWriter
	r     *http.Request
	store Store
	sess  *Session
	saved bool
}

func (sw *saveWriter) save() {
	if sw.saved {
		return
	}
	sw.saved = true
	sw.store.Save(sw.ResponseWriter, sw.r, sw.sess)
}

func (sw *saveWriter) WriteHeader(code int) {
	sw.save()
	sw.ResponseWriter.WriteHeader(code)
}

func (sw *saveWriter) Write(b []byte) (int, error) {
	sw.save()
	return sw.ResponseWriter.Write(b)
}
