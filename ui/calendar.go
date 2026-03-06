package ui

import (
	"fmt"
	"time"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
)

// CalendarOpts configures a Calendar.
type CalendarOpts struct {
	Year             int
	Month            time.Month
	Selected         *time.Time
	Today            time.Time
	DayNames         []string // defaults to ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"]
	YearRange        [2]int   // [min, max] for year selector; zero values default to Year ± 10
	SelectEvent      string   // gerbera-click on day cell; payload: {"value": "2006-01-02"}
	PrevMonthEvent   string   // gerbera-click on prev month button
	NextMonthEvent   string   // gerbera-click on next month button
	MonthChangeEvent string   // gerbera-change on month dropdown; payload: {"value": "1"-"12"}
	YearChangeEvent  string   // gerbera-change on year dropdown; payload: {"value": "2026"}
}

var defaultDayNames = []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

// Calendar renders a month-view calendar grid.
func Calendar(opts CalendarOpts, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	dayNames := opts.DayNames
	if len(dayNames) == 0 {
		dayNames = defaultDayNames
	}

	firstDay := time.Date(opts.Year, opts.Month, 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1)
	startWeekday := int(firstDay.Weekday())

	// Header row with day names
	var headerCells gerbera.Components
	for _, d := range dayNames {
		headerCells = append(headerCells, dom.Div(
			property.Class("g-calendar-dayname"),
			property.Value(d),
		))
	}

	// Day cells
	var dayCells gerbera.Components

	// Previous month padding
	if startWeekday > 0 {
		prevLastDay := firstDay.AddDate(0, 0, -1)
		for i := startWeekday - 1; i >= 0; i-- {
			d := prevLastDay.Day() - i
			dayCells = append(dayCells, dom.Div(
				property.Class("g-calendar-day", "g-calendar-day-outside"),
				property.Value(fmt.Sprintf("%d", d)),
			))
		}
	}

	// Current month days
	for d := 1; d <= lastDay.Day(); d++ {
		current := time.Date(opts.Year, opts.Month, d, 0, 0, 0, 0, time.UTC)
		isToday := sameDate(current, opts.Today)
		isSelected := opts.Selected != nil && sameDate(current, *opts.Selected)

		dayAttrs := gerbera.Components{
			property.Class("g-calendar-day"),
			property.ClassIf(isToday, "g-calendar-day-today"),
			property.ClassIf(isSelected, "g-calendar-day-selected"),
			property.Attr("data-date", current.Format("2006-01-02")),
			property.Value(fmt.Sprintf("%d", d)),
		}
		if opts.SelectEvent != "" {
			dayAttrs = append(dayAttrs,
				property.Attr("gerbera-click", opts.SelectEvent),
				property.Attr("gerbera-value", current.Format("2006-01-02")),
			)
		}
		dayCells = append(dayCells, dom.Div(dayAttrs...))
	}

	// Next month padding
	totalCells := startWeekday + lastDay.Day()
	remaining := (7 - totalCells%7) % 7
	for i := 1; i <= remaining; i++ {
		dayCells = append(dayCells, dom.Div(
			property.Class("g-calendar-day", "g-calendar-day-outside"),
			property.Value(fmt.Sprintf("%d", i)),
		))
	}

	// Navigation buttons (only when events are set)
	hasNav := opts.PrevMonthEvent != "" || opts.NextMonthEvent != ""

	var prevBtn gerbera.ComponentFunc
	if hasNav {
		prevBtnAttrs := gerbera.Components{
			property.Class("g-btn", "g-btn-sm", "g-calendar-nav"),
			property.Attr("type", "button"),
			property.AriaLabel("Previous month"),
			property.Value("\u2039"),
		}
		if opts.PrevMonthEvent != "" {
			prevBtnAttrs = append(prevBtnAttrs, property.Attr("gerbera-click", opts.PrevMonthEvent))
		}
		prevBtn = dom.Button(prevBtnAttrs...)
	}

	var nextBtn gerbera.ComponentFunc
	if hasNav {
		nextBtnAttrs := gerbera.Components{
			property.Class("g-btn", "g-btn-sm", "g-calendar-nav"),
			property.Attr("type", "button"),
			property.AriaLabel("Next month"),
			property.Value("\u203a"),
		}
		if opts.NextMonthEvent != "" {
			nextBtnAttrs = append(nextBtnAttrs, property.Attr("gerbera-click", opts.NextMonthEvent))
		}
		nextBtn = dom.Button(nextBtnAttrs...)
	}

	// Month select options
	var monthOpts gerbera.Components
	for m := time.January; m <= time.December; m++ {
		optAttrs := gerbera.Components{
			property.Attr("value", fmt.Sprintf("%d", int(m))),
			property.Value(m.String()),
		}
		if m == opts.Month {
			optAttrs = append(optAttrs, property.Attr("selected", "selected"))
		}
		monthOpts = append(monthOpts, dom.Option(optAttrs...))
	}

	// Year select options
	yearMin, yearMax := opts.YearRange[0], opts.YearRange[1]
	if yearMin == 0 {
		yearMin = opts.Year - 10
	}
	if yearMax == 0 {
		yearMax = opts.Year + 10
	}
	var yearOpts gerbera.Components
	for y := yearMin; y <= yearMax; y++ {
		optAttrs := gerbera.Components{
			property.Attr("value", fmt.Sprintf("%d", y)),
			property.Value(fmt.Sprintf("%d", y)),
		}
		if y == opts.Year {
			optAttrs = append(optAttrs, property.Attr("selected", "selected"))
		}
		yearOpts = append(yearOpts, dom.Option(optAttrs...))
	}

	monthSelectAttrs := append(gerbera.Components{
		property.Class("g-calendar-select"),
		property.Name("calendar-month"),
		property.AriaLabel("Month"),
	}, monthOpts...)
	if opts.MonthChangeEvent != "" {
		monthSelectAttrs = append(monthSelectAttrs,
			property.Attr("value", fmt.Sprintf("%d", int(opts.Month))),
			property.Attr("gerbera-change", opts.MonthChangeEvent),
		)
	}
	monthSelect := dom.Select(monthSelectAttrs...)

	yearSelectAttrs := append(gerbera.Components{
		property.Class("g-calendar-select"),
		property.Name("calendar-year"),
		property.AriaLabel("Year"),
	}, yearOpts...)
	if opts.YearChangeEvent != "" {
		yearSelectAttrs = append(yearSelectAttrs,
			property.Attr("value", fmt.Sprintf("%d", opts.Year)),
			property.Attr("gerbera-change", opts.YearChangeEvent),
		)
	}
	yearSelect := dom.Select(yearSelectAttrs...)

	wrapAttrs := gerbera.Components{
		property.Class("g-calendar"),
		property.Role("grid"),
		property.AriaLabel(fmt.Sprintf("%s %d", opts.Month.String(), opts.Year)),
	}
	wrapAttrs = append(wrapAttrs, extra...)

	// Header with nav + selectors
	headerChildren := gerbera.Components{
		property.Class("g-calendar-header"),
	}
	if prevBtn != nil {
		headerChildren = append(headerChildren, prevBtn)
	}
	headerChildren = append(headerChildren,
		dom.Div(property.Class("g-calendar-selectors"),
			monthSelect,
			yearSelect,
		),
	)
	if nextBtn != nil {
		headerChildren = append(headerChildren, nextBtn)
	}
	wrapAttrs = append(wrapAttrs, dom.Div(headerChildren...))

	// Day names row
	wrapAttrs = append(wrapAttrs,
		dom.Div(append(gerbera.Components{property.Class("g-calendar-weekdays")}, headerCells...)...),
	)

	// Day grid
	wrapAttrs = append(wrapAttrs,
		dom.Div(append(gerbera.Components{property.Class("g-calendar-grid")}, dayCells...)...),
	)

	return dom.Div(wrapAttrs...)
}

func sameDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
