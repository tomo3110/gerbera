package components

import (
	"strings"
	"testing"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/tui"
)

func renderToString(c gerbera.ComponentFunc) (string, error) {
	root := &gerbera.Element{TagName: "root"}
	if err := c(root); err != nil {
		return "", err
	}
	return tui.RenderString(root.Children[0]), nil
}

func TestProgressBar(t *testing.T) {
	out, err := renderToString(ProgressBar(50, 100, 10))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "█") || !strings.Contains(out, "░") {
		t.Errorf("expected progress bar chars, got %q", out)
	}
	if !strings.Contains(out, "50%") {
		t.Errorf("expected 50%%, got %q", out)
	}
}

func TestProgressBarEmpty(t *testing.T) {
	out, err := renderToString(ProgressBar(0, 100, 10))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "0%") {
		t.Errorf("expected 0%%, got %q", out)
	}
}

func TestProgressBarFull(t *testing.T) {
	out, err := renderToString(ProgressBar(100, 100, 10))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "100%") {
		t.Errorf("expected 100%%, got %q", out)
	}
}

func TestSpinner(t *testing.T) {
	out, err := renderToString(Spinner(0))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "⠋") {
		t.Errorf("expected spinner frame 0, got %q", out)
	}

	out, err = renderToString(Spinner(1))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "⠙") {
		t.Errorf("expected spinner frame 1, got %q", out)
	}
}

func TestTabs(t *testing.T) {
	out, err := renderToString(Tabs([]string{"Tab1", "Tab2", "Tab3"}, 0))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Tab1") || !strings.Contains(out, "Tab2") || !strings.Contains(out, "Tab3") {
		t.Errorf("expected all tab labels, got %q", out)
	}
}

func TestKeyHelp(t *testing.T) {
	out, err := renderToString(KeyHelp([][2]string{{"k", "up"}, {"j", "down"}, {"q", "quit"}}))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "k") || !strings.Contains(out, "up") {
		t.Errorf("expected key help content, got %q", out)
	}
}

func TestStatusBar(t *testing.T) {
	out, err := renderToString(StatusBar("left", "center", "right", 40))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "left") || !strings.Contains(out, "center") || !strings.Contains(out, "right") {
		t.Errorf("expected status bar content, got %q", out)
	}
}

func TestFormField(t *testing.T) {
	out, err := renderToString(FormField("Name:", "Alice", "Enter name", false))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Name:") || !strings.Contains(out, "Alice") {
		t.Errorf("expected form field content, got %q", out)
	}
}

func TestFormFieldPlaceholder(t *testing.T) {
	out, err := renderToString(FormField("Name:", "", "Enter name", false))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Enter name") {
		t.Errorf("expected placeholder, got %q", out)
	}
}
