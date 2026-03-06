package live

import (
	"github.com/tomo3110/gerbera"
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
//	func page() []g.ComponentFunc {
//	    return []g.ComponentFunc{
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
//	func adminPage() []g.ComponentFunc {
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
			for _, opt := range opts {
				opt(n)
			}
		},
	)
}
