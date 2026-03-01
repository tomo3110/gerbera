package live

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	gp "github.com/tomo3110/gerbera/property"
)

type testView struct {
	Count int
}

func (v *testView) Mount(params Params) error {
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
	h := Handler(func() View { return &testView{} })

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
	h := Handler(func() View { return &testView{} })

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	ct := w.Result().Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("expected text/html content type, got %s", ct)
	}
}

func TestHandlerWSWithoutSession(t *testing.T) {
	h := Handler(func() View { return &testView{} })

	req := httptest.NewRequest("GET", "/?gerbera-ws=1&session=invalid", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusGone {
		t.Errorf("expected 410 for expired/invalid session, got %d", w.Result().StatusCode)
	}
}

func TestHandlerWithLang(t *testing.T) {
	h := Handler(func() View { return &testView{} }, WithLang("en"))

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
			if err := fn(el); err != nil {
				t.Fatal(err)
			}
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
	if err := fn(el); err != nil {
		t.Fatal(err)
	}
	if el.Attr["gerbera-key"] != "Enter" {
		t.Errorf("expected gerbera-key=Enter, got %s", el.Attr["gerbera-key"])
	}
}

func TestThrottle(t *testing.T) {
	el := &g.Element{TagName: "div", Attr: make(g.AttrMap)}
	fn := Throttle(200)
	if err := fn(el); err != nil {
		t.Fatal(err)
	}
	if el.Attr["gerbera-throttle"] != "200" {
		t.Errorf("expected gerbera-throttle=200, got %s", el.Attr["gerbera-throttle"])
	}
}

func TestBuildTree(t *testing.T) {
	components := []g.ComponentFunc{
		gd.Head(gd.Title("Test")),
		gd.Body(gd.H1(gp.Value("Hello"))),
	}

	tree, err := buildTree("ja", "test-id", components)
	if err != nil {
		t.Fatal(err)
	}
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
