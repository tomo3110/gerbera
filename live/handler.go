package live

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/diff"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/session"
)

type handlerConfig struct {
	lang         string
	sessionTTL   time.Duration
	debug        bool
	middlewares  []func(http.Handler) http.Handler
	checkOrigin  func(r *http.Request) bool
	sessionStore session.Store
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

// WithSessionStore sets the session store for push-based session invalidation.
// If the store implements session.BrokerStore, WebSocket connections will
// automatically subscribe to session invalidation events and close when
// the session is destroyed or expires.
func WithSessionStore(store session.Store) Option {
	return func(c *handlerConfig) { c.sessionStore = store }
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
// The context.Context carries request-scoped values (e.g. authenticated user
// set by middleware via r.Context()).
func Handler(viewFactory func(context.Context) View, opts ...Option) http.Handler {
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

func handleHTTP(w http.ResponseWriter, r *http.Request, viewFactory func(context.Context) View, store *sessionStore, cfg *handlerConfig, dlog *debugLogger) {
	view := viewFactory(r.Context())

	sess := store.create(view)
	dlog.sessionCreated(sess.ID)

	params := Params{
		Path:  r.URL.Path,
		Query: r.URL.Query(),
		Conn: ConnInfo{
			LiveSession: sess,
			RemoteAddr:  r.RemoteAddr,
			UserAgent:   r.Header.Get("User-Agent"),
		},
	}
	if httpSess := session.FromContext(r.Context()); httpSess != nil {
		params.Conn.Session = httpSess
	}
	if err := view.Mount(params); err != nil {
		store.remove(sess.ID)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	components := view.Render()
	components = appendScriptComponents(components, cfg.debug)

	tree := buildTree(cfg.lang, sess.ID, sess.CSRFToken, components)

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
	defer store.remove(sessionID)

	dlog.sessionConnected(sessionID)
	defer dlog.sessionDisconnected(sessionID)

	// Create WSTransport
	var wsOpts []WSTransportOption
	if cfg.debug {
		wsOpts = append(wsOpts, WithWSDebug(sessionID, cfg.sessionTTL))
	}
	transport := NewWSTransport(conn, wsOpts...)
	defer transport.Close()

	// Build ViewLoopConfig
	loopCfg := ViewLoopConfig{
		SessionID:  sessionID,
		CSRFToken:  sess.CSRFToken,
		Lang:       cfg.lang,
		Debug:      cfg.debug,
		SessionTTL: cfg.sessionTTL,
		InfoCh:     sess.infoCh,
		StopTick:   sess.stopTick,
		DLog:       dlog,
	}

	// Broker subscription setup
	if httpSess := session.FromContext(r.Context()); httpSess != nil {
		if bs, ok := cfg.sessionStore.(session.BrokerStore); ok {
			loopCfg.Broker = bs.Broker()
			loopCfg.HTTPSessionID = httpSess.ID
		}
	}

	// Run the view lifecycle loop
	if err := ViewLoop(sess.View, transport, loopCfg); err != nil && err != ErrSessionExpired {
		dlog.handleError(sessionID, "ViewLoop", err)
	}
}

// renderAndDiff renders the view and computes patches against the stored tree.
// Must be called with sess.mu held.
// When debug is true, it also marshals the View state for the debug panel.
func renderAndDiff(sess *Session, cfg *handlerConfig) ([]diff.Patch, []jsCommand, json.RawMessage) {
	components := sess.View.Render()
	components = appendScriptComponents(components, cfg.debug)
	lang := ""
	if sess.tree != nil && sess.tree.Attr != nil {
		lang = sess.tree.Attr["lang"]
	}
	newTree := buildTree(lang, sess.ID, sess.CSRFToken, components)

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

	return patches, cmds, viewState
}

// appendScriptComponents appends gerbera JS (and debug JS when enabled)
// as a second <body> component. Both handleHTTP and renderAndDiff call this
// so that old and new trees always share the same structure, preventing
// spurious diff patches for the script body.
func appendScriptComponents(components []gerbera.ComponentFunc, debug bool) []gerbera.ComponentFunc {
	if debug {
		debugHTML := renderDebugPanelHTML()
		escaped := escapeForJSString(debugHTML)
		debugJS := strings.Replace(gerberaDebugJS,
			`/*__GERBERA_DEBUG_HTML__*/""`,
			`"`+escaped+`"`, 1)
		return append(components, dom.Body(
			gerbera.Literal(fmt.Sprintf("<script>%s</script>", gerberaJS)),
			gerbera.Literal(fmt.Sprintf("<script>%s</script>", debugJS)),
		))
	}
	return append(components, dom.Body(
		gerbera.Literal(fmt.Sprintf("<script>%s</script>", gerberaJS)),
	))
}

// buildTree creates the full <html> Element tree from view components.
func buildTree(lang, sessionID, csrfToken string, components []gerbera.ComponentFunc) *gerbera.Element {
	root := &gerbera.Element{
		TagName:    "html",
		Attr:       gerbera.AttrMap{"lang": lang, "gerbera-session": sessionID},
		ClassNames: make(gerbera.ClassMap),
		Children:   make([]*gerbera.Element, 0),
	}

	root = gerbera.Parse(root, components...)

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

	return root
}
