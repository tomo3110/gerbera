package live

import (
	"fmt"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/property"
)

// Component creates a mount point for a sub-LiveView within a page.
// The component ID is used to identify the component's DOM container.
// The component will maintain its own WebSocket connection and state.
//
// Usage in Render():
//
//	gl.Component("chat-widget", "/chat")
//
// This renders a <div> with a gerbera-component attribute that the client-side
// JS will use to mount an independent LiveView.
func Component(id, path string) gerbera.ComponentFunc {
	return gerbera.Tag("div",
		property.ID("gerbera-component-"+id),
		property.Attr("gerbera-component", path),
		property.Attr("gerbera-component-id", id),
	)
}

// ComponentInline creates an inline component placeholder that loads via iframe.
// This is simpler than the WebSocket-based Component but provides full isolation.
func ComponentInline(id, path string, width, height string) gerbera.ComponentFunc {
	return gerbera.Tag("iframe",
		property.ID("gerbera-component-"+id),
		property.Attr("src", path),
		property.Attr("style", fmt.Sprintf("width:%s;height:%s;border:none;", width, height)),
		property.Attr("loading", "lazy"),
	)
}
