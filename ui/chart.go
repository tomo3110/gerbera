package ui

import (
	"fmt"
	"math"
	"sort"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/property"
)

// DataPoint is a single labeled value.
type DataPoint struct {
	Label string
	Value float64
}

// Series is a named data set.
type Series struct {
	Name   string
	Color  string // empty → auto palette
	Points []DataPoint
}

// ChartOpts holds common chart options.
type ChartOpts struct {
	Width       int    // SVG width (default 600)
	Height      int    // SVG height (default 400)
	Title       string
	ShowGrid    bool
	ShowLegend  bool
	ShowTooltip bool // <title> native tooltip
	XLabel      string
	YLabel      string
	Colors      []string // color palette override
}

// HistogramOpts holds histogram-specific settings.
type HistogramOpts struct {
	ChartOpts
	BinCount int // number of bins (default 10)
}

// Default color palette (10 colors).
var defaultChartColors = []string{
	"#3b82f6", "#ef4444", "#22c55e", "#f59e0b", "#8b5cf6",
	"#06b6d4", "#ec4899", "#14b8a6", "#f97316", "#6366f1",
}

// ---------- linear scale ----------

type linearScale struct {
	Min, Max           float64
	RangeStart, RangeEnd int
}

func newLinearScale(dataMin, dataMax float64, rangeStart, rangeEnd int) linearScale {
	if dataMax == dataMin {
		dataMax = dataMin + 1
	}
	return linearScale{Min: dataMin, Max: dataMax, RangeStart: rangeStart, RangeEnd: rangeEnd}
}

func (s linearScale) Apply(val float64) float64 {
	ratio := (val - s.Min) / (s.Max - s.Min)
	return float64(s.RangeStart) + ratio*float64(s.RangeEnd-s.RangeStart)
}

// ---------- helpers ----------

func chartDefaults(opts *ChartOpts) {
	if opts.Width == 0 {
		opts.Width = 600
	}
	if opts.Height == 0 {
		opts.Height = 400
	}
	if len(opts.Colors) == 0 {
		opts.Colors = defaultChartColors
	}
}

func seriesColor(opts ChartOpts, s Series, idx int) string {
	if s.Color != "" {
		return s.Color
	}
	return opts.Colors[idx%len(opts.Colors)]
}

func computeMinMax(series []Series) (min, max float64) {
	first := true
	for _, s := range series {
		for _, p := range s.Points {
			if first || p.Value < min {
				min = p.Value
			}
			if first || p.Value > max {
				max = p.Value
			}
			first = false
		}
	}
	if min > 0 {
		min = 0
	}
	if max == min {
		max = min + 1
	}
	return
}

func generateTicks(min, max float64, count int) []float64 {
	if count <= 0 {
		count = 5
	}
	step := (max - min) / float64(count)
	ticks := make([]float64, 0, count+1)
	for i := 0; i <= count; i++ {
		ticks = append(ticks, min+float64(i)*step)
	}
	return ticks
}

type chartPadding struct {
	Top, Left, Bottom, Right int
}

func chartPad(opts ChartOpts) chartPadding {
	p := chartPadding{Top: 40, Left: 60, Bottom: 50, Right: 20}
	if opts.Title != "" {
		p.Top = 55
	}
	if opts.YLabel != "" {
		p.Left = 75
	}
	if opts.ShowLegend {
		p.Right = 140
	}
	return p
}

// ---------- SVG element builders ----------

func svgRoot(w, h int, children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := []gerbera.ComponentFunc{
		property.Class("g-chart"),
		property.Attr("viewBox", fmt.Sprintf("0 0 %d %d", w, h)),
		property.Attr("xmlns", "http://www.w3.org/2000/svg"),
	}
	attrs = append(attrs, children...)
	return gerbera.Tag("svg", attrs...)
}

func svgG(class string, children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := []gerbera.ComponentFunc{property.Class(class)}
	attrs = append(attrs, children...)
	return gerbera.Tag("g", attrs...)
}

func svgRect(x, y, w, h float64, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := []gerbera.ComponentFunc{
		property.Attr("x", ff(x)),
		property.Attr("y", ff(y)),
		property.Attr("width", ff(w)),
		property.Attr("height", ff(h)),
	}
	attrs = append(attrs, extra...)
	return gerbera.Tag("rect", attrs...)
}

