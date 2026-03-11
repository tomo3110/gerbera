package live

import (
	"net/url"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/assets"
	"github.com/tomo3110/gerbera/property"
)

// Script returns a ComponentFunc that injects the Gerbera client-side
// JavaScript required for LiveMount to work.
//
// When using LiveMount within an SSR page rendered by g.Handler or
// g.HandlerFunc, you must include Script() in the page's <body> so
// that the client can detect [gerbera-live] elements and establish
// WebSocket connections.
//
// Script() is not needed when the entire page is served by gl.Handler
// (full LiveView), because gl.Handler injects the script automatically.
//
// Usage:
//
//	func page() g.Components {
//	    return g.Components{
//	        gd.Head(gd.Title("My Page")),
//	        gd.Body(
//	            gl.LiveMount("/live/widget"),
//	            gl.Script(),
//	        ),
//	    }
//	}
func Script() gerbera.ComponentFunc {
	return func(parent gerbera.Node) {
		child := parent.AppendElement("script")
		child.SetText(gerberaJS)
	}
}

// MountOption configures a LiveMount element.
type MountOption func(gerbera.Node)

// WithMountID sets a custom ID on the mount point element.
// By default, an auto-generated ID is used.
func WithMountID(id string) MountOption {
	return func(n gerbera.Node) {
		n.SetAttribute("id", id)
	}
}

// WithMountClass adds a CSS class to the mount point element.
func WithMountClass(class string) MountOption {
	return func(n gerbera.Node) {
		n.AddClass(class)
	}
}

// MultiplexAttr returns a ComponentFunc that marks the page for WebSocket
// multiplexing by setting the gerbera-multiplex attribute on the root element.
// Include this in your SSR page components when using g.Serve with g.WithMultiplex.
//
// Usage:
//
//	g.Handler(gl.MultiplexAttr(), gd.Head(...), gd.Body(...))
func MultiplexAttr() gerbera.ComponentFunc {
	return func(parent gerbera.Node) {
		parent.SetAttribute("gerbera-multiplex", "/_gerbera/ws")
	}
}

// WithIndependentConnection marks this LiveMount to always use its own
// independent WebSocket connection, even when the page uses multiplexing.
// This is useful for components that need a dedicated connection (e.g. for
// reliability or isolation from other components on the page).
func WithIndependentConnection() MountOption {
	return func(n gerbera.Node) {
		n.SetAttribute("gerbera-multiplex", "false")
	}
}

// LiveMount creates a mount point for embedding a LiveView within an SSR page.
//
// When the client-side JavaScript detects a [gerbera-live] element, it
// establishes a WebSocket connection to the specified path and renders
// the LiveView inside the element.
//
// Each LiveMount creates an independent WebSocket connection with its own
// View instance and state. Multiple LiveMounts can coexist on the same page.
//
// Usage in an SSR layout:
//
//	func adminPage() g.Components {
//	    return layout("Admin",
//	        gd.Header(gl.LiveMount("/admin/notifications")),
//	        gd.Main(gl.LiveMount("/admin/orders")),
//	    )
//	}
//
//	// Route setup:
//	mux.Handle("GET /admin", g.Handler(adminPage()...))
//	mux.Handle("/admin/notifications", live.Handler(notifFactory))
//	mux.Handle("/admin/orders", live.Handler(orderFactory))
func LiveMount(path string, opts ...MountOption) gerbera.ComponentFunc {
	return gerbera.Tag("div",
		property.Attr("gerbera-live", path),
		property.Attr("gerbera-live-id", generateID()),
		func(n gerbera.Node) {
			// Register gerbera.js for automatic injection (deduplicated)
			jsURL, _ := url.Parse(assets.JSPath())
			assets.RequireScript(n, jsURL)

			for _, opt := range opts {
				opt(n)
			}
		},
	)
}
