package style

import (
	"fmt"

	g "github.com/tomo3110/gerbera"
)

func setAttr(key, value string) g.ComponentFunc {
	return func(el *g.Element) error {
		if el.Attr == nil {
			el.Attr = make(map[string]string)
		}
		el.Attr[key] = value
		return nil
	}
}

// Color

func FgColor(color string) g.ComponentFunc { return setAttr("tui-fg", color) }
func BgColor(color string) g.ComponentFunc { return setAttr("tui-bg", color) }

// Text decoration

func Bold(enabled bool) g.ComponentFunc          { return setAttr("tui-bold", fmt.Sprint(enabled)) }
func Italic(enabled bool) g.ComponentFunc        { return setAttr("tui-italic", fmt.Sprint(enabled)) }
func Underline(enabled bool) g.ComponentFunc     { return setAttr("tui-underline", fmt.Sprint(enabled)) }
func Strikethrough(enabled bool) g.ComponentFunc { return setAttr("tui-strikethrough", fmt.Sprint(enabled)) }
func Faint(enabled bool) g.ComponentFunc         { return setAttr("tui-faint", fmt.Sprint(enabled)) }
func Reverse(enabled bool) g.ComponentFunc       { return setAttr("tui-reverse", fmt.Sprint(enabled)) }

// Layout

func Width(w int) g.ComponentFunc     { return setAttr("tui-width", fmt.Sprint(w)) }
func Height(h int) g.ComponentFunc    { return setAttr("tui-height", fmt.Sprint(h)) }
func MaxWidth(w int) g.ComponentFunc  { return setAttr("tui-max-width", fmt.Sprint(w)) }
func MaxHeight(h int) g.ComponentFunc { return setAttr("tui-max-height", fmt.Sprint(h)) }

func Padding(top, right, bottom, left int) g.ComponentFunc {
	return setAttr("tui-padding", fmt.Sprintf("%d %d %d %d", top, right, bottom, left))
}

func Margin(top, right, bottom, left int) g.ComponentFunc {
	return setAttr("tui-margin", fmt.Sprintf("%d %d %d %d", top, right, bottom, left))
}

// Border

func Border(style string) g.ComponentFunc      { return setAttr("tui-border", style) }
func BorderColor(color string) g.ComponentFunc { return setAttr("tui-border-color", color) }
func BorderLabel(label string) g.ComponentFunc { return setAttr("tui-border-label", label) }

// Alignment

func Align(alignment string) g.ComponentFunc  { return setAttr("tui-align", alignment) }
func VAlign(alignment string) g.ComponentFunc { return setAttr("tui-valign", alignment) }