func svgCircle(cx, cy, r float64, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := []gerbera.ComponentFunc{
		property.Attr("cx", ff(cx)),
		property.Attr("cy", ff(cy)),
		property.Attr("r", ff(r)),
	}
	attrs = append(attrs, extra...)
	return gerbera.Tag("circle", attrs...)
}

func svgLine(x1, y1, x2, y2 float64, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := []gerbera.ComponentFunc{
		property.Attr("x1", ff(x1)),
		property.Attr("y1", ff(y1)),
		property.Attr("x2", ff(x2)),
		property.Attr("y2", ff(y2)),
	}
	attrs = append(attrs, extra...)
	return gerbera.Tag("line", attrs...)
}

func svgText(x, y float64, text string, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := []gerbera.ComponentFunc{
		property.Attr("x", ff(x)),
		property.Attr("y", ff(y)),
		property.Value(text),
	}
	attrs = append(attrs, extra...)
	return gerbera.Tag("text", attrs...)
}

func svgPath(d string, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := []gerbera.ComponentFunc{
		property.Attr("d", d),
	}
	attrs = append(attrs, extra...)
	return gerbera.Tag("path", attrs...)
}

func svgPolyline(points string, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := []gerbera.ComponentFunc{
		property.Attr("points", points),
	}
	attrs = append(attrs, extra...)
	return gerbera.Tag("polyline", attrs...)
}

func svgTitle(text string) gerbera.ComponentFunc {
	return gerbera.Tag("title", property.Value(text))
}

// ff formats a float for SVG attributes.
func ff(v float64) string {
	return fmt.Sprintf("%.1f", v)
}

// ---------- grid / axes / legend / title ----------

func renderGrid(opts ChartOpts, pad chartPadding, yScale linearScale, ticks []float64) gerbera.ComponentFunc {
	if !opts.ShowGrid {
		return nil
	}
	plotW := float64(opts.Width - pad.Left - pad.Right)
	var lines []gerbera.ComponentFunc
	for _, t := range ticks {
		y := yScale.Apply(t)
		lines = append(lines, svgLine(float64(pad.Left), y, float64(pad.Left)+plotW, y))
	}
	return svgG("g-chart-grid", lines...)
}

func renderAxes(opts ChartOpts, pad chartPadding, yScale linearScale, ticks []float64, xLabels []string) gerbera.ComponentFunc {
	plotW := float64(opts.Width - pad.Left - pad.Right)
	var children []gerbera.ComponentFunc

	// Y axis line
	children = append(children, svgLine(
		float64(pad.Left), float64(pad.Top),
		float64(pad.Left), float64(opts.Height-pad.Bottom),
	))
	// X axis line
	children = append(children, svgLine(
		float64(pad.Left), float64(opts.Height-pad.Bottom),
		float64(pad.Left)+plotW, float64(opts.Height-pad.Bottom),
	))

	// Y axis tick labels
	for _, t := range ticks {
		y := yScale.Apply(t)
		children = append(children, svgText(float64(pad.Left-8), y+4, fmt.Sprintf("%.0f", t),
			property.Attr("text-anchor", "end"),
		))
	}

	// X axis labels
	if len(xLabels) > 0 {
		step := plotW / float64(len(xLabels))
		for i, lbl := range xLabels {
			x := float64(pad.Left) + step*float64(i) + step/2
			children = append(children, svgText(x, float64(opts.Height-pad.Bottom+20), lbl,
				property.Attr("text-anchor", "middle"),
			))
		}
	}

	// Y label
	if opts.YLabel != "" {
		children = append(children, svgText(15, float64(opts.Height/2), opts.YLabel,
			property.Attr("text-anchor", "middle"),
			property.Attr("transform", fmt.Sprintf("rotate(-90,15,%d)", opts.Height/2)),
		))
	}

	// X label
	if opts.XLabel != "" {
		children = append(children, svgText(float64(opts.Width/2), float64(opts.Height-5), opts.XLabel,
			property.Attr("text-anchor", "middle"),
		))
	}

	return svgG("g-chart-axes", children...)
}

