package app

import (
	"testing"

	"github.com/tomo3110/gerbera"
)

func TestOnKey(t *testing.T) {
	el := &gerbera.Element{TagName: "box"}
	if err := OnKey("k", "increment")(el); err != nil {
		t.Fatal(err)
	}
	if el.Attr["tui-on-key-k"] != "increment" {
		t.Errorf("want tui-on-key-k=increment, got %s", el.Attr["tui-on-key-k"])
	}
}

func TestFocusable(t *testing.T) {
	el := &gerbera.Element{TagName: "input"}
	if err := Focusable(0)(el); err != nil {
		t.Fatal(err)
	}
	if el.Attr["tui-focusable"] != "0" {
		t.Errorf("want tui-focusable=0, got %s", el.Attr["tui-focusable"])
	}
}

func TestParseInput(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{"enter", []byte{0x0d}, "enter"},
		{"tab", []byte{0x09}, "tab"},
		{"backspace", []byte{0x7f}, "backspace"},
		{"escape", []byte{0x1b}, "esc"},
		{"ctrl+c", []byte{0x03}, "ctrl+c"},
		{"ctrl+a", []byte{0x01}, "ctrl+a"},
		{"ctrl+z", []byte{0x1a}, "ctrl+z"},
		{"letter a", []byte("a"), "a"},
		{"letter z", []byte("z"), "z"},
		{"arrow up", []byte{0x1b, '[', 'A'}, "up"},
		{"arrow down", []byte{0x1b, '[', 'B'}, "down"},
		{"arrow right", []byte{0x1b, '[', 'C'}, "right"},
		{"arrow left", []byte{0x1b, '[', 'D'}, "left"},
		{"home", []byte{0x1b, '[', 'H'}, "home"},
		{"end", []byte{0x1b, '[', 'F'}, "end"},
		{"delete", []byte{0x1b, '[', '3', '~'}, "delete"},
		{"pgup", []byte{0x1b, '[', '5', '~'}, "pgup"},
		{"pgdown", []byte{0x1b, '[', '6', '~'}, "pgdown"},
		{"shift+tab", []byte{0x1b, '[', 'Z'}, "shift+tab"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := parseInput(tt.input)
			if ev.Key != tt.want {
				t.Errorf("parseInput(%v) key = %q, want %q", tt.input, ev.Key, tt.want)
			}
			if ev.Type != EventKey {
				t.Errorf("expected EventKey type")
			}
		})
	}
}

func TestParseInputUTF8(t *testing.T) {
	ev := parseInput([]byte("あ"))
	if ev.Key != "あ" {
		t.Errorf("expected key=あ, got %q", ev.Key)
	}
	if len(ev.Runes) != 1 || ev.Runes[0] != 'あ' {
		t.Errorf("expected rune あ, got %v", ev.Runes)
	}
}

func TestParseInputIME(t *testing.T) {
	// IME confirmation sends multiple runes in a single read.
	ev := parseInput([]byte("日本語"))
	if ev.Key != "日本語" {
		t.Errorf("expected key=日本語, got %q", ev.Key)
	}
	if len(ev.Runes) != 3 {
		t.Fatalf("expected 3 runes, got %d", len(ev.Runes))
	}
	want := []rune{'日', '本', '語'}
	for i, r := range ev.Runes {
		if r != want[i] {
			t.Errorf("rune[%d] = %q, want %q", i, r, want[i])
		}
	}
}

func TestParseInputMultiASCII(t *testing.T) {
	// Pasted ASCII text arrives as multiple bytes in one read.
	ev := parseInput([]byte("hello"))
	if ev.Key != "hello" {
		t.Errorf("expected key=hello, got %q", ev.Key)
	}
	if len(ev.Runes) != 5 {
		t.Errorf("expected 5 runes, got %d", len(ev.Runes))
	}
}
