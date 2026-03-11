package live

// MultiplexTransport implements the Transport interface for a single View
// within a multiplexed WebSocket connection. The multiplexConn dispatches
// incoming events to the correct MultiplexTransport via eventCh, and
// outbound messages are funnelled through outCh to the shared write goroutine.
type MultiplexTransport struct {
	viewID  string
	eventCh chan eventMsg         // receives events from multiplexConn's read loop
	outCh   chan<- outboundMsg    // sends to multiplexConn's write goroutine
	done    chan struct{}
}

type eventMsg struct {
	name    string
	payload Payload
	err     error
}

type outboundMsg struct {
	viewID string
	msg    Message
}

// newMultiplexTransport creates a transport for a single view within a multiplex conn.
func newMultiplexTransport(viewID string, outCh chan<- outboundMsg) *MultiplexTransport {
	return &MultiplexTransport{
		viewID:  viewID,
		eventCh: make(chan eventMsg, 32),
		outCh:   outCh,
		done:    make(chan struct{}),
	}
}

// Send delivers a message by wrapping it with the view_id and sending
// it to the shared write goroutine.
func (t *MultiplexTransport) Send(msg Message) error {
	select {
	case <-t.done:
		return ErrConnClosed
	default:
	}
	select {
	case t.outCh <- outboundMsg{viewID: t.viewID, msg: msg}:
		return nil
	case <-t.done:
		return ErrConnClosed
	}
}

// Receive blocks until the next event for this view arrives.
func (t *MultiplexTransport) Receive() (string, Payload, error) {
	select {
	case ev := <-t.eventCh:
		return ev.name, ev.payload, ev.err
	case <-t.done:
		// Drain remaining
		select {
		case ev := <-t.eventCh:
			return ev.name, ev.payload, ev.err
		default:
			return "", nil, ErrConnClosed
		}
	}
}

// Close signals that this transport is done.
func (t *MultiplexTransport) Close() error {
	select {
	case <-t.done:
	default:
		close(t.done)
	}
	return nil
}