func renderLegend(series []Series, opts ChartOpts, pad chartPadding) gerbera.ComponentFunc {
	if !opts.ShowLegend || len(series) == 0 {
		return nil
	}
	x := float64(opts.Width - pad.Right + 10)
	var items []gerbera.ComponentFunc
	for i, s := range series {
		y := float64(pad.Top + i*20)
		color := seriesColor(opts, s, i)
		items = append(items,
			svgRect(x, y, 12, 12, property.Attr("fill", color), property.Attr("rx", "2")),
			svgText(x+16, y+10, s.Name, property.Attr("font-size", "11")),
		)
	}
	return svgG("g-chart-legend", items...)
}

func renderChartTitle(title string, w int) gerbera.ComponentFunc {
	if title == "" {
		return nil
	}
	return svgText(float64(w/2), 24, title, property.Class("g-chart-title"))
}

// ---------- empty data fallback ----------

func renderEmpty(opts ChartOpts, extra []gerbera.ComponentFunc) gerbera.ComponentFunc {
	children := []gerbera.ComponentFunc{
		svgText(float64(opts.Width/2), float64(opts.Height/2), "No data",
			property.Attr("text-anchor", "middle"),
			property.Attr("fill", "var(--g-text-tertiary)"),
			property.Attr("font-size", "14"),
		),
	}
	children = append(children, extra...)
	return svgRoot(opts.Width, opts.Height, children...)
}

// ---------- unique X labels from series ----------

func collectXLabels(series []Series) []string {
	seen := map[string]bool{}
	var labels []string
	for _, s := range series {
		for _, p := range s.Points {
			if !seen[p.Label] {
				seen[p.Label] = true
				labels = append(labels, p.Label)
			}
		}
	}
	return labels
}

// ---------- describeArc for PieChart ----------

func describeArc(cx, cy, r, startAngle, endAngle float64) string {
	x1 := cx + r*math.Cos(startAngle)
	y1 := cy + r*math.Sin(startAngle)
	x2 := cx + r*math.Cos(endAngle)
	y2 := cy + r*math.Sin(endAngle)
	largeArc := 0
	if endAngle-startAngle > math.Pi {
		largeArc = 1
	}
	return fmt.Sprintf("M %s %s L %s %s A %s %s 0 %d 1 %s %s Z",
		ff(cx), ff(cy), ff(x1), ff(y1), ff(r), ff(r), largeArc, ff(x2), ff(y2))
}

// ========== Public chart functions ==========

// LineChart renders a line chart with optional data points.
func LineChart(series []Series, opts ChartOpts, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	chartDefaults(&opts)
	if len(series) == 0 {
		return renderEmpty(opts, extra)
	}

	pad := chartPad(opts)
	xLabels := collectXLabels(series)
	if len(xLabels) == 0 {
		return renderEmpty(opts, extra)
	}
	min, max := computeMinMax(series)
	ticks := generateTicks(min, max, 5)
	yScale := newLinearScale(min, max, opts.Height-pad.Bottom, pad.Top)
	plotW := float64(opts.Width - pad.Left - pad.Right)
	step := plotW / float64(len(xLabels))

	// label → x index
	labelIdx := map[string]int{}
	for i, l := range xLabels {
		labelIdx[l] = i
	}

	var children []gerbera.ComponentFunc
	if g := renderGrid(opts, pad, yScale, ticks); g != nil {
		children = append(children, g)
	}
	children = append(children, renderAxes(opts, pad, yScale, ticks, xLabels))

	// Data lines
	var dataChildren []gerbera.ComponentFunc
	for si, s := range series {
		color := seriesColor(opts, s, si)
		pts := ""
		var circles []gerbera.ComponentFunc
		for _, p := range s.Points {
			idx := labelIdx[p.Label]
			x := float64(pad.Left) + step*float64(idx) + step/2
			y := yScale.Apply(p.Value)
			if pts != "" {
				pts += " "
			}
			pts += fmt.Sprintf("%s,%s", ff(x), ff(y))
			circleExtra := []gerbera.ComponentFunc{
				property.Class("g-chart-point"),
				property.Attr("fill", color),
			}
			if opts.ShowTooltip {
				circleExtra = append(circleExtra, svgTitle(fmt.Sprintf("%s: %s = %s", s.Name, p.Label, ff(p.Value))))
			}
			circles = append(circles, svgCircle(x, y, 4, circleExtra...))
		}
		dataChildren = append(dataChildren,
			svgPolyline(pts,
				property.Class("g-chart-line"),
				property.Attr("stroke", color),
			),
		)
		dataChildren = append(dataChildren, circles...)
	}
	children = append(children, svgG("g-chart-data", dataChildren...))

	if l := renderLegend(series, opts, pad); l != nil {
		children = append(children, l)
	}
	if t := renderChartTitle(opts.Title, opts.Width); t != nil {
		children = append(children, t)
	}
	children = append(children, extra...)
	return svgRoot(opts.Width, opts.Height, children...)
}

