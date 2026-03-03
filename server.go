package gerbera

import "net/http"

func NewServeMux(c ...ComponentFunc) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := ExecuteTemplate(w, "ja", c...); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
		}
	})
	return mux
}

// Handler returns an http.HandlerFunc that renders static page content.
// It sets the Content-Type header and renders the components as HTML.
//
// The components are fixed at handler creation time and do not change
// per request. For dynamic content that depends on the request, use
// HandlerFunc instead.
//
// Unlike gl.Handler (LiveView), Handler is stateless and completes
// in a single HTTP round-trip.
//
// Usage:
//
//	mux.Handle("GET /about", g.Handler(aboutPage()...))
func Handler(components ...ComponentFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		ExecuteTemplate(w, "en", components...)
	}
}

// HandlerFunc returns an http.HandlerFunc that dynamically generates page
// content based on the request. It sets the Content-Type header and renders
// the components returned by fn as HTML.
//
// Use HandlerFunc when the page depends on path parameters, query strings,
// session data, or other request-specific information.
//
// Unlike gl.Handler (LiveView), HandlerFunc is stateless and completes
// in a single HTTP round-trip.
//
// Usage:
//
//	mux.Handle("GET /users/{id}", g.HandlerFunc(func(r *http.Request) []g.ComponentFunc {
//	    id := r.PathValue("id")
//	    user, _ := userRepo.GetByID(id)
//	    return userPage(user)
//	}))
func HandlerFunc(fn func(r *http.Request) []ComponentFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		components := fn(r)
		ExecuteTemplate(w, "en", components...)
	}
}
