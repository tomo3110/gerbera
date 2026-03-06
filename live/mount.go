package live

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/property"
)

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
