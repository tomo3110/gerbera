package live

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	g "github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
)

// --- Test helpers ---

type testMuxView struct {
	mu       sync.Mutex
	text     string
	mounted  bool
	unmounted bool
}

func (v *testMuxView) Mount(params Params) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.mounted = true
	v.text = "hello"
	return nil
}

func (v *testMuxView) Render() g.Components {
	v.mu.Lock()
	defer v.mu.Unlock()
	return g.Components{
		dom.Body(dom.Div(func(n g.Node) { n.SetText(v.text) })),
	}
}

func (v *testMuxView) HandleEvent(event string, payload Payload) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if event == "update" {
		v.text = payload["value"]
	}
	return nil
}

func (v *testMuxView) Unmount() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.unmounted = true
}

func setupMuxServer(t *testing.T) (*httptest.Server, *ViewRegistry) {
	t.Helper()
	registry := NewViewRegistry()
	registry.Register("/live/test", func(_ context.Context) View {
		return &testMuxView{}
	})

	handler := NewMultiplexHandler(registry)
	mux := http.NewServeMux()
	mux.Handle("/_gerbera/ws", handler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server, registry
}

func dialMux(t *testing.T, server *httptest.Server) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/_gerbera/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func sendMuxMsg(t *testing.T, conn *websocket.Conn, msg MultiplexClientMessage) {
	t.Helper()
	if err := conn.WriteJSON(msg); err != nil {
		t.Fatalf("write failed: %v", err)
	}
}

func readMuxMsg(t *testing.T, conn *websocket.Conn) MultiplexServerMessage {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var msg MultiplexServerMessage
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("read failed: %v", err)
	}
	return msg
}

// --- MultiplexTransport unit tests ---

func TestMultiplexTransport_SendReceive(t *testing.T) {
	outCh := make(chan outboundMsg, 10)
	transport := newMultiplexTransport("v1", outCh)
	defer transport.Close()

	// Push an event into the transport
	transport.eventCh <- eventMsg{name: "click", payload: Payload{"id": "42"}}

	// Receive should return it
	event, payload, err := transport.Receive()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event != "click" {
		t.Errorf("expected event 'click', got %q", event)
	}
	if payload["id"] != "42" {
		t.Errorf("expected payload id '42', got %q", payload["id"])
	}
}

