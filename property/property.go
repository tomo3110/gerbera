package property

import (
	"fmt"
	"html"

	"github.com/tomo3110/gerbera"
)

func Class(c ...string) gerbera.ComponentFunc {
	return func(el *gerbera.Element) error {
		el.ClassNames = make(gerbera.ClassMap, len(c))
		for _, s := range c {
			el.ClassNames[s] = false
		}
		return nil
	}
}

func Attr(key string, val any) gerbera.ComponentFunc {
	return func(el *gerbera.Element) error {
		if el.Attr == nil {
			el.Attr = make(map[string]string)
		}
		el.Attr[key] = html.EscapeString(fmt.Sprint(val))
		return nil
	}
}

func ID(text string) gerbera.ComponentFunc {
	return Attr("id", text)
}

func Name(text string) gerbera.ComponentFunc {
	return Attr("name", text)
}

func Value(value any) gerbera.ComponentFunc {
	return func(el *gerbera.Element) error {
		el.Value = html.EscapeString(fmt.Sprint(value))
		return nil
	}
}

// ClassIf conditionally adds a CSS class name to the element.
// If condition is true, the class is added; otherwise it is not.
func ClassIf(condition bool, className string) gerbera.ComponentFunc {
	return func(el *gerbera.Element) error {
		if condition {
			if el.ClassNames == nil {
				el.ClassNames = make(gerbera.ClassMap)
			}
			el.ClassNames[className] = false
		}
		return nil
	}
}

// ClassMap conditionally adds CSS class names based on a map.
// Each key is a class name and the bool value determines if it should be included.
func ClassMap(classes map[string]bool) gerbera.ComponentFunc {
	return func(el *gerbera.Element) error {
		if el.ClassNames == nil {
			el.ClassNames = make(gerbera.ClassMap)
		}
		for name, active := range classes {
			if active {
				el.ClassNames[name] = false
			}
		}
		return nil
	}
}

// Href sets the href attribute.
func Href(url string) gerbera.ComponentFunc {
	return Attr("href", url)
}

// Src sets the src attribute.
func Src(url string) gerbera.ComponentFunc {
	return Attr("src", url)
}

// Type sets the type attribute.
func Type(t string) gerbera.ComponentFunc {
	return Attr("type", t)
}

// Placeholder sets the placeholder attribute.
func Placeholder(text string) gerbera.ComponentFunc {
	return Attr("placeholder", text)
}

// Disabled sets the disabled attribute.
func Disabled(disabled bool) gerbera.ComponentFunc {
	return func(el *gerbera.Element) error {
		if disabled {
			if el.Attr == nil {
				el.Attr = make(map[string]string)
			}
			el.Attr["disabled"] = "disabled"
		}
		return nil
	}
}

// Required sets the required attribute.
func Required(required bool) gerbera.ComponentFunc {
	return func(el *gerbera.Element) error {
		if required {
			if el.Attr == nil {
				el.Attr = make(map[string]string)
			}
			el.Attr["required"] = "required"
		}
		return nil
	}
}

// Readonly sets the readonly attribute.
func Readonly(readonly bool) gerbera.ComponentFunc {
	return func(el *gerbera.Element) error {
		if readonly {
			if el.Attr == nil {
				el.Attr = make(map[string]string)
			}
			el.Attr["readonly"] = "readonly"
		}
		return nil
	}
}

// For sets the "for" attribute (e.g., on <label> elements).
func For(id string) gerbera.ComponentFunc {
	return Attr("for", id)
}

// Title sets the title attribute (tooltip).
func Title(text string) gerbera.ComponentFunc {
	return Attr("title", text)
}

// DataAttr sets a data-* attribute.
func DataAttr(key, value string) gerbera.ComponentFunc {
	return Attr("data-"+key, value)
}

// Tabindex sets the tabindex attribute.
func Tabindex(index int) gerbera.ComponentFunc {
	return Attr("tabindex", index)
}

// --- ARIA Accessibility Helpers ---

// AriaLabel sets the aria-label attribute.
func AriaLabel(label string) gerbera.ComponentFunc {
	return Attr("aria-label", label)
}

// AriaDescribedBy sets the aria-describedby attribute.
func AriaDescribedBy(id string) gerbera.ComponentFunc {
	return Attr("aria-describedby", id)
}

// AriaLabelledBy sets the aria-labelledby attribute.
func AriaLabelledBy(id string) gerbera.ComponentFunc {
	return Attr("aria-labelledby", id)
}

// Role sets the role attribute.
func Role(role string) gerbera.ComponentFunc {
	return Attr("role", role)
}

// AriaHidden sets the aria-hidden attribute.
func AriaHidden(hidden bool) gerbera.ComponentFunc {
	if hidden {
		return Attr("aria-hidden", "true")
	}
	return Attr("aria-hidden", "false")
}

// AriaExpanded sets the aria-expanded attribute.
func AriaExpanded(expanded bool) gerbera.ComponentFunc {
	if expanded {
		return Attr("aria-expanded", "true")
	}
	return Attr("aria-expanded", "false")
}

// AriaSelected sets the aria-selected attribute.
func AriaSelected(selected bool) gerbera.ComponentFunc {
	if selected {
		return Attr("aria-selected", "true")
	}
	return Attr("aria-selected", "false")
}

// AriaDisabled sets the aria-disabled attribute.
func AriaDisabled(disabled bool) gerbera.ComponentFunc {
	if disabled {
		return Attr("aria-disabled", "true")
	}
	return Attr("aria-disabled", "false")
}

// AriaLive sets the aria-live attribute for live regions.
// Common values: "polite", "assertive", "off".
func AriaLive(mode string) gerbera.ComponentFunc {
	return Attr("aria-live", mode)
}

// AriaAtomic sets the aria-atomic attribute.
func AriaAtomic(atomic bool) gerbera.ComponentFunc {
	if atomic {
		return Attr("aria-atomic", "true")
	}
	return Attr("aria-atomic", "false")
}

// AriaControls sets the aria-controls attribute.
func AriaControls(id string) gerbera.ComponentFunc {
	return Attr("aria-controls", id)
}

// AriaCurrent sets the aria-current attribute.
// Common values: "page", "step", "location", "date", "time", "true".
func AriaCurrent(value string) gerbera.ComponentFunc {
	return Attr("aria-current", value)
}

// AriaHasPopup sets the aria-haspopup attribute.
func AriaHasPopup(value string) gerbera.ComponentFunc {
	return Attr("aria-haspopup", value)
}

// AriaInvalid sets the aria-invalid attribute.
func AriaInvalid(invalid bool) gerbera.ComponentFunc {
	if invalid {
		return Attr("aria-invalid", "true")
	}
	return Attr("aria-invalid", "false")
}

// AriaRequired sets the aria-required attribute.
func AriaRequired(required bool) gerbera.ComponentFunc {
	if required {
		return Attr("aria-required", "true")
	}
	return Attr("aria-required", "false")
}

// AriaValueNow sets the aria-valuenow attribute (for range widgets).
func AriaValueNow(value any) gerbera.ComponentFunc {
	return Attr("aria-valuenow", value)
}

// AriaValueMin sets the aria-valuemin attribute.
func AriaValueMin(value any) gerbera.ComponentFunc {
	return Attr("aria-valuemin", value)
}

// AriaValueMax sets the aria-valuemax attribute.
func AriaValueMax(value any) gerbera.ComponentFunc {
	return Attr("aria-valuemax", value)
}
