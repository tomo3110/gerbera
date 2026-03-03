package live

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tomo3110/gerbera/diff"
)

func setupWSPair(t *testing.T) (*websocket.Conn, *websocket.Conn) {
	t.Helper()
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	serverConn := make(chan *websocket.Conn, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade: %v", err)
		}
		serverConn <- c
	}))
	t.Cleanup(srv.Close)

	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { client.Close() })

	server := <-serverConn
	t.Cleanup(func() { server.Close() })

	return server, client
}

func TestWSTransportSendNonDebug(t *testing.T) {
	server, client := setupWSPair(t)

	tr := NewWSTransport(server)
	defer tr.Close()

	patches := []diff.Patch{
		{Op: diff.OpReplace, Path: []int{0, 1, 0}, Value: "Count: 1"},
	}

	if err := tr.Send(Message{Patches: patches}); err != nil {
		t.Fatalf("Send: %v", err)
	}

	// Read from client side
	_, raw, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}

	// Non-debug without JS commands sends patches as a plain array
	var got []diff.Patch
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(got))
	}
	if got[0].Value != "Count: 1" {
		t.Errorf("expected patch value 'Count: 1', got %q", got[0].Value)
	}
}

func TestWSTransportSendWithJSCommands(t *testing.T) {
	server, client := setupWSPair(t)

	tr := NewWSTransport(server)
	defer tr.Close()

	patches := []diff.Patch{
		{Op: diff.OpReplace, Path: []int{0}, Value: "x"},
	}
	cmds := []jsCommand{
		{Cmd: "focus", Target: "#input"},
	}

	if err := tr.Send(Message{Patches: patches, JSCommands: cmds}); err != nil {
		t.Fatalf("Send: %v", err)
	}

	_, raw, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}

	var got wsMessage
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(got.JSCommands) != 1 {
		t.Fatalf("expected 1 JS command, got %d", len(got.JSCommands))
	}
	if got.JSCommands[0].Cmd != "focus" {
		t.Errorf("expected cmd 'focus', got %q", got.JSCommands[0].Cmd)
	}
}

func TestWSTransportSendDebug(t *testing.T) {
	server, client := setupWSPair(t)

	tr := NewWSTransport(server, WithWSDebug("sess-1", 5*time.Minute))
	defer tr.Close()

	patches := []diff.Patch{
		{Op: diff.OpReplace, Path: []int{0}, Value: "v"},
	}

	if err := tr.Send(Message{
		Patches:   patches,
		EventName: "click",
		Duration:  42 * time.Millisecond,
	}); err != nil {
		t.Fatalf("Send: %v", err)
	}

	_, raw, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}

	var got debugMessage
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Debug == nil {
		t.Fatal("expected debug metadata")
	}
	if got.Debug.Event != "click" {
		t.Errorf("expected event 'click', got %q", got.Debug.Event)
	}
	if got.Debug.SessionID != "sess-1" {
		t.Errorf("expected sessionId 'sess-1', got %q", got.Debug.SessionID)
	}
}

func TestWSTransportSendSkipsEmpty(t *testing.T) {
	server, _ := setupWSPair(t)

	tr := NewWSTransport(server)
	defer tr.Close()

	// Empty message with no debug — should be a no-op
	if err := tr.Send(Message{}); err != nil {
		t.Fatalf("Send: %v", err)
	}
}

func TestWSTransportReceive(t *testing.T) {
	server, client := setupWSPair(t)

	tr := NewWSTransport(server)
	defer tr.Close()

	// Send a wsEvent from the client side
	ev := wsEvent{Name: "inc", Payload: Payload{"value": "1"}}
	if err := client.WriteJSON(ev); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	name, payload, err := tr.Receive()
	if err != nil {
		t.Fatalf("Receive: %v", err)
	}
	if name != "inc" {
		t.Errorf("expected event 'inc', got %q", name)
	}
	if payload["value"] != "1" {
		t.Errorf("expected payload value '1', got %q", payload["value"])
	}
}

func TestWSTransportReceiveOnClose(t *testing.T) {
	server, client := setupWSPair(t)

	tr := NewWSTransport(server)

	// Close the client to trigger connection closure
	client.Close()

	_, _, err := tr.Receive()
	if err == nil {
		t.Error("expected error on closed connection")
	}

	tr.Close()
}
