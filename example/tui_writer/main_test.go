package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tomo3110/gerbera/tui/app"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func mountTestWriter(t *testing.T) (*app.TestView, *WriterView) {
	t.Helper()
	view := &WriterView{}
	tv := app.NewTestView(view)
	params := app.Params{
		"width":  "80",
		"height": "24",
	}
	if err := tv.Mount(params); err != nil {
		t.Fatal(err)
	}
	return tv, view
}

func TestWriterViewMount(t *testing.T) {
	tv, view := mountTestWriter(t)

	out, err := tv.RenderString()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, "No file open") {
		t.Errorf("expected 'No file open' in initial render, got:\n%s", out)
	}

	if view.width != 80 {
		t.Errorf("expected width=80, got %d", view.width)
	}
	if view.height != 24 {
		t.Errorf("expected height=24, got %d", view.height)
	}
}

func TestWriterViewNavigateTree(t *testing.T) {
	tv, view := mountTestWriter(t)

	if len(view.visibleNodes) == 0 {
		t.Skip("no entries in current directory")
	}

	initialCursor := view.treeCursor
	if initialCursor != 0 {
		t.Errorf("expected initial cursor=0, got %d", initialCursor)
	}

	// Move down
	_, err := tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "j"})
	if err != nil {
		t.Fatal(err)
	}
	if len(view.visibleNodes) > 1 && view.treeCursor != 1 {
		t.Errorf("expected cursor=1 after j, got %d", view.treeCursor)
	}

	// Move up
	_, err = tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "k"})
	if err != nil {
		t.Fatal(err)
	}
	if view.treeCursor != 0 {
		t.Errorf("expected cursor=0 after k, got %d", view.treeCursor)
	}

	// Can't go above 0
	_, err = tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "k"})
	if err != nil {
		t.Fatal(err)
	}
	if view.treeCursor != 0 {
		t.Errorf("expected cursor=0 after extra k, got %d", view.treeCursor)
	}
}

func TestWriterViewResize(t *testing.T) {
	tv, view := mountTestWriter(t)

	_, err := tv.SimulateEvent(app.Event{
		Type:   app.EventResize,
		Width:  120,
		Height: 40,
	})
	if err != nil {
		t.Fatal(err)
	}

	if view.width != 120 {
		t.Errorf("expected width=120 after resize, got %d", view.width)
	}
	if view.height != 40 {
		t.Errorf("expected height=40 after resize, got %d", view.height)
	}
}

func TestWriterViewSearchMode(t *testing.T) {
	tv, view := mountTestWriter(t)

	if len(view.visibleNodes) == 0 {
		t.Skip("no entries in current directory")
	}

	// Enter search mode
	_, err := tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "/"})
	if err != nil {
		t.Fatal(err)
	}
	if view.mode != modeSearch {
		t.Errorf("expected modeSearch after /, got %d", view.mode)
	}

	// Type a character
	_, err = tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "m", Runes: []rune{'m'}})
	if err != nil {
		t.Fatal(err)
	}
	if view.searchQuery != "m" {
		t.Errorf("expected searchQuery='m', got %q", view.searchQuery)
	}

	// Cancel with Esc
	_, err = tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "esc"})
	if err != nil {
		t.Fatal(err)
	}
	if view.mode != modeTree {
		t.Errorf("expected modeTree after Esc, got %d", view.mode)
	}
}

