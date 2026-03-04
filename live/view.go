package live

import (
	"net/url"
	"time"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/session"
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

// SessionExpiredHandler is an optional interface that Views can implement
// to perform cleanup before the connection is closed due to session invalidation.
// If not implemented, the connection is closed immediately.
//
// Typical uses include showing a toast notification, auto-saving drafts,
// or dispatching a client-side redirect.
type SessionExpiredHandler interface {
	// OnSessionExpired is called when the session is invalidated
	// (e.g. by logout or expiry). The view's state changes are rendered
	// and sent to the client before the connection is closed.
	OnSessionExpired() error
}

// Patcher is an optional interface for Views that synchronize state with URL parameters.
// HandleParams is called when the browser's back/forward buttons change the URL
// (popstate event). Views that use PushPatch should implement this interface
// to restore state when the user navigates through browser history.
type Patcher interface {
	HandleParams(params url.Values) error
}

// Unmounter is an optional interface that Views can implement
// to perform cleanup when the WebSocket connection is closed.
// Unmount is called automatically at the end of ViewLoop.
type Unmounter interface {
	Unmount()
}

// ConnInfo holds connection-level information available at Mount time.
type ConnInfo struct {
	Session     *session.Session
	LiveSession *Session // LiveView session (for SendInfo)
	RemoteAddr  string
	UserAgent   string
}

// Params holds URL query parameters and connection info passed to Mount.
type Params struct {
	Query url.Values
	Conn  ConnInfo
}

// Get returns the first value for the given query parameter key.
func (p Params) Get(key string) string {
	return p.Query.Get(key)
}

// Payload holds the event data sent from the browser.
type Payload map[string]string
