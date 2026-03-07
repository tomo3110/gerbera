package ui

import (
	"fmt"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
)

// TimePickerOpts configures a TimePicker.
type TimePickerOpts struct {
	Use24H      bool   // true=24-hour format (default), false=12-hour format with AM/PM
	ShowSec     bool   // show seconds field
	Disabled    bool
	ChangeEvent string // gerbera-click on buttons, gerbera-change on inputs
}

// TimePicker renders a time picker with hour, minute, and optional second fields.
// Each field has increment/decrement buttons for value adjustment.
func TimePicker(name string, hour, minute, second int, opts TimePickerOpts, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	hour, minute, second = clampTime(hour, minute, second, opts.Use24H)

	displayHour := hour
	ampm := "AM"
	if !opts.Use24H {
		displayHour, ampm = to12Hour(hour)
	}

	hourMax := 23
	if !opts.Use24H {
		hourMax = 12
	}
	hourMin := 0
	if !opts.Use24H {
		hourMin = 1
	}

	// Format the current value for the payload
	timeValue := FormatTime(hour, minute, second, opts.ShowSec)

	wrapAttrs := gerbera.Components{
		property.Class("g-timepicker"),
		property.Role("group"),
		property.AriaLabel(name),
	}
	if opts.ChangeEvent != "" {
		wrapAttrs = append(wrapAttrs, property.Attr("data-value", timeValue))
	}
	wrapAttrs = append(wrapAttrs, extra...)

	// Hour unit
	wrapAttrs = append(wrapAttrs, timeUnit("Hour", displayHour, hourMin, hourMax, opts.Disabled, opts.ChangeEvent, "hour-up", "hour-down"))

	// Separator
	wrapAttrs = append(wrapAttrs, dom.Span(
		property.Class("g-timepicker-sep"),
		property.Value(":"),
	))

	// Minute unit
	wrapAttrs = append(wrapAttrs, timeUnit("Minute", minute, 0, 59, opts.Disabled, opts.ChangeEvent, "minute-up", "minute-down"))

	if opts.ShowSec {
		wrapAttrs = append(wrapAttrs, dom.Span(
			property.Class("g-timepicker-sep"),
			property.Value(":"),
		))
		wrapAttrs = append(wrapAttrs, timeUnit("Second", second, 0, 59, opts.Disabled, opts.ChangeEvent, "second-up", "second-down"))
	}

	// AM/PM toggle for 12-hour format
	if !opts.Use24H {
		amBtnAttrs := gerbera.Components{
			property.Class("g-timepicker-ampm-btn"),
			property.Attr("type", "button"),
			property.Attr("aria-pressed", boolStr(ampm == "AM")),
			property.Disabled(opts.Disabled),
			property.Value("AM"),
		}
		pmBtnAttrs := gerbera.Components{
			property.Class("g-timepicker-ampm-btn"),
			property.Attr("type", "button"),
			property.Attr("aria-pressed", boolStr(ampm == "PM")),
			property.Disabled(opts.Disabled),
			property.Value("PM"),
		}
		if opts.ChangeEvent != "" {
			amBtnAttrs = append(amBtnAttrs,
				property.Attr("gerbera-click", opts.ChangeEvent),
				property.Attr("gerbera-value", "am"),
			)
			pmBtnAttrs = append(pmBtnAttrs,
				property.Attr("gerbera-click", opts.ChangeEvent),
				property.Attr("gerbera-value", "pm"),
			)
		}
		wrapAttrs = append(wrapAttrs, dom.Div(
			property.Class("g-timepicker-ampm"),
			dom.Button(amBtnAttrs...),
			dom.Button(pmBtnAttrs...),
		))
	}

	return dom.Div(wrapAttrs...)
}

func timeUnit(label string, value, min, max int, disabled bool, changeEvent, upValue, downValue string) gerbera.ComponentFunc {
	upBtnAttrs := gerbera.Components{
		property.Class("g-timepicker-btn", "g-timepicker-up"),
		property.Attr("type", "button"),
		property.Attr("tabindex", "-1"),
		property.AriaLabel(fmt.Sprintf("Increase %s", label)),
		property.Disabled(disabled),
		property.Value("\u25B2"),
	}
	if changeEvent != "" {
		upBtnAttrs = append(upBtnAttrs,
			property.Attr("gerbera-click", changeEvent),
			property.Attr("gerbera-value", upValue),
		)
	}

	downBtnAttrs := gerbera.Components{
		property.Class("g-timepicker-btn", "g-timepicker-down"),
		property.Attr("type", "button"),
		property.Attr("tabindex", "-1"),
		property.AriaLabel(fmt.Sprintf("Decrease %s", label)),
		property.Disabled(disabled),
		property.Value("\u25BC"),
	}
	if changeEvent != "" {
		downBtnAttrs = append(downBtnAttrs,
			property.Attr("gerbera-click", changeEvent),
			property.Attr("gerbera-value", downValue),
		)
	}

	inputAttrs := gerbera.Components{
		property.Class("g-timepicker-field"),
		property.Attr("type", "text"),
		property.Attr("inputmode", "numeric"),
		property.Attr("size", "2"),
		property.Attr("value", fmt.Sprintf("%02d", value)),
		property.Disabled(disabled),
	}
	if changeEvent != "" {
		inputAttrs = append(inputAttrs, property.Attr("gerbera-change", changeEvent))
	}

	return dom.Div(
		property.Class("g-timepicker-unit"),
		property.Role("spinbutton"),
		property.AriaLabel(label),
		property.AriaValueNow(value),
		property.AriaValueMin(min),
		property.AriaValueMax(max),
		dom.Button(upBtnAttrs...),
		dom.Input(inputAttrs...),
		dom.Button(downBtnAttrs...),
	)
}

func clampTime(hour, minute, second int, use24H bool) (int, int, int) {
	if use24H {
		hour = hour % 24
	} else {
		hour = hour % 24
	}
	if hour < 0 {
		hour = 0
	}
	minute = minute % 60
	if minute < 0 {
		minute = 0
	}
	second = second % 60
	if second < 0 {
		second = 0
	}
	return hour, minute, second
}

func to12Hour(hour24 int) (int, string) {
	ampm := "AM"
	h := hour24 % 24
	if h >= 12 {
		ampm = "PM"
	}
	h = h % 12
	if h == 0 {
		h = 12
	}
	return h, ampm
}

// FormatTime formats hour, minute, second as "HH:MM" or "HH:MM:SS".
func FormatTime(hour, minute, second int, showSec bool) string {
	if showSec {
		return fmt.Sprintf("%02d:%02d:%02d", hour, minute, second)
	}
	return fmt.Sprintf("%02d:%02d", hour, minute)
}