func TestEditorBuffer(t *testing.T) {
	t.Run("insertRune", func(t *testing.T) {
		buf := newEditorBuffer()
		buf.insertRune('H')
		buf.insertRune('i')
		if buf.content() != "Hi" {
			t.Errorf("expected 'Hi', got %q", buf.content())
		}
		if buf.curCol != 2 {
			t.Errorf("expected curCol=2, got %d", buf.curCol)
		}
	})

	t.Run("backspace", func(t *testing.T) {
		buf := newEditorBuffer()
		buf.insertRune('A')
		buf.insertRune('B')
		buf.insertRune('C')
		buf.backspace()
		if buf.content() != "AB" {
			t.Errorf("expected 'AB', got %q", buf.content())
		}
		if buf.curCol != 2 {
			t.Errorf("expected curCol=2, got %d", buf.curCol)
		}
	})

	t.Run("backspace at beginning of line joins lines", func(t *testing.T) {
		buf := newEditorBuffer()
		buf.insertRune('A')
		buf.newLine()
		buf.insertRune('B')
		buf.curCol = 0
		buf.backspace()
		if buf.content() != "AB" {
			t.Errorf("expected 'AB', got %q", buf.content())
		}
		if buf.curRow != 0 {
			t.Errorf("expected curRow=0, got %d", buf.curRow)
		}
		if buf.curCol != 1 {
			t.Errorf("expected curCol=1, got %d", buf.curCol)
		}
	})

	t.Run("newLine", func(t *testing.T) {
		buf := newEditorBuffer()
		buf.insertRune('A')
		buf.insertRune('B')
		buf.curCol = 1
		buf.newLine()
		if len(buf.lines) != 2 {
			t.Errorf("expected 2 lines, got %d", len(buf.lines))
		}
		if buf.lines[0] != "A" {
			t.Errorf("expected first line 'A', got %q", buf.lines[0])
		}
		if buf.lines[1] != "B" {
			t.Errorf("expected second line 'B', got %q", buf.lines[1])
		}
		if buf.curRow != 1 || buf.curCol != 0 {
			t.Errorf("expected cursor at (1,0), got (%d,%d)", buf.curRow, buf.curCol)
		}
	})

	t.Run("dirty flag", func(t *testing.T) {
		buf := newEditorBuffer()
		if buf.dirty {
			t.Error("expected dirty=false initially")
		}
		buf.insertRune('X')
		if !buf.dirty {
			t.Error("expected dirty=true after insert")
		}
	})

	t.Run("load resets state", func(t *testing.T) {
		buf := newEditorBuffer()
		buf.insertRune('X')
		buf.load("Hello\nWorld")
		if buf.curRow != 0 || buf.curCol != 0 {
			t.Errorf("expected cursor at (0,0), got (%d,%d)", buf.curRow, buf.curCol)
		}
		if buf.dirty {
			t.Error("expected dirty=false after load")
		}
		if len(buf.lines) != 2 {
			t.Errorf("expected 2 lines, got %d", len(buf.lines))
		}
	})
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		want     string
	}{
		{"ascii short", "Hello", 10, "Hello"},
		{"ascii exact", "Hello", 5, "Hello"},
		{"ascii truncate", "Hello World", 5, "Hello"},
		{"empty", "", 5, ""},
		{"cjk truncate", "あいう", 4, "あい"},
		{"cjk exact", "あい", 4, "あい"},
		{"mixed", "Aあ", 3, "Aあ"},
		{"mixed truncate", "Aあい", 3, "Aあ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxWidth)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxWidth, got, tt.want)
			}
		})
	}
}

func TestWriterViewTabToEditor(t *testing.T) {
	tv, view := mountTestWriter(t)

	_, err := tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "tab"})
	if err != nil {
		t.Fatal(err)
	}
	if view.mode != modeTree {
		t.Errorf("expected modeTree when no file open, got %d", view.mode)
	}
}

func TestWriterViewModeDisplay(t *testing.T) {
	tv, view := mountTestWriter(t)

	out, err := tv.RenderString()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, "TREE") {
		t.Errorf("expected 'TREE' in status bar, got:\n%s", out)
	}

	_, err = tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "/"})
	if err != nil {
		t.Fatal(err)
	}
	out, err = tv.RenderString()
	if err != nil {
		t.Fatal(err)
	}
	if view.mode != modeSearch {
		t.Errorf("expected modeSearch, got %d", view.mode)
	}
}

