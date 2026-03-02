package app

import (
	g "github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/tui"
)

// TestView is a test helper that drives a View through its lifecycle
// without a real terminal.
type TestView struct {
	View     View
	rendered bool
}

// NewTestView creates a new TestView wrapper around a View instance.
func NewTestView(view View) *TestView {
	return &TestView{View: view}
}

// Mount calls View.Mount with the given params and performs the initial render.
func (tv *TestView) Mount(params Params) error {
	if params == nil {
		params = Params{}
	}
	if err := tv.View.Mount(params); err != nil {
		return err
	}
	tv.rendered = true
	return nil
}

// SimulateEvent simulates a terminal event and returns the rendered output.
func (tv *TestView) SimulateEvent(event Event) (string, error) {
	if err := tv.View.HandleEvent(event); err != nil {
		return "", err
	}
	return tv.RenderString()
}

// RenderString returns the current rendered output as a string.
func (tv *TestView) RenderString() (string, error) {
	components := tv.View.Render()
	root := &g.Element{TagName: "root"}
	for _, c := range components {
		if err := c(root); err != nil {
			return "", err
		}
	}

	var result string
	for _, child := range root.Children {
		result += tui.RenderString(child) + "\n"
	}
	return result, nil
}
