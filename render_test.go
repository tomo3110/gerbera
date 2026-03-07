package gerbera

import (
	"bytes"
	"strings"
	"testing"
)

func TestRender_ScriptAutoInjection(t *testing.T) {
	root := &Element{
		TagName:    "html",
		Attr:       AttrMap{"lang": "en"},
		ChildElems: []*Element{},
	}
	root.ensureMeta()

	head := &Element{TagName: "head", meta: root.meta}
	body := &Element{TagName: "body", meta: root.meta}
	root.ChildElems = append(root.ChildElems, head, body)

	// Register a script via meta
	scripts := map[string]bool{"/_gerbera/js/gerbera.abc123.js": true}
	root.SetMeta(metaKeyScripts, scripts)

	var buf bytes.Buffer
	if err := Render(&buf, root); err != nil {
		t.Fatal(err)
	}

	html := buf.String()
	if !strings.Contains(html, `<script src="/_gerbera/js/gerbera.abc123.js" defer></script>`) {
		t.Errorf("expected script tag in output, got:\n%s", html)
	}
}

func TestRender_StyleSheetAutoInjection(t *testing.T) {
	root := &Element{
		TagName:    "html",
		Attr:       AttrMap{"lang": "en"},
		ChildElems: []*Element{},
	}
	root.ensureMeta()

	head := &Element{TagName: "head", meta: root.meta}
	body := &Element{TagName: "body", meta: root.meta}
	root.ChildElems = append(root.ChildElems, head, body)

	styles := map[string]bool{"/_gerbera/css/gerbera.def456.css": true}
	root.SetMeta(metaKeyStyleSheets, styles)

	var buf bytes.Buffer
	if err := Render(&buf, root); err != nil {
		t.Fatal(err)
	}

	html := buf.String()
	if !strings.Contains(html, `<link rel="stylesheet" href="/_gerbera/css/gerbera.def456.css">`) {
		t.Errorf("expected link tag in output, got:\n%s", html)
	}
}

func TestRender_NoInjectionWithoutMeta(t *testing.T) {
	root := &Element{
		TagName:    "html",
		Attr:       AttrMap{"lang": "en"},
		ChildElems: []*Element{},
	}

	head := &Element{TagName: "head"}
	body := &Element{TagName: "body"}
	root.ChildElems = append(root.ChildElems, head, body)

	var buf bytes.Buffer
	if err := Render(&buf, root); err != nil {
		t.Fatal(err)
	}

	html := buf.String()
	if strings.Contains(html, "<script") {
		t.Errorf("should not contain script tag, got:\n%s", html)
	}
	if strings.Contains(html, "<link") {
		t.Errorf("should not contain link tag, got:\n%s", html)
	}
}

func TestRender_ScriptBeforeBodyClose(t *testing.T) {
	root := &Element{
		TagName:    "html",
		Attr:       AttrMap{"lang": "en"},
		ChildElems: []*Element{},
	}
	root.ensureMeta()

	head := &Element{TagName: "head", meta: root.meta}
	body := &Element{TagName: "body", meta: root.meta, ChildElems: []*Element{
		{TagName: "h1", Value: "Hello"},
	}}
	root.ChildElems = append(root.ChildElems, head, body)

	scripts := map[string]bool{"/app.js": true}
	root.SetMeta(metaKeyScripts, scripts)

	var buf bytes.Buffer
	if err := Render(&buf, root); err != nil {
		t.Fatal(err)
	}

	html := buf.String()
	// Script should appear before </body>
	scriptIdx := strings.Index(html, `<script src="/app.js"`)
	bodyCloseIdx := strings.Index(html, "</body>")
	if scriptIdx == -1 || bodyCloseIdx == -1 {
		t.Fatalf("missing script or body close tag, got:\n%s", html)
	}
	if scriptIdx > bodyCloseIdx {
		t.Error("script tag should appear before </body>")
	}
}

func TestRender_StyleSheetBeforeHeadClose(t *testing.T) {
	root := &Element{
		TagName:    "html",
		Attr:       AttrMap{"lang": "en"},
		ChildElems: []*Element{},
	}
	root.ensureMeta()

	head := &Element{TagName: "head", meta: root.meta, ChildElems: []*Element{
		{TagName: "title", Value: "Test"},
	}}
	body := &Element{TagName: "body", meta: root.meta}
	root.ChildElems = append(root.ChildElems, head, body)

	styles := map[string]bool{"/style.css": true}
	root.SetMeta(metaKeyStyleSheets, styles)

	var buf bytes.Buffer
	if err := Render(&buf, root); err != nil {
		t.Fatal(err)
	}

	html := buf.String()
	linkIdx := strings.Index(html, `<link rel="stylesheet"`)
	headCloseIdx := strings.Index(html, "</head>")
	if linkIdx == -1 || headCloseIdx == -1 {
		t.Fatalf("missing link or head close tag, got:\n%s", html)
	}
	if linkIdx > headCloseIdx {
		t.Error("link tag should appear before </head>")
	}
}