func TestTreeExpandCollapse(t *testing.T) {
	tv, view := mountTestWriter(t)

	// Find first directory in visible nodes
	dirIdx := -1
	for i, n := range view.visibleNodes {
		if n.isDir {
			dirIdx = i
			break
		}
	}
	if dirIdx < 0 {
		t.Skip("no directory found in tree")
	}

	// Navigate to the directory
	for i := 0; i < dirIdx; i++ {
		tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "j"})
	}

	dirNode := view.visibleNodes[dirIdx]
	if dirNode.expanded {
		t.Fatal("expected directory to start collapsed")
	}

	countBefore := len(view.visibleNodes)

	// Expand with Enter
	_, err := tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "enter"})
	if err != nil {
		t.Fatal(err)
	}

	if !dirNode.expanded {
		t.Error("expected directory to be expanded after Enter")
	}

	countAfter := len(view.visibleNodes)
	if dirNode.loaded && len(dirNode.children) > 0 && countAfter <= countBefore {
		t.Errorf("expected more visible nodes after expand, before=%d after=%d", countBefore, countAfter)
	}

	// Collapse with h
	_, err = tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "h"})
	if err != nil {
		t.Fatal(err)
	}

	if dirNode.expanded {
		t.Error("expected directory to be collapsed after h")
	}

	countCollapsed := len(view.visibleNodes)
	if countCollapsed != countBefore {
		t.Errorf("expected visible count to restore after collapse, before=%d after=%d", countBefore, countCollapsed)
	}
}

func TestTreeNodeLoadChildren(t *testing.T) {
	// Create a temp directory structure
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "subdir", "child.txt"), []byte("world"), 0644)

	node := &treeNode{
		name:  "testdir",
		path:  tmpDir,
		isDir: true,
		depth: 0,
	}

	node.loadChildren()

	if !node.loaded {
		t.Error("expected loaded=true after loadChildren")
	}
	if len(node.children) != 2 {
		t.Errorf("expected 2 children, got %d", len(node.children))
	}

	// Directories should come first
	if len(node.children) >= 2 {
		if !node.children[0].isDir {
			t.Error("expected first child to be directory (sorted)")
		}
		if node.children[1].isDir {
			t.Error("expected second child to be file")
		}
	}
}

func TestFlattenVisible(t *testing.T) {
	root := []*treeNode{
		{name: "dir1", isDir: true, expanded: true, children: []*treeNode{
			{name: "a.txt", isDir: false, depth: 1},
			{name: "b.txt", isDir: false, depth: 1},
		}},
		{name: "dir2", isDir: true, expanded: false, children: []*treeNode{
			{name: "c.txt", isDir: false, depth: 1},
		}},
		{name: "file.go", isDir: false},
	}

	visible := flattenVisible(root)

	// dir1, a.txt, b.txt, dir2, file.go (dir2 collapsed, so c.txt hidden)
	if len(visible) != 5 {
		t.Errorf("expected 5 visible nodes, got %d", len(visible))
		for i, n := range visible {
			t.Logf("  [%d] %s (dir=%v)", i, n.name, n.isDir)
		}
	}

	expected := []string{"dir1", "a.txt", "b.txt", "dir2", "file.go"}
	for i, name := range expected {
		if i < len(visible) && visible[i].name != name {
			t.Errorf("visible[%d] = %q, want %q", i, visible[i].name, name)
		}
	}
}

func TestExpandTabs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no tabs", "hello", "hello"},
		{"tab at start", "\thello", "    hello"},
		{"tab after 1 char", "a\tb", "a   b"},
		{"tab after 2 chars", "ab\tc", "ab  c"},
		{"tab after 4 chars", "abcd\te", "abcd    e"},
		{"multiple tabs", "\t\t", "        "},
		{"mixed", "a\tb\tc", "a   b   c"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandTabs(tt.input, tabWidth)
			if got != tt.want {
				t.Errorf("expandTabs(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildCursorLine(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		curCol    int
		wantBef   string
		wantCur   string
		wantAfter string
	}{
		{
			name: "cursor on normal char",
			line: "hello", curCol: 2,
			wantBef: "he", wantCur: "l", wantAfter: "lo",
		},
		{
			name: "cursor at end",
			line: "hi", curCol: 2,
			wantBef: "hi", wantCur: " ", wantAfter: "",
		},
		{
			name: "cursor on tab at start",
			line: "\thello", curCol: 0,
			wantBef: "", wantCur: " ", wantAfter: "   hello",
		},
		{
			name: "cursor after tab",
			line: "\thello", curCol: 1,
			wantBef: "    ", wantCur: "h", wantAfter: "ello",
		},
		{
			name: "cursor on tab after char",
			line: "a\tb", curCol: 1,
			wantBef: "a", wantCur: " ", wantAfter: "  b",
		},
		{
			name: "cursor on char after tab",
			line: "a\tb", curCol: 2,
			wantBef: "a   ", wantCur: "b", wantAfter: "",
		},
		{
			name: "empty line cursor at 0",
			line: "", curCol: 0,
			wantBef: "", wantCur: " ", wantAfter: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runes := []rune(tt.line)
			bef, cur, aft := buildCursorLine(runes, tt.curCol, tabWidth)
			if bef != tt.wantBef {
				t.Errorf("before: got %q, want %q", bef, tt.wantBef)
			}
			if cur != tt.wantCur {
				t.Errorf("cursor: got %q, want %q", cur, tt.wantCur)
			}
			if aft != tt.wantAfter {
				t.Errorf("after: got %q, want %q", aft, tt.wantAfter)
			}
		})
	}
}

