package live

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/diff"
	"github.com/tomo3110/gerbera/dom"
)

type handlerConfig struct {
	lang        string
	sessionTTL  time.Duration
	debug       bool
	middlewares []func(http.Handler) http.Handler
	checkOrigin func(r *http.Request) bool
}

// Option configures the live handler.
type Option func(*handlerConfig)

// WithLang sets the HTML lang attribute (default "ja").
func WithLang(lang string) Option {
	return func(c *handlerConfig) { c.lang = lang }
}

// WithSessionTTL sets the session timeout (default 5 minutes).
func WithSessionTTL(d time.Duration) Option {
	return func(c *handlerConfig) { c.sessionTTL = d }
}

// WithDebug enables the browser DevPanel and server-side structured logging.
func WithDebug() Option {
	return func(c *handlerConfig) { c.debug = true }
}

// WithMiddleware adds HTTP middleware to the handler chain.
// Middleware is applied in the order provided.
// This can be used for authentication, logging, CORS, etc.
func WithMiddleware(mw func(http.Handler) http.Handler) Option {
	return func(c *handlerConfig) {
		c.middlewares = append(c.middlewares, mw)
	}
}

// WithCheckOrigin sets a custom Origin header check function for WebSocket upgrades.
// By default, the Origin header is validated against the request Host.
func WithCheckOrigin(fn func(r *http.Request) bool) Option {
	return func(c *handlerConfig) { c.checkOrigin = fn }
}

type wsEvent struct {
	Name    string  `json:"e"`
	Payload Payload `json:"p"`
}

// defaultCheckOrigin validates that the Origin header matches the request Host.
// Non-browser clients (no Origin header) are allowed.
func defaultCheckOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true // non-browser clients
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return u.Host == r.Host
}

// Handler returns an http.Handler for a LiveView.
// viewFactory is called once per session to create a new View instance.
func Handler(viewFactory func() View, opts ...Option) http.Handler {
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
	upgrader := &websocket.Upgrader{
		CheckOrigin: checkOrigin,
	}

	dlog := newDebugLogger(cfg.debug)
	store := newSessionStore(cfg.sessionTTL, func(id string) {
		dlog.sessionExpired(id)
	})

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("gerbera-ws") == "1" {
			handleWS(w, r, store, cfg, dlog, upgrader)
			return
		}
		if r.URL.Query().Get("gerbera-upload") == "1" && r.Method == http.MethodPost {
			handleUpload(w, r, store, dlog)
			return
		}
		handleHTTP(w, r, viewFactory, store, cfg, dlog)
	})

	// Apply middleware in reverse order so they execute in the order provided
	for i := len(cfg.middlewares) - 1; i >= 0; i-- {
		handler = cfg.middlewares[i](handler)
	}

	return handler
}

