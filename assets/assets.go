package assets

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
	"net/http"

	"github.com/tomo3110/gerbera"
)

func init() {
	gerbera.RegisterAssetHandler(Handler())
}

//go:embed files/gerbera.js
var gerberaJS []byte

//go:embed files/gerbera.css
var gerberaCSS []byte

var jsHash = computeHash(gerberaJS)
var cssHash = computeHash(gerberaCSS)

func computeHash(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h[:8])
}

// JSPath returns the URL path for gerbera.js with a content hash for cache busting.
func JSPath() string {
	return "/_gerbera/js/gerbera." + jsHash + ".js"
}

// CSSPath returns the URL path for gerbera.css with a content hash for cache busting.
func CSSPath() string {
	return "/_gerbera/css/gerbera." + cssHash + ".css"
}

// Handler returns an http.Handler that serves gerbera's static assets.
// File names contain content hashes, enabling immutable caching.
func Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET "+JSPath(), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.Write(gerberaJS)
	})

	mux.HandleFunc("GET "+CSSPath(), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.Write(gerberaCSS)
	})

	return mux
}
