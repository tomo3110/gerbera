package app

import (
	"fmt"
	"io"
	"unicode/utf8"

	"golang.org/x/sys/unix"
)

// EnableRawMode puts the terminal into raw mode and returns a restore function.
func EnableRawMode(fd int) (restore func(), err error) {
	orig, err := unix.IoctlGetTermios(fd, ioctlReadTermios)
	if err != nil {
		return nil, fmt.Errorf("get termios: %w", err)
	}

	raw := *orig
	// cfmakeraw equivalent
	raw.Iflag &^= unix.BRKINT | unix.ICRNL | unix.INPCK | unix.ISTRIP | unix.IXON
	raw.Oflag &^= unix.OPOST
	raw.Cflag &^= unix.CSIZE | unix.PARENB
	raw.Cflag |= unix.CS8
	raw.Lflag &^= unix.ECHO | unix.ICANON | unix.IEXTEN | unix.ISIG
	raw.Cc[unix.VMIN] = 1
	raw.Cc[unix.VTIME] = 0

	if err := unix.IoctlSetTermios(fd, ioctlWriteTermios, &raw); err != nil {
		return nil, fmt.Errorf("set raw mode: %w", err)
	}

	return func() {
		unix.IoctlSetTermios(fd, ioctlWriteTermios, orig)
	}, nil
}

// GetTerminalSize returns the terminal width and height.
func GetTerminalSize(fd int) (width, height int, err error) {
	ws, err := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)
	if err != nil {
		return 0, 0, err
	}
	return int(ws.Col), int(ws.Row), nil
}

// ClearScreen clears the terminal and moves cursor to home position.
func ClearScreen(w io.Writer) {
	io.WriteString(w, "\033[2J\033[H")
}

// MoveCursor moves the cursor to the given row and column (1-based).
func MoveCursor(w io.Writer, row, col int) {
	fmt.Fprintf(w, "\033[%d;%dH", row, col)
}

// HideCursor hides the cursor.
func HideCursor(w io.Writer) {
	io.WriteString(w, "\033[?25l")
}

// ShowCursor shows the cursor.
func ShowCursor(w io.Writer) {
	io.WriteString(w, "\033[?25h")
}

// ReadEvent reads a single input event from the reader.
func ReadEvent(r io.Reader) (Event, error) {
	buf := make([]byte, 64)
	n, err := r.Read(buf)
	if err != nil {
		return Event{}, err
	}
	buf = buf[:n]

	return parseInput(buf), nil
}

func parseInput(buf []byte) Event {
	if len(buf) == 0 {
		return Event{Type: EventKey, Key: ""}
	}

	b := buf[0]

	// Ctrl characters
	if b <= 0x1a && b != 0x09 && b != 0x0d && b != 0x1b {
		key := "ctrl+" + string(rune('a'+b-1))
		return Event{Type: EventKey, Key: key}
	}

	switch b {
	case 0x0d:
		return Event{Type: EventKey, Key: "enter"}
	case 0x09:
		return Event{Type: EventKey, Key: "tab"}
	case 0x7f:
		return Event{Type: EventKey, Key: "backspace"}
	case 0x1b:
		return parseEscapeSequence(buf)
	}

	// Normal UTF-8 characters.
	// IME confirmation may deliver multiple runes in a single read,
	// so decode all runes from the buffer.
	var runes []rune
	for len(buf) > 0 {
		r, size := utf8.DecodeRune(buf)
		if r == utf8.RuneError {
			break
		}
		runes = append(runes, r)
		buf = buf[size:]
	}
	if len(runes) > 0 {
		return Event{Type: EventKey, Key: string(runes), Runes: runes}
	}

	return Event{Type: EventKey, Key: ""}
}

func parseEscapeSequence(buf []byte) Event {
	if len(buf) == 1 {
		return Event{Type: EventKey, Key: "esc"}
	}

	if buf[1] == '[' {
		return parseCSI(buf[2:])
	}

	// Alt+key
	if len(buf) == 2 {
		return Event{Type: EventKey, Key: "alt+" + string(rune(buf[1]))}
	}

	return Event{Type: EventKey, Key: "esc"}
}

func parseCSI(buf []byte) Event {
	if len(buf) == 0 {
		return Event{Type: EventKey, Key: "esc"}
	}

	switch buf[0] {
	case 'A':
		return Event{Type: EventKey, Key: "up"}
	case 'B':
		return Event{Type: EventKey, Key: "down"}
	case 'C':
		return Event{Type: EventKey, Key: "right"}
	case 'D':
		return Event{Type: EventKey, Key: "left"}
	case 'H':
		return Event{Type: EventKey, Key: "home"}
	case 'F':
		return Event{Type: EventKey, Key: "end"}
	case 'Z':
		return Event{Type: EventKey, Key: "shift+tab"}
	}

	// Extended sequences like ESC [ 3 ~
	if len(buf) >= 2 && buf[1] == '~' {
		switch buf[0] {
		case '1':
			return Event{Type: EventKey, Key: "home"}
		case '2':
			return Event{Type: EventKey, Key: "insert"}
		case '3':
			return Event{Type: EventKey, Key: "delete"}
		case '4':
			return Event{Type: EventKey, Key: "end"}
		case '5':
			return Event{Type: EventKey, Key: "pgup"}
		case '6':
			return Event{Type: EventKey, Key: "pgdown"}
		}
	}

	return Event{Type: EventKey, Key: "esc"}
}