// ColumnChart renders a vertical bar chart.
func ColumnChart(series []Series, opts ChartOpts, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	chartDefaults(&opts)
	if len(series) == 0 {
		return renderEmpty(opts, extra)
	}

	pad := chartPad(opts)
	xLabels := collectXLabels(series)
	if len(xLabels) == 0 {
		return renderEmpty(opts, extra)
	}
	min, max := computeMinMax(series)
	ticks := generateTicks(min, max, 5)
	yScale := newLinearScale(min, max, opts.Height-pad.Bottom, pad.Top)
	plotW := float64(opts.Width - pad.Left - pad.Right)
	groupW := plotW / float64(len(xLabels))
	barW := groupW * 0.7 / float64(len(series))
	baseY := yScale.Apply(0)

	labelIdx := map[string]int{}
	for i, l := range xLabels {
		labelIdx[l] = i
	}

	var children []gerbera.ComponentFunc
	if g := renderGrid(opts, pad, yScale, ticks); g != nil {
		children = append(children, g)
	}
	children = append(children, renderAxes(opts, pad, yScale, ticks, xLabels))

	var dataChildren []gerbera.ComponentFunc
	for si, s := range series {
		color := seriesColor(opts, s, si)
		for _, p := range s.Points {
			idx := labelIdx[p.Label]
			groupX := float64(pad.Left) + groupW*float64(idx) + groupW*0.15
			x := groupX + barW*float64(si)
			y := yScale.Apply(p.Value)
			h := baseY - y
			if h < 0 {
				y = baseY
				h = -h
			}
			rectExtra := []gerbera.ComponentFunc{
				property.Attr("fill", color),
				property.Attr("rx", "2"),
			}
			if opts.ShowTooltip {
				rectExtra = append(rectExtra, svgTitle(fmt.Sprintf("%s: %s = %s", s.Name, p.Label, ff(p.Value))))
			}
			dataChildren = append(dataChildren, svgRect(x, y, barW, h, rectExtra...))
		}
	}
	children = append(children, svgG("g-chart-data", dataChildren...))

	if l := renderLegend(series, opts, pad); l != nil {
		children = append(children, l)
	}
	if t := renderChartTitle(opts.Title, opts.Width); t != nil {
		children = append(children, t)
	}
	children = append(children, extra...)
	return svgRoot(opts.Width, opts.Height, children...)
}

