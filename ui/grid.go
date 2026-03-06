package ui

import (
	"fmt"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
)

// Row renders a horizontal flex container (gap: 8px).
func Row(children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := gerbera.Components{property.Class("g-row")}
	attrs = append(attrs, children...)
	return dom.Div(attrs...)
}

// Column renders a vertical flex container (gap: 8px).
func Column(children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := gerbera.Components{property.Class("g-col")}
	attrs = append(attrs, children...)
	return dom.Div(attrs...)
}

// Stack renders a vertical stack with gap: 16px — for page-level content stacking.
func Stack(children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := gerbera.Components{property.Class("g-stack")}
	attrs = append(attrs, children...)
	return dom.Div(attrs...)
}

// HStack renders a horizontal stack (gap: 8px, wrap, align-center) — for button groups etc.
func HStack(children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := gerbera.Components{property.Class("g-hstack")}
	attrs = append(attrs, children...)
	return dom.Div(attrs...)
}

// VStack renders a vertical stack with center alignment (gap: 8px) — for icon+label combos.
func VStack(children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := gerbera.Components{property.Class("g-vstack")}
	attrs = append(attrs, children...)
	return dom.Div(attrs...)
}

// Center renders a flex container that centers its children both horizontally and vertically.
func Center(children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := gerbera.Components{property.Class("g-center")}
	attrs = append(attrs, children...)
	return dom.Div(attrs...)
}

// Container renders a max-width container (960px) with auto margins.
func Container(children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := gerbera.Components{property.Class("g-container")}
	attrs = append(attrs, children...)
	return dom.Div(attrs...)
}

// ContainerNarrow renders a narrow container (640px).
func ContainerNarrow(children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := gerbera.Components{property.Class("g-container", "g-container-narrow")}
	attrs = append(attrs, children...)
	return dom.Div(attrs...)
}

// ContainerWide renders a wide container (1280px).
func ContainerWide(children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := gerbera.Components{property.Class("g-container", "g-container-wide")}
	attrs = append(attrs, children...)
	return dom.Div(attrs...)
}

// Grid renders a CSS Grid container (gap: 24px).
func Grid(children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := gerbera.Components{property.Class("g-grid")}
	attrs = append(attrs, children...)
	return dom.Div(attrs...)
}

// GridCols2–GridCols6 set the number of grid columns.
var (
	GridCols2 gerbera.ComponentFunc = property.ClassIf(true, "g-grid-2")
	GridCols3 gerbera.ComponentFunc = property.ClassIf(true, "g-grid-3")
	GridCols4 gerbera.ComponentFunc = property.ClassIf(true, "g-grid-4")
	GridCols5 gerbera.ComponentFunc = property.ClassIf(true, "g-grid-5")
	GridCols6 gerbera.ComponentFunc = property.ClassIf(true, "g-grid-6")
)

// GridAutoFill returns a modifier that sets grid-template-columns to repeat(auto-fill, minmax(min, 1fr)).
func GridAutoFill(minWidth string) gerbera.ComponentFunc {
	return property.Attr("style", fmt.Sprintf("grid-template-columns:repeat(auto-fill,minmax(%s,1fr))", minWidth))
}

// GridAutoFit returns a modifier that sets grid-template-columns to repeat(auto-fit, minmax(min, 1fr)).
func GridAutoFit(minWidth string) gerbera.ComponentFunc {
	return property.Attr("style", fmt.Sprintf("grid-template-columns:repeat(auto-fit,minmax(%s,1fr))", minWidth))
}

// GridSpan returns a modifier that makes an element span multiple grid columns.
func GridSpan(cols int) gerbera.ComponentFunc {
	return property.Attr("style", fmt.Sprintf("grid-column:span %d", cols))
}

// GridRowSpan returns a modifier that makes an element span multiple grid rows.
func GridRowSpan(rows int) gerbera.ComponentFunc {
	return property.Attr("style", fmt.Sprintf("grid-row:span %d", rows))
}

// Spacer renders a flex spacer that fills available space.
func Spacer() gerbera.ComponentFunc {
	return dom.Div(property.Class("g-spacer"))
}

// SpaceY renders a vertical spacer div. size: "xs"|"sm"|"md"|"lg"|"xl".
func SpaceY(size string) gerbera.ComponentFunc {
	return dom.Div(property.Class("g-space-y-" + size))
}

// Gap modifiers
var (
	GapNone gerbera.ComponentFunc = property.ClassIf(true, "g-gap-none")
	GapXs   gerbera.ComponentFunc = property.ClassIf(true, "g-gap-xs")
	GapSm   gerbera.ComponentFunc = property.ClassIf(true, "g-gap-sm")
	GapMd   gerbera.ComponentFunc = property.ClassIf(true, "g-gap-md")
	GapLg   gerbera.ComponentFunc = property.ClassIf(true, "g-gap-lg")
	GapXl   gerbera.ComponentFunc = property.ClassIf(true, "g-gap-xl")
)

// Wrap enables flex-wrap.
var Wrap gerbera.ComponentFunc = property.ClassIf(true, "g-wrap")

// justify-content modifiers
var (
	JustifyStart   gerbera.ComponentFunc = property.ClassIf(true, "g-justify-start")
	JustifyCenter  gerbera.ComponentFunc = property.ClassIf(true, "g-justify-center")
	JustifyEnd     gerbera.ComponentFunc = property.ClassIf(true, "g-justify-end")
	JustifyBetween gerbera.ComponentFunc = property.ClassIf(true, "g-justify-between")
	JustifyAround  gerbera.ComponentFunc = property.ClassIf(true, "g-justify-around")
)

// align-items modifiers
var (
	AlignStart    gerbera.ComponentFunc = property.ClassIf(true, "g-align-start")
	AlignCenter   gerbera.ComponentFunc = property.ClassIf(true, "g-align-center")
	AlignEnd      gerbera.ComponentFunc = property.ClassIf(true, "g-align-end")
	AlignStretch  gerbera.ComponentFunc = property.ClassIf(true, "g-align-stretch")
	AlignBaseline gerbera.ComponentFunc = property.ClassIf(true, "g-align-baseline")
)

// Flex child modifiers
var (
	Grow    gerbera.ComponentFunc = property.ClassIf(true, "g-grow")
	Shrink0 gerbera.ComponentFunc = property.ClassIf(true, "g-shrink-0")
)
