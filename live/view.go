package live

import "github.com/tomo3110/gerbera"

// View represents a LiveView component.
// Users implement this interface to define stateful, real-time pages.
type View interface {
	// Mount is called once when the session is created.
	Mount(params Params) error

	// Render returns the component tree for the current state.
	// The returned slice is mounted as children of the <html> root.
	Render() []gerbera.ComponentFunc

	// HandleEvent processes a user event sent via WebSocket.
	HandleEvent(event string, payload Payload) error
}

// Params holds URL query parameters passed to Mount.
type Params map[string]string

// Payload holds the event data sent from the browser.
type Payload map[string]string
