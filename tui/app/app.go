package app

import (
	"os"
	"time"

	g "github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/tui"
)

// Params holds parameters passed to View.Mount.
type Params map[string]string

// View is the TUI version of live.View.
// The same Mount/Render/HandleEvent pattern adapted for terminal I/O.
type View interface {
	// Mount is called once when the app starts.
	Mount(params Params) error

	// Render returns the component tree for the current state.
	Render() []g.ComponentFunc

	// HandleEvent processes a terminal input event.
	HandleEvent(event Event) error
}

// TickerView is an optional interface for Views that need periodic updates.
type TickerView interface {
	View
	// TickInterval returns the interval between ticks.
	TickInterval() time.Duration
	// HandleTick is called on each tick.
	HandleTick() error
}

// Option configures the Run function.
type Option func(*runConfig)

type runConfig struct {
	params Params
}

// WithParams sets the initial parameters for Mount.
func WithParams(params Params) Option {
	return func(c *runConfig) {
		c.params = params
	}
}

// Run starts the interactive TUI event loop.
func Run(viewFactory func() View, opts ...Option) error {
	cfg := &runConfig{
		params: Params{},
	}
	for _, opt := range opts {
		opt(cfg)
	}

	view := viewFactory()

	// Enable raw mode
	restore, err := EnableRawMode(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer restore()

	// Enable alternate screen buffer
	os.Stdout.WriteString("\033[?1049h")
	defer os.Stdout.WriteString("\033[?1049l")

	HideCursor(os.Stdout)
	defer ShowCursor(os.Stdout)

	// Mount
	if err := view.Mount(cfg.params); err != nil {
		return err
	}

	// Initial render
	if err := renderView(view); err != nil {
		return err
	}

	// Event channel
	eventCh := make(chan Event, 16)
	errCh := make(chan error, 1)

	// Input reader goroutine
	go func() {
		for {
			ev, err := ReadEvent(os.Stdin)
			if err != nil {
				errCh <- err
				return
			}
			eventCh <- ev
		}
	}()

	// Ticker setup
	var tickCh <-chan time.Time
	if tv, ok := view.(TickerView); ok {
		interval := tv.TickInterval()
		if interval > 0 {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			tickCh = ticker.C
		}
	}

	// Main event loop
	for {
		select {
		case ev := <-eventCh:
			// Ctrl+C or "q" to quit (if not explicitly handled)
			if ev.Key == "ctrl+c" {
				return nil
			}

			if err := view.HandleEvent(ev); err != nil {
				return err
			}

			if err := renderView(view); err != nil {
				return err
			}

		case <-tickCh:
			if tv, ok := view.(TickerView); ok {
				if err := tv.HandleTick(); err != nil {
					return err
				}
				if err := renderView(view); err != nil {
					return err
				}
			}

		case err := <-errCh:
			return err
		}
	}
}

func renderView(view View) error {
	components := view.Render()
	root := &g.Element{TagName: "root"}
	for _, c := range components {
		if err := c(root); err != nil {
			return err
		}
	}

	ClearScreen(os.Stdout)

	for _, child := range root.Children {
		if err := tui.Render(os.Stdout, child); err != nil {
			return err
		}
		os.Stdout.WriteString("\n")
	}
	return nil
}
