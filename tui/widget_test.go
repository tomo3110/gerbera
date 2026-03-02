package tui

import (
	"testing"

	"github.com/tomo3110/gerbera"
)

func testCommon(t *testing.T, c gerbera.ComponentFunc, wantTagName string) {
	t.Helper()
	parent := &gerbera.Element{TagName: "root"}
	if err := c(parent); err != nil {
		t.Error(err.Error())
	}
	target := parent.Children[0]
	if target == nil {
		t.Error("要素が正常に追加されていません")
	}
	if target.TagName != wantTagName {
		t.Errorf("要素名が異なります: want = %s, result = %s\n", wantTagName, target.TagName)
	}
}

func TestBox(t *testing.T)         { testCommon(t, Box(), "box") }
func TestVBox(t *testing.T)        { testCommon(t, VBox(), "vbox") }
func TestHBox(t *testing.T)        { testCommon(t, HBox(), "hbox") }
func TestText(t *testing.T)        { testCommon(t, Text(), "text") }
func TestHeader(t *testing.T)      { testCommon(t, Header(), "header") }
func TestParagraph(t *testing.T)   { testCommon(t, Paragraph(), "paragraph") }
func TestList(t *testing.T)        { testCommon(t, List(), "list") }
func TestListItem(t *testing.T)    { testCommon(t, ListItem(), "list-item") }
func TestTable(t *testing.T)       { testCommon(t, Table(), "table") }
func TestTableRow(t *testing.T)    { testCommon(t, TableRow(), "table-row") }
func TestTableCell(t *testing.T)   { testCommon(t, TableCell(), "table-cell") }
func TestTableHeader(t *testing.T) { testCommon(t, TableHeader(), "table-header") }
func TestInput(t *testing.T)       { testCommon(t, Input(), "input") }
func TestButton(t *testing.T)      { testCommon(t, Button(), "button") }
func TestCheckbox(t *testing.T)    { testCommon(t, Checkbox(), "checkbox") }
func TestDivider(t *testing.T)     { testCommon(t, Divider(), "divider") }
func TestSpacer(t *testing.T)      { testCommon(t, Spacer(), "spacer") }
func TestProgressBar(t *testing.T) { testCommon(t, ProgressBar(), "progress") }
func TestSpinner(t *testing.T)     { testCommon(t, Spinner(), "spinner") }
func TestStatusBar(t *testing.T)   { testCommon(t, StatusBar(), "statusbar") }

func TestWidgetWithChildren(t *testing.T) {
	parent := &gerbera.Element{TagName: "root"}
	c := Box(
		Text(gerbera.Literal("hello")),
		Text(gerbera.Literal("world")),
	)
	if err := c(parent); err != nil {
		t.Fatal(err)
	}
	box := parent.Children[0]
	if len(box.Children) != 2 {
		t.Errorf("子要素数が異なります: want = 2, result = %d", len(box.Children))
	}
	if box.Children[0].TagName != "text" {
		t.Errorf("子要素のタグ名が異なります: want = text, result = %s", box.Children[0].TagName)
	}
	if box.Children[0].Value != "hello" {
		t.Errorf("子要素の値が異なります: want = hello, result = %s", box.Children[0].Value)
	}
}
