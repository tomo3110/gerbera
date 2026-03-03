package live

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWithDebugOption(t *testing.T) {
	cfg := &handlerConfig{
		lang:       "ja",
		sessionTTL: 5 * time.Minute,
	}
	WithDebug()(cfg)
	if !cfg.debug {
		t.Error("expected debug to be true after WithDebug()")
	}
}

func TestDebugHTTPInjectsDebugScript(t *testing.T) {
	h := Handler(func(_ context.Context) View { return &testView{} }, WithDebug())
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	body := w.Body.String()
	// gerbera-debug-host is unique to gerbera_debug.js (the DevPanel)
	if !strings.Contains(body, "gerbera-debug-host") {
		t.Error("expected debug panel script in response when debug is enabled")
	}
}

func TestNonDebugHTTPDoesNotInjectDebugScript(t *testing.T) {
	h := Handler(func(_ context.Context) View { return &testView{} })
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	body := w.Body.String()
	// gerbera-debug-host is unique to gerbera_debug.js (the DevPanel)
	if strings.Contains(body, "gerbera-debug-host") {
		t.Error("debug panel script should not be present when debug is disabled")
	}
}

func TestDebugMessageEnvelope(t *testing.T) {
	patchJSON, _ := json.Marshal([]map[string]interface{}{
		{"op": "text", "path": []int{0, 1}, "val": "hello"},
	})

	msg := debugMessage{
		Patches: patchJSON,
		Debug: &debugMeta{
			Event:      "inc",
			Payload:    Payload{"value": "1"},
			PatchCount: 1,
			DurationMS: 5,
			ViewState:  json.RawMessage(`{"Count":1}`),
			SessionID:  "abc123",
			SessionTTL: "5m0s",
			Timestamp:  time.Now().UnixMilli(),
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal("expected object envelope, got:", string(data))
	}
	if _, ok := raw["patches"]; !ok {
		t.Error("expected 'patches' key in envelope")
	}
	if _, ok := raw["debug"]; !ok {
		t.Error("expected 'debug' key in envelope")
	}
}

func TestDebugLoggerNoOpWhenDisabled(t *testing.T) {
	dlog := newDebugLogger(false)
	// These should not panic
	dlog.eventReceived("id", "test", Payload{})
	dlog.patchesGenerated("id", 5, time.Millisecond)
	dlog.sessionCreated("id")
	dlog.sessionConnected("id")
	dlog.sessionDisconnected("id")
	dlog.sessionExpired("id")
	dlog.handleError("id", "ctx", fmt.Errorf("test"))
}

func TestDebugLoggerEnabledDoesNotPanic(t *testing.T) {
	dlog := newDebugLogger(true)
	dlog.eventReceived("id", "test", Payload{"key": "val"})
	dlog.patchesGenerated("id", 3, 10*time.Millisecond)
	dlog.sessionCreated("id")
	dlog.sessionConnected("id")
	dlog.sessionDisconnected("id")
	dlog.sessionExpired("id")
	dlog.handleError("id", "ctx", fmt.Errorf("test error"))
}

func TestViewStateSerialization(t *testing.T) {
	view := &testView{Count: 42}
	data, err := json.Marshal(view)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m["Count"] != float64(42) {
		t.Errorf("expected Count=42, got %v", m["Count"])
	}
}
