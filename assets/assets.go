package assets

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/tomo3110/gerbera"
)

func init() {
	gerbera.RegisterAssetHandler(Handler())
}

//go:embed files/gerbera.js
var gerberaJS []byte

//go:embed files/gerbera_debug.js
var gerberaDebugJS []byte

//go:embed files/gerbera.css
var gerberaCSS []byte

var jsHash = computeHash(gerberaJS)
var debugJSHash = computeHash(gerberaDebugJS)
var cssHash = computeHash(gerberaCSS)

func computeHash(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h[:8])
}

// JSString returns the gerbera.js content as a string.
// Used by the live package to inline the script in full LiveView pages.
func JSString() string {
	return string(gerberaJS)
}

// DebugJSString returns the gerbera_debug.js content as a string.
// Used by the live package to inline the debug panel script.
func DebugJSString() string {
	return string(gerberaDebugJS)
}

// JSPath returns the URL path for gerbera.js with a content hash for cache busting.
func JSPath() string {
	return "/_gerbera/js/gerbera." + jsHash + ".js"
}

// DebugJSPath returns the URL path for gerbera_debug.js with a content hash for cache busting.
func DebugJSPath() string {
	return "/_gerbera/js/gerbera_debug." + debugJSHash + ".js"
}

// CSSPath returns the URL path for gerbera.css with a content hash for cache busting.
func CSSPath() string {
	return "/_gerbera/css/gerbera." + cssHash + ".css"
}

var debugHTMLProvider func() string

// RegisterDebugHTMLProvider sets the callback that provides debug panel HTML.
// Called by the live package to register its debug panel renderer.
func RegisterDebugHTMLProvider(fn func() string) {
	debugHTMLProvider = fn
}

func escapeForJSString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "</script>", `<\/script>`)
	return s
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

	mux.HandleFunc("GET "+DebugJSPath(), func(w http.ResponseWriter, r *http.Request) {
		js := string(gerberaDebugJS)
		if debugHTMLProvider != nil {
			debugHTML := debugHTMLProvider()
			escaped := escapeForJSString(debugHTML)
			js = strings.Replace(js, `/*__GERBERA_DEBUG_HTML__*/""`, `"`+escaped+`"`, 1)
		}
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "no-cache")
		w.Write([]byte(js))
	})

	return mux
}
