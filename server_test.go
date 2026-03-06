package gerbera

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_ContentType(t *testing.T) {
	h := Handler(Tag("head"), Tag("body"))
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if ct := w.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/html; charset=utf-8", ct)
	}
}

func TestHandler_RendersHTML(t *testing.T) {
	h := Handler(Tag("head"), Tag("body", Tag("h1", Literal("Hello"))))
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	body := w.Body.String()
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("response should contain DOCTYPE")
	}
	if !strings.Contains(body, "<h1>Hello</h1>") {
		t.Errorf("response should contain <h1>Hello</h1>, got: %s", body)
	}
}

func TestHandler_StaticContent(t *testing.T) {
	h := Handler(Tag("head"), Tag("body", Tag("p", Literal("static"))))

	// Multiple requests should return the same content
	for i := 0; i < 3; i++ {
		r := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if !strings.Contains(w.Body.String(), "<p>static</p>") {
			t.Errorf("request %d: response should contain <p>static</p>", i)
		}
	}
}

func TestHandlerFunc_ContentType(t *testing.T) {
	h := HandlerFunc(func(r *http.Request) Components {
		return Components{Tag("head"), Tag("body")}
	})
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if ct := w.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/html; charset=utf-8", ct)
	}
}

func TestHandlerFunc_DynamicContent(t *testing.T) {
	h := HandlerFunc(func(r *http.Request) Components {
		name := r.URL.Query().Get("name")
		if name == "" {
			name = "World"
		}
		return Components{
			Tag("head"),
			Tag("body", Tag("h1", Literal("Hello, "+name))),
		}
	})

	tests := []struct {
		path string
		want string
	}{
		{"/", "Hello, World"},
		{"/?name=Alice", "Hello, Alice"},
		{"/?name=Bob", "Hello, Bob"},
	}
	for _, tt := range tests {
		r := httptest.NewRequest("GET", tt.path, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if !strings.Contains(w.Body.String(), tt.want) {
			t.Errorf("path %s: want %q in body, got: %s", tt.path, tt.want, w.Body.String())
		}
	}
}

func TestHandlerFunc_WithMiddleware(t *testing.T) {
	called := false
	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			next.ServeHTTP(w, r)
		})
	}

	h := mw(HandlerFunc(func(r *http.Request) Components {
		return Components{Tag("head"), Tag("body")}
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if !called {
		t.Error("middleware should have been called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandler_ImplementsHTTPHandler(t *testing.T) {
	var h http.Handler = Handler(Tag("head"), Tag("body"))
	_ = h
}

func TestHandlerFunc_ImplementsHTTPHandler(t *testing.T) {
	var h http.Handler = HandlerFunc(func(r *http.Request) Components {
		return nil
	})
	_ = h
}

func TestHandler_UsesLangEn(t *testing.T) {
	h := Handler(Tag("head"), Tag("body"))
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	body := w.Body.String()
	if !strings.Contains(body, `lang="en"`) {
		t.Errorf("Handler should use lang=en, got: %s", body)
	}
}

func TestHandlerFunc_UsesLangEn(t *testing.T) {
	h := HandlerFunc(func(r *http.Request) Components {
		return Components{Tag("head"), Tag("body")}
	})
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	body := w.Body.String()
	if !strings.Contains(body, `lang="en"`) {
		t.Errorf("HandlerFunc should use lang=en, got: %s", body)
	}
}

func TestHandler_WithServeMux(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("GET /about", Handler(
		Tag("head", Tag("title", Literal("About"))),
		Tag("body", Tag("h1", Literal("About Page"))),
	))
	mux.Handle("GET /contact", Handler(
		Tag("head", Tag("title", Literal("Contact"))),
		Tag("body", Tag("h1", Literal("Contact Page"))),
	))

	// Test /about
	r := httptest.NewRequest("GET", "/about", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if !strings.Contains(w.Body.String(), "About Page") {
		t.Errorf("/about should render About Page, got: %s", w.Body.String())
	}

	// Test /contact
	r = httptest.NewRequest("GET", "/contact", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if !strings.Contains(w.Body.String(), "Contact Page") {
		t.Errorf("/contact should render Contact Page, got: %s", w.Body.String())
	}
}