func TestFindParentIndex(t *testing.T) {
	nodes := []*treeNode{
		{name: "dir1", path: "/root/dir1", isDir: true},
		{name: "a.txt", path: "/root/dir1/a.txt", isDir: false},
		{name: "b.txt", path: "/root/dir1/b.txt", isDir: false},
		{name: "file.go", path: "/root/file.go", isDir: false},
	}

	// Parent of a.txt (idx=1) should be dir1 (idx=0)
	pi := findParentIndex(nodes, 1)
	if pi != 0 {
		t.Errorf("expected parent index 0, got %d", pi)
	}

	// Parent of b.txt (idx=2) should be dir1 (idx=0)
	pi = findParentIndex(nodes, 2)
	if pi != 0 {
		t.Errorf("expected parent index 0, got %d", pi)
	}

	// Parent of dir1 (idx=0) should be -1 (no parent visible)
	pi = findParentIndex(nodes, 0)
	if pi != -1 {
		t.Errorf("expected parent index -1, got %d", pi)
	}
}

// ---- New tests for encoding, temp files, new modes ----

func TestDetectEncoding(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantIdx int
	}{
		{
			name:    "UTF-8 BOM",
			data:    append([]byte{0xEF, 0xBB, 0xBF}, []byte("hello")...),
			wantIdx: 1, // UTF-8 BOM
		},
		{
			name:    "UTF-16LE BOM",
			data:    []byte{0xFF, 0xFE, 'h', 0, 'i', 0},
			wantIdx: 5, // UTF-16LE
		},
		{
			name:    "UTF-16BE BOM",
			data:    []byte{0xFE, 0xFF, 0, 'h', 0, 'i'},
			wantIdx: 6, // UTF-16BE
		},
		{
			name:    "plain UTF-8",
			data:    []byte("Hello, World!"),
			wantIdx: 0, // UTF-8
		},
		{
			name:    "UTF-8 with Japanese",
			data:    []byte("こんにちは"),
			wantIdx: 0, // UTF-8
		},
		{
			name:    "empty",
			data:    []byte{},
			wantIdx: 0, // UTF-8
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectEncoding(tt.data)
			if got != tt.wantIdx {
				t.Errorf("detectEncoding() = %d (%s), want %d (%s)",
					got, supportedEncodings[got].name,
					tt.wantIdx, supportedEncodings[tt.wantIdx].name)
			}
		})
	}
}

func TestDecodeEncodeRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		encIdx int
		text   string
	}{
		{"UTF-8", 0, "Hello, World!"},
		{"UTF-8 Japanese", 0, "こんにちは世界"},
		{"Shift-JIS", 2, "こんにちは"},
		{"EUC-JP", 3, "テスト"},
		{"ISO-8859-1", 4, "Hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode then decode
			encoded := encodeString(tt.text, tt.encIdx)
			decoded := decodeBytes(encoded, tt.encIdx)
			if decoded != tt.text {
				t.Errorf("round-trip failed: encoded→decoded = %q, want %q", decoded, tt.text)
			}
		})
	}
}

