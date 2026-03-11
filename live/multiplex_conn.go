package live

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tomo3110/gerbera/session"
)

// muxViewEntry tracks a view and its transport within a multiplexed connection.
type muxViewEntry struct {
	vs        *viewSession
	transport *MultiplexTransport
}

// multiplexConn manages a single multiplexed WebSocket connection.
// It dispatches incoming messages to per-view transports and fans-in
// outbound messages from all views to a single write goroutine.
type multiplexConn struct {
	conn     *websocket.Conn
	registry *ViewRegistry
	cfg      *handlerConfig
	dlog     *debugLogger
	req      *http.Request // original upgrade request (for session, etc.)

	mu    sync.Mutex
	views map[string]*muxViewEntry

	outCh chan outboundMsg
	done  chan struct{}
}

func newMultiplexConn(conn *websocket.Conn, registry *ViewRegistry, cfg *handlerConfig, dlog *debugLogger, r *http.Request) *multiplexConn {
	return &multiplexConn{
		conn:     conn,
		registry: registry,
		cfg:      cfg,
		dlog:     dlog,
		req:      r,
		views:    make(map[string]*muxViewEntry),
		outCh:    make(chan outboundMsg, 64),
		done:     make(chan struct{}),
	}
}

// run starts the read and write goroutines and blocks until the connection closes.
func (mc *multiplexConn) run() {
	defer mc.conn.Close()
	defer mc.cleanupAll()

	// Start write goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		mc.writeLoop()
	}()

	// Read loop (blocks)
	mc.readLoop()

	// Signal write loop to stop
	close(mc.done)
	wg.Wait()
}

func (mc *multiplexConn) readLoop() {
	for {
		_, raw, err := mc.conn.ReadMessage()
		if err != nil {
			return
		}

		var msg MultiplexClientMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "mount":
			mc.handleMount(msg)
		case "event":
			mc.dispatchEvent(msg.ViewID, msg.Event, msg.Payload)
		case "params_change":
			mc.dispatchEvent(msg.ViewID, "gerbera:params_change", msg.Payload)
		case "unmount":
			mc.handleUnmount(msg)
		}
	}
}

func (mc *multiplexConn) writeLoop() {
	for {
		select {
		case out := <-mc.outCh:
			serverMsg := mc.buildServerMessage(out)
			if err := mc.conn.WriteJSON(serverMsg); err != nil {
				return
			}
		case <-mc.done:
			// Drain remaining messages
			for {
				select {
				case out := <-mc.outCh:
					serverMsg := mc.buildServerMessage(out)
					mc.conn.WriteJSON(serverMsg)
				default:
					return
				}
			}
		}
	}
}

func (mc *multiplexConn) buildServerMessage(out outboundMsg) MultiplexServerMessage {
	msg := out.msg
	serverMsg := MultiplexServerMessage{
		Type:     "update",
		ViewID:   out.viewID,
		Commands: msg.JSCommands,
	}
	if len(msg.Patches) > 0 {
		patchJSON, _ := json.Marshal(msg.Patches)
		serverMsg.Patches = patchJSON
	}
	return serverMsg
}