func TestMultiplexTransport_Send(t *testing.T) {
	outCh := make(chan outboundMsg, 10)
	transport := newMultiplexTransport("v1", outCh)
	defer transport.Close()

	msg := Message{EventName: "test"}
	if err := transport.Send(msg); err != nil {
		t.Fatalf("send failed: %v", err)
	}

	select {
	case out := <-outCh:
		if out.viewID != "v1" {
			t.Errorf("expected viewID 'v1', got %q", out.viewID)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for outbound message")
	}
}

func TestMultiplexTransport_Close(t *testing.T) {
	outCh := make(chan outboundMsg, 10)
	transport := newMultiplexTransport("v1", outCh)

	transport.Close()

	// Receive should return ErrConnClosed after Close
	_, _, err := transport.Receive()
	if err != ErrConnClosed {
		t.Errorf("expected ErrConnClosed, got %v", err)
	}

	// Send should return ErrConnClosed after Close
	err = transport.Send(Message{})
	if err != ErrConnClosed {
		t.Errorf("expected ErrConnClosed from Send, got %v", err)
	}
}

// --- multiplexConn integration tests ---

func TestMultiplexConn_MountAndUpdate(t *testing.T) {
	server, _ := setupMuxServer(t)
	conn := dialMux(t, server)

	// Mount a view
	sendMuxMsg(t, conn, MultiplexClientMessage{
		Type:   "mount",
		ViewID: "v1",
		Path:   "/live/test",
	})

	// Read mounted response
	msg := readMuxMsg(t, conn)
	if msg.Type != "mounted" {
		t.Fatalf("expected type 'mounted', got %q", msg.Type)
	}
	if msg.ViewID != "v1" {
		t.Errorf("expected viewID 'v1', got %q", msg.ViewID)
	}
	if msg.HTML == "" {
		t.Error("expected non-empty HTML in mounted response")
	}
	if msg.SessionID == "" {
		t.Error("expected non-empty session_id")
	}
	if msg.CSRF == "" {
		t.Error("expected non-empty csrf")
	}

	// Send an event
	sendMuxMsg(t, conn, MultiplexClientMessage{
		Type:    "event",
		ViewID:  "v1",
		Event:   "update",
		Payload: Payload{"value": "world"},
	})

	// Read update response
	msg = readMuxMsg(t, conn)
	if msg.Type != "update" {
		t.Fatalf("expected type 'update', got %q", msg.Type)
	}
	if msg.ViewID != "v1" {
		t.Errorf("expected viewID 'v1', got %q", msg.ViewID)
	}
	if len(msg.Patches) == 0 {
		t.Error("expected patches in update response")
	}
}

func TestMultiplexConn_MultipleViews(t *testing.T) {
	registry := NewViewRegistry()
	registry.Register("/live/a", func(_ context.Context) View { return &testMuxView{} })
	registry.Register("/live/b", func(_ context.Context) View { return &testMuxView{} })

	handler := NewMultiplexHandler(registry)
	mux := http.NewServeMux()
	mux.Handle("/_gerbera/ws", handler)
	server := httptest.NewServer(mux)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/_gerbera/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	// Mount two views
	sendMuxMsg(t, conn, MultiplexClientMessage{Type: "mount", ViewID: "v1", Path: "/live/a"})
	msg := readMuxMsg(t, conn)
	if msg.Type != "mounted" || msg.ViewID != "v1" {
		t.Fatalf("expected mounted v1, got type=%s viewID=%s", msg.Type, msg.ViewID)
	}

	sendMuxMsg(t, conn, MultiplexClientMessage{Type: "mount", ViewID: "v2", Path: "/live/b"})
	msg = readMuxMsg(t, conn)
	if msg.Type != "mounted" || msg.ViewID != "v2" {
		t.Fatalf("expected mounted v2, got type=%s viewID=%s", msg.Type, msg.ViewID)
	}

	// Send events to both
	sendMuxMsg(t, conn, MultiplexClientMessage{
		Type: "event", ViewID: "v1", Event: "update", Payload: Payload{"value": "alpha"},
	})
	msg = readMuxMsg(t, conn)
	if msg.ViewID != "v1" {
		t.Errorf("expected update for v1, got %s", msg.ViewID)
	}

	sendMuxMsg(t, conn, MultiplexClientMessage{
		Type: "event", ViewID: "v2", Event: "update", Payload: Payload{"value": "beta"},
	})
	msg = readMuxMsg(t, conn)
	if msg.ViewID != "v2" {
		t.Errorf("expected update for v2, got %s", msg.ViewID)
	}
}

func TestMultiplexConn_Unmount(t *testing.T) {
	registry := NewViewRegistry()
	viewPtr := &testMuxView{}
	registry.Register("/live/test", func(_ context.Context) View {
		return viewPtr
	})

	handler := NewMultiplexHandler(registry)
	mux := http.NewServeMux()
	mux.Handle("/_gerbera/ws", handler)
	server := httptest.NewServer(mux)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/_gerbera/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	// Mount
	sendMuxMsg(t, conn, MultiplexClientMessage{Type: "mount", ViewID: "v1", Path: "/live/test"})
	msg := readMuxMsg(t, conn)
	if msg.Type != "mounted" {
		t.Fatal("expected mounted")
	}

	// Unmount
	sendMuxMsg(t, conn, MultiplexClientMessage{Type: "unmount", ViewID: "v1"})

	// Give ViewLoop time to exit and call Unmount
	time.Sleep(200 * time.Millisecond)

	viewPtr.mu.Lock()
	unmounted := viewPtr.unmounted
	viewPtr.mu.Unlock()
	if !unmounted {
		t.Error("expected Unmount to be called")
	}
}

func TestMultiplexConn_UnknownPath(t *testing.T) {
	server, _ := setupMuxServer(t)
	conn := dialMux(t, server)

	sendMuxMsg(t, conn, MultiplexClientMessage{
		Type:   "mount",
		ViewID: "v1",
		Path:   "/live/nonexistent",
	})

	msg := readMuxMsg(t, conn)
	if msg.Type != "error" {
		t.Errorf("expected error type, got %q", msg.Type)
	}
}

func TestMultiplexConn_ConnectionClose_CleansUpViews(t *testing.T) {
	registry := NewViewRegistry()
	viewPtr := &testMuxView{}
	registry.Register("/live/test", func(_ context.Context) View {
		return viewPtr
	})

	handler := NewMultiplexHandler(registry)
	mux := http.NewServeMux()
	mux.Handle("/_gerbera/ws", handler)
	server := httptest.NewServer(mux)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/_gerbera/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}

	sendMuxMsg(t, conn, MultiplexClientMessage{Type: "mount", ViewID: "v1", Path: "/live/test"})
	_ = readMuxMsg(t, conn)

	// Close the connection
	conn.Close()

	// Give cleanup time
	time.Sleep(300 * time.Millisecond)

	viewPtr.mu.Lock()
	unmounted := viewPtr.unmounted
	viewPtr.mu.Unlock()
	if !unmounted {
		t.Error("expected Unmount to be called after connection close")
	}
}

// --- ViewRegistry tests ---

func TestViewRegistry_RegisterAndLookup(t *testing.T) {
	registry := NewViewRegistry()
	factory := func(_ context.Context) View { return &testMuxView{} }
	registry.Register("/live/test", factory)

	if registry.lookup("/live/test") == nil {
		t.Error("expected to find registered factory")
	}
	if registry.lookup("/live/other") != nil {
		t.Error("expected nil for unregistered path")
	}
}

// --- MultiplexServerMessage JSON format tests ---

func TestMultiplexServerMessage_MarshalJSON(t *testing.T) {
	msg := MultiplexServerMessage{
		Type:      "mounted",
		ViewID:    "v1",
		HTML:      "<div>hello</div>",
		SessionID: "sess123",
		CSRF:      "csrf456",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded["type"] != "mounted" {
		t.Errorf("expected type 'mounted', got %v", decoded["type"])
	}
	if decoded["view_id"] != "v1" {
		t.Errorf("expected view_id 'v1', got %v", decoded["view_id"])
	}
	if decoded["html"] != "<div>hello</div>" {
		t.Errorf("unexpected html: %v", decoded["html"])
	}
}
