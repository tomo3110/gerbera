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

	store := newSessionStore(cfg.sessionTTL)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("gerbera-ws") == "1" {
			handleWS(w, r, store)
			return
		}
		handleHTTP(w, r, viewFactory, store, cfg)
	})
}

func handleHTTP(w http.ResponseWriter, r *http.Request, viewFactory func() View, store *sessionStore, cfg *handlerConfig) {
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

	components := view.Render()
	// Append the gerbera.js script tag
	components = append(components, dom.Body(
		gerbera.Literal(fmt.Sprintf("<script>%s</script>", gerberaJS)),
	))

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

func handleWS(w http.ResponseWriter, r *http.Request, store *sessionStore) {
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

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var ev wsEvent
		if err := json.Unmarshal(msg, &ev); err != nil {
			continue
		}

		sess.mu.Lock()

		if err := sess.View.HandleEvent(ev.Name, ev.Payload); err != nil {
			sess.mu.Unlock()
			continue
		}

		components := sess.View.Render()
		// Re-derive lang and session ID from existing tree
		lang := ""
		if sess.tree != nil && sess.tree.Attr != nil {
			lang = sess.tree.Attr["lang"]
		}
		newTree, err := buildTree(lang, sess.ID, components)
		if err != nil {
			sess.mu.Unlock()
			continue
		}

		patches := diff.Diff(sess.tree, newTree)
		sess.tree = newTree
		sess.mu.Unlock()

		if len(patches) > 0 {
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
