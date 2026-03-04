package live

import (
	"net/url"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/diff"
)

// TestView is a test helper that drives a View through its lifecycle
// without a real WebSocket connection.
type TestView struct {
	View     View
	tree     *gerbera.Element
	Patches  []diff.Patch
	Rendered bool
}

// NewTestView creates a new TestView wrapper around a View instance.
func NewTestView(view View) *TestView {
	return &TestView{View: view}
}

// Mount calls View.Mount with the given params and renders the initial tree.
func (tv *TestView) Mount(params Params) error {
	if params.Query == nil {
		params.Query = make(url.Values)
	}
	if err := tv.View.Mount(params); err != nil {
		return err
	}
	components := tv.View.Render()
	tree, err := buildTree("en", "test-session", "", components)
	if err != nil {
		return err
	}
	tv.tree = tree
	tv.Rendered = true
	return nil
}

// SimulateEvent simulates a user event and returns the resulting patches.
func (tv *TestView) SimulateEvent(event string, payload Payload) ([]diff.Patch, error) {
	if payload == nil {
		payload = Payload{}
	}
	if err := tv.View.HandleEvent(event, payload); err != nil {
		return nil, err
	}

	components := tv.View.Render()
	newTree, err := buildTree("en", "test-session", "", components)
	if err != nil {
		return nil, err
	}

	patches := diff.Diff(tv.tree, newTree)
	tv.tree = newTree
	tv.Patches = patches
	return patches, nil
}

// SimulateTick simulates a tick event for TickerView implementations.
func (tv *TestView) SimulateTick() ([]diff.Patch, error) {
	tickerView, ok := tv.View.(TickerView)
	if !ok {
		return nil, nil
	}

	if err := tickerView.HandleTick(); err != nil {
		return nil, err
	}

	components := tv.View.Render()
	newTree, err := buildTree("en", "test-session", "", components)
	if err != nil {
		return nil, err
	}

	patches := diff.Diff(tv.tree, newTree)
	tv.tree = newTree
	tv.Patches = patches
	return patches, nil
}

// SimulateInfo simulates an info message for InfoReceiver implementations.
func (tv *TestView) SimulateInfo(msg any) ([]diff.Patch, error) {
	infoReceiver, ok := tv.View.(InfoReceiver)
	if !ok {
		return nil, nil
	}

	if err := infoReceiver.HandleInfo(msg); err != nil {
		return nil, err
	}

	components := tv.View.Render()
	newTree, err := buildTree("en", "test-session", "", components)
	if err != nil {
		return nil, err
	}

	patches := diff.Diff(tv.tree, newTree)
	tv.tree = newTree
	tv.Patches = patches
	return patches, nil
}

// SimulateParams simulates a browser back/forward navigation for Patcher implementations.
// The path parameter is the URL path (e.g. "/profile"), and params are the query parameters.
func (tv *TestView) SimulateParams(path string, params url.Values) ([]diff.Patch, error) {
	patcher, ok := tv.View.(Patcher)
	if !ok {
		return nil, nil
	}

	if err := patcher.HandleParams(path, params); err != nil {
		return nil, err
	}

	components := tv.View.Render()
	newTree, err := buildTree("en", "test-session", "", components)
	if err != nil {
		return nil, err
	}

	patches := diff.Diff(tv.tree, newTree)
	tv.tree = newTree
	tv.Patches = patches
	return patches, nil
}

// RenderHTML returns the current rendered HTML as a string.
func (tv *TestView) RenderHTML() (string, error) {
	if tv.tree == nil {
		return "", nil
	}
	return diff.RenderFragment(tv.tree)
}

// HasPatch returns true if the last operation produced a patch with the given operation type.
func (tv *TestView) HasPatch(op diff.OpType) bool {
	for _, p := range tv.Patches {
		if p.Op == op {
			return true
		}
	}
	return false
}

// PatchCount returns the number of patches from the last operation.
func (tv *TestView) PatchCount() int {
	return len(tv.Patches)
}
