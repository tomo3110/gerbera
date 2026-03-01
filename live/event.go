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
