package live

import (
	"errors"
	"sync"
)

// TestTransport is an in-memory Transport for testing ViewLoop without
// a real WebSocket connection.
type TestTransport struct {
	mu       sync.Mutex
	events   chan testEvent
	Messages []Message // all messages sent by ViewLoop
	closed   bool
}

type testEvent struct {
	event   string
	payload Payload
}

// NewTestTransport creates a TestTransport ready for use with ViewLoop.
func NewTestTransport() *TestTransport {
	return &TestTransport{
		events: make(chan testEvent, 32),
	}
}

// Send records the message for later inspection.
func (t *TestTransport) Send(msg Message) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return errors.New("transport closed")
	}
	t.Messages = append(t.Messages, msg)
	return nil
}

// Receive blocks until PushEvent is called or the transport is closed.
func (t *TestTransport) Receive() (string, Payload, error) {
	ev, ok := <-t.events
	if !ok {
		return "", nil, errConnClosed
	}
	return ev.event, ev.payload, nil
}

// Close shuts down the transport, unblocking any pending Receive call.
func (t *TestTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.closed {
		t.closed = true
		close(t.events)
	}
	return nil
}

// PushEvent injects an event into the transport as if the client sent it.
func (t *TestTransport) PushEvent(event string, payload Payload) {
	t.events <- testEvent{event: event, payload: payload}
}

// LastMessage returns the most recently sent Message, or an empty Message
// if nothing has been sent yet.
func (t *TestTransport) LastMessage() Message {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.Messages) == 0 {
		return Message{}
	}
	return t.Messages[len(t.Messages)-1]
}
