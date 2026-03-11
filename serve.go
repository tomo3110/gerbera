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

// serveConfig holds options for Serve.
type serveConfig struct {
	multiplexHandler http.Handler
}

// ServeOption configures the Serve function.
type ServeOption func(*serveConfig)

// WithMultiplex adds a WebSocket multiplexing endpoint at /_gerbera/ws.
// Pass a MultiplexHandler created by live.NewMultiplexHandler.
func WithMultiplex(h http.Handler) ServeOption {
	return func(c *serveConfig) {
		c.multiplexHandler = h
	}
}

// Serve wraps an http.Handler to serve gerbera's static assets under /_gerbera/.
// Other requests are delegated to the wrapped handler.
//
// Use Serve when your application uses LiveView (LiveMount) or UI components
// that require external JS/CSS files. For pure SSR sites without LiveView,
// Serve is not needed.
//
//	http.ListenAndServe(":8800", g.Serve(mux))
//	http.ListenAndServe(":8800", g.Serve(mux, g.WithMultiplex(muxHandler)))
func Serve(handler http.Handler, opts ...ServeOption) http.Handler {
	var cfg serveConfig
	for _, o := range opts {
		o(&cfg)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/_gerbera/ws" && cfg.multiplexHandler != nil {
			cfg.multiplexHandler.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/_gerbera/") && assetHandler != nil {
			assetHandler.ServeHTTP(w, r)
			return
		}
		handler.ServeHTTP(w, r)
	})
}
