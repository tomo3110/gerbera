package live

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/tomo3110/gerbera/diff"
)

// Message is the unit of data that a Transport sends to the client.
// It bundles DOM patches, JS commands, and optional debug metadata.
type Message struct {
	Patches      []diff.Patch
	JSCommands   []JSCommand
	ViewState    json.RawMessage // debug panel (nil when debug is off)
	EventName    string          // debug panel
	EventPayload Payload         // debug panel
	Duration     time.Duration   // debug panel
}

// Transport abstracts the delivery of DOM patches and reception of
// browser events, decoupling the View lifecycle loop from a specific
// wire protocol (WebSocket, IPC, SSE, tests, etc.).
type Transport interface {
	// Send delivers a message (patches + JS commands + debug info) to the client.
	Send(msg Message) error

	// Receive blocks until the next client event arrives.
	// It returns the event name, its payload, and any error.
	// A non-nil error signals that the connection is closed.
	Receive() (event string, payload Payload, err error)

	// Close shuts down the transport connection.
	Close() error
}

// ErrSessionExpired is returned by ViewLoop when the HTTP session is
// invalidated (e.g. logout) while a LiveView connection is active.
var ErrSessionExpired = errors.New("session expired")
