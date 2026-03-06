package live

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	gp "github.com/tomo3110/gerbera/property"
)

type testView struct {
	Count int
}

func (v *testView) Mount(_ Params) error {
	v.Count = 0
	return nil
}

func (v *testView) Render() []g.ComponentFunc {
	return []g.ComponentFunc{
		gd.Head(gd.Title("Test")),
		gd.Body(
			gd.H1(gp.Value(fmt.Sprintf("Count: %d", v.Count))),
			gd.Button(Click("inc"), gp.Value("+")),
		),
	}
}

func (v *testView) HandleEvent(event string, payload Payload) error {
	if event == "inc" {
		v.Count++
	}
	return nil
}

func TestHandlerHTTPRender(t *testing.T) {
	h := Handler(func(_ context.Context) View { return &testView{} })

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := w.Body.String()

	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("expected DOCTYPE in response")
	}
	if !strings.Contains(body, "gerbera-session") {
		t.Error("expected gerbera-session attribute in response")
	}
	if !strings.Contains(body, "Count: 0") {
		t.Error("expected 'Count: 0' in initial render")
	}
	if !strings.Contains(body, "gerbera-click") {
		t.Error("expected gerbera-click attribute in response")
	}
	if !strings.Contains(body, "<script>") {
		t.Error("expected embedded script tag")
	}
}

func TestHandlerHTTPContentType(t *testing.T) {
	h := Handler(func(_ context.Context) View { return &testView{} })

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	ct := w.Result().Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("expected text/html content type, got %s", ct)
	}
}

func TestHandlerWSWithoutSession(t *testing.T) {
	h := Handler(func(_ context.Context) View { return &testView{} })

	req := httptest.NewRequest("GET", "/?gerbera-ws=1&session=invalid", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusGone {
		t.Errorf("expected 410 for expired/invalid session, got %d", w.Result().StatusCode)
	}
}

func TestHandlerWithLang(t *testing.T) {
	h := Handler(func(_ context.Context) View { return &testView{} }, WithLang("en"))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `lang="en"`) {
		t.Error("expected lang=en in response")
	}
}

func TestEventBindings(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(string) g.ComponentFunc
		attr     string
		event    string
	}{
		{"Click", Click, "gerbera-click", "test"},
		{"Input", Input, "gerbera-input", "test"},
		{"Change", Change, "gerbera-change", "test"},
		{"Submit", Submit, "gerbera-submit", "test"},
		{"Focus", Focus, "gerbera-focus", "test"},
		{"Blur", Blur, "gerbera-blur", "test"},
		{"Keydown", Keydown, "gerbera-keydown", "test"},
		{"Scroll", Scroll, "gerbera-scroll", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			el := &g.Element{
				TagName: "button",
				Attr:    make(g.AttrMap),
			}
			fn := tt.fn(tt.event)
			fn(el)
			if el.Attr[tt.attr] != tt.event {
				t.Errorf("expected %s=%s, got %s", tt.attr, tt.event, el.Attr[tt.attr])
			}
		})
	}
}

func TestKeyBinding(t *testing.T) {
	el := &g.Element{
		TagName: "input",
		Attr:    make(g.AttrMap),
	}
	fn := Key("Enter")
	fn(el)
	if el.Attr["gerbera-key"] != "Enter" {
		t.Errorf("expected gerbera-key=Enter, got %s", el.Attr["gerbera-key"])
	}
}

func TestThrottle(t *testing.T) {
	el := &g.Element{TagName: "div", Attr: make(g.AttrMap)}
	fn := Throttle(200)
	fn(el)
	if el.Attr["gerbera-throttle"] != "200" {
		t.Errorf("expected gerbera-throttle=200, got %s", el.Attr["gerbera-throttle"])
	}
}

func TestBuildTree(t *testing.T) {
	components := []g.ComponentFunc{
		gd.Head(gd.Title("Test")),
		gd.Body(gd.H1(gp.Value("Hello"))),
	}

	tree := buildTree("ja", "test-id", "", components)
	if tree.TagName != "html" {
		t.Errorf("expected html root, got %s", tree.TagName)
	}
	if tree.Attr["gerbera-session"] != "test-id" {
		t.Error("expected gerbera-session attribute")
	}
	if len(tree.Children) != 2 {
		t.Errorf("expected 2 children (head, body), got %d", len(tree.Children))
	}
}

func TestBuildTreeWithCSRFToken(t *testing.T) {
	components := []g.ComponentFunc{
		gd.Head(gd.Title("Test")),
		gd.Body(gd.H1(gp.Value("Hello"))),
	}

	tree := buildTree("ja", "test-id", "csrf-token-abc", components)

	// Find <head> and check that first child is the CSRF meta tag
	var head *g.Element
	for _, child := range tree.Children {
		if child.TagName == "head" {
			head = child
			break
		}
	}
	if head == nil {
		t.Fatal("expected <head> element")
	}
	if len(head.Children) == 0 {
		t.Fatal("expected children in <head>")
	}
	meta := head.Children[0]
	if meta.TagName != "meta" {
		t.Errorf("expected meta tag, got %s", meta.TagName)
	}
	if meta.Attr["name"] != "gerbera-csrf" {
		t.Error("expected meta name=gerbera-csrf")
	}
	if meta.Attr["content"] != "csrf-token-abc" {
		t.Errorf("expected csrf-token-abc, got %s", meta.Attr["content"])
	}
}

