package ui

import (
	"strconv"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/expr"
	"github.com/tomo3110/gerbera/property"
)

// InfiniteScrollView represents the display mode of InfiniteScroll content.
type InfiniteScrollView string

const (
	InfiniteScrollList InfiniteScrollView = "list"
	InfiniteScrollGrid InfiniteScrollView = "grid"
)

// InfiniteScrollOpts configures an InfiniteScroll component.
type InfiniteScrollOpts struct {
	View          InfiniteScrollView
	Loading       bool
	ShowToggle    bool
	LoadMoreEvent string // gerbera-scroll on content area
	ToggleEvent   string // gerbera-click on toggle buttons; payload: {"value": "list"|"grid"}
	ThrottleMs    int    // throttle for scroll event; default 200
}

// InfiniteScroll renders a scrollable container with optional loading indicator and view toggle.
func InfiniteScroll(opts InfiniteScrollOpts, children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	if opts.View == "" {
		opts.View = InfiniteScrollList
	}
	if opts.ThrottleMs <= 0 {
		opts.ThrottleMs = 200
	}

	wrapAttrs := []gerbera.ComponentFunc{
		property.Class("g-infinitescroll"),
	}

	// Toolbar with view toggle
	if opts.ShowToggle {
		listPressed := boolStr(opts.View == InfiniteScrollList)
		gridPressed := boolStr(opts.View == InfiniteScrollGrid)

		listBtn := []gerbera.ComponentFunc{
			property.Class("g-infinitescroll-toggle"),
			property.ClassIf(opts.View == InfiniteScrollList, "g-infinitescroll-toggle-active"),
			property.Attr("type", "button"),
			property.Attr("aria-pressed", listPressed),
			property.AriaLabel("List view"),
			property.Value("\u2630"),
		}
		if opts.ToggleEvent != "" {
			listBtn = append(listBtn,
				property.Attr("gerbera-click", opts.ToggleEvent),
				property.Attr("gerbera-value", "list"),
			)
		} else {
			listBtn = append(listBtn, property.Attr("data-value", "list"))
		}

		gridBtn := []gerbera.ComponentFunc{
			property.Class("g-infinitescroll-toggle"),
			property.ClassIf(opts.View == InfiniteScrollGrid, "g-infinitescroll-toggle-active"),
			property.Attr("type", "button"),
			property.Attr("aria-pressed", gridPressed),
			property.AriaLabel("Grid view"),
			property.Value("\u25A6"),
		}
		if opts.ToggleEvent != "" {
			gridBtn = append(gridBtn,
				property.Attr("gerbera-click", opts.ToggleEvent),
				property.Attr("gerbera-value", "grid"),
			)
		} else {
			gridBtn = append(gridBtn, property.Attr("data-value", "grid"))
		}

		toolbar := dom.Div(
			property.Class("g-infinitescroll-toolbar"),
			dom.Button(listBtn...),
			dom.Button(gridBtn...),
		)
		wrapAttrs = append(wrapAttrs, toolbar)
	}

	// Content area
	contentAttrs := []gerbera.ComponentFunc{
		property.Class("g-infinitescroll-content"),
		property.ClassIf(opts.View == InfiniteScrollGrid, "g-infinitescroll-grid"),
		property.AriaLive("polite"),
	}
	if opts.LoadMoreEvent != "" {
		contentAttrs = append(contentAttrs,
			property.Attr("gerbera-scroll", opts.LoadMoreEvent),
			property.Attr("gerbera-throttle", strconv.Itoa(opts.ThrottleMs)),
		)
	}
	contentAttrs = append(contentAttrs, children...)
	wrapAttrs = append(wrapAttrs, dom.Div(contentAttrs...))

	// Loader
	wrapAttrs = append(wrapAttrs,
		expr.If(opts.Loading,
			dom.Div(
				property.Class("g-infinitescroll-loader"),
				property.Role("status"),
				property.AriaLabel("Loading more items"),
				Spinner("sm"),
			),
		),
	)

	return dom.Div(wrapAttrs...)
}
