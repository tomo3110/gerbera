package tui

import g "github.com/tomo3110/gerbera"

// Layout

func Box(children ...g.ComponentFunc) g.ComponentFunc  { return g.Tag("box", children...) }
func VBox(children ...g.ComponentFunc) g.ComponentFunc { return g.Tag("vbox", children...) }
func HBox(children ...g.ComponentFunc) g.ComponentFunc { return g.Tag("hbox", children...) }

// Text

func Text(children ...g.ComponentFunc) g.ComponentFunc      { return g.Tag("text", children...) }
func Header(children ...g.ComponentFunc) g.ComponentFunc    { return g.Tag("header", children...) }
func Paragraph(children ...g.ComponentFunc) g.ComponentFunc { return g.Tag("paragraph", children...) }

// List

func List(children ...g.ComponentFunc) g.ComponentFunc     { return g.Tag("list", children...) }
func ListItem(children ...g.ComponentFunc) g.ComponentFunc { return g.Tag("list-item", children...) }

// Table

func Table(children ...g.ComponentFunc) g.ComponentFunc       { return g.Tag("table", children...) }
func TableRow(children ...g.ComponentFunc) g.ComponentFunc    { return g.Tag("table-row", children...) }
func TableCell(children ...g.ComponentFunc) g.ComponentFunc   { return g.Tag("table-cell", children...) }
func TableHeader(children ...g.ComponentFunc) g.ComponentFunc { return g.Tag("table-header", children...) }

// Interactive

func Input(children ...g.ComponentFunc) g.ComponentFunc    { return g.Tag("input", children...) }
func Button(children ...g.ComponentFunc) g.ComponentFunc   { return g.Tag("button", children...) }
func Checkbox(children ...g.ComponentFunc) g.ComponentFunc { return g.Tag("checkbox", children...) }

// Visual

func Divider(children ...g.ComponentFunc) g.ComponentFunc     { return g.Tag("divider", children...) }
func Spacer(children ...g.ComponentFunc) g.ComponentFunc      { return g.Tag("spacer", children...) }
func ProgressBar(children ...g.ComponentFunc) g.ComponentFunc { return g.Tag("progress", children...) }
func Spinner(children ...g.ComponentFunc) g.ComponentFunc     { return g.Tag("spinner", children...) }
func StatusBar(children ...g.ComponentFunc) g.ComponentFunc   { return g.Tag("statusbar", children...) }