func handleHTTP(w http.ResponseWriter, r *http.Request, viewFactory func() View, store *sessionStore, cfg *handlerConfig, dlog *debugLogger) {
	view := viewFactory()

	params := make(Params)
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	if err := view.Mount(params); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sess := store.create(view)
	dlog.sessionCreated(sess.ID)

	components := view.Render()
	if cfg.debug {
		debugHTML := renderDebugPanelHTML()
		escaped := escapeForJSString(debugHTML)
		debugJS := strings.Replace(gerberaDebugJS,
			`/*__GERBERA_DEBUG_HTML__*/""`,
			`"`+escaped+`"`, 1)
		components = append(components, dom.Body(
			gerbera.Literal(fmt.Sprintf("<script>%s</script>", gerberaJS)),
			gerbera.Literal(fmt.Sprintf("<script>%s</script>", debugJS)),
		))
	} else {
		components = append(components, dom.Body(
			gerbera.Literal(fmt.Sprintf("<script>%s</script>", gerberaJS)),
		))
	}

	tree, err := buildTree(cfg.lang, sess.ID, sess.CSRFToken, components)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sess.mu.Lock()
	sess.tree = tree
	sess.mu.Unlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := gerbera.Render(w, tree); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleWS(w http.ResponseWriter, r *http.Request, store *sessionStore, cfg *handlerConfig, dlog *debugLogger, upgrader *websocket.Upgrader) {
	sessionID := r.URL.Query().Get("session")
	sess := store.get(sessionID)
	if sess == nil {
		// Session expired or not found - signal client to reload for session recovery
		http.Error(w, "session_expired", http.StatusGone)
		return
	}

	// Validate CSRF token
	csrfToken := r.URL.Query().Get("csrf")
	if subtle.ConstantTimeCompare([]byte(csrfToken), []byte(sess.CSRFToken)) != 1 {
		http.Error(w, "invalid CSRF token", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	defer store.remove(sessionID)

	dlog.sessionConnected(sessionID)
	defer dlog.sessionDisconnected(sessionID)

	// Channel for client WebSocket messages
	msgCh := make(chan []byte, 32)
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			msgCh <- msg
		}
	}()

	// Start ticker if view implements TickerView
	var tickCh <-chan time.Time
	if tv, ok := sess.View.(TickerView); ok {
		if interval := tv.TickInterval(); interval > 0 {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			tickCh = ticker.C
		}
	}

	var wsMu sync.Mutex // protects conn.WriteJSON

	for {
		select {
		case <-doneCh:
			close(sess.stopTick)
			return

		case msg := <-msgCh:
			var ev wsEvent
			if err := json.Unmarshal(msg, &ev); err != nil {
				dlog.handleError(sessionID, "unmarshal event", err)
				continue
			}

			dlog.eventReceived(sessionID, ev.Name, ev.Payload)

			if err := processEvent(sess, conn, cfg, dlog, sessionID, ev.Name, ev.Payload, &wsMu); err != nil {
				return
			}

		case <-tickCh:
			tv := sess.View.(TickerView)

			var diffStart time.Time
			if cfg.debug {
				diffStart = time.Now()
			}

			sess.mu.Lock()
			if err := tv.HandleTick(); err != nil {
				dlog.handleError(sessionID, "HandleTick", err)
				sess.mu.Unlock()
				continue
			}
			patches, jsCommands, viewState, err := renderAndDiff(sess, cfg)
			sess.mu.Unlock()
			if err != nil {
				dlog.handleError(sessionID, "renderAndDiff", err)
				continue
			}

			var duration time.Duration
			if cfg.debug {
				duration = time.Since(diffStart)
			}
			dlog.patchesGenerated(sessionID, len(patches), duration)

			wsMu.Lock()
			err = sendPatches(conn, patches, jsCommands, viewState, cfg, dlog, sessionID, "tick", nil, duration)
			wsMu.Unlock()
			if err != nil {
				return
			}

		case info := <-sess.infoCh:
			ir, ok := sess.View.(InfoReceiver)
			if !ok {
				continue
			}

			var diffStart time.Time
			if cfg.debug {
				diffStart = time.Now()
			}

			sess.mu.Lock()
			if err := ir.HandleInfo(info); err != nil {
				dlog.handleError(sessionID, "HandleInfo", err)
				sess.mu.Unlock()
				continue
			}
			patches, jsCommands, viewState, err := renderAndDiff(sess, cfg)
			sess.mu.Unlock()
			if err != nil {
				dlog.handleError(sessionID, "renderAndDiff", err)
				continue
			}

			var duration time.Duration
			if cfg.debug {
				duration = time.Since(diffStart)
			}
			dlog.patchesGenerated(sessionID, len(patches), duration)

			wsMu.Lock()
			err = sendPatches(conn, patches, jsCommands, viewState, cfg, dlog, sessionID, "info", nil, duration)
			wsMu.Unlock()
			if err != nil {
				return
			}
		}
	}
}

// processEvent handles a client WebSocket event: HandleEvent + render + diff + send.
func processEvent(sess *Session, conn *websocket.Conn, cfg *handlerConfig, dlog *debugLogger, sessionID, eventName string, payload Payload, wsMu *sync.Mutex) error {
	var diffStart time.Time
	if cfg.debug {
		diffStart = time.Now()
	}

	sess.mu.Lock()

	if err := sess.View.HandleEvent(eventName, payload); err != nil {
		dlog.handleError(sessionID, "HandleEvent", err)
		sess.mu.Unlock()
		return nil
	}

	patches, jsCommands, viewState, err := renderAndDiff(sess, cfg)
	sess.mu.Unlock()
	if err != nil {
		dlog.handleError(sessionID, "renderAndDiff", err)
		return nil
	}

	var duration time.Duration
	if cfg.debug {
		duration = time.Since(diffStart)
	}
	dlog.patchesGenerated(sessionID, len(patches), duration)

	wsMu.Lock()
	defer wsMu.Unlock()
	return sendPatches(conn, patches, jsCommands, viewState, cfg, dlog, sessionID, eventName, payload, duration)
}

// renderAndDiff renders the view and computes patches against the stored tree.
// Must be called with sess.mu held.
// When debug is true, it also marshals the View state for the debug panel.
func renderAndDiff(sess *Session, cfg *handlerConfig) ([]diff.Patch, []jsCommand, json.RawMessage, error) {
	components := sess.View.Render()
	lang := ""
	if sess.tree != nil && sess.tree.Attr != nil {
		lang = sess.tree.Attr["lang"]
	}
	newTree, err := buildTree(lang, sess.ID, sess.CSRFToken, components)
	if err != nil {
		return nil, nil, nil, err
	}

	patches := diff.Diff(sess.tree, newTree)
	sess.tree = newTree

	// Collect JS commands if view implements JSCommander
	var cmds []jsCommand
	if jc, ok := sess.View.(JSCommander); ok {
		cmds = jc.DrainCommands()
	}

	// Marshal View state for debug panel while still under lock
	var viewState json.RawMessage
	if cfg.debug {
		viewState, _ = json.Marshal(sess.View)
	}

	return patches, cmds, viewState, nil
}

// sendPatches sends patches (and optionally JS commands) to the client.
// Returns non-nil error if the connection should be closed.
func sendPatches(conn *websocket.Conn, patches []diff.Patch, jsCommands []jsCommand, viewState json.RawMessage, cfg *handlerConfig, dlog *debugLogger, sessionID, eventName string, payload Payload, duration time.Duration) error {
	if len(patches) == 0 && len(jsCommands) == 0 && !cfg.debug {
		return nil
	}

	if cfg.debug {
		patchJSON, _ := json.Marshal(patches)
		envelope := debugMessage{
			Patches:    patchJSON,
			JSCommands: jsCommands,
			Debug: &debugMeta{
				Event:      eventName,
				Payload:    payload,
				PatchCount: len(patches),
				DurationMS: duration.Milliseconds(),
				ViewState:  viewState,
				SessionID:  sessionID,
				SessionTTL: cfg.sessionTTL.String(),
				Timestamp:  time.Now().UnixMilli(),
			},
		}
		if err := conn.WriteJSON(envelope); err != nil {
			return err
		}
	} else {
		msg := wsMessage{
			Patches:    patches,
			JSCommands: jsCommands,
		}
		if len(jsCommands) == 0 {
			// Backward compatible: send patches directly as array
			if err := conn.WriteJSON(patches); err != nil {
				return err
			}
		} else {
			if err := conn.WriteJSON(msg); err != nil {
				return err
			}
		}
		_ = msg // suppress unused
	}
	return nil
}

// buildTree creates the full <html> Element tree from view components.
func buildTree(lang, sessionID, csrfToken string, components []gerbera.ComponentFunc) (*gerbera.Element, error) {
	root := &gerbera.Element{
		TagName:    "html",
		Attr:       gerbera.AttrMap{"lang": lang, "gerbera-session": sessionID},
		ClassNames: make(gerbera.ClassMap),
		Children:   make([]*gerbera.Element, 0),
	}

	root, err := gerbera.Parse(root, components...)
	if err != nil {
		return nil, err
	}

	// Inject CSRF meta tag into <head> if a token is provided
	if csrfToken != "" {
		for _, child := range root.Children {
			if child.TagName == "head" {
				meta := &gerbera.Element{
					TagName:    "meta",
					Attr:       gerbera.AttrMap{"name": "gerbera-csrf", "content": csrfToken},
					ClassNames: make(gerbera.ClassMap),
					Children:   make([]*gerbera.Element, 0),
				}
				child.Children = append([]*gerbera.Element{meta}, child.Children...)
				break
			}
		}
	}

	return root, nil
}
