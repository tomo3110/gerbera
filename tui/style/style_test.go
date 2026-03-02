package style

import (
	"testing"

	"github.com/tomo3110/gerbera"
)

func TestFgColor(t *testing.T) {
	el := &gerbera.Element{TagName: "text"}
	if err := FgColor("red")(el); err != nil {
		t.Fatal(err)
	}
	if el.Attr["tui-fg"] != "red" {
		t.Errorf("want tui-fg=red, got %s", el.Attr["tui-fg"])
	}
}

func TestBgColor(t *testing.T) {
	el := &gerbera.Element{TagName: "text"}
	if err := BgColor("#00ff00")(el); err != nil {
		t.Fatal(err)
	}
	if el.Attr["tui-bg"] != "#00ff00" {
		t.Errorf("want tui-bg=#00ff00, got %s", el.Attr["tui-bg"])
	}
}

func TestBold(t *testing.T) {
	el := &gerbera.Element{TagName: "text"}
	if err := Bold(true)(el); err != nil {
		t.Fatal(err)
	}
	if el.Attr["tui-bold"] != "true" {
		t.Errorf("want tui-bold=true, got %s", el.Attr["tui-bold"])
	}
}

func TestWidth(t *testing.T) {
	el := &gerbera.Element{TagName: "box"}
	if err := Width(40)(el); err != nil {
		t.Fatal(err)
	}
	if el.Attr["tui-width"] != "40" {
		t.Errorf("want tui-width=40, got %s", el.Attr["tui-width"])
	}
}

func TestPadding(t *testing.T) {
	el := &gerbera.Element{TagName: "box"}
	if err := Padding(1, 2, 3, 4)(el); err != nil {
		t.Fatal(err)
	}
	if el.Attr["tui-padding"] != "1 2 3 4" {
		t.Errorf("want tui-padding='1 2 3 4', got %s", el.Attr["tui-padding"])
	}
}

func TestMargin(t *testing.T) {
	el := &gerbera.Element{TagName: "box"}
	if err := Margin(1, 2, 3, 4)(el); err != nil {
		t.Fatal(err)
	}
	if el.Attr["tui-margin"] != "1 2 3 4" {
		t.Errorf("want tui-margin='1 2 3 4', got %s", el.Attr["tui-margin"])
	}
}

func TestBorder(t *testing.T) {
	el := &gerbera.Element{TagName: "box"}
	if err := Border("rounded")(el); err != nil {
		t.Fatal(err)
	}
	if el.Attr["tui-border"] != "rounded" {
		t.Errorf("want tui-border=rounded, got %s", el.Attr["tui-border"])
	}
}

func TestAlign(t *testing.T) {
	el := &gerbera.Element{TagName: "box"}
	if err := Align("center")(el); err != nil {
		t.Fatal(err)
	}
	if el.Attr["tui-align"] != "center" {
		t.Errorf("want tui-align=center, got %s", el.Attr["tui-align"])
	}
}

func TestMultipleStyles(t *testing.T) {
	el := &gerbera.Element{TagName: "text"}
	styles := []gerbera.ComponentFunc{
		FgColor("red"),
		Bold(true),
		Underline(true),
	}
	for _, s := range styles {
		if err := s(el); err != nil {
			t.Fatal(err)
		}
	}
	if el.Attr["tui-fg"] != "red" {
		t.Errorf("want tui-fg=red, got %s", el.Attr["tui-fg"])
	}
	if el.Attr["tui-bold"] != "true" {
		t.Errorf("want tui-bold=true, got %s", el.Attr["tui-bold"])
	}
	if el.Attr["tui-underline"] != "true" {
		t.Errorf("want tui-underline=true, got %s", el.Attr["tui-underline"])
	}
}
