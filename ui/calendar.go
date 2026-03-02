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
	Year      int
	Month     time.Month
	Selected  *time.Time
	Today     time.Time
	DayNames  []string // defaults to ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"]
	YearRange [2]int   // [min, max] for year selector; zero values default to Year ± 10
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
	var headerCells []gerbera.ComponentFunc
	for _, d := range dayNames {
		headerCells = append(headerCells, dom.Div(
			property.Class("g-calendar-dayname"),
			property.Value(d),
		))
	}

	// Day cells
	var dayCells []gerbera.ComponentFunc

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

		dayCells = append(dayCells, dom.Div(
			property.Class("g-calendar-day"),
			property.ClassIf(isToday, "g-calendar-day-today"),
			property.ClassIf(isSelected, "g-calendar-day-selected"),
			property.Attr("data-date", current.Format("2006-01-02")),
			property.Value(fmt.Sprintf("%d", d)),
		))
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

	// Month select options
	var monthOpts []gerbera.ComponentFunc
	for m := time.January; m <= time.December; m++ {
		optAttrs := []gerbera.ComponentFunc{
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
	var yearOpts []gerbera.ComponentFunc
	for y := yearMin; y <= yearMax; y++ {
		optAttrs := []gerbera.ComponentFunc{
			property.Attr("value", fmt.Sprintf("%d", y)),
			property.Value(fmt.Sprintf("%d", y)),
		}
		if y == opts.Year {
			optAttrs = append(optAttrs, property.Attr("selected", "selected"))
		}
		yearOpts = append(yearOpts, dom.Option(optAttrs...))
	}

	monthSelect := dom.Select(append([]gerbera.ComponentFunc{
		property.Class("g-calendar-select"),
		property.Name("calendar-month"),
		property.AriaLabel("Month"),
	}, monthOpts...)...)

	yearSelect := dom.Select(append([]gerbera.ComponentFunc{
		property.Class("g-calendar-select"),
		property.Name("calendar-year"),
		property.AriaLabel("Year"),
	}, yearOpts...)...)

	wrapAttrs := []gerbera.ComponentFunc{
		property.Class("g-calendar"),
		property.Role("grid"),
		property.AriaLabel(fmt.Sprintf("%s %d", opts.Month.String(), opts.Year)),
	}
	wrapAttrs = append(wrapAttrs, extra...)

	// Header with month/year selectors
	wrapAttrs = append(wrapAttrs,
		dom.Div(
			property.Class("g-calendar-header"),
			dom.Div(property.Class("g-calendar-selectors"),
				monthSelect,
				yearSelect,
			),
		),
	)

	// Day names row
	wrapAttrs = append(wrapAttrs,
		dom.Div(append([]gerbera.ComponentFunc{property.Class("g-calendar-weekdays")}, headerCells...)...),
	)

	// Day grid
	wrapAttrs = append(wrapAttrs,
		dom.Div(append([]gerbera.ComponentFunc{property.Class("g-calendar-grid")}, dayCells...)...),
	)

	return dom.Div(wrapAttrs...)
}

func sameDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
