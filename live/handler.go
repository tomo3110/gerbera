package live

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/diff"
	"github.com/tomo3110/gerbera/dom"
)

type handlerConfig struct {
	lang       string
	sessionTTL time.Duration
	debug      bool
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

type wsEvent struct {
	Name    string  `json:"e"`
	Payload Payload `json:"p"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
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

	dlog := newDebugLogger(cfg.debug)
	store := newSessionStore(cfg.sessionTTL, func(id string) {
		dlog.sessionExpired(id)
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("gerbera-ws") == "1" {
			handleWS(w, r, store, cfg, dlog)
			return
		}
		handleHTTP(w, r, viewFactory, store, cfg, dlog)
	})
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
		components = append(components, dom.Body(
			gerbera.Literal(fmt.Sprintf("<script>%s</script>", gerberaJS)),
			gerbera.Literal(fmt.Sprintf("<script>%s</script>", gerberaDebugJS)),
		))
	} else {
		components = append(components, dom.Body(
			gerbera.Literal(fmt.Sprintf("<script>%s</script>", gerberaJS)),
		))
	}

	tree, err := buildTree(cfg.lang, sess.ID, components)
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

func handleWS(w http.ResponseWriter, r *http.Request, store *sessionStore, cfg *handlerConfig, dlog *debugLogger) {
	sessionID := r.URL.Query().Get("session")
	sess := store.get(sessionID)
	if sess == nil {
		http.Error(w, "session not found", http.StatusNotFound)
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

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var ev wsEvent
		if err := json.Unmarshal(msg, &ev); err != nil {
			dlog.handleError(sessionID, "unmarshal event", err)
			continue
		}

		dlog.eventReceived(sessionID, ev.Name, ev.Payload)

		var diffStart time.Time
		if cfg.debug {
			diffStart = time.Now()
		}

		sess.mu.Lock()

		if err := sess.View.HandleEvent(ev.Name, ev.Payload); err != nil {
			dlog.handleError(sessionID, "HandleEvent", err)
			sess.mu.Unlock()
			continue
		}

		components := sess.View.Render()
		lang := ""
		if sess.tree != nil && sess.tree.Attr != nil {
			lang = sess.tree.Attr["lang"]
		}
		newTree, err := buildTree(lang, sess.ID, components)
		if err != nil {
			dlog.handleError(sessionID, "buildTree", err)
			sess.mu.Unlock()
			continue
		}

		patches := diff.Diff(sess.tree, newTree)
		sess.tree = newTree

		var viewStateJSON json.RawMessage
		if cfg.debug {
			viewStateJSON, _ = json.Marshal(sess.View)
		}

		sess.mu.Unlock()

		var duration time.Duration
		if cfg.debug {
			duration = time.Since(diffStart)
		}
		dlog.patchesGenerated(sessionID, len(patches), duration)

		if len(patches) == 0 && !cfg.debug {
			continue
		}

		if cfg.debug {
			patchJSON, _ := json.Marshal(patches)
			envelope := debugMessage{
				Patches: patchJSON,
				Debug: &debugMeta{
					Event:      ev.Name,
					Payload:    ev.Payload,
					PatchCount: len(patches),
					DurationMS: duration.Milliseconds(),
					ViewState:  viewStateJSON,
					SessionID:  sessionID,
					SessionTTL: cfg.sessionTTL.String(),
					Timestamp:  time.Now().UnixMilli(),
				},
			}
			if err := conn.WriteJSON(envelope); err != nil {
				break
			}
		} else {
			if err := conn.WriteJSON(patches); err != nil {
				break
			}
		}
	}
}

// buildTree creates the full <html> Element tree from view components.
func buildTree(lang, sessionID string, components []gerbera.ComponentFunc) (*gerbera.Element, error) {
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
	return root, nil
}