// BarChart renders a horizontal bar chart (X/Y swapped from ColumnChart).
func BarChart(series []Series, opts ChartOpts, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	chartDefaults(&opts)
	if len(series) == 0 {
		return renderEmpty(opts, extra)
	}

	pad := chartPad(opts)
	yLabels := collectXLabels(series)
	if len(yLabels) == 0 {
		return renderEmpty(opts, extra)
	}
	min, max := computeMinMax(series)
	ticks := generateTicks(min, max, 5)
	xScale := newLinearScale(min, max, pad.Left, opts.Width-pad.Right)
	plotH := float64(opts.Height - pad.Top - pad.Bottom)
	groupH := plotH / float64(len(yLabels))
	barH := groupH * 0.7 / float64(len(series))
	baseX := xScale.Apply(0)

	labelIdx := map[string]int{}
	for i, l := range yLabels {
		labelIdx[l] = i
	}

	var children []gerbera.ComponentFunc

	// Horizontal grid
	if opts.ShowGrid {
		var gridLines []gerbera.ComponentFunc
		for _, t := range ticks {
			x := xScale.Apply(t)
			gridLines = append(gridLines, svgLine(x, float64(pad.Top), x, float64(opts.Height-pad.Bottom)))
		}
		children = append(children, svgG("g-chart-grid", gridLines...))
	}

	// Axes
	var axisChildren []gerbera.ComponentFunc
	axisChildren = append(axisChildren, svgLine(
		float64(pad.Left), float64(pad.Top),
		float64(pad.Left), float64(opts.Height-pad.Bottom),
	))
	axisChildren = append(axisChildren, svgLine(
		float64(pad.Left), float64(opts.Height-pad.Bottom),
		float64(opts.Width-pad.Right), float64(opts.Height-pad.Bottom),
	))
	// X tick labels
	for _, t := range ticks {
		x := xScale.Apply(t)
		axisChildren = append(axisChildren, svgText(x, float64(opts.Height-pad.Bottom+20), fmt.Sprintf("%.0f", t),
			property.Attr("text-anchor", "middle"),
		))
	}
	// Y labels
	for i, lbl := range yLabels {
		y := float64(pad.Top) + groupH*float64(i) + groupH/2
		axisChildren = append(axisChildren, svgText(float64(pad.Left-8), y+4, lbl,
			property.Attr("text-anchor", "end"),
		))
	}
	children = append(children, svgG("g-chart-axes", axisChildren...))

	// Data bars
	var dataChildren []gerbera.ComponentFunc
	for si, s := range series {
		color := seriesColor(opts, s, si)
		for _, p := range s.Points {
			idx := labelIdx[p.Label]
			groupY := float64(pad.Top) + groupH*float64(idx) + groupH*0.15
			y := groupY + barH*float64(si)
			x := xScale.Apply(p.Value)
			w := x - baseX
			rx := baseX
			if w < 0 {
				rx = x
				w = -w
			}
			rectExtra := []gerbera.ComponentFunc{
				property.Attr("fill", color),
				property.Attr("rx", "2"),
			}
			if opts.ShowTooltip {
				rectExtra = append(rectExtra, svgTitle(fmt.Sprintf("%s: %s = %s", s.Name, p.Label, ff(p.Value))))
			}
			dataChildren = append(dataChildren, svgRect(rx, y, w, barH, rectExtra...))
		}
	}
	children = append(children, svgG("g-chart-data", dataChildren...))

	if l := renderLegend(series, opts, pad); l != nil {
		children = append(children, l)
	}
	if t := renderChartTitle(opts.Title, opts.Width); t != nil {
		children = append(children, t)
	}
	children = append(children, extra...)
	return svgRoot(opts.Width, opts.Height, children...)
}

// PieChart renders a pie chart.
func PieChart(data []DataPoint, opts ChartOpts, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	chartDefaults(&opts)
	if len(data) == 0 {
		return renderEmpty(opts, extra)
	}

	total := 0.0
	for _, d := range data {
		total += d.Value
	}
	if total == 0 {
		return renderEmpty(opts, extra)
	}

	cx := float64(opts.Width) / 2
	cy := float64(opts.Height) / 2
	r := math.Min(cx, cy) - 40

	var dataChildren []gerbera.ComponentFunc

	// Single slice 100% → use circle
	if len(data) == 1 {
		color := opts.Colors[0]
		sliceExtra := []gerbera.ComponentFunc{
			property.Class("g-chart-slice"),
			property.Attr("fill", color),
		}
		if opts.ShowTooltip {
			sliceExtra = append(sliceExtra, svgTitle(fmt.Sprintf("%s = %s", data[0].Label, ff(data[0].Value))))
		}
		dataChildren = append(dataChildren, svgCircle(cx, cy, r, sliceExtra...))
	} else {
		startAngle := -math.Pi / 2
		for i, d := range data {
			color := opts.Colors[i%len(opts.Colors)]
			sweep := (d.Value / total) * 2 * math.Pi
			endAngle := startAngle + sweep

			pathD := describeArc(cx, cy, r, startAngle, endAngle)
			sliceExtra := []gerbera.ComponentFunc{
				property.Class("g-chart-slice"),
				property.Attr("fill", color),
			}
			if opts.ShowTooltip {
				pct := d.Value / total * 100
				sliceExtra = append(sliceExtra, svgTitle(fmt.Sprintf("%s = %s (%.0f%%)", d.Label, ff(d.Value), pct)))
			}
			dataChildren = append(dataChildren, svgPath(pathD, sliceExtra...))
			startAngle = endAngle
		}
	}

	var children []gerbera.ComponentFunc
	children = append(children, svgG("g-chart-data", dataChildren...))

	// Legend for pie
	if opts.ShowLegend {
		var legendItems []gerbera.ComponentFunc
		lx := float64(opts.Width) - 130.0
		for i, d := range data {
			ly := float64(40 + i*20)
			color := opts.Colors[i%len(opts.Colors)]
			legendItems = append(legendItems,
				svgRect(lx, ly, 12, 12, property.Attr("fill", color), property.Attr("rx", "2")),
				svgText(lx+16, ly+10, d.Label, property.Attr("font-size", "11")),
			)
		}
		children = append(children, svgG("g-chart-legend", legendItems...))
	}

	if t := renderChartTitle(opts.Title, opts.Width); t != nil {
		children = append(children, t)
	}
	children = append(children, extra...)
	return svgRoot(opts.Width, opts.Height, children...)
}

