package live

import (
	"encoding/json"
	"log/slog"
	"time"
)

// debugLogger wraps slog for gerbera-live debug logging.
// When enabled is false, all methods are no-ops with zero overhead.
type debugLogger struct {
	enabled bool
	logger  *slog.Logger
}

func newDebugLogger(enabled bool) *debugLogger {
	if !enabled {
		return &debugLogger{enabled: false}
	}
	return &debugLogger{
		enabled: true,
		logger:  slog.Default().With("component", "gerbera-live"),
	}
}

func (d *debugLogger) eventReceived(sessionID, event string, payload Payload) {
	if !d.enabled {
		return
	}
	d.logger.Info("event received",
		"session_id", sessionID,
		"event", event,
		"payload", payload,
	)
}

func (d *debugLogger) patchesGenerated(sessionID string, count int, duration time.Duration) {
	if !d.enabled {
		return
	}
	d.logger.Info("patches generated",
		"session_id", sessionID,
		"patch_count", count,
		"duration_ms", duration.Milliseconds(),
	)
}

func (d *debugLogger) sessionCreated(sessionID string) {
	if !d.enabled {
		return
	}
	d.logger.Info("session created", "session_id", sessionID)
}

func (d *debugLogger) sessionConnected(sessionID string) {
	if !d.enabled {
		return
	}
	d.logger.Info("websocket connected", "session_id", sessionID)
}

func (d *debugLogger) sessionDisconnected(sessionID string) {
	if !d.enabled {
		return
	}
	d.logger.Info("websocket disconnected", "session_id", sessionID)
}

func (d *debugLogger) sessionExpired(sessionID string) {
	if !d.enabled {
		return
	}
	d.logger.Info("session expired", "session_id", sessionID)
}

func (d *debugLogger) handleError(sessionID, context string, err error) {
	if !d.enabled {
		return
	}
	d.logger.Error("error",
		"session_id", sessionID,
		"context", context,
		"error", err.Error(),
	)
}

// debugMessage is the WebSocket envelope sent when debug mode is enabled.
type debugMessage struct {
	Patches json.RawMessage `json:"patches"`
	Debug   *debugMeta      `json:"debug,omitempty"`
}

// debugMeta carries debug information sent to the browser DevPanel.
type debugMeta struct {
	Event      string          `json:"event"`
	Payload    Payload         `json:"payload"`
	PatchCount int             `json:"patchCount"`
	DurationMS int64           `json:"durationMs"`
	ViewState  json.RawMessage `json:"viewState"`
	SessionID  string          `json:"sessionId"`
	SessionTTL string          `json:"sessionTtl"`
	Timestamp  int64           `json:"timestamp"`
}
