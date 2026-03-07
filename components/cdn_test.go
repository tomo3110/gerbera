package components

import (
	"bytes"
	"testing"

	"github.com/tomo3110/gerbera"
)

// buildHead creates a head element, applies the given ComponentFuncs, and returns it.
func buildHead(t *testing.T, children ...gerbera.ComponentFunc) *gerbera.Element {
	t.Helper()
	head := &gerbera.Element{TagName: "head"}
	for _, cf := range children {
		cf(head)
	}
	return head
}

// hasChild returns true if parent has a child matching tagName with all specified attrs.
func hasChild(parent *gerbera.Element, tagName string, attrs map[string]string) bool {
	for _, child := range parent.ChildElems {
		if child.TagName != tagName {
			continue
		}
		match := true
		for k, v := range attrs {
			if child.Attr[k] != v {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// countChildren counts children matching tagName and the given attrs.
func countChildren(parent *gerbera.Element, tagName string, attrs map[string]string) int {
	count := 0
	for _, child := range parent.ChildElems {
		if child.TagName != tagName {
			continue
		}
		match := true
		for k, v := range attrs {
			if child.Attr[k] != v {
				match = false
				break
			}
		}
		if match {
			count++
		}
	}
	return count
}

func TestBootstrapCSS_AddsElements(t *testing.T) {
	head := buildHead(t, BootstrapCSS())

	// charset meta
	if !hasChild(head, "meta", map[string]string{"charset": "utf-8"}) {
		t.Error("missing charset meta")
	}
	// viewport meta
	if !hasChild(head, "meta", map[string]string{"name": "viewport"}) {
		t.Error("missing viewport meta")
	}
	// CSS link
	if !hasChild(head, "link", map[string]string{
		"rel":  "stylesheet",
		"href": "https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css",
	}) {
		t.Error("missing Bootstrap CSS link")
	}
	// JS script
	if !hasChild(head, "script", map[string]string{
		"src": "https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/js/bootstrap.bundle.min.js",
	}) {
		t.Error("missing Bootstrap JS script")
	}
	// CSS link should have integrity
	if !hasChild(head, "link", map[string]string{
		"integrity": "sha384-QWTKZyjpPEjISv5WaRU9OFeRpok6YctnYmDr5pNlyT2bRjXh0JMhjY6hW+ALEwIH",
	}) {
		t.Error("missing CSS integrity hash")
	}
	// JS script should have integrity
	if !hasChild(head, "script", map[string]string{
		"integrity": "sha384-YvpcrYf0tY3lHB60NNkmXc5s9fDVZLESaAA55NDzOxhy9GkcIdslK1eN7N6jIeHz",
	}) {
		t.Error("missing JS integrity hash")
	}
}

func TestMaterializeCSS_AddsElements(t *testing.T) {
	head := buildHead(t, MaterializeCSS())

	if !hasChild(head, "meta", map[string]string{"charset": "utf-8"}) {
		t.Error("missing charset meta")
	}
	if !hasChild(head, "meta", map[string]string{"name": "viewport"}) {
		t.Error("missing viewport meta")
	}
	if !hasChild(head, "link", map[string]string{
		"href": "https://cdnjs.cloudflare.com/ajax/libs/materialize/1.0.0/css/materialize.min.css",
	}) {
		t.Error("missing Materialize CSS link")
	}
	if !hasChild(head, "script", map[string]string{
		"src": "https://cdnjs.cloudflare.com/ajax/libs/materialize/1.0.0/js/materialize.min.js",
	}) {
		t.Error("missing Materialize JS script")
	}
}

func TestTailwindCSS_AddsElements(t *testing.T) {
	head := buildHead(t, TailwindCSS())

	if !hasChild(head, "meta", map[string]string{"charset": "utf-8"}) {
		t.Error("missing charset meta")
	}
	if !hasChild(head, "meta", map[string]string{"name": "viewport"}) {
		t.Error("missing viewport meta")
	}
	if !hasChild(head, "script", map[string]string{
		"src": "https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4",
	}) {
		t.Error("missing Tailwind CSS script")
	}
}

func TestEnsureDefaults_NoDoubleCharset(t *testing.T) {
	// Pre-add a charset meta before calling BootstrapCSS
	charsetMeta := &gerbera.Element{
		TagName: "meta",
		Attr:    gerbera.AttrMap{"charset": "utf-8"},
	}
	head := &gerbera.Element{
		TagName:  "head",
		ChildElems: []*gerbera.Element{charsetMeta},
	}
	BootstrapCSS()(head)
	count := countChildren(head, "meta", map[string]string{"charset": "utf-8"})
	if count != 1 {
		t.Errorf("charset meta count: got %d, want 1", count)
	}
}

func TestEnsureDefaults_NoDoubleViewport(t *testing.T) {
	// Pre-add a viewport meta before calling MaterializeCSS
	viewportMeta := &gerbera.Element{
		TagName: "meta",
		Attr: gerbera.AttrMap{
			"name":    "viewport",
			"content": "width=device-width, initial-scale=1",
		},
	}
	head := &gerbera.Element{
		TagName:  "head",
		ChildElems: []*gerbera.Element{viewportMeta},
	}
	MaterializeCSS()(head)
	count := countChildren(head, "meta", map[string]string{"name": "viewport"})
	if count != 1 {
		t.Errorf("viewport meta count: got %d, want 1", count)
	}
}

func TestEnsureDefaults_BothPresent(t *testing.T) {
	charsetMeta := &gerbera.Element{
		TagName: "meta",
		Attr:    gerbera.AttrMap{"charset": "utf-8"},
	}
	viewportMeta := &gerbera.Element{
		TagName: "meta",
		Attr: gerbera.AttrMap{
			"name":    "viewport",
			"content": "width=device-width, initial-scale=1",
		},
	}
	head := &gerbera.Element{
		TagName:  "head",
		ChildElems: []*gerbera.Element{charsetMeta, viewportMeta},
	}
	before := len(head.ChildElems)
	TailwindCSS()(head)
	// Should only have added the Tailwind script (no new meta tags)
	added := len(head.ChildElems) - before
	if added != 1 {
		t.Errorf("expected 1 new child (script), got %d", added)
	}
}

func TestBootstrapCSS_CharsetPrepended(t *testing.T) {
	// title element first, then BootstrapCSS
	title := &gerbera.Element{TagName: "title", Value: "Test"}
	head := &gerbera.Element{
		TagName:  "head",
		ChildElems: []*gerbera.Element{title},
	}
	BootstrapCSS()(head)
	// charset should be the first child
	first := head.ChildElems[0]
	if first.TagName != "meta" || first.Attr["charset"] != "utf-8" {
		t.Errorf("first child should be charset meta, got %s with attr %v", first.TagName, first.Attr)
	}
}

func TestBootstrapCSS_CustomVersion(t *testing.T) {
	head := buildHead(t, BootstrapCSS("5.2.0"))

	// CSS link with custom version
	if !hasChild(head, "link", map[string]string{
		"rel":  "stylesheet",
		"href": "https://cdn.jsdelivr.net/npm/bootstrap@5.2.0/dist/css/bootstrap.min.css",
	}) {
		t.Error("missing Bootstrap 5.2.0 CSS link")
	}
	// JS script with custom version
	if !hasChild(head, "script", map[string]string{
		"src": "https://cdn.jsdelivr.net/npm/bootstrap@5.2.0/dist/js/bootstrap.bundle.min.js",
	}) {
		t.Error("missing Bootstrap 5.2.0 JS script")
	}
	// SRI hash should NOT be present for custom version
	for _, child := range head.ChildElems {
		if child.TagName == "link" || child.TagName == "script" {
			if _, ok := child.Attr["integrity"]; ok {
				t.Errorf("integrity attribute should not be present for custom version on %s", child.TagName)
			}
			if _, ok := child.Attr["crossorigin"]; ok {
				t.Errorf("crossorigin attribute should not be present for custom version on %s", child.TagName)
			}
		}
	}
}

func TestMaterializeCSS_CustomVersion(t *testing.T) {
	head := buildHead(t, MaterializeCSS("0.100.2"))

	if !hasChild(head, "link", map[string]string{
		"href": "https://cdnjs.cloudflare.com/ajax/libs/materialize/0.100.2/css/materialize.min.css",
	}) {
		t.Error("missing Materialize 0.100.2 CSS link")
	}
	if !hasChild(head, "script", map[string]string{
		"src": "https://cdnjs.cloudflare.com/ajax/libs/materialize/0.100.2/js/materialize.min.js",
	}) {
		t.Error("missing Materialize 0.100.2 JS script")
	}
}

func TestTailwindCSS_CustomVersion(t *testing.T) {
	head := buildHead(t, TailwindCSS("4.1"))

	if !hasChild(head, "script", map[string]string{
		"src": "https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4.1",
	}) {
		t.Error("missing Tailwind 4.1 script")
	}
}

func TestBootstrapCSS_IntegrationWithRender(t *testing.T) {
	// Build a head with Title + BootstrapCSS, render to HTML and check output
	head := &gerbera.Element{TagName: "head"}

	// Simulate dom.Title("My Page") — Tag("title", Value("My Page"))
	titleEl := &gerbera.Element{TagName: "title", Value: "My Page"}
	head.ChildElems = append(head.ChildElems, titleEl)

	BootstrapCSS()(head)

	var buf bytes.Buffer
	if err := gerbera.Render(&buf, head); err != nil {
		t.Fatalf("render error: %v", err)
	}
	html := buf.String()

	checks := []string{
		`charset="utf-8"`,
		`name="viewport"`,
		`<title>My Page</title>`,
		`bootstrap@5.3.3/dist/css/bootstrap.min.css`,
		`bootstrap@5.3.3/dist/js/bootstrap.bundle.min.js`,
	}
	for _, s := range checks {
		if !bytes.Contains([]byte(html), []byte(s)) {
			t.Errorf("rendered HTML missing %q", s)
		}
	}
}