// ScatterPlot renders a scatter chart.
func ScatterPlot(series []Series, opts ChartOpts, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	chartDefaults(&opts)
	if len(series) == 0 {
		return renderEmpty(opts, extra)
	}

	pad := chartPad(opts)
	xLabels := collectXLabels(series)
	if len(xLabels) == 0 {
		return renderEmpty(opts, extra)
	}
	min, max := computeMinMax(series)
	ticks := generateTicks(min, max, 5)
	yScale := newLinearScale(min, max, opts.Height-pad.Bottom, pad.Top)
	plotW := float64(opts.Width - pad.Left - pad.Right)
	step := plotW / float64(len(xLabels))

	labelIdx := map[string]int{}
	for i, l := range xLabels {
		labelIdx[l] = i
	}

	var children []gerbera.ComponentFunc
	if g := renderGrid(opts, pad, yScale, ticks); g != nil {
		children = append(children, g)
	}
	children = append(children, renderAxes(opts, pad, yScale, ticks, xLabels))

	var dataChildren []gerbera.ComponentFunc
	for si, s := range series {
		color := seriesColor(opts, s, si)
		for _, p := range s.Points {
			idx := labelIdx[p.Label]
			x := float64(pad.Left) + step*float64(idx) + step/2
			y := yScale.Apply(p.Value)
			circleExtra := []gerbera.ComponentFunc{
				property.Class("g-chart-point"),
				property.Attr("fill", color),
			}
			if opts.ShowTooltip {
				circleExtra = append(circleExtra, svgTitle(fmt.Sprintf("%s: %s = %s", s.Name, p.Label, ff(p.Value))))
			}
			dataChildren = append(dataChildren, svgCircle(x, y, 5, circleExtra...))
		}
	}
	children = append(children, svgG("g-chart-data", dataChildren...))

	if l := renderLegend(series, opts, pad); l != nil {
		children = append(children, l)
	}
	if t := renderChartTitle(opts.Title, opts.Width); t != nil {
		children = append(children, t)
	}
	children = append(children, extra...)
	return svgRoot(opts.Width, opts.Height, children...)
}

// Histogram renders a histogram from raw values.
func Histogram(values []float64, opts HistogramOpts, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	chartDefaults(&opts.ChartOpts)
	if len(values) == 0 {
		return renderEmpty(opts.ChartOpts, extra)
	}
	binCount := opts.BinCount
	if binCount <= 0 {
		binCount = 10
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)
	vMin, vMax := sorted[0], sorted[len(sorted)-1]
	if vMax == vMin {
		vMax = vMin + 1
	}
	binWidth := (vMax - vMin) / float64(binCount)

	// Build bins
	bins := make([]DataPoint, binCount)
	counts := make([]float64, binCount)
	for _, v := range sorted {
		idx := int((v - vMin) / binWidth)
		if idx >= binCount {
			idx = binCount - 1
		}
		counts[idx]++
	}
	for i := 0; i < binCount; i++ {
		lo := vMin + float64(i)*binWidth
		bins[i] = DataPoint{
			Label: fmt.Sprintf("%.0f-%.0f", lo, lo+binWidth),
			Value: counts[i],
		}
	}

	// Reuse ColumnChart
	s := Series{Name: "Frequency", Points: bins}
	return ColumnChart([]Series{s}, opts.ChartOpts, extra...)
}

