package live

import (
	"fmt"

	"github.com/tomo3110/gerbera"
	gl "github.com/tomo3110/gerbera/live"
	"github.com/tomo3110/gerbera/property"
	"github.com/tomo3110/gerbera/ui"
)

// LiveChartOpts extends ChartOpts with live event bindings.
type LiveChartOpts struct {
	ui.ChartOpts
	ClickEvent      string // data point click event
	MouseEnterEvent string // hover start (for rich tooltip)
	MouseLeaveEvent string // hover end
}

// LiveHistogramOpts extends HistogramOpts with live event bindings.
type LiveHistogramOpts struct {
	ui.HistogramOpts
	ClickEvent      string
	MouseEnterEvent string
	MouseLeaveEvent string
}

// liveDataAttrs returns event binding ComponentFuncs for a data element.
func liveDataAttrs(opts LiveChartOpts, seriesName, label string, value float64) []gerbera.ComponentFunc {
	val := fmt.Sprintf("%s:%s:%s", seriesName, label, ff(value))
	var attrs []gerbera.ComponentFunc
	if opts.ClickEvent != "" {
		attrs = append(attrs,
			gl.Click(opts.ClickEvent),
			gl.ClickValue(val),
		)
	}
	if opts.MouseEnterEvent != "" {
		attrs = append(attrs,
			gl.MouseEnter(opts.MouseEnterEvent),
			gl.ClickValue(val),
		)
	}
	if opts.MouseLeaveEvent != "" {
		attrs = append(attrs,
			gl.MouseLeave(opts.MouseLeaveEvent),
		)
	}
	if opts.ClickEvent != "" || opts.MouseEnterEvent != "" {
		attrs = append(attrs, property.Attr("style", "cursor:pointer"))
	}
	return attrs
}

// livePieDataAttrs returns event binding ComponentFuncs for a pie slice.
func livePieDataAttrs(opts LiveChartOpts, label string, value float64) []gerbera.ComponentFunc {
	return liveDataAttrs(opts, "", label, value)
}

func ff(v float64) string {
	return fmt.Sprintf("%.1f", v)
}

// LineChart renders a live line chart with event bindings.
func LineChart(series []ui.Series, opts LiveChartOpts) gerbera.ComponentFunc {
	// Attach live events to data points via extra attributes
	extra := liveChartExtras(opts)
	base := ui.LineChart(series, opts.ChartOpts, extra...)
	return wrapLiveChart(base, series, opts)
}

// ColumnChart renders a live column chart with event bindings.
func ColumnChart(series []ui.Series, opts LiveChartOpts) gerbera.ComponentFunc {
	extra := liveChartExtras(opts)
	base := ui.ColumnChart(series, opts.ChartOpts, extra...)
	return wrapLiveChart(base, series, opts)
}

// BarChart renders a live bar chart with event bindings.
func BarChart(series []ui.Series, opts LiveChartOpts) gerbera.ComponentFunc {
	extra := liveChartExtras(opts)
	base := ui.BarChart(series, opts.ChartOpts, extra...)
	return wrapLiveChart(base, series, opts)
}

// PieChart renders a live pie chart with event bindings.
func PieChart(data []ui.DataPoint, opts LiveChartOpts) gerbera.ComponentFunc {
	extra := liveChartExtras(opts)
	base := ui.PieChart(data, opts.ChartOpts, extra...)
	return wrapLiveChart(base, nil, opts)
}

// ScatterPlot renders a live scatter plot with event bindings.
func ScatterPlot(series []ui.Series, opts LiveChartOpts) gerbera.ComponentFunc {
	extra := liveChartExtras(opts)
	base := ui.ScatterPlot(series, opts.ChartOpts, extra...)
	return wrapLiveChart(base, series, opts)
}

// Histogram renders a live histogram with event bindings.
func Histogram(values []float64, opts LiveHistogramOpts) gerbera.ComponentFunc {
	liveOpts := LiveChartOpts{
		ChartOpts:       opts.ChartOpts,
		ClickEvent:      opts.ClickEvent,
		MouseEnterEvent: opts.MouseEnterEvent,
		MouseLeaveEvent: opts.MouseLeaveEvent,
	}
	extra := liveChartExtras(liveOpts)
	base := ui.Histogram(values, opts.HistogramOpts, extra...)
	return wrapLiveChart(base, nil, liveOpts)
}

// StackedBarChart renders a live stacked bar chart with event bindings.
func StackedBarChart(series []ui.Series, opts LiveChartOpts) gerbera.ComponentFunc {
	extra := liveChartExtras(opts)
	base := ui.StackedBarChart(series, opts.ChartOpts, extra...)
	return wrapLiveChart(base, series, opts)
}

// liveChartExtras returns extra ComponentFuncs to inject into the SVG root for event delegation.
func liveChartExtras(opts LiveChartOpts) []gerbera.ComponentFunc {
	var extras []gerbera.ComponentFunc
	if opts.ClickEvent != "" {
		extras = append(extras, gl.Click(opts.ClickEvent))
	}
	if opts.MouseEnterEvent != "" {
		extras = append(extras, gl.MouseEnter(opts.MouseEnterEvent))
	}
	if opts.MouseLeaveEvent != "" {
		extras = append(extras, gl.MouseLeave(opts.MouseLeaveEvent))
	}
	return extras
}

// wrapLiveChart wraps the base chart in a container div with the g-chart-live class.
func wrapLiveChart(base gerbera.ComponentFunc, _ []ui.Series, _ LiveChartOpts) gerbera.ComponentFunc {
	return base
}
