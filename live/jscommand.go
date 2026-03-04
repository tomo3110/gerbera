package live

import "sync"

// jsCommand represents a server-to-client command.
type jsCommand struct {
	Cmd    string            `json:"cmd"`
	Target string            `json:"target,omitempty"`
	Args   map[string]string `json:"args,omitempty"`
}

// wsMessage is the message envelope when JS commands are present.
type wsMessage struct {
	Patches    any         `json:"patches"`
	JSCommands []jsCommand `json:"js_commands,omitempty"`
}

// JSCommander is an optional interface for Views that issue JS commands.
type JSCommander interface {
	DrainCommands() []jsCommand
}

// CommandQueue is a helper that Views can embed to gain JS command support.
type CommandQueue struct {
	mu   sync.Mutex
	cmds []jsCommand
}

// PushCommand queues a JS command to be sent to the client after the next render.
func (q *CommandQueue) PushCommand(cmd string, target string, args map[string]string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.cmds = append(q.cmds, jsCommand{Cmd: cmd, Target: target, Args: args})
}

// ScrollTo queues a scrollTo command for the given CSS selector.
func (q *CommandQueue) ScrollTo(selector string, top, left string) {
	q.PushCommand("scroll_to", selector, map[string]string{"top": top, "left": left})
}

// ScrollIntoPct queues a percentage-based scroll command.
// pct should be a value between 0.0 and 1.0.
func (q *CommandQueue) ScrollIntoPct(selector string, pct string) {
	q.PushCommand("scroll_into_pct", selector, map[string]string{"pct": pct})
}

// Focus queues a focus command for the given CSS selector.
func (q *CommandQueue) Focus(selector string) {
	q.PushCommand("focus", selector, nil)
}

// Blur queues a blur command for the given CSS selector.
func (q *CommandQueue) Blur(selector string) {
	q.PushCommand("blur", selector, nil)
}

// SetAttribute queues an attribute set command.
func (q *CommandQueue) SetAttribute(selector, key, value string) {
	q.PushCommand("set_attr", selector, map[string]string{"key": key, "value": value})
}

// RemoveAttribute queues an attribute remove command.
func (q *CommandQueue) RemoveAttribute(selector, key string) {
	q.PushCommand("remove_attr", selector, map[string]string{"key": key})
}

// AddClass queues a CSS class add command.
func (q *CommandQueue) AddClass(selector, className string) {
	q.PushCommand("add_class", selector, map[string]string{"class": className})
}

// RemoveClass queues a CSS class remove command.
func (q *CommandQueue) RemoveClass(selector, className string) {
	q.PushCommand("remove_class", selector, map[string]string{"class": className})
}

// ToggleClass queues a CSS class toggle command.
func (q *CommandQueue) ToggleClass(selector, className string) {
	q.PushCommand("toggle_class", selector, map[string]string{"class": className})
}

// SetProperty queues a DOM property set command (e.g., value, checked).
func (q *CommandQueue) SetProperty(selector, key, value string) {
	q.PushCommand("set_prop", selector, map[string]string{"key": key, "value": value})
}

// Dispatch queues a custom event dispatch on the target element.
func (q *CommandQueue) Dispatch(selector, event string) {
	q.PushCommand("dispatch", selector, map[string]string{"event": event})
}

// Show queues a show command (sets display to "").
func (q *CommandQueue) Show(selector string) {
	q.PushCommand("show", selector, nil)
}

// Hide queues a hide command (sets display to "none").
func (q *CommandQueue) Hide(selector string) {
	q.PushCommand("hide", selector, nil)
}

// Toggle queues a visibility toggle command.
func (q *CommandQueue) Toggle(selector string) {
	q.PushCommand("toggle", selector, nil)
}

// Navigate queues a client-side navigation to the given URL.
func (q *CommandQueue) Navigate(url string) {
	q.PushCommand("navigate", "", map[string]string{"url": url})
}

// PushPatch pushes a new browser history entry without a full page reload.
// The URL changes and the state is preserved. Use this for in-page navigation
// where the View handles the state change (e.g. tab switches, filters).
func (q *CommandQueue) PushPatch(path string) {
	q.PushCommand("push_patch", "", map[string]string{"path": path})
}

// ReplacePatch replaces the current browser history entry without adding a new one.
// Use this when the URL should reflect state but the user shouldn't be able
// to "go back" to the previous URL (e.g. sort order changes).
func (q *CommandQueue) ReplacePatch(path string) {
	q.PushCommand("replace_patch", "", map[string]string{"path": path})
}

// PushNavigate performs a full navigation to a new URL, closing the WebSocket
// connection and loading the new page. Use this for cross-page navigation
// where a different View handles the destination.
func (q *CommandQueue) PushNavigate(path string) {
	q.PushCommand("push_navigate", "", map[string]string{"path": path})
}

// DrainCommands returns all queued commands and clears the queue.
func (q *CommandQueue) DrainCommands() []jsCommand {
	q.mu.Lock()
	defer q.mu.Unlock()
	cmds := q.cmds
	q.cmds = nil
	return cmds
}
