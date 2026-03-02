package app

import (
	"fmt"

	g "github.com/tomo3110/gerbera"
)

// EventType represents the type of terminal event.
type EventType int

const (
	EventKey    EventType = iota // Keyboard input
	EventResize                  // Terminal resize
)

// Event represents a terminal input event.
type Event struct {
	Type   EventType // Event type
	Key    string    // Key name: "enter", "a", "ctrl+c", "esc", etc.
	Runes  []rune    // Input runes for character keys
	Width  int       // Terminal width (for EventResize)
	Height int       // Terminal height (for EventResize)
}

// OnKey sets a tui-on-key attribute to map a key press to an event name.
func OnKey(key, event string) g.ComponentFunc {
	return func(el *g.Element) error {
		if el.Attr == nil {
			el.Attr = make(map[string]string)
		}
		el.Attr[fmt.Sprintf("tui-on-key-%s", key)] = event
		return nil
	}
}

// Focusable marks an element as focusable with a tab index.
func Focusable(tabIndex int) g.ComponentFunc {
	return func(el *g.Element) error {
		if el.Attr == nil {
			el.Attr = make(map[string]string)
		}
		el.Attr["tui-focusable"] = fmt.Sprint(tabIndex)
		return nil
	}
}
