package live

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSTransportOption configures a WSTransport.
type WSTransportOption func(*wsTransportConfig)

type wsTransportConfig struct {
	debug      bool
	sessionID  string
	sessionTTL time.Duration
}

// WithWSDebug enables debug envelope format in Send.
func WithWSDebug(sessionID string, sessionTTL time.Duration) WSTransportOption {
	return func(c *wsTransportConfig) {
		c.debug = true
		c.sessionID = sessionID
		c.sessionTTL = sessionTTL
	}
}

// WSTransport implements Transport over a gorilla/websocket connection.
type WSTransport struct {
	conn   *websocket.Conn
	mu     sync.Mutex   // protects conn.WriteJSON
	msgCh  chan []byte   // incoming raw messages from read goroutine
	doneCh chan struct{} // closed when read goroutine exits
	cfg    wsTransportConfig
}

// NewWSTransport creates a WSTransport and starts a background read goroutine.
func NewWSTransport(conn *websocket.Conn, opts ...WSTransportOption) *WSTransport {
	t := &WSTransport{
		conn:   conn,
		msgCh:  make(chan []byte, 32),
		doneCh: make(chan struct{}),
	}
	for _, o := range opts {
		o(&t.cfg)
	}
	go t.readLoop()
	return t
}

func (t *WSTransport) readLoop() {
	defer close(t.doneCh)
	for {
		_, msg, err := t.conn.ReadMessage()
		if err != nil {
			return
		}
		t.msgCh <- msg
	}
}

// Send serialises a Message and writes it to the WebSocket connection.
// The format mirrors the original sendPatches behaviour: debug mode uses
// a debugMessage envelope; non-debug mode sends patches as a plain array
// when there are no JS commands, or a wsMessage envelope otherwise.
func (t *WSTransport) Send(msg Message) error {
	if len(msg.Patches) == 0 && len(msg.JSCommands) == 0 && !t.cfg.debug {
		return nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cfg.debug {
		patchJSON, _ := json.Marshal(msg.Patches)
		envelope := debugMessage{
			Patches:    patchJSON,
			JSCommands: msg.JSCommands,
			Debug: &debugMeta{
				Event:      msg.EventName,
				Payload:    msg.EventPayload,
				PatchCount: len(msg.Patches),
				DurationMS: msg.Duration.Milliseconds(),
				ViewState:  msg.ViewState,
				SessionID:  t.cfg.sessionID,
				SessionTTL: t.cfg.sessionTTL.String(),
				Timestamp:  time.Now().UnixMilli(),
			},
		}
		return t.conn.WriteJSON(envelope)
	}

	if len(msg.JSCommands) == 0 {
		return t.conn.WriteJSON(msg.Patches)
	}
	return t.conn.WriteJSON(wsMessage{
		Patches:    msg.Patches,
		JSCommands: msg.JSCommands,
	})
}

// ErrConnClosed is a sentinel indicating the transport connection was closed.
// Custom Transport implementations should return this error from Receive
// when the underlying connection is terminated.
var ErrConnClosed = errors.New("connection closed")

// Receive blocks until the next client event arrives or the connection closes.
func (t *WSTransport) Receive() (string, Payload, error) {
	// Prioritise buffered messages over doneCh to avoid dropping
	// events that were already read before the connection closed.
	select {
	case raw := <-t.msgCh:
		return t.unmarshal(raw)
	default:
	}

	select {
	case raw := <-t.msgCh:
		return t.unmarshal(raw)
	case <-t.doneCh:
		// Drain remaining messages after the read goroutine exits.
		select {
		case raw := <-t.msgCh:
			return t.unmarshal(raw)
		default:
			return "", nil, ErrConnClosed
		}
	}
}

func (t *WSTransport) unmarshal(raw []byte) (string, Payload, error) {
	var ev wsEvent
	if err := json.Unmarshal(raw, &ev); err != nil {
		return "", nil, err
	}
	return ev.Name, ev.Payload, nil
}

// Close closes the underlying WebSocket connection.
func (t *WSTransport) Close() error {
	return t.conn.Close()
}
