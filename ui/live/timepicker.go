package live

import (
	"fmt"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	gl "github.com/tomo3110/gerbera/live"
	"github.com/tomo3110/gerbera/property"
	"github.com/tomo3110/gerbera/ui"
)

// TimePickerOpts configures a live TimePicker.
type TimePickerOpts struct {
	Name        string
	Hour        int
	Minute      int
	Second      int
	Use24H      bool // true=24-hour format, false=12-hour format with AM/PM
	ShowSec     bool // show seconds field
	Disabled    bool
	ChangeEvent string // event fired on any change; payload: {"value": "HH:MM" or "HH:MM:SS"}
}

// TimePicker renders a live time picker with hour, minute, and optional second fields.
// Each field has increment/decrement buttons bound to ChangeEvent.
func TimePicker(opts TimePickerOpts) gerbera.ComponentFunc {
	hour, minute, second := clampTime(opts.Hour, opts.Minute, opts.Second)

	displayHour := hour
	ampm := "AM"
	if !opts.Use24H {
		displayHour, ampm = to12Hour(hour)
	}

	hourMax := 23
	hourMin := 0
	if !opts.Use24H {
		hourMax = 12
		hourMin = 1
	}

	// Format the current value for the payload
	timeValue := ui.FormatTime(hour, minute, second, opts.ShowSec)

	wrapAttrs := []gerbera.ComponentFunc{
		property.Class("g-timepicker"),
		property.Role("group"),
		property.AriaLabel(opts.Name),
		property.Attr("data-value", timeValue),
	}

	// Hour unit
	wrapAttrs = append(wrapAttrs, liveTimeUnit("Hour", displayHour, hourMin, hourMax, opts.Disabled, opts.ChangeEvent, "hour-up", "hour-down"))

	// Separator
	wrapAttrs = append(wrapAttrs, dom.Span(
		property.Class("g-timepicker-sep"),
		property.Value(":"),
	))

	// Minute unit
	wrapAttrs = append(wrapAttrs, liveTimeUnit("Minute", minute, 0, 59, opts.Disabled, opts.ChangeEvent, "minute-up", "minute-down"))

	if opts.ShowSec {
		wrapAttrs = append(wrapAttrs, dom.Span(
			property.Class("g-timepicker-sep"),
			property.Value(":"),
		))
		wrapAttrs = append(wrapAttrs, liveTimeUnit("Second", second, 0, 59, opts.Disabled, opts.ChangeEvent, "second-up", "second-down"))
	}

	// AM/PM toggle for 12-hour format
	if !opts.Use24H {
		amBtnAttrs := []gerbera.ComponentFunc{
			property.Class("g-timepicker-ampm-btn"),
			property.Attr("type", "button"),
			property.Attr("aria-pressed", boolStr(ampm == "AM")),
			property.Disabled(opts.Disabled),
			property.Value("AM"),
		}
		pmBtnAttrs := []gerbera.ComponentFunc{
			property.Class("g-timepicker-ampm-btn"),
			property.Attr("type", "button"),
			property.Attr("aria-pressed", boolStr(ampm == "PM")),
			property.Disabled(opts.Disabled),
			property.Value("PM"),
		}
		if opts.ChangeEvent != "" {
			amBtnAttrs = append(amBtnAttrs, gl.Click(opts.ChangeEvent), gl.ClickValue("am"))
			pmBtnAttrs = append(pmBtnAttrs, gl.Click(opts.ChangeEvent), gl.ClickValue("pm"))
		}
		wrapAttrs = append(wrapAttrs, dom.Div(
			property.Class("g-timepicker-ampm"),
			dom.Button(amBtnAttrs...),
			dom.Button(pmBtnAttrs...),
		))
	}

	return dom.Div(wrapAttrs...)
}

func liveTimeUnit(label string, value, min, max int, disabled bool, changeEvent, upValue, downValue string) gerbera.ComponentFunc {
	upBtnAttrs := []gerbera.ComponentFunc{
		property.Class("g-timepicker-btn", "g-timepicker-up"),
		property.Attr("type", "button"),
		property.Attr("tabindex", "-1"),
		property.AriaLabel(fmt.Sprintf("Increase %s", label)),
		property.Disabled(disabled),
		property.Value("\u25B2"),
	}
	if changeEvent != "" {
		upBtnAttrs = append(upBtnAttrs, gl.Click(changeEvent), gl.ClickValue(upValue))
	}

	downBtnAttrs := []gerbera.ComponentFunc{
		property.Class("g-timepicker-btn", "g-timepicker-down"),
		property.Attr("type", "button"),
		property.Attr("tabindex", "-1"),
		property.AriaLabel(fmt.Sprintf("Decrease %s", label)),
		property.Disabled(disabled),
		property.Value("\u25BC"),
	}
	if changeEvent != "" {
		downBtnAttrs = append(downBtnAttrs, gl.Click(changeEvent), gl.ClickValue(downValue))
	}

	inputAttrs := []gerbera.ComponentFunc{
		property.Class("g-timepicker-field"),
		property.Attr("type", "text"),
		property.Attr("inputmode", "numeric"),
		property.Attr("size", "2"),
		property.Attr("value", fmt.Sprintf("%02d", value)),
		property.Disabled(disabled),
	}
	if changeEvent != "" {
		inputAttrs = append(inputAttrs, gl.Change(changeEvent))
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

func clampTime(hour, minute, second int) (int, int, int) {
	hour = hour % 24
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

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
