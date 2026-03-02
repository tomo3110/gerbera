package live

import (
	"fmt"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/expr"
	gl "github.com/tomo3110/gerbera/live"
	"github.com/tomo3110/gerbera/property"
)

// Column defines a table column for DataTable.
type Column struct {
	Key      string
	Label    string
	Sortable bool
}

// DataTableOpts configures a DataTable widget.
type DataTableOpts struct {
	Columns   []Column
	Rows      [][]string
	SortCol   string // current sort column key
	SortDir   string // "asc" or "desc"
	SortEvent string // event name for sort (payload: {"value": key})
	Page      int    // current page (0-based)
	PageSize  int    // rows per page
	Total     int    // total row count
	PageEvent string // page change event (payload: {"value": "N"})
}

// DataTable renders a sortable, paginated data table.
func DataTable(opts DataTableOpts) gerbera.ComponentFunc {
	// Build header row
	var ths []gerbera.ComponentFunc
	for _, col := range opts.Columns {
		attrs := []gerbera.ComponentFunc{}
		if col.Sortable && opts.SortEvent != "" {
			attrs = append(attrs,
				property.Class("g-sortable"),
				property.ClassIf(opts.SortCol == col.Key, "g-sort-active"),
				gl.Click(opts.SortEvent),
				gl.ClickValue(col.Key),
			)
			indicator := ""
			if opts.SortCol == col.Key {
				if opts.SortDir == "asc" {
					indicator = " \u25b2"
				} else {
					indicator = " \u25bc"
				}
			}
			attrs = append(attrs,
				dom.Span(property.Value(col.Label)),
				dom.Span(property.Class("g-sort-indicator"), property.Value(indicator)),
			)
		} else {
			attrs = append(attrs, property.Value(col.Label))
		}
		ths = append(ths, dom.Th(attrs...))
	}

	// Build body rows
	var trs []gerbera.ComponentFunc
	for _, row := range opts.Rows {
		var tds []gerbera.ComponentFunc
		for _, cell := range row {
			tds = append(tds, dom.Td(property.Value(cell)))
		}
		trs = append(trs, dom.Tr(tds...))
	}

	table := dom.Table(
		property.Class("g-table"),
		dom.Thead(dom.Tr(ths...)),
		dom.Tbody(trs...),
	)

	// Pagination footer
	if opts.PageSize > 0 && opts.Total > 0 && opts.PageEvent != "" {
		totalPages := (opts.Total + opts.PageSize - 1) / opts.PageSize
		start := opts.Page*opts.PageSize + 1
		end := start + opts.PageSize - 1
		if end > opts.Total {
			end = opts.Total
		}

		var pageButtons []gerbera.ComponentFunc
		for i := 0; i < totalPages; i++ {
			pageButtons = append(pageButtons, dom.Button(
				property.Class("g-datatable-page"),
				property.ClassIf(i == opts.Page, "g-datatable-page-active"),
				gl.Click(opts.PageEvent),
				gl.ClickValue(fmt.Sprintf("%d", i)),
				property.Value(fmt.Sprintf("%d", i+1)),
			))
		}

		footer := dom.Div(
			property.Class("g-datatable-footer"),
			dom.Span(property.Value(fmt.Sprintf("%d\u2013%d / %d", start, end, opts.Total))),
			dom.Div(
				append([]gerbera.ComponentFunc{property.Class("g-datatable-pages")}, pageButtons...)...,
			),
		)

		return dom.Div(
			property.Class("g-card"),
			table,
			footer,
		)
	}

	// No pagination - simple table in card
	return expr.If(opts.PageSize > 0,
		dom.Div(property.Class("g-card"), table),
		table,
	)
}