func TestDetectShiftJIS(t *testing.T) {
	// Encode "テスト" in Shift-JIS
	sjisData, _, err := transform.Bytes(japanese.ShiftJIS.NewEncoder(), []byte("テスト"))
	if err != nil {
		t.Fatal(err)
	}
	idx := detectEncoding(sjisData)
	// Should detect as Shift-JIS (2) — the data is invalid UTF-8
	if idx != 2 {
		t.Errorf("expected Shift-JIS (2), got %d (%s)", idx, supportedEncodings[idx].name)
	}
}

func TestNewFileMode(t *testing.T) {
	tv, view := mountTestWriter(t)

	// Press + to enter new file mode
	_, err := tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "+"})
	if err != nil {
		t.Fatal(err)
	}
	if view.mode != modeNewFile {
		t.Errorf("expected modeNewFile after +, got %d", view.mode)
	}

	// Type a filename
	for _, r := range "test" {
		_, err = tv.SimulateEvent(app.Event{Type: app.EventKey, Key: string(r), Runes: []rune{r}})
		if err != nil {
			t.Fatal(err)
		}
	}
	if view.newFileName != "test" {
		t.Errorf("expected newFileName='test', got %q", view.newFileName)
	}

	// Backspace
	_, err = tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "backspace"})
	if err != nil {
		t.Fatal(err)
	}
	if view.newFileName != "tes" {
		t.Errorf("expected newFileName='tes', got %q", view.newFileName)
	}

	// Cancel with Esc
	_, err = tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "esc"})
	if err != nil {
		t.Fatal(err)
	}
	if view.mode != modeTree {
		t.Errorf("expected modeTree after Esc, got %d", view.mode)
	}
	if view.newFileName != "" {
		t.Errorf("expected newFileName cleared after Esc, got %q", view.newFileName)
	}
}

func TestEncodingMode(t *testing.T) {
	tv, view := mountTestWriter(t)

	// Need to open a file first
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(tmpFile, []byte("hello"), 0644)

	view.currentFile = tmpFile
	view.rawBytes = []byte("hello")
	view.fileEncoding = 0
	view.editor.load("hello")

	// Press e to enter encoding mode
	_, err := tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "e"})
	if err != nil {
		t.Fatal(err)
	}
	if view.mode != modeEncoding {
		t.Errorf("expected modeEncoding after e, got %d", view.mode)
	}
	if view.encodingCursor != 0 {
		t.Errorf("expected encodingCursor=0 (current encoding), got %d", view.encodingCursor)
	}

	// Move down with j
	_, err = tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "j"})
	if err != nil {
		t.Fatal(err)
	}
	if view.encodingCursor != 1 {
		t.Errorf("expected encodingCursor=1 after j, got %d", view.encodingCursor)
	}

	// Move up with k
	_, err = tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "k"})
	if err != nil {
		t.Fatal(err)
	}
	if view.encodingCursor != 0 {
		t.Errorf("expected encodingCursor=0 after k, got %d", view.encodingCursor)
	}

	// Cancel with Esc
	_, err = tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "esc"})
	if err != nil {
		t.Fatal(err)
	}
	if view.mode != modeTree {
		t.Errorf("expected modeTree after Esc, got %d", view.mode)
	}
}

func TestEncodingModeNoFileOpen(t *testing.T) {
	tv, view := mountTestWriter(t)

	// Press e without file open — should stay in tree mode
	_, err := tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "e"})
	if err != nil {
		t.Fatal(err)
	}
	if view.mode != modeTree {
		t.Errorf("expected modeTree when no file open, got %d", view.mode)
	}
}

func TestSaveDiscardWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(tmpFile, []byte("original"), 0644)

	view := &WriterView{}
	tv := app.NewTestView(view)
	if err := tv.Mount(app.Params{"width": "80", "height": "24"}); err != nil {
		t.Fatal(err)
	}

	// Open file directly
	view.openFile(tmpFile)
	if view.editor.content() != "original" {
		t.Fatalf("expected 'original', got %q", view.editor.content())
	}

	// Edit the file
	view.editor.insertRune('X')
	if !view.editor.dirty {
		t.Fatal("expected dirty after edit")
	}

	// Tab to tree (should auto-save temp)
	_, err := tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "tab"})
	if err != nil {
		t.Fatal(err)
	}
	if view.mode != modeTree {
		t.Fatalf("expected modeTree, got %d", view.mode)
	}

	// Verify temp file exists
	if !hasTempFile(view.tempDir, tmpFile) {
		t.Error("expected temp file to exist after Tab")
	}

	// Discard with u
	_, err = tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "u"})
	if err != nil {
		t.Fatal(err)
	}

	// Verify content is restored
	if view.editor.content() != "original" {
		t.Errorf("expected 'original' after discard, got %q", view.editor.content())
	}
	if view.editor.dirty {
		t.Error("expected dirty=false after discard")
	}

	// Verify temp file is removed
	if hasTempFile(view.tempDir, tmpFile) {
		t.Error("expected temp file to be removed after discard")
	}
}

func TestSaveWithW(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(tmpFile, []byte("original"), 0644)

	view := &WriterView{}
	tv := app.NewTestView(view)
	if err := tv.Mount(app.Params{"width": "80", "height": "24"}); err != nil {
		t.Fatal(err)
	}

	// Open and edit
	view.openFile(tmpFile)
	view.editor.load("modified content")
	view.editor.dirty = true

	// Tab to tree
	_, err := tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "tab"})
	if err != nil {
		t.Fatal(err)
	}

	// Save with w
	_, err = tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "w"})
	if err != nil {
		t.Fatal(err)
	}

	// Verify file on disk
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "modified content" {
		t.Errorf("expected 'modified content' on disk, got %q", string(data))
	}

	// Verify dirty flag cleared
	if view.editor.dirty {
		t.Error("expected dirty=false after save")
	}

	// Verify temp file removed
	if hasTempFile(view.tempDir, tmpFile) {
		t.Error("expected temp file removed after save")
	}
}

func TestTempFileManagement(t *testing.T) {
	tempDir, err := initTempDir()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanupTempDir(tempDir)

	// Verify directory exists
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Fatal("expected temp dir to exist")
	}

	testPath := "/some/file/path.txt"
	testData := []byte("test content")

	// Save temp file
	if err := saveTempFile(tempDir, testPath, testData); err != nil {
		t.Fatal(err)
	}

	// Verify it exists
	if !hasTempFile(tempDir, testPath) {
		t.Error("expected temp file to exist")
	}

	// Load and verify
	loaded := loadTempFile(tempDir, testPath)
	if string(loaded) != "test content" {
		t.Errorf("expected 'test content', got %q", string(loaded))
	}

	// Remove
	removeTempFile(tempDir, testPath)
	if hasTempFile(tempDir, testPath) {
		t.Error("expected temp file to be removed")
	}

	// Cleanup
	cleanupTempDir(tempDir)
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Error("expected temp dir to be removed after cleanup")
	}
}

func TestStatusBarShowsEncoding(t *testing.T) {
	tv, view := mountTestWriter(t)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(tmpFile, []byte("hello"), 0644)

	view.openFile(tmpFile)
	if view.mode != modeEditor {
		t.Fatalf("expected modeEditor, got %d", view.mode)
	}

	out, err := tv.RenderString()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, "UTF-8") {
		t.Errorf("expected 'UTF-8' in status bar, got:\n%s", out)
	}

	// Also check Ln/Col is shown
	if !strings.Contains(out, "Ln") {
		t.Errorf("expected 'Ln' in status bar, got:\n%s", out)
	}
}

func TestCtrlSRemovedFromEditor(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(tmpFile, []byte("original"), 0644)

	view := &WriterView{}
	tv := app.NewTestView(view)
	if err := tv.Mount(app.Params{"width": "80", "height": "24"}); err != nil {
		t.Fatal(err)
	}

	view.openFile(tmpFile)
	view.editor.insertRune('X')

	// Press Ctrl+S — should NOT save (no handler for it)
	_, err := tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "ctrl+s"})
	if err != nil {
		t.Fatal(err)
	}

	// File on disk should be unchanged
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "original" {
		t.Errorf("expected 'original' on disk (Ctrl+S should not save), got %q", string(data))
	}

	// Dirty flag should still be set
	if !view.editor.dirty {
		t.Error("expected dirty=true (Ctrl+S should not clear dirty)")
	}
}

