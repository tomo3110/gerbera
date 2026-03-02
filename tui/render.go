package tui

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
	g "github.com/tomo3110/gerbera"
)

var bufPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

// Render writes the TUI representation of an Element tree to w.
func Render(w io.Writer, el *g.Element) error {
	s := renderElement(el)
	_, err := io.WriteString(w, s)
	return err
}

// RenderString returns the TUI representation of an Element tree as a string.
func RenderString(el *g.Element) string {
	return renderElement(el)
}

func renderElement(el *g.Element) string {
	style := buildStyle(el)

	switch el.TagName {
	case "vbox", "box":
		return renderVBox(el, style)
	case "hbox":
		return renderHBox(el, style)
	case "text", "header", "paragraph":
		return renderText(el, style)
	case "divider":
		return renderDivider(el, style)
	case "spacer":
		return renderSpacer(el, style)
	case "list":
		return renderList(el, style)
	case "list-item":
		return renderText(el, style)
	case "table":
		return renderTable(el, style)
	case "table-row", "table-header":
		return renderTableRow(el, style)
	case "table-cell":
		return renderText(el, style)
	case "progress":
		return renderProgress(el, style)
	case "spinner":
		return renderSpinner(el, style)
	case "statusbar":
		return renderStatusBar(el, style)
	case "input":
		return renderInput(el, style)
	case "button":
		return renderButton(el, style)
	case "checkbox":
		return renderCheckbox(el, style)
	default:
		return renderVBox(el, style)
	}
}

func renderVBox(el *g.Element, style lipgloss.Style) string {
	if len(el.Children) == 0 {
		return style.Render(el.Value)
	}
	parts := make([]string, 0, len(el.Children))
	for _, child := range el.Children {
		parts = append(parts, renderElement(child))
	}
	align := getAlignH(el)
	return style.Render(lipgloss.JoinVertical(align, parts...))
}

func renderHBox(el *g.Element, style lipgloss.Style) string {
	if len(el.Children) == 0 {
		return style.Render(el.Value)
	}
	parts := make([]string, 0, len(el.Children))
	for _, child := range el.Children {
		parts = append(parts, renderElement(child))
	}
	align := getAlignV(el)
	return style.Render(lipgloss.JoinHorizontal(align, parts...))
}

func renderText(el *g.Element, style lipgloss.Style) string {
	val := el.Value
	for _, child := range el.Children {
		val += renderElement(child)
	}
	return style.Render(val)
}

func renderDivider(el *g.Element, style lipgloss.Style) string {
	w := getAttrInt(el, "tui-width", 40)
	char := "─"
	return style.Render(strings.Repeat(char, w))
}

func renderSpacer(el *g.Element, style lipgloss.Style) string {
	h := getAttrInt(el, "tui-height", 1)
	lines := make([]string, h)
	for i := range lines {
		lines[i] = ""
	}
	return style.Render(strings.Join(lines, "\n"))
}

func renderList(el *g.Element, style lipgloss.Style) string {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)

	for i, child := range el.Children {
		if i > 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString("• ")
		buf.WriteString(renderElement(child))
	}
	return style.Render(buf.String())
}

func renderTable(el *g.Element, style lipgloss.Style) string {
	if len(el.Children) == 0 {
		return style.Render("")
	}

	// Collect all rows
	var rows [][]string
	for _, row := range el.Children {
		var cells []string
		for _, cell := range row.Children {
			cells = append(cells, getTextContent(cell))
		}
		rows = append(rows, cells)
	}

	// Find max columns and column widths
	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}
	if maxCols == 0 {
		return style.Render("")
	}

	colWidths := make([]int, maxCols)
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)

	for ri, row := range rows {
		if ri > 0 {
			buf.WriteByte('\n')
		}
		for ci := 0; ci < maxCols; ci++ {
			if ci > 0 {
				buf.WriteString("  ")
			}
			cell := ""
			if ci < len(row) {
				cell = row[ci]
			}
			buf.WriteString(cell)
			// Pad to column width
			for pad := len(cell); pad < colWidths[ci]; pad++ {
				buf.WriteByte(' ')
			}
		}
	}
	return style.Render(buf.String())
}

