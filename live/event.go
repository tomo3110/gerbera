package live

import (
	"fmt"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/property"
)

// Click binds a click event to the element.
func Click(event string) gerbera.ComponentFunc {
	return property.Attr("gerbera-click", event)
}

// ClickValue sets the gerbera-value attribute on the element.
// When a click event fires, this value is sent in the payload as "value".
func ClickValue(value string) gerbera.ComponentFunc {
	return property.Attr("gerbera-value", value)
}

// Input binds an input event to the element.
func Input(event string) gerbera.ComponentFunc {
	return property.Attr("gerbera-input", event)
}

// Change binds a change event to the element.
func Change(event string) gerbera.ComponentFunc {
	return property.Attr("gerbera-change", event)
}

// Submit binds a submit event to the element.
func Submit(event string) gerbera.ComponentFunc {
	return property.Attr("gerbera-submit", event)
}

// Focus binds a focus event to the element.
func Focus(event string) gerbera.ComponentFunc {
	return property.Attr("gerbera-focus", event)
}

// Blur binds a blur event to the element.
func Blur(event string) gerbera.ComponentFunc {
	return property.Attr("gerbera-blur", event)
}

// Keydown binds a keydown event to the element.
func Keydown(event string) gerbera.ComponentFunc {
	return property.Attr("gerbera-keydown", event)
}

// Key sets a key filter for keydown events.
func Key(key string) gerbera.ComponentFunc {
	return property.Attr("gerbera-key", key)
}

// Scroll binds a scroll event to the element.
// Default throttle is 100ms. Use Throttle() to customize.
func Scroll(event string) gerbera.ComponentFunc {
	return property.Attr("gerbera-scroll", event)
}

// Throttle sets the throttle interval in milliseconds for scroll events.
func Throttle(ms int) gerbera.ComponentFunc {
	return property.Attr("gerbera-throttle", fmt.Sprintf("%d", ms))
}

// DblClick binds a double-click event to the element.
func DblClick(event string) gerbera.ComponentFunc {
	return property.Attr("gerbera-dblclick", event)
}

// MouseEnter binds a mouseenter event to the element.
func MouseEnter(event string) gerbera.ComponentFunc {
	return property.Attr("gerbera-mouseenter", event)
}

// MouseLeave binds a mouseleave event to the element.
func MouseLeave(event string) gerbera.ComponentFunc {
	return property.Attr("gerbera-mouseleave", event)
}

// TouchStart binds a touchstart event to the element.
func TouchStart(event string) gerbera.ComponentFunc {
	return property.Attr("gerbera-touchstart", event)
}

// TouchEnd binds a touchend event to the element.
func TouchEnd(event string) gerbera.ComponentFunc {
	return property.Attr("gerbera-touchend", event)
}

// TouchMove binds a touchmove event to the element.
func TouchMove(event string) gerbera.ComponentFunc {
	return property.Attr("gerbera-touchmove", event)
}

// Debounce sets the debounce interval in milliseconds for any event.
// When set, the event handler will wait for the specified duration after
// the last event before sending to the server.
func Debounce(ms int) gerbera.ComponentFunc {
	return property.Attr("gerbera-debounce", fmt.Sprintf("%d", ms))
}

// Hook attaches a lifecycle hook to the element.
// The hook name must be registered on the client via Gerbera.registerHook().
// Hook callbacks: mounted(el), updated(el), destroyed(el), disconnected(el).
func Hook(name string) gerbera.ComponentFunc {
	return property.Attr("gerbera-hook", name)
}

// LiveLink marks an anchor element for client-side navigation.
// Clicking the link will push state and send a navigation event
// instead of performing a full page reload.
func LiveLink(href string) gerbera.ComponentFunc {
	return func(n gerbera.Node) {
		n.SetAttribute("href", href)
		n.SetAttribute("gerbera-live-link", href)
	}
}

// Loading marks an element to receive loading state CSS class.
// When an event is being processed, elements matching the selector
// receive the "gerbera-loading" class on the <html> element.
func Loading(cssClass string) gerbera.ComponentFunc {
	return property.Attr("gerbera-loading-class", cssClass)
}
