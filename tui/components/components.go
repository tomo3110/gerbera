package components

import (
	"fmt"
	"strings"

	g "github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/tui"
	ts "github.com/tomo3110/gerbera/tui/style"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// ProgressBar renders a text-based progress bar.
//
//	█████░░░░░ 50/100 (50%)
func ProgressBar(current, total, width int) g.ComponentFunc {
	if total <= 0 {
		total = 100
	}
	ratio := float64(current) / float64(total)
	if ratio > 1 {
		ratio = 1
	}
	if ratio < 0 {
		ratio = 0
	}
	filled := int(ratio * float64(width))
	empty := width - filled
	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	label := fmt.Sprintf(" %d/%d (%d%%)", current, total, int(ratio*100))
	return tui.Text(g.Literal(bar + label))
}

// Spinner renders a braille spinner at the given frame index.
// Use with TickerView to animate.
func Spinner(frame int) g.ComponentFunc {
	idx := frame % len(spinnerFrames)
	return tui.Text(g.Literal(spinnerFrames[idx]))
}

// Tabs renders a tab bar with the active tab highlighted.
//
//	[Tab1] Tab2 Tab3
func Tabs(labels []string, activeIndex int) g.ComponentFunc {
	children := make([]g.ComponentFunc, len(labels))
	for i, label := range labels {
		if i == activeIndex {
			children[i] = tui.Text(ts.Bold(true), ts.Reverse(true), g.Literal(" "+label+" "))
		} else {
			children[i] = tui.Text(g.Literal(" "+label+" "))
		}
	}
	return tui.HBox(children...)
}

// KeyHelp renders a key binding help bar.
//
//	k increment  j decrement  q quit
func KeyHelp(bindings [][2]string) g.ComponentFunc {
	parts := make([]g.ComponentFunc, len(bindings))
	for i, b := range bindings {
		parts[i] = tui.Text(ts.Bold(true), g.Literal(b[0]+" "), g.Literal(b[1]))
	}
	return tui.HBox(append([]g.ComponentFunc{ts.Faint(true)}, parts...)...)
}

// StatusBar renders a status bar spanning the given width.
func StatusBar(left, center, right string, width int) g.ComponentFunc {
	centerPad := width - len(left) - len(right) - len(center)
	if centerPad < 0 {
		centerPad = 0
	}
	leftPad := centerPad / 2
	rightPad := centerPad - leftPad

	content := left + strings.Repeat(" ", leftPad) + center + strings.Repeat(" ", rightPad) + right
	return tui.Text(ts.Reverse(true), ts.Width(width), g.Literal(content))
}

// FormField renders a labeled form input field.
func FormField(label, value, placeholder string, focused bool) g.ComponentFunc {
	display := value
	if display == "" {
		display = placeholder
	}

	var inputStyle []g.ComponentFunc
	if focused {
		inputStyle = append(inputStyle, ts.Underline(true), ts.Bold(true))
	}
	inputStyle = append(inputStyle, g.Literal(display))

	return tui.VBox(
		tui.Text(ts.Bold(true), g.Literal(label)),
		tui.Text(inputStyle...),
	)
}
