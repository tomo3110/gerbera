package app

import (
	"fmt"
	"strings"
	"testing"

	g "github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/tui"
)

// testCounterView is a simple View for testing.
type testCounterView struct {
	Count int
}

func (v *testCounterView) Mount(params Params) error {
	v.Count = 0
	return nil
}

func (v *testCounterView) Render() []g.ComponentFunc {
	return []g.ComponentFunc{
		tui.Box(
			tui.Text(g.Literal(fmt.Sprintf("Count: %d", v.Count))),
		),
	}
}

func (v *testCounterView) HandleEvent(event Event) error {
	switch event.Key {
	case "k":
		v.Count++
	case "j":
		v.Count--
	}
	return nil
}

func TestTestViewMount(t *testing.T) {
	tv := NewTestView(&testCounterView{})
	if err := tv.Mount(nil); err != nil {
		t.Fatal(err)
	}
	out, err := tv.RenderString()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Count: 0") {
		t.Errorf("expected 'Count: 0', got %q", out)
	}
}

func TestTestViewSimulateEvent(t *testing.T) {
	tv := NewTestView(&testCounterView{})
	if err := tv.Mount(nil); err != nil {
		t.Fatal(err)
	}

	// Increment
	out, err := tv.SimulateEvent(Event{Type: EventKey, Key: "k"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Count: 1") {
		t.Errorf("expected 'Count: 1', got %q", out)
	}

	// Increment again
	out, err = tv.SimulateEvent(Event{Type: EventKey, Key: "k"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Count: 2") {
		t.Errorf("expected 'Count: 2', got %q", out)
	}

	// Decrement
	out, err = tv.SimulateEvent(Event{Type: EventKey, Key: "j"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Count: 1") {
		t.Errorf("expected 'Count: 1', got %q", out)
	}
}
