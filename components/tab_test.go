package components

import (
	"strings"
	"testing"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/property"
)

// helper builds a parent element and applies the ComponentFunc.
func buildTab(t *testing.T, cf gerbera.ComponentFunc) *gerbera.Element {
	t.Helper()
	parent := &gerbera.Element{TagName: "div", Children: make([]*gerbera.Element, 0)}
	if err := cf(parent); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parent.Children) == 0 {
		t.Fatal("expected at least one child element")
	}
	return parent.Children[0]
}

func TestTabs_Structure(t *testing.T) {
	tabs := []Tab{
		{Label: "Tab A", Content: property.Value("Content A")},
		{Label: "Tab B", Content: property.Value("Content B")},
	}
	wrapper := buildTab(t, Tabs("test", 0, tabs))

	// Wrapper should be a div with class gerbera-tabs
	if wrapper.TagName != "div" {
		t.Errorf("wrapper tag: got %q, want %q", wrapper.TagName, "div")
	}
	if _, ok := wrapper.ClassNames["gerbera-tabs"]; !ok {
		t.Error("wrapper missing class gerbera-tabs")
	}
	if wrapper.Attr["id"] != "test" {
		t.Errorf("wrapper id: got %q, want %q", wrapper.Attr["id"], "test")
	}

	// Should have 3 children: tablist + 2 panels
	if len(wrapper.Children) != 3 {
		t.Fatalf("wrapper children: got %d, want 3", len(wrapper.Children))
	}

	tablist := wrapper.Children[0]
	if tablist.Attr["role"] != "tablist" {
		t.Errorf("tablist role: got %q, want %q", tablist.Attr["role"], "tablist")
	}

	// Tablist should have 2 buttons
	if len(tablist.Children) != 2 {
		t.Fatalf("tablist children: got %d, want 2", len(tablist.Children))
	}
	for _, btn := range tablist.Children {
		if btn.TagName != "button" {
			t.Errorf("tab button tag: got %q, want %q", btn.TagName, "button")
		}
		if btn.Attr["role"] != "tab" {
			t.Errorf("tab button role: got %q, want %q", btn.Attr["role"], "tab")
		}
	}
}

func TestTabs_ActiveTab(t *testing.T) {
	tabs := []Tab{
		{Label: "First", Content: property.Value("Panel 1")},
		{Label: "Second", Content: property.Value("Panel 2")},
		{Label: "Third", Content: property.Value("Panel 3")},
	}
	wrapper := buildTab(t, Tabs("t", 1, tabs))

	tablist := wrapper.Children[0]

	// Check aria-selected
	for i, btn := range tablist.Children {
		got := btn.Attr["aria-selected"]
		if i == 1 {
			if got != "true" {
				t.Errorf("tab %d aria-selected: got %q, want %q", i, got, "true")
			}
			if _, ok := btn.ClassNames["gerbera-tab-active"]; !ok {
				t.Errorf("tab %d missing gerbera-tab-active class", i)
			}
			if btn.Attr["tabindex"] != "0" {
				t.Errorf("active tab tabindex: got %q, want %q", btn.Attr["tabindex"], "0")
			}
		} else {
			if got != "false" {
				t.Errorf("tab %d aria-selected: got %q, want %q", i, got, "false")
			}
			if _, ok := btn.ClassNames["gerbera-tab-active"]; ok {
				t.Errorf("tab %d should not have gerbera-tab-active class", i)
			}
			if btn.Attr["tabindex"] != "-1" {
				t.Errorf("inactive tab %d tabindex: got %q, want %q", i, btn.Attr["tabindex"], "-1")
			}
		}
	}

	// Check panels: only activeIndex=1 should not have hidden
	for i := 0; i < 3; i++ {
		panel := wrapper.Children[i+1] // panels start after tablist
		_, hasHidden := panel.Attr["hidden"]
		if i == 1 {
			if hasHidden {
				t.Errorf("panel %d should not have hidden attribute", i)
			}
		} else {
			if !hasHidden {
				t.Errorf("panel %d should have hidden attribute", i)
			}
		}
	}
}

