package gerbera

import "net/http"

// NewServeMux creates an http.ServeMux that serves a single page.
//
// Deprecated: Use Handler or HandlerFunc with http.ServeMux instead.
// NewServeMux always uses lang="ja" and binds to "/", which limits flexibility.
//
//	// Before:
//	http.ListenAndServe(":8800", g.NewServeMux(components...))
//
//	// After:
//	mux := http.NewServeMux()
//	mux.Handle("GET /", g.Handler(components...))
//	http.ListenAndServe(":8800", mux)
func NewServeMux(c ...ComponentFunc) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := ExecuteTemplate(w, "ja", c...); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
		}
	})
	return mux
}

// Handler returns an http.Handler that renders static page content.
// The components are fixed at handler creation time and do not change
// per request. For dynamic content that depends on the request, use
// HandlerFunc instead.
//
// Unlike live.Handler (LiveView), Handler is stateless and completes
// in a single HTTP round-trip.
//
// Usage:
//
//	mux.Handle("GET /", g.Handler(homePage()...))
//	mux.Handle("GET /about", g.Handler(aboutPage()...))
func Handler(components ...ComponentFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := ExecuteTemplate(w, "en", components...); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

// HandlerFunc returns an http.Handler that dynamically generates page
// content based on the request.
//
// Use HandlerFunc when the page depends on path parameters, query strings,
// session data, or other request-specific information.
//
// Unlike live.Handler (LiveView), HandlerFunc is stateless and completes
// in a single HTTP round-trip.
//
// Usage:
//
//	mux.Handle("GET /users/{id}", g.HandlerFunc(func(r *http.Request) []g.ComponentFunc {
//	    id := r.PathValue("id")
//	    user, _ := userRepo.GetByID(id)
//	    return userPage(user)
//	}))
func HandlerFunc(fn func(r *http.Request) []ComponentFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		components := fn(r)
		if err := ExecuteTemplate(w, "en", components...); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}
