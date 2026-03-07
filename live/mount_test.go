package live

import (
	"testing"

	g "github.com/tomo3110/gerbera"
)

func TestLiveMount_BasicAttributes(t *testing.T) {
	parent := &g.Element{TagName: "div", ChildElems: make([]*g.Element, 0)}
	LiveMount("/admin/notifications")(parent)

	if len(parent.ChildElems) == 0 {
		t.Fatal("expected at least one child element")
	}
	div := parent.ChildElems[0]
	if div.TagName != "div" {
		t.Errorf("tag: got %q, want %q", div.TagName, "div")
	}
	if div.Attr["gerbera-live"] != "/admin/notifications" {
		t.Errorf("gerbera-live: got %q, want %q", div.Attr["gerbera-live"], "/admin/notifications")
	}
	if div.Attr["gerbera-live-id"] == "" {
		t.Error("gerbera-live-id should not be empty")
	}
}

func TestLiveMount_UniqueIDs(t *testing.T) {
	parent := &g.Element{TagName: "div", ChildElems: make([]*g.Element, 0)}
	LiveMount("/a")(parent)
	LiveMount("/b")(parent)

	if len(parent.ChildElems) != 2 {
		t.Fatalf("expected 2 children, got %d", len(parent.ChildElems))
	}
	id1 := parent.ChildElems[0].Attr["gerbera-live-id"]
	id2 := parent.ChildElems[1].Attr["gerbera-live-id"]
	if id1 == id2 {
		t.Errorf("LiveMount IDs should be unique, both got %q", id1)
	}
}

func TestLiveMount_WithMountID(t *testing.T) {
	parent := &g.Element{TagName: "div", ChildElems: make([]*g.Element, 0)}
	LiveMount("/chat", WithMountID("chat-panel"))(parent)

	div := parent.ChildElems[0]
	if div.Attr["id"] != "chat-panel" {
		t.Errorf("id: got %q, want %q", div.Attr["id"], "chat-panel")
	}
}

func TestLiveMount_WithMountClass(t *testing.T) {
	parent := &g.Element{TagName: "div", ChildElems: make([]*g.Element, 0)}
	LiveMount("/sidebar", WithMountClass("sidebar-panel"))(parent)

	div := parent.ChildElems[0]
	if _, ok := div.ClassNames["sidebar-panel"]; !ok {
		t.Error("expected class sidebar-panel")
	}
}

func TestLiveMount_MultipleOptions(t *testing.T) {
	parent := &g.Element{TagName: "div", ChildElems: make([]*g.Element, 0)}
	LiveMount("/widget",
		WithMountID("my-widget"),
		WithMountClass("widget-container"),
	)(parent)

	div := parent.ChildElems[0]
	if div.Attr["id"] != "my-widget" {
		t.Errorf("id: got %q, want %q", div.Attr["id"], "my-widget")
	}
	if _, ok := div.ClassNames["widget-container"]; !ok {
		t.Error("expected class widget-container")
	}
	if div.Attr["gerbera-live"] != "/widget" {
		t.Errorf("gerbera-live: got %q, want %q", div.Attr["gerbera-live"], "/widget")
	}
}