func TestCSRFTokenInResponse(t *testing.T) {
	h := Handler(func(_ context.Context) View { return &testView{} })

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `name="gerbera-csrf"`) {
		t.Error("expected gerbera-csrf meta tag in response")
	}
	if !strings.Contains(body, `content="`) {
		t.Error("expected content attribute in CSRF meta tag")
	}
}

func TestCSRFTokenDiffersFromSession(t *testing.T) {
	store := newSessionStore(5*time.Minute, nil)
	view := &testView{}
	sess := store.create(view)

	if sess.CSRFToken == "" {
		t.Error("expected non-empty CSRF token")
	}
	if sess.CSRFToken == sess.ID {
		t.Error("CSRF token should differ from session ID")
	}
}

func TestWSRejectsWithoutCSRF(t *testing.T) {
	h := Handler(func(_ context.Context) View { return &testView{} })

	// First, create a session via HTTP
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	body := w.Body.String()
	// Extract session ID from response
	idx := strings.Index(body, `gerbera-session="`)
	if idx == -1 {
		t.Fatal("could not find session ID in response")
	}
	start := idx + len(`gerbera-session="`)
	end := strings.Index(body[start:], `"`)
	sessionID := body[start : start+end]

	// Try WebSocket without CSRF token
	req2 := httptest.NewRequest("GET", "/?gerbera-ws=1&session="+sessionID, nil)
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)

	if w2.Result().StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 for missing CSRF, got %d", w2.Result().StatusCode)
	}
}

func TestWSRejectsWithWrongCSRF(t *testing.T) {
	h := Handler(func(_ context.Context) View { return &testView{} })

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	body := w.Body.String()
	idx := strings.Index(body, `gerbera-session="`)
	if idx == -1 {
		t.Fatal("could not find session ID in response")
	}
	start := idx + len(`gerbera-session="`)
	end := strings.Index(body[start:], `"`)
	sessionID := body[start : start+end]

	// Try WebSocket with wrong CSRF token
	req2 := httptest.NewRequest("GET", "/?gerbera-ws=1&session="+sessionID+"&csrf=wrong-token", nil)
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)

	if w2.Result().StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 for wrong CSRF, got %d", w2.Result().StatusCode)
	}
}

func TestUploadRejectsWithoutCSRF(t *testing.T) {
	h := Handler(func(_ context.Context) View { return &testView{} })

	// Create a session
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	body := w.Body.String()
	idx := strings.Index(body, `gerbera-session="`)
	if idx == -1 {
		t.Fatal("could not find session ID in response")
	}
	start := idx + len(`gerbera-session="`)
	end := strings.Index(body[start:], `"`)
	sessionID := body[start : start+end]

	// Try upload without CSRF token
	req2 := httptest.NewRequest("POST", "/?gerbera-upload=1&session="+sessionID+"&event=upload", nil)
	req2.Header.Set("Content-Type", "multipart/form-data; boundary=----test")
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)

	if w2.Result().StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 for missing CSRF on upload, got %d", w2.Result().StatusCode)
	}
}

func TestDefaultCheckOrigin(t *testing.T) {
	tests := []struct {
		name     string
		origin   string
		host     string
		expected bool
	}{
		{"no origin header", "", "localhost:8080", true},
		{"matching origin", "http://localhost:8080", "localhost:8080", true},
		{"mismatched origin", "http://evil.com", "localhost:8080", false},
		{"invalid origin URL", "://invalid", "localhost:8080", false},
		{"matching HTTPS", "https://example.com", "example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			r.Host = tt.host
			if tt.origin != "" {
				r.Header.Set("Origin", tt.origin)
			}
			if got := defaultCheckOrigin(r); got != tt.expected {
				t.Errorf("defaultCheckOrigin() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWithCheckOriginOption(t *testing.T) {
	called := false
	customCheck := func(r *http.Request) bool {
		called = true
		return true
	}

	cfg := &handlerConfig{}
	WithCheckOrigin(customCheck)(cfg)

	if cfg.checkOrigin == nil {
		t.Error("expected checkOrigin to be set")
	}

	r := httptest.NewRequest("GET", "/", nil)
	cfg.checkOrigin(r)
	if !called {
		t.Error("expected custom checkOrigin to be called")
	}
}

type ctxKey string

func TestHandlerPassesContextToFactory(t *testing.T) {
	var receivedCtx context.Context

	h := Handler(func(ctx context.Context) View {
		receivedCtx = ctx
		return &testView{}
	})

	ctx := context.WithValue(context.Background(), ctxKey("user"), "alice")
	req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if receivedCtx == nil {
		t.Fatal("expected factory to receive context.Context, got nil")
	}
	if receivedCtx.Value(ctxKey("user")) != "alice" {
		t.Errorf("expected context value user=alice, got %v", receivedCtx.Value(ctxKey("user")))
	}
}