func TestCreateNewFile(t *testing.T) {
	tmpDir := t.TempDir()

	view := &WriterView{}
	tv := app.NewTestView(view)
	if err := tv.Mount(app.Params{"width": "80", "height": "24"}); err != nil {
		t.Fatal(err)
	}

	// Override rootDir to use temp directory
	view.rootDir = tmpDir
	view.loadRootDir()

	// Enter new file mode
	_, err := tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "+"})
	if err != nil {
		t.Fatal(err)
	}

	// Type filename
	for _, r := range "hello.txt" {
		tv.SimulateEvent(app.Event{Type: app.EventKey, Key: string(r), Runes: []rune{r}})
	}

	// Confirm
	_, err = tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "enter"})
	if err != nil {
		t.Fatal(err)
	}

	// Verify file was created
	newPath := filepath.Join(tmpDir, "hello.txt")
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Error("expected new file to be created on disk")
	}

	// Verify we switched to editor mode
	if view.mode != modeEditor {
		t.Errorf("expected modeEditor after creating file, got %d", view.mode)
	}

	// Verify the file is opened
	if view.currentFile != newPath {
		t.Errorf("expected currentFile=%q, got %q", newPath, view.currentFile)
	}

	// Verify encoding is UTF-8
	if view.fileEncoding != 0 {
		t.Errorf("expected UTF-8 encoding (0), got %d", view.fileEncoding)
	}
}

func TestCreateNewFileDuplicate(t *testing.T) {
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "existing.txt")
	os.WriteFile(existingFile, []byte("data"), 0644)

	view := &WriterView{}
	tv := app.NewTestView(view)
	if err := tv.Mount(app.Params{"width": "80", "height": "24"}); err != nil {
		t.Fatal(err)
	}

	view.rootDir = tmpDir
	view.loadRootDir()

	// Try to create a file with existing name
	view.mode = modeNewFile
	view.newFileName = "existing.txt"
	_, err := tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "enter"})
	if err != nil {
		t.Fatal(err)
	}

	// Should show error and stay in tree mode (createNewFile sets statusMsg)
	if !strings.Contains(view.statusMsg, "already exists") {
		t.Errorf("expected 'already exists' error, got %q", view.statusMsg)
	}
}

func TestEncodingSelectAndReopen(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(tmpFile, []byte("hello"), 0644)

	view := &WriterView{}
	tv := app.NewTestView(view)
	if err := tv.Mount(app.Params{"width": "80", "height": "24"}); err != nil {
		t.Fatal(err)
	}

	// Open file
	view.openFile(tmpFile)

	// Go to tree
	tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "tab"})

	// Enter encoding mode
	tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "e"})
	if view.mode != modeEncoding {
		t.Fatalf("expected modeEncoding, got %d", view.mode)
	}

	// Select a different encoding (move to Shift-JIS = index 2)
	tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "j"}) // 1
	tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "j"}) // 2
	tv.SimulateEvent(app.Event{Type: app.EventKey, Key: "enter"})

	if view.fileEncoding != 2 {
		t.Errorf("expected encoding 2 (Shift-JIS), got %d", view.fileEncoding)
	}
	if view.mode != modeEditor {
		t.Errorf("expected modeEditor after encoding selection, got %d", view.mode)
	}
}

func TestDirtyFileMarkerInTree(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(tmpFile, []byte("hello"), 0644)

	view := &WriterView{}
	tv := app.NewTestView(view)
	if err := tv.Mount(app.Params{"width": "80", "height": "24"}); err != nil {
		t.Fatal(err)
	}

	view.rootDir = tmpDir
	view.loadRootDir()

	// Mark the file as dirty
	view.dirtyFiles[tmpFile] = true

	out, err := tv.RenderString()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, "[+]") {
		t.Errorf("expected '[+]' dirty marker in tree, got:\n%s", out)
	}
}
