package live

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tomo3110/gerbera/diff"
)

// --- Message types for the multiplex protocol ---

// MultiplexClientMessage is the wire format for client-to-server messages
// on the multiplexed WebSocket connection.
type MultiplexClientMessage struct {
	Type    string  `json:"type"`               // mount, event, params_change, unmount
	ViewID  string  `json:"view_id"`
	Path    string  `json:"path,omitempty"`
	Event   string  `json:"event,omitempty"`
	Payload Payload `json:"payload,omitempty"`
}

// MultiplexServerMessage is the wire format for server-to-client messages
// on the multiplexed WebSocket connection.
type MultiplexServerMessage struct {
	Type      string          `json:"type"`                  // mounted, update, error
	ViewID    string          `json:"view_id"`
	HTML      string          `json:"html,omitempty"`        // initial HTML for mounted
	Patches   json.RawMessage `json:"patches,omitempty"`
	Commands  []JSCommand     `json:"js_commands,omitempty"`
	SessionID string          `json:"session_id,omitempty"`  // for upload support
	CSRF      string          `json:"csrf,omitempty"`        // for upload support
}

// --- ViewRegistry ---

// ViewRegistry maps path prefixes to View factory functions.
// Register View factories before creating a MultiplexHandler.
type ViewRegistry struct {
	mu        sync.RWMutex
	factories map[string]func(context.Context) View
}

// NewViewRegistry creates an empty ViewRegistry.
func NewViewRegistry() *ViewRegistry {
	return &ViewRegistry{
		factories: make(map[string]func(context.Context) View),
	}
}

// Register adds a View factory for the given path.
func (r *ViewRegistry) Register(path string, factory func(context.Context) View) {
	r.mu.Lock()
	r.factories[path] = factory
	r.mu.Unlock()
}

// lookup returns the factory for a path, or nil if not found.
func (r *ViewRegistry) lookup(path string) func(context.Context) View {
	r.mu.RLock()
	f := r.factories[path]
	r.mu.RUnlock()
	return f
}

// --- viewSession: per-View state within a multiplexed connection ---

type viewSession struct {
	viewID string
	path   string
	view   View
	sess   *Session
	cancel context.CancelFunc
}

// --- MultiplexHandler ---

// MultiplexHandler serves the multiplexed WebSocket endpoint.
// It upgrades the HTTP connection and delegates to multiplexConn.
type MultiplexHandler struct {
	registry *ViewRegistry
	cfg      *handlerConfig
	upgrader *websocket.Upgrader
	dlog     *debugLogger
}

// NewMultiplexHandler creates a new MultiplexHandler backed by the given registry.
func NewMultiplexHandler(registry *ViewRegistry, opts ...Option) *MultiplexHandler {
	cfg := &handlerConfig{
		lang:       "ja",
		sessionTTL: 5 * time.Minute,
	}
	for _, o := range opts {
		o(cfg)
	}

	checkOrigin := cfg.checkOrigin
	if checkOrigin == nil {
		checkOrigin = defaultCheckOrigin
	}

	return &MultiplexHandler{
		registry: registry,
		cfg:      cfg,
		upgrader: &websocket.Upgrader{CheckOrigin: checkOrigin},
		dlog:     newDebugLogger(cfg.debug),
	}
}

func (h *MultiplexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	mc := newMultiplexConn(conn, h.registry, h.cfg, h.dlog, r)
	mc.run()
}

// --- renderAndDiffForMux renders a view's body and returns HTML/patches ---

// renderInitialHTML performs the initial Mount + Render for a multiplexed view
// and returns the body's inner HTML string.
func renderInitialHTML(vs *viewSession, cfg *handlerConfig) (string, error) {
	components := vs.view.Render()
	tree := buildTree(cfg.lang, vs.sess.ID, vs.sess.CSRFToken, components)
	vs.sess.mu.Lock()
	vs.sess.tree = tree
	vs.sess.mu.Unlock()

	// Render the body content as an HTML fragment
	var bodyHTML string
	for _, child := range tree.ChildElems {
		if child.TagName == "body" {
			bodyHTML, _ = diff.RenderFragment(child)
			break
		}
	}
	return bodyHTML, nil
}