func TestTabs_AriaLinks(t *testing.T) {
	tabs := []Tab{
		{Label: "A", Content: property.Value("P")},
		{Label: "B", Content: property.Value("Q")},
	}
	wrapper := buildTab(t, Tabs("my-tabs", 0, tabs))

	tablist := wrapper.Children[0]
	for i := 0; i < 2; i++ {
		btn := tablist.Children[i]
		panel := wrapper.Children[i+1]

		expectedTabID := "my-tabs-tab-" + string(rune('0'+i))
		expectedPanelID := "my-tabs-panel-" + string(rune('0'+i))

		if btn.Attr["id"] != expectedTabID {
			t.Errorf("tab %d id: got %q, want %q", i, btn.Attr["id"], expectedTabID)
		}
		if btn.Attr["aria-controls"] != expectedPanelID {
			t.Errorf("tab %d aria-controls: got %q, want %q", i, btn.Attr["aria-controls"], expectedPanelID)
		}
		if panel.Attr["id"] != expectedPanelID {
			t.Errorf("panel %d id: got %q, want %q", i, panel.Attr["id"], expectedPanelID)
		}
		if panel.Attr["aria-labelledby"] != expectedTabID {
			t.Errorf("panel %d aria-labelledby: got %q, want %q", i, panel.Attr["aria-labelledby"], expectedTabID)
		}
	}
}

func TestTabs_ButtonAttrs(t *testing.T) {
	tabs := []Tab{
		{
			Label:   "Click Me",
			Content: property.Value("content"),
			ButtonAttrs: []gerbera.ComponentFunc{
				property.Attr("gerbera-click", "switch-tab"),
				property.Attr("gerbera-value", "0"),
			},
		},
	}
	wrapper := buildTab(t, Tabs("btn-test", 0, tabs))

	btn := wrapper.Children[0].Children[0]
	if btn.Attr["gerbera-click"] != "switch-tab" {
		t.Errorf("button gerbera-click: got %q, want %q", btn.Attr["gerbera-click"], "switch-tab")
	}
	if btn.Attr["gerbera-value"] != "0" {
		t.Errorf("button gerbera-value: got %q, want %q", btn.Attr["gerbera-value"], "0")
	}
}

func TestTabs_EmptyTabs(t *testing.T) {
	// Should not panic
	wrapper := buildTab(t, Tabs("empty", 0, nil))
	if wrapper.TagName != "div" {
		t.Errorf("wrapper tag: got %q, want %q", wrapper.TagName, "div")
	}
	// Only tablist child, no panels
	if len(wrapper.Children) != 1 {
		t.Errorf("wrapper children: got %d, want 1 (just tablist)", len(wrapper.Children))
	}
}

func TestTabs_OutOfRangeIndex(t *testing.T) {
	tabs := []Tab{
		{Label: "Only", Content: property.Value("content")},
	}
	// activeIndex=5 is out of range — should not panic, all panels hidden
	wrapper := buildTab(t, Tabs("oor", 5, tabs))
	panel := wrapper.Children[1]
	if _, hasHidden := panel.Attr["hidden"]; !hasHidden {
		t.Error("panel should have hidden attribute when activeIndex is out of range")
	}

	// Button should not be marked active
	btn := wrapper.Children[0].Children[0]
	if btn.Attr["aria-selected"] != "false" {
		t.Errorf("button aria-selected: got %q, want %q", btn.Attr["aria-selected"], "false")
	}
}

func TestTabs_NegativeIndex(t *testing.T) {
	tabs := []Tab{
		{Label: "A", Content: property.Value("a")},
	}
	// Should not panic with negative index
	wrapper := buildTab(t, Tabs("neg", -1, tabs))
	panel := wrapper.Children[1]
	if _, hasHidden := panel.Attr["hidden"]; !hasHidden {
		t.Error("panel should have hidden attribute when activeIndex is negative")
	}
}

func TestTabsDefaultCSS(t *testing.T) {
	parent := &gerbera.Element{TagName: "head", Children: make([]*gerbera.Element, 0)}
	if err := TabsDefaultCSS()(parent); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parent.Children) == 0 {
		t.Fatal("expected a child element")
	}
	style := parent.Children[0]
	if style.TagName != "style" {
		t.Errorf("tag: got %q, want %q", style.TagName, "style")
	}
	if !strings.Contains(style.Value, "gerbera-tab") {
		t.Error("CSS should contain gerbera-tab rules")
	}
}

func TestTabs_WrapperAttrs(t *testing.T) {
	tabs := []Tab{
		{Label: "A", Content: property.Value("a")},
	}
	wrapper := buildTab(t, Tabs("w", 0, tabs, property.Attr("data-custom", "yes")))
	if wrapper.Attr["data-custom"] != "yes" {
		t.Errorf("wrapper data-custom: got %q, want %q", wrapper.Attr["data-custom"], "yes")
	}
}

func TestTabs_NilContent(t *testing.T) {
	tabs := []Tab{
		{Label: "Empty Tab"},
	}
	// Should not panic with nil Content
	wrapper := buildTab(t, Tabs("nil-content", 0, tabs))
	panel := wrapper.Children[1]
	if panel.Attr["role"] != "tabpanel" {
		t.Errorf("panel role: got %q, want %q", panel.Attr["role"], "tabpanel")
	}
}
