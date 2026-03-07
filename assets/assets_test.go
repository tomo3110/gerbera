package assets

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	g "github.com/tomo3110/gerbera"
)

func TestJSPath_ContainsHash(t *testing.T) {
	path := JSPath()
	if !strings.HasPrefix(path, "/_gerbera/js/gerbera.") {
		t.Errorf("JSPath() = %q, want prefix /_gerbera/js/gerbera.", path)
	}
	if !strings.HasSuffix(path, ".js") {
		t.Errorf("JSPath() = %q, want suffix .js", path)
	}
}

func TestCSSPath_ContainsHash(t *testing.T) {
	path := CSSPath()
	if !strings.HasPrefix(path, "/_gerbera/css/gerbera.") {
		t.Errorf("CSSPath() = %q, want prefix /_gerbera/css/gerbera.", path)
	}
	if !strings.HasSuffix(path, ".css") {
		t.Errorf("CSSPath() = %q, want suffix .css", path)
	}
}

func TestHandler_ServesJS(t *testing.T) {
	h := Handler()
	r := httptest.NewRequest("GET", JSPath(), nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/javascript" {
		t.Errorf("Content-Type = %q, want application/javascript", ct)
	}
	if cc := w.Header().Get("Cache-Control"); !strings.Contains(cc, "immutable") {
		t.Errorf("Cache-Control = %q, want immutable", cc)
	}
	if w.Body.Len() == 0 {
		t.Error("response body should not be empty")
	}
}

func TestHandler_ServesCSS(t *testing.T) {
	h := Handler()
	r := httptest.NewRequest("GET", CSSPath(), nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/css" {
		t.Errorf("Content-Type = %q, want text/css", ct)
	}
	if cc := w.Header().Get("Cache-Control"); !strings.Contains(cc, "immutable") {
		t.Errorf("Cache-Control = %q, want immutable", cc)
	}
}

func TestRequireScript_Deduplication(t *testing.T) {
	root := &g.Element{TagName: "html"}
	root.SetMeta("_init", true) // ensure meta is initialized

	path1, _ := url.Parse("/static/app.js")
	path2, _ := url.Parse("/static/app.js")
	path3, _ := url.Parse("/static/other.js")

	RequireScript(root, path1)
	RequireScript(root, path2) // duplicate
	RequireScript(root, path3)

	scripts := Scripts(root)
	if len(scripts) != 2 {
		t.Errorf("Scripts() returned %d items, want 2", len(scripts))
	}
}

func TestRequireStyleSheet_Deduplication(t *testing.T) {
	root := &g.Element{TagName: "html"}

	path1, _ := url.Parse("/static/style.css")
	path2, _ := url.Parse("/static/style.css")

	RequireStyleSheet(root, path1)
	RequireStyleSheet(root, path2) // duplicate

	sheets := StyleSheets(root)
	if len(sheets) != 1 {
		t.Errorf("StyleSheets() returned %d items, want 1", len(sheets))
	}
}

func TestRequireScript_SharedAcrossTree(t *testing.T) {
	root := &g.Element{TagName: "html"}
	child := root.AppendElement("body")

	path, _ := url.Parse("/static/app.js")
	RequireScript(child, path)

	// Should be visible from root
	scripts := Scripts(root)
	if len(scripts) != 1 {
		t.Errorf("Scripts() from root returned %d items, want 1", len(scripts))
	}
}
