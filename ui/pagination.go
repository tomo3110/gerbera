package ui

import (
	"fmt"
	"strconv"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
)

// PaginationOpts configures a Pagination component.
type PaginationOpts struct {
	Page      int    // 0-based current page
	PageSize  int
	Total     int
	PageEvent string // gerbera-click on page buttons; payload: {"value": "N"} (0-based)
}

// Pagination renders a page navigation bar with ellipsis for large page counts.
func Pagination(opts PaginationOpts, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	if opts.PageSize <= 0 {
		opts.PageSize = 10
	}
	totalPages := (opts.Total + opts.PageSize - 1) / opts.PageSize
	if totalPages < 1 {
		totalPages = 1
	}
	if opts.Page < 0 {
		opts.Page = 0
	}
	if opts.Page >= totalPages {
		opts.Page = totalPages - 1
	}

	startItem := opts.Page*opts.PageSize + 1
	endItem := startItem + opts.PageSize - 1
	if endItem > opts.Total {
		endItem = opts.Total
	}
	if opts.Total == 0 {
		startItem = 0
	}
	info := dom.Span(
		property.Class("g-pagination-info"),
		property.Value(fmt.Sprintf("%d\u2013%d of %d", startItem, endItem, opts.Total)),
	)

	// Build page buttons
	pages := PaginationPages(opts.Page, totalPages)
	var pageButtons gerbera.Components
	// Prev button
	prevAttrs := gerbera.Components{
		property.Class("g-pagination-btn", "g-pagination-prev"),
		property.Attr("type", "button"),
		property.AriaLabel("Previous page"),
		property.Disabled(opts.Page == 0),
		property.Value("\u2039"),
	}
	if opts.PageEvent != "" && opts.Page > 0 {
		prevAttrs = append(prevAttrs,
			property.Attr("gerbera-click", opts.PageEvent),
			property.Attr("gerbera-value", strconv.Itoa(opts.Page-1)),
		)
	}
	pageButtons = append(pageButtons, dom.Button(prevAttrs...))

	for _, p := range pages {
		if p == -1 {
			pageButtons = append(pageButtons, dom.Span(
				property.Class("g-pagination-ellipsis"),
				property.Value("\u2026"),
			))
		} else {
			isActive := p == opts.Page
			btnAttrs := gerbera.Components{
				property.Class("g-pagination-btn"),
				property.ClassIf(isActive, "g-pagination-active"),
				property.Attr("type", "button"),
				property.Value(strconv.Itoa(p + 1)),
			}
			if isActive {
				btnAttrs = append(btnAttrs, property.AriaCurrent("page"))
			}
			if opts.PageEvent != "" && !isActive {
				btnAttrs = append(btnAttrs,
					property.Attr("gerbera-click", opts.PageEvent),
					property.Attr("gerbera-value", strconv.Itoa(p)),
				)
			} else if opts.PageEvent == "" {
				btnAttrs = append(btnAttrs, property.Attr("data-page", strconv.Itoa(p)))
			}
			pageButtons = append(pageButtons, dom.Button(btnAttrs...))
		}
	}

	// Next button
	nextAttrs := gerbera.Components{
		property.Class("g-pagination-btn", "g-pagination-next"),
		property.Attr("type", "button"),
		property.AriaLabel("Next page"),
		property.Disabled(opts.Page >= totalPages-1),
		property.Value("\u203A"),
	}
	if opts.PageEvent != "" && opts.Page < totalPages-1 {
		nextAttrs = append(nextAttrs,
			property.Attr("gerbera-click", opts.PageEvent),
			property.Attr("gerbera-value", strconv.Itoa(opts.Page+1)),
		)
	}
	pageButtons = append(pageButtons, dom.Button(nextAttrs...))

	wrapAttrs := gerbera.Components{
		property.Class("g-pagination"),
		property.Role("navigation"),
		property.AriaLabel("Pagination"),
	}
	wrapAttrs = append(wrapAttrs, extra...)
	wrapAttrs = append(wrapAttrs,
		info,
		dom.Div(append(gerbera.Components{property.Class("g-pagination-pages")}, pageButtons...)...),
	)

	return dom.Nav(wrapAttrs...)
}

// PaginationPages returns the page indices to display.
// -1 indicates an ellipsis. Exported for use by the live package.
func PaginationPages(current, total int) []int {
	if total <= 7 {
		pages := make([]int, total)
		for i := range pages {
			pages[i] = i
		}
		return pages
	}

	pages := []int{}
	seen := map[int]bool{}

	add := func(p int) {
		if p >= 0 && p < total && !seen[p] {
			seen[p] = true
			pages = append(pages, p)
		}
	}

	// Always show first
	add(0)
	// current-1, current, current+1
	add(current - 1)
	add(current)
	add(current + 1)
	// Always show last
	add(total - 1)

	// Sort
	for i := 0; i < len(pages)-1; i++ {
		for j := i + 1; j < len(pages); j++ {
			if pages[i] > pages[j] {
				pages[i], pages[j] = pages[j], pages[i]
			}
		}
	}

	// Insert ellipsis for gaps
	var result []int
	for i, p := range pages {
		if i > 0 && p-pages[i-1] > 1 {
			result = append(result, -1)
		}
		result = append(result, p)
	}
	return result
}
