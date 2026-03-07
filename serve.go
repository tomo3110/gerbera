package gerbera

import (
	"net/http"
	"strings"
)

// assetHandler is set by the assets package via RegisterAssetHandler
// to avoid an import cycle (gerbera → assets → gerbera).
var assetHandler http.Handler

// RegisterAssetHandler is called by the assets package's init() to register
// the static asset handler. This avoids a circular import.
func RegisterAssetHandler(h http.Handler) {
	assetHandler = h
}

// Serve wraps an http.Handler to serve gerbera's static assets under /_gerbera/.
// Other requests are delegated to the wrapped handler.
//
// Use Serve when your application uses LiveView (LiveMount) or UI components
// that require external JS/CSS files. For pure SSR sites without LiveView,
// Serve is not needed.
//
//	http.ListenAndServe(":8800", g.Serve(mux))
func Serve(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/_gerbera/") && assetHandler != nil {
			assetHandler.ServeHTTP(w, r)
			return
		}
		handler.ServeHTTP(w, r)
	})
}