func (mc *multiplexConn) handleMount(msg MultiplexClientMessage) {
	// Parse path to get the base path for registry lookup
	parsedURL, _ := url.Parse(msg.Path)
	factory := mc.registry.lookup(parsedURL.Path)
	if factory == nil {
		mc.sendError(msg.ViewID, "unknown path: "+msg.Path)
		return
	}

	ctx, cancel := context.WithCancel(mc.req.Context())
	view := factory(ctx)

	sess := &Session{
		ID:        generateID(),
		CSRFToken: generateID(),
		View:      view,
		infoCh:    make(chan any, 32),
		stopTick:  make(chan struct{}),
	}

	// Build mount params
	params := Params{
		Path:  parsedURL.Path,
		Query: parsedURL.Query(),
		Conn: ConnInfo{
			LiveSession: sess,
			RemoteAddr:  mc.req.RemoteAddr,
			UserAgent:   mc.req.Header.Get("User-Agent"),
		},
	}

	// Propagate HTTP session if available
	if httpSess := session.FromContext(mc.req.Context()); httpSess != nil {
		params.Conn.Session = httpSess
	}

	if err := view.Mount(params); err != nil {
		cancel()
		mc.sendError(msg.ViewID, "mount failed: "+err.Error())
		return
	}

	vs := &viewSession{
		viewID: msg.ViewID,
		path:   msg.Path,
		view:   view,
		sess:   sess,
		cancel: cancel,
	}

	// Render initial HTML
	bodyHTML, err := renderInitialHTML(vs, mc.cfg)
	if err != nil {
		cancel()
		mc.sendError(msg.ViewID, "render failed: "+err.Error())
		return
	}

	transport := newMultiplexTransport(msg.ViewID, mc.outCh)

	mc.mu.Lock()
	mc.views[msg.ViewID] = &muxViewEntry{vs: vs, transport: transport}
	mc.mu.Unlock()

	// Send mounted response (direct write since ViewLoop hasn't started yet)
	mc.conn.WriteJSON(MultiplexServerMessage{
		Type:      "mounted",
		ViewID:    msg.ViewID,
		HTML:      bodyHTML,
		SessionID: sess.ID,
		CSRF:      sess.CSRFToken,
	})

	// Start ViewLoop in a goroutine
	go func() {
		loopCfg := ViewLoopConfig{
			SessionID:  sess.ID,
			CSRFToken:  sess.CSRFToken,
			Lang:       mc.cfg.lang,
			Debug:      mc.cfg.debug,
			SessionTTL: mc.cfg.sessionTTL,
			InfoCh:     sess.infoCh,
			StopTick:   sess.stopTick,
			DLog:       mc.dlog,
		}

		// Broker subscription
		if httpSess := session.FromContext(mc.req.Context()); httpSess != nil {
			if bs, ok := mc.cfg.sessionStore.(session.BrokerStore); ok {
				loopCfg.Broker = bs.Broker()
				loopCfg.HTTPSessionID = httpSess.ID
			}
		}

		if err := ViewLoop(view, transport, loopCfg); err != nil && err != ErrSessionExpired {
			mc.dlog.handleError(sess.ID, "ViewLoop", err)
		}

		transport.Close()

		mc.mu.Lock()
		delete(mc.views, msg.ViewID)
		mc.mu.Unlock()
	}()
}

func (mc *multiplexConn) dispatchEvent(viewID, event string, payload Payload) {
	mc.mu.Lock()
	entry := mc.views[viewID]
	mc.mu.Unlock()

	if entry == nil {
		return
	}

	// Push event into the transport's event channel for ViewLoop to consume
	select {
	case entry.transport.eventCh <- eventMsg{name: event, payload: payload}:
	default:
		// Event channel full, drop event
	}
}

func (mc *multiplexConn) handleUnmount(msg MultiplexClientMessage) {
	mc.mu.Lock()
	entry := mc.views[msg.ViewID]
	mc.mu.Unlock()

	if entry == nil {
		return
	}

	// Close the transport, which will cause ViewLoop to exit
	entry.transport.Close()
	entry.vs.cancel()
}

func (mc *multiplexConn) sendError(viewID, errMsg string) {
	mc.conn.WriteJSON(MultiplexServerMessage{
		Type:   "error",
		ViewID: viewID,
	})
}

func (mc *multiplexConn) cleanupAll() {
	mc.mu.Lock()
	entries := make([]*muxViewEntry, 0, len(mc.views))
	for _, e := range mc.views {
		entries = append(entries, e)
	}
	mc.mu.Unlock()

	for _, e := range entries {
		e.transport.Close()
		e.vs.cancel()
	}

	// Give ViewLoops time to run Unmount
	time.Sleep(50 * time.Millisecond)
}
