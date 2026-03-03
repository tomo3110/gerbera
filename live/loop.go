package live

import (
	"time"

	"github.com/tomo3110/gerbera/session"
)

// ViewLoopConfig holds the parameters needed to run a ViewLoop.
type ViewLoopConfig struct {
	Params        Params
	SessionID     string        // LiveView session ID (debug)
	CSRFToken     string
	Lang          string
	Debug         bool
	SessionTTL    time.Duration // debug panel
	Broker        *session.Broker
	HTTPSessionID string // Broker subscription key
	InfoCh        <-chan any
	StopTick      chan struct{} // closed on loop exit
	DLog          *debugLogger
}

// ViewLoop drives a View's lifecycle over a Transport.
// It blocks until the Transport is closed, the session is invalidated,
// or an unrecoverable send error occurs.
func ViewLoop(view View, transport Transport, cfg ViewLoopConfig) error {
	// --- event receive goroutine ---
	type eventMsg struct {
		name    string
		payload Payload
		err     error
	}
	eventCh := make(chan eventMsg, 1)
	go func() {
		for {
			name, payload, err := transport.Receive()
			eventCh <- eventMsg{name, payload, err}
			if err != nil {
				return
			}
		}
	}()

	// --- ticker ---
	var tickCh <-chan time.Time
	if tv, ok := view.(TickerView); ok {
		if interval := tv.TickInterval(); interval > 0 {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			tickCh = ticker.C
		}
	}

	// --- broker subscription ---
	var invalidatedCh <-chan struct{}
	if cfg.Broker != nil && cfg.HTTPSessionID != "" {
		ch, unsub := cfg.Broker.Subscribe(cfg.HTTPSessionID)
		defer unsub()
		invalidatedCh = ch
	}

	// Shared state for renderAndDiff.
	sess := &Session{
		ID:        cfg.SessionID,
		CSRFToken: cfg.CSRFToken,
		View:      view,
	}

	// Build the initial tree so that the first diff has a baseline.
	components := view.Render()
	initTree, err := buildTree(cfg.Lang, cfg.SessionID, cfg.CSRFToken, components)
	if err != nil {
		return err
	}
	sess.tree = initTree

	hcfg := &handlerConfig{
		debug:      cfg.Debug,
		sessionTTL: cfg.SessionTTL,
	}

	for {
		select {
		case ev := <-eventCh:
			if ev.err != nil {
				// Connection closed.
				if cfg.StopTick != nil {
					close(cfg.StopTick)
				}
				return nil
			}

			if cfg.DLog != nil {
				cfg.DLog.eventReceived(cfg.SessionID, ev.name, ev.payload)
			}

			if err := loopProcessEvent(sess, transport, hcfg, cfg.DLog, cfg.SessionID, ev.name, ev.payload); err != nil {
				return err
			}

		case <-tickCh:
			tv := view.(TickerView)

			var diffStart time.Time
			if cfg.Debug {
				diffStart = time.Now()
			}

			sess.mu.Lock()
			if err := tv.HandleTick(); err != nil {
				if cfg.DLog != nil {
					cfg.DLog.handleError(cfg.SessionID, "HandleTick", err)
				}
				sess.mu.Unlock()
				continue
			}
			patches, jsCommands, viewState, err := renderAndDiff(sess, hcfg)
			sess.mu.Unlock()
			if err != nil {
				if cfg.DLog != nil {
					cfg.DLog.handleError(cfg.SessionID, "renderAndDiff", err)
				}
				continue
			}

			var duration time.Duration
			if cfg.Debug {
				duration = time.Since(diffStart)
			}
			if cfg.DLog != nil {
				cfg.DLog.patchesGenerated(cfg.SessionID, len(patches), duration)
			}

			if err := transport.Send(Message{
				Patches:    patches,
				JSCommands: jsCommands,
				ViewState:  viewState,
				EventName:  "tick",
				Duration:   duration,
			}); err != nil {
				return err
			}

		case info := <-cfg.InfoCh:
			ir, ok := view.(InfoReceiver)
			if !ok {
				continue
			}

			var diffStart time.Time
			if cfg.Debug {
				diffStart = time.Now()
			}

			sess.mu.Lock()
			if err := ir.HandleInfo(info); err != nil {
				if cfg.DLog != nil {
					cfg.DLog.handleError(cfg.SessionID, "HandleInfo", err)
				}
				sess.mu.Unlock()
				continue
			}
			patches, jsCommands, viewState, err := renderAndDiff(sess, hcfg)
			sess.mu.Unlock()
			if err != nil {
				if cfg.DLog != nil {
					cfg.DLog.handleError(cfg.SessionID, "renderAndDiff", err)
				}
				continue
			}

			var duration time.Duration
			if cfg.Debug {
				duration = time.Since(diffStart)
			}
			if cfg.DLog != nil {
				cfg.DLog.patchesGenerated(cfg.SessionID, len(patches), duration)
			}

			if err := transport.Send(Message{
				Patches:    patches,
				JSCommands: jsCommands,
				ViewState:  viewState,
				EventName:  "info",
				Duration:   duration,
			}); err != nil {
				return err
			}

		case <-invalidatedCh:
			if h, ok := view.(SessionExpiredHandler); ok {
				sess.mu.Lock()
				h.OnSessionExpired()
				patches, jsCommands, viewState, err := renderAndDiff(sess, hcfg)
				sess.mu.Unlock()
				if err == nil {
					transport.Send(Message{
						Patches:    patches,
						JSCommands: jsCommands,
						ViewState:  viewState,
						EventName:  "session_expired",
					})
					time.Sleep(100 * time.Millisecond)
				}
			}
			return ErrSessionExpired
		}
	}
}

// loopProcessEvent handles a single client event inside ViewLoop.
func loopProcessEvent(sess *Session, transport Transport, cfg *handlerConfig, dlog *debugLogger, sessionID, eventName string, payload Payload) error {
	var diffStart time.Time
	if cfg.debug {
		diffStart = time.Now()
	}

	sess.mu.Lock()
	if err := sess.View.HandleEvent(eventName, payload); err != nil {
		if dlog != nil {
			dlog.handleError(sessionID, "HandleEvent", err)
		}
		sess.mu.Unlock()
		return nil
	}
	patches, jsCommands, viewState, err := renderAndDiff(sess, cfg)
	sess.mu.Unlock()
	if err != nil {
		if dlog != nil {
			dlog.handleError(sessionID, "renderAndDiff", err)
		}
		return nil
	}

	var duration time.Duration
	if cfg.debug {
		duration = time.Since(diffStart)
	}
	if dlog != nil {
		dlog.patchesGenerated(sessionID, len(patches), duration)
	}

	return transport.Send(Message{
		Patches:      patches,
		JSCommands:   jsCommands,
		ViewState:    viewState,
		EventName:    eventName,
		EventPayload: payload,
		Duration:     duration,
	})
}
