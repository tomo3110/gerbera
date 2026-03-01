package live

import (
	"time"

	"github.com/tomo3110/gerbera"
)

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

// TickerView is an optional interface that Views can implement
// to receive periodic server-side tick events.
// HandleTick is called at the interval returned by TickInterval().
type TickerView interface {
	View
	// TickInterval returns the interval between ticks.
	// Return 0 to disable ticking.
	TickInterval() time.Duration
	// HandleTick is called on each tick. Update state here and
	// the view will be re-rendered and patches sent to the client.
	HandleTick() error
}

// InfoReceiver is an optional interface that Views can implement
// to receive arbitrary messages sent via SendInfo.
type InfoReceiver interface {
	View
	// HandleInfo processes a server-side info message.
	HandleInfo(msg any) error
}

// Params holds URL query parameters passed to Mount.
type Params map[string]string

// Payload holds the event data sent from the browser.
type Payload map[string]string
