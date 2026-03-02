package tui

import (
	"bytes"
	"strings"
	"testing"

	g "github.com/tomo3110/gerbera"
	ts "github.com/tomo3110/gerbera/tui/style"
)

func TestRenderText(t *testing.T) {
	root := &g.Element{TagName: "root"}
	c := Text(g.Literal("hello"))
	if err := c(root); err != nil {
		t.Fatal(err)
	}
	el, err := g.Parse(root, c)
	if err != nil {
		t.Fatal(err)
	}

	result := RenderString(el.Children[len(el.Children)-1])
	if !strings.Contains(result, "hello") {
		t.Errorf("expected output to contain 'hello', got %q", result)
	}
}

func TestRenderVBox(t *testing.T) {
	root := &g.Element{TagName: "root"}
	c := VBox(
		Text(g.Literal("line1")),
		Text(g.Literal("line2")),
	)
	if err := c(root); err != nil {
		t.Fatal(err)
	}

	result := RenderString(root.Children[0])
	if !strings.Contains(result, "line1") || !strings.Contains(result, "line2") {
		t.Errorf("expected output to contain line1 and line2, got %q", result)
	}
}

func TestRenderHBox(t *testing.T) {
	root := &g.Element{TagName: "root"}
	c := HBox(
		Text(g.Literal("left")),
		Text(g.Literal("right")),
	)
	if err := c(root); err != nil {
		t.Fatal(err)
	}

	result := RenderString(root.Children[0])
	if !strings.Contains(result, "left") || !strings.Contains(result, "right") {
		t.Errorf("expected output to contain left and right, got %q", result)
	}
}

func TestRenderList(t *testing.T) {
	root := &g.Element{TagName: "root"}
	c := List(
		ListItem(g.Literal("item1")),
		ListItem(g.Literal("item2")),
		ListItem(g.Literal("item3")),
	)
	if err := c(root); err != nil {
		t.Fatal(err)
	}

	result := RenderString(root.Children[0])
	if !strings.Contains(result, "•") {
		t.Errorf("expected list items with bullet markers, got %q", result)
	}
	if !strings.Contains(result, "item1") || !strings.Contains(result, "item2") || !strings.Contains(result, "item3") {
		t.Errorf("expected all items in output, got %q", result)
	}
}

func TestRenderDivider(t *testing.T) {
	root := &g.Element{TagName: "root"}
	c := Divider(ts.Width(10))
	if err := c(root); err != nil {
		t.Fatal(err)
	}

	result := RenderString(root.Children[0])
	if !strings.Contains(result, "──────────") {
		t.Errorf("expected divider line, got %q", result)
	}
}

func TestRenderProgress(t *testing.T) {
	el := &g.Element{
		TagName: "progress",
		Attr: map[string]string{
			"tui-current": "50",
			"tui-total":   "100",
			"tui-width":   "10",
		},
	}

	result := RenderString(el)
	if !strings.Contains(result, "█") && !strings.Contains(result, "░") {
		t.Errorf("expected progress bar characters, got %q", result)
	}
	if !strings.Contains(result, "50%") {
		t.Errorf("expected percentage, got %q", result)
	}
}

func TestRenderSpinner(t *testing.T) {
	el := &g.Element{
		TagName: "spinner",
		Value:   "loading",
		Attr: map[string]string{
			"tui-frame": "0",
		},
	}

	result := RenderString(el)
	if !strings.Contains(result, "⠋") {
		t.Errorf("expected spinner frame, got %q", result)
	}
	if !strings.Contains(result, "loading") {
		t.Errorf("expected label, got %q", result)
	}
}

func TestRenderButton(t *testing.T) {
	root := &g.Element{TagName: "root"}
	c := Button(g.Literal("OK"))
	if err := c(root); err != nil {
		t.Fatal(err)
	}

	result := RenderString(root.Children[0])
	if !strings.Contains(result, "OK") {
		t.Errorf("expected button text, got %q", result)
	}
	if !strings.Contains(result, "[") || !strings.Contains(result, "]") {
		t.Errorf("expected button brackets, got %q", result)
	}
}

func TestRenderCheckbox(t *testing.T) {
	el := &g.Element{
		TagName: "checkbox",
		Value:   "option",
		Attr:    map[string]string{"tui-checked": "true"},
	}
	result := RenderString(el)
	if !strings.Contains(result, "[x]") {
		t.Errorf("expected checked box, got %q", result)
	}

	el.Attr["tui-checked"] = "false"
	result = RenderString(el)
	if !strings.Contains(result, "[ ]") {
		t.Errorf("expected unchecked box, got %q", result)
	}
}

func TestRenderWithStyle(t *testing.T) {
	root := &g.Element{TagName: "root"}
	c := Text(ts.Bold(true), g.Literal("bold text"))
	if err := c(root); err != nil {
		t.Fatal(err)
	}

	result := RenderString(root.Children[0])
	if !strings.Contains(result, "bold text") {
		t.Errorf("expected styled text, got %q", result)
	}
}

func TestRenderToWriter(t *testing.T) {
	root := &g.Element{TagName: "root"}
	c := Text(g.Literal("writer test"))
	if err := c(root); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := Render(&buf, root.Children[0]); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "writer test") {
		t.Errorf("expected 'writer test' in output, got %q", buf.String())
	}
}

func TestRenderTable(t *testing.T) {
	root := &g.Element{TagName: "root"}
	c := Table(
		TableRow(
			TableCell(g.Literal("Name")),
			TableCell(g.Literal("Age")),
		),
		TableRow(
			TableCell(g.Literal("Alice")),
			TableCell(g.Literal("30")),
		),
	)
	if err := c(root); err != nil {
		t.Fatal(err)
	}

	result := RenderString(root.Children[0])
	if !strings.Contains(result, "Name") || !strings.Contains(result, "Alice") {
		t.Errorf("expected table content, got %q", result)
	}
}

func TestBuildStyleBorder(t *testing.T) {
	el := &g.Element{
		TagName: "box",
		Attr: map[string]string{
			"tui-border": "rounded",
		},
	}
	// Should not panic
	_ = buildStyle(el)
}

func TestParseSpacing(t *testing.T) {
	tests := []struct {
		input string
		want  [4]int
	}{
		{"5", [4]int{5, 5, 5, 5}},
		{"1 2", [4]int{1, 2, 1, 2}},
		{"1 2 3", [4]int{1, 2, 3, 2}},
		{"1 2 3 4", [4]int{1, 2, 3, 4}},
	}
	for _, tt := range tests {
		result := parseSpacing(tt.input)
		if result != tt.want {
			t.Errorf("parseSpacing(%q) = %v, want %v", tt.input, result, tt.want)
		}
	}
}
