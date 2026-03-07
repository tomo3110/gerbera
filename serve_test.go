package gerbera

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServe_DelegatesToHandler(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("inner"))
	})

	h := Serve(inner)
	r := httptest.NewRequest("GET", "/hello", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Body.String() != "inner" {
		t.Errorf("body = %q, want inner", w.Body.String())
	}
}

func TestServe_HandlesGerberaPrefix(t *testing.T) {
	// Register a test asset handler
	oldHandler := assetHandler
	assetHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("asset"))
	})
	defer func() { assetHandler = oldHandler }()

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("inner"))
	})

	h := Serve(inner)
	r := httptest.NewRequest("GET", "/_gerbera/js/test.js", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Body.String() != "asset" {
		t.Errorf("body = %q, want asset", w.Body.String())
	}
}

func TestServe_NilAssetHandler(t *testing.T) {
	oldHandler := assetHandler
	assetHandler = nil
	defer func() { assetHandler = oldHandler }()

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("inner"))
	})

	h := Serve(inner)
	r := httptest.NewRequest("GET", "/_gerbera/js/test.js", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	// Falls through to inner when no asset handler is registered
	if w.Body.String() != "inner" {
		t.Errorf("body = %q, want inner", w.Body.String())
	}
}