// StackedBarChart renders a horizontal stacked bar chart.
func StackedBarChart(series []Series, opts ChartOpts, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	chartDefaults(&opts)
	if len(series) == 0 {
		return renderEmpty(opts, extra)
	}

	pad := chartPad(opts)
	yLabels := collectXLabels(series)
	if len(yLabels) == 0 {
		return renderEmpty(opts, extra)
	}

	// Compute stacked max
	labelIdx := map[string]int{}
	for i, l := range yLabels {
		labelIdx[l] = i
	}
	stackMax := make([]float64, len(yLabels))
	for _, s := range series {
		for _, p := range s.Points {
			idx := labelIdx[p.Label]
			stackMax[idx] += p.Value
		}
	}
	max := 0.0
	for _, v := range stackMax {
		if v > max {
			max = v
		}
	}
	if max == 0 {
		max = 1
	}

	ticks := generateTicks(0, max, 5)
	xScale := newLinearScale(0, max, pad.Left, opts.Width-pad.Right)
	plotH := float64(opts.Height - pad.Top - pad.Bottom)
	barH := plotH / float64(len(yLabels)) * 0.7

	var children []gerbera.ComponentFunc

	// Grid
	if opts.ShowGrid {
		var gridLines []gerbera.ComponentFunc
		for _, t := range ticks {
			x := xScale.Apply(t)
			gridLines = append(gridLines, svgLine(x, float64(pad.Top), x, float64(opts.Height-pad.Bottom)))
		}
		children = append(children, svgG("g-chart-grid", gridLines...))
	}

	// Axes
	var axisChildren []gerbera.ComponentFunc
	axisChildren = append(axisChildren, svgLine(
		float64(pad.Left), float64(pad.Top),
		float64(pad.Left), float64(opts.Height-pad.Bottom),
	))
	axisChildren = append(axisChildren, svgLine(
		float64(pad.Left), float64(opts.Height-pad.Bottom),
		float64(opts.Width-pad.Right), float64(opts.Height-pad.Bottom),
	))
	for _, t := range ticks {
		x := xScale.Apply(t)
		axisChildren = append(axisChildren, svgText(x, float64(opts.Height-pad.Bottom+20), fmt.Sprintf("%.0f", t),
			property.Attr("text-anchor", "middle"),
		))
	}
	rowH := plotH / float64(len(yLabels))
	for i, lbl := range yLabels {
		y := float64(pad.Top) + rowH*float64(i) + rowH/2
		axisChildren = append(axisChildren, svgText(float64(pad.Left-8), y+4, lbl,
			property.Attr("text-anchor", "end"),
		))
	}
	children = append(children, svgG("g-chart-axes", axisChildren...))

	// Stacked data
	offsets := make([]float64, len(yLabels))
	var dataChildren []gerbera.ComponentFunc
	for si, s := range series {
		color := seriesColor(opts, s, si)
		for _, p := range s.Points {
			idx := labelIdx[p.Label]
			y := float64(pad.Top) + rowH*float64(idx) + (rowH-barH)/2
			x := xScale.Apply(offsets[idx])
			w := xScale.Apply(offsets[idx]+p.Value) - x
			rectExtra := []gerbera.ComponentFunc{
				property.Attr("fill", color),
				property.Attr("rx", "2"),
			}
			if opts.ShowTooltip {
				rectExtra = append(rectExtra, svgTitle(fmt.Sprintf("%s: %s = %s", s.Name, p.Label, ff(p.Value))))
			}
			dataChildren = append(dataChildren, svgRect(x, y, w, barH, rectExtra...))
			offsets[idx] += p.Value
		}
	}
	children = append(children, svgG("g-chart-data", dataChildren...))

	if l := renderLegend(series, opts, pad); l != nil {
		children = append(children, l)
	}
	if t := renderChartTitle(opts.Title, opts.Width); t != nil {
		children = append(children, t)
	}
	children = append(children, extra...)
	return svgRoot(opts.Width, opts.Height, children...)
}