func renderTableRow(el *g.Element, style lipgloss.Style) string {
	parts := make([]string, 0, len(el.Children))
	for _, child := range el.Children {
		parts = append(parts, renderElement(child))
	}
	return style.Render(strings.Join(parts, "  "))
}

func renderProgress(el *g.Element, style lipgloss.Style) string {
	current := getAttrInt(el, "tui-current", 0)
	total := getAttrInt(el, "tui-total", 100)
	w := getAttrInt(el, "tui-width", 20)

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

	filled := int(ratio * float64(w))
	empty := w - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	label := fmt.Sprintf(" %d/%d (%d%%)", current, total, int(ratio*100))
	return style.Render(bar + label)
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func renderSpinner(el *g.Element, style lipgloss.Style) string {
	frame := getAttrInt(el, "tui-frame", 0)
	idx := frame % len(spinnerFrames)
	label := el.Value
	if label != "" {
		return style.Render(spinnerFrames[idx] + " " + label)
	}
	return style.Render(spinnerFrames[idx])
}

func renderStatusBar(el *g.Element, style lipgloss.Style) string {
	w := getAttrInt(el, "tui-width", 80)
	parts := make([]string, 0, len(el.Children))
	for _, child := range el.Children {
		parts = append(parts, renderElement(child))
	}
	content := strings.Join(parts, " ")
	if len(content) < w {
		content += strings.Repeat(" ", w-len(content))
	}
	return style.Render(content)
}

func renderInput(el *g.Element, style lipgloss.Style) string {
	value := el.Value
	placeholder := el.Attr["tui-placeholder"]
	if value == "" && placeholder != "" {
		return style.Render("[" + placeholder + "]")
	}
	return style.Render("[" + value + "]")
}

func renderButton(el *g.Element, style lipgloss.Style) string {
	return style.Render("[ " + el.Value + " ]")
}

func renderCheckbox(el *g.Element, style lipgloss.Style) string {
	checked := el.Attr["tui-checked"] == "true"
	label := el.Value
	if checked {
		return style.Render("[x] " + label)
	}
	return style.Render("[ ] " + label)
}

// buildStyle reads tui-* attributes and constructs a lipgloss.Style.
func buildStyle(el *g.Element) lipgloss.Style {
	s := lipgloss.NewStyle()
	if el.Attr == nil {
		return s
	}

	if v, ok := el.Attr["tui-fg"]; ok {
		s = s.Foreground(lipgloss.Color(v))
	}
	if v, ok := el.Attr["tui-bg"]; ok {
		s = s.Background(lipgloss.Color(v))
	}
	if el.Attr["tui-bold"] == "true" {
		s = s.Bold(true)
	}
	if el.Attr["tui-italic"] == "true" {
		s = s.Italic(true)
	}
	if el.Attr["tui-underline"] == "true" {
		s = s.Underline(true)
	}
	if el.Attr["tui-strikethrough"] == "true" {
		s = s.Strikethrough(true)
	}
	if el.Attr["tui-faint"] == "true" {
		s = s.Faint(true)
	}
	if el.Attr["tui-reverse"] == "true" {
		s = s.Reverse(true)
	}
	if v, ok := el.Attr["tui-width"]; ok {
		if w, err := strconv.Atoi(v); err == nil {
			s = s.Width(w)
		}
	}
	if v, ok := el.Attr["tui-height"]; ok {
		if h, err := strconv.Atoi(v); err == nil {
			s = s.Height(h)
		}
	}
	if v, ok := el.Attr["tui-max-width"]; ok {
		if w, err := strconv.Atoi(v); err == nil {
			s = s.MaxWidth(w)
		}
	}
	if v, ok := el.Attr["tui-max-height"]; ok {
		if h, err := strconv.Atoi(v); err == nil {
			s = s.MaxHeight(h)
		}
	}
	if v, ok := el.Attr["tui-padding"]; ok {
		parts := parseSpacing(v)
		s = s.PaddingTop(parts[0]).PaddingRight(parts[1]).PaddingBottom(parts[2]).PaddingLeft(parts[3])
	}
	if v, ok := el.Attr["tui-margin"]; ok {
		parts := parseSpacing(v)
		s = s.MarginTop(parts[0]).MarginRight(parts[1]).MarginBottom(parts[2]).MarginLeft(parts[3])
	}
	if v, ok := el.Attr["tui-border"]; ok {
		var border lipgloss.Border
		switch v {
		case "single":
			border = lipgloss.NormalBorder()
		case "double":
			border = lipgloss.DoubleBorder()
		case "rounded":
			border = lipgloss.RoundedBorder()
		case "thick":
			border = lipgloss.ThickBorder()
		case "hidden":
			border = lipgloss.HiddenBorder()
		default:
			border = lipgloss.NormalBorder()
		}
		s = s.BorderStyle(border).BorderTop(true).BorderBottom(true).BorderLeft(true).BorderRight(true)
	}
	if v, ok := el.Attr["tui-border-color"]; ok {
		s = s.BorderForeground(lipgloss.Color(v))
	}
	if v, ok := el.Attr["tui-align"]; ok {
		s = s.AlignHorizontal(parseAlignH(v))
	}
	if v, ok := el.Attr["tui-valign"]; ok {
		s = s.AlignVertical(parseAlignV(v))
	}

	return s
}

func parseSpacing(v string) [4]int {
	parts := strings.Fields(v)
	var result [4]int
	switch len(parts) {
	case 1:
		n, _ := strconv.Atoi(parts[0])
		result = [4]int{n, n, n, n}
	case 2:
		tb, _ := strconv.Atoi(parts[0])
		lr, _ := strconv.Atoi(parts[1])
		result = [4]int{tb, lr, tb, lr}
	case 3:
		t, _ := strconv.Atoi(parts[0])
		lr, _ := strconv.Atoi(parts[1])
		b, _ := strconv.Atoi(parts[2])
		result = [4]int{t, lr, b, lr}
	case 4:
		t, _ := strconv.Atoi(parts[0])
		r, _ := strconv.Atoi(parts[1])
		b, _ := strconv.Atoi(parts[2])
		l, _ := strconv.Atoi(parts[3])
		result = [4]int{t, r, b, l}
	}
	return result
}

func parseAlignH(v string) lipgloss.Position {
	switch v {
	case "center":
		return lipgloss.Center
	case "right":
		return lipgloss.Right
	default:
		return lipgloss.Left
	}
}

func parseAlignV(v string) lipgloss.Position {
	switch v {
	case "middle", "center":
		return lipgloss.Center
	case "bottom":
		return lipgloss.Bottom
	default:
		return lipgloss.Top
	}
}

func getAlignH(el *g.Element) lipgloss.Position {
	if el.Attr != nil {
		if v, ok := el.Attr["tui-align"]; ok {
			return parseAlignH(v)
		}
	}
	return lipgloss.Left
}

func getAlignV(el *g.Element) lipgloss.Position {
	if el.Attr != nil {
		if v, ok := el.Attr["tui-valign"]; ok {
			return parseAlignV(v)
		}
	}
	return lipgloss.Top
}

func getAttrInt(el *g.Element, key string, defaultVal int) int {
	if el.Attr != nil {
		if v, ok := el.Attr[key]; ok {
			if n, err := strconv.Atoi(v); err == nil {
				return n
			}
		}
	}
	return defaultVal
}

func getTextContent(el *g.Element) string {
	if el.Value != "" {
		return el.Value
	}
	parts := make([]string, 0, len(el.Children))
	for _, child := range el.Children {
		parts = append(parts, getTextContent(child))
	}
	return strings.Join(parts, "")
}
