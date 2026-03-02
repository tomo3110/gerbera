package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	g "github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/tui"
	"github.com/tomo3110/gerbera/tui/app"
	tc "github.com/tomo3110/gerbera/tui/components"
	ts "github.com/tomo3110/gerbera/tui/style"
	"github.com/mattn/go-runewidth"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// mode represents the active input mode.
type mode int

const (
	modeTree     mode = iota // File tree navigation
	modeSearch               // File search
	modeEditor               // Text editor
	modeNewFile              // New file name input
	modeEncoding             // Encoding selection
)

const maxFileSize = 1 << 20 // 1MB

// globalTempDir is set by main() for cleanup on exit.
var globalTempDir string

// ---------------- Encoding Support ----------------

type encodingInfo struct {
	name string
	enc  encoding.Encoding // nil = UTF-8
}

var supportedEncodings = []encodingInfo{
	{"UTF-8", nil},
	{"UTF-8 BOM", unicode.UTF8BOM},
	{"Shift-JIS", japanese.ShiftJIS},
	{"EUC-JP", japanese.EUCJP},
	{"ISO-8859-1", charmap.ISO8859_1},
	{"UTF-16LE", unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)},
	{"UTF-16BE", unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)},
}

// detectEncoding returns the index into supportedEncodings for the given data.
func detectEncoding(data []byte) int {
	// BOM check
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return 1 // UTF-8 BOM
	}
	if len(data) >= 2 {
		if data[0] == 0xFF && data[1] == 0xFE {
			return 5 // UTF-16LE
		}
		if data[0] == 0xFE && data[1] == 0xFF {
			return 6 // UTF-16BE
		}
	}

	// Valid UTF-8?
	if utf8.Valid(data) {
		return 0 // UTF-8
	}

	// Try Shift-JIS
	if _, _, err := transform.Bytes(japanese.ShiftJIS.NewDecoder(), data); err == nil {
		return 2 // Shift-JIS
	}

	// Try EUC-JP
	if _, _, err := transform.Bytes(japanese.EUCJP.NewDecoder(), data); err == nil {
		return 3 // EUC-JP
	}

	// Fallback to ISO-8859-1
	return 4
}

// decodeBytes converts raw bytes to a UTF-8 string using the given encoding index.
func decodeBytes(data []byte, encIdx int) string {
	enc := supportedEncodings[encIdx].enc
	if enc == nil {
		return string(data)
	}
	decoded, _, err := transform.Bytes(enc.NewDecoder(), data)
	if err != nil {
		return string(data)
	}
	return string(decoded)
}

// encodeString converts a UTF-8 string to bytes in the given encoding.
func encodeString(s string, encIdx int) []byte {
	enc := supportedEncodings[encIdx].enc
	if enc == nil {
		return []byte(s)
	}
	encoded, _, err := transform.Bytes(enc.NewEncoder(), []byte(s))
	if err != nil {
		return []byte(s)
	}
	return encoded
}

// ---------------- Editor Buffer ----------------

// editorBuffer holds the state of the text editor.
type editorBuffer struct {
	lines   []string
	curRow  int
	curCol  int
	scrollY int
	dirty   bool
}

func newEditorBuffer() *editorBuffer {
	return &editorBuffer{
		lines: []string{""},
	}
}

func (b *editorBuffer) load(content string) {
	if content == "" {
		b.lines = []string{""}
	} else {
		b.lines = strings.Split(content, "\n")
	}
	b.curRow = 0
	b.curCol = 0
	b.scrollY = 0
	b.dirty = false
}

func (b *editorBuffer) insertRune(r rune) {
	line := b.lines[b.curRow]
	runes := []rune(line)
	if b.curCol > len(runes) {
		b.curCol = len(runes)
	}
	runes = append(runes[:b.curCol], append([]rune{r}, runes[b.curCol:]...)...)
	b.lines[b.curRow] = string(runes)
	b.curCol++
	b.dirty = true
}

func (b *editorBuffer) backspace() {
	if b.curCol > 0 {
		line := b.lines[b.curRow]
		runes := []rune(line)
		b.curCol--
		runes = append(runes[:b.curCol], runes[b.curCol+1:]...)
		b.lines[b.curRow] = string(runes)
		b.dirty = true
	} else if b.curRow > 0 {
		prevLine := b.lines[b.curRow-1]
		b.curCol = len([]rune(prevLine))
		b.lines[b.curRow-1] = prevLine + b.lines[b.curRow]
		b.lines = append(b.lines[:b.curRow], b.lines[b.curRow+1:]...)
		b.curRow--
		b.dirty = true
	}
}

func (b *editorBuffer) newLine() {
	line := b.lines[b.curRow]
	runes := []rune(line)
	if b.curCol > len(runes) {
		b.curCol = len(runes)
	}
	before := string(runes[:b.curCol])
	after := string(runes[b.curCol:])
	b.lines[b.curRow] = before
	newLines := make([]string, len(b.lines)+1)
	copy(newLines, b.lines[:b.curRow+1])
	newLines[b.curRow+1] = after
	copy(newLines[b.curRow+2:], b.lines[b.curRow+1:])
	b.lines = newLines
	b.curRow++
	b.curCol = 0
	b.dirty = true
}

func (b *editorBuffer) content() string {
	return strings.Join(b.lines, "\n")
}

// ---------------- Tree Node ----------------

// treeNode represents a file or directory in the expandable tree.
type treeNode struct {
	name     string
	path     string // absolute path
	isDir    bool
	expanded bool
	loaded   bool // children loaded from disk
	depth    int
	children []*treeNode
}

// loadChildren reads the directory and populates children.
func (n *treeNode) loadChildren() {
	if n.loaded || !n.isDir {
		return
	}
	n.loaded = true

	dirEntries, err := os.ReadDir(n.path)
	if err != nil {
		return
	}

	for _, e := range dirEntries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		n.children = append(n.children, &treeNode{
			name:  name,
			path:  filepath.Join(n.path, name),
			isDir: e.IsDir(),
			depth: n.depth + 1,
		})
	}
	sortNodes(n.children)
}

func sortNodes(nodes []*treeNode) {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].isDir != nodes[j].isDir {
			return nodes[i].isDir
		}
		return nodes[i].name < nodes[j].name
	})
}

// flattenVisible builds a flat slice of visible nodes from the tree.
func flattenVisible(roots []*treeNode) []*treeNode {
	var result []*treeNode
	var walk func(nodes []*treeNode)
	walk = func(nodes []*treeNode) {
		for _, n := range nodes {
			result = append(result, n)
			if n.isDir && n.expanded {
				walk(n.children)
			}
		}
	}
	walk(roots)
	return result
}

// findParentIndex returns the index in visible of the parent directory node,
// or -1 if not found.
func findParentIndex(visible []*treeNode, idx int) int {
	if idx < 0 || idx >= len(visible) {
		return -1
	}
	target := visible[idx]
	parentPath := filepath.Dir(target.path)
	for i := idx - 1; i >= 0; i-- {
		if visible[i].path == parentPath && visible[i].isDir {
			return i
		}
	}
	return -1
}

// ---------------- Temp File Management ----------------

// tempFilePath returns the path to the temp file for the given file path.
func tempFilePath(tempDir, filePath string) string {
	h := sha256.Sum256([]byte(filePath))
	return filepath.Join(tempDir, fmt.Sprintf("%x", h[:8]))
}

// initTempDir creates a temp directory for this process.
func initTempDir() (string, error) {
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("gerbera-writer-%d", os.Getpid()))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// cleanupTempDir removes the temp directory.
func cleanupTempDir(dir string) {
	if dir != "" {
		os.RemoveAll(dir)
	}
}

// saveTempFile saves content to the temp file for the given file path.
func saveTempFile(tempDir, filePath string, data []byte) error {
	if tempDir == "" {
		return nil
	}
	return os.WriteFile(tempFilePath(tempDir, filePath), data, 0644)
}

// loadTempFile loads content from the temp file, returns nil if not found.
func loadTempFile(tempDir, filePath string) []byte {
	if tempDir == "" {
		return nil
	}
	data, err := os.ReadFile(tempFilePath(tempDir, filePath))
	if err != nil {
		return nil
	}
	return data
}

// removeTempFile removes the temp file for the given file path.
func removeTempFile(tempDir, filePath string) {
	if tempDir == "" {
		return
	}
	os.Remove(tempFilePath(tempDir, filePath))
}

// hasTempFile checks if a temp file exists for the given file path.
func hasTempFile(tempDir, filePath string) bool {
	if tempDir == "" {
		return false
	}
	_, err := os.Stat(tempFilePath(tempDir, filePath))
	return err == nil
}

// ---------------- Writer View ----------------

// WriterView is the main TUI writer application view.
type WriterView struct {
	width  int
	height int

	mode mode

	// File tree
	rootDir      string
	rootNodes    []*treeNode
	visibleNodes []*treeNode
	treeCursor   int
	treeScrollY  int

	// Search
	searchQuery   string
	searchResults []*treeNode

	// Editor
	editor      *editorBuffer
	currentFile string

	// Encoding
	fileEncoding   int              // index into supportedEncodings
	rawBytes       []byte           // raw bytes from disk
	encodingCursor int              // cursor in encoding selection

	// Temp files
	tempDir    string
	dirtyFiles map[string]bool // tracks files with unsaved changes

	// New file creation
	newFileName string

	// Status
	statusMsg string
}

func (v *WriterView) Mount(params app.Params) error {
	if w, err := strconv.Atoi(params["width"]); err == nil {
		v.width = w
	}
	if v.width == 0 {
		v.width = 80
	}
	if h, err := strconv.Atoi(params["height"]); err == nil {
		v.height = h
	}
	if v.height == 0 {
		v.height = 24
	}

	v.mode = modeTree
	v.editor = newEditorBuffer()
	v.dirtyFiles = make(map[string]bool)

	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}
	v.rootDir = dir
	v.loadRootDir()

	// Initialize temp directory
	tempDir, err := initTempDir()
	if err == nil {
		v.tempDir = tempDir
		globalTempDir = tempDir
	}

	return nil
}

func (v *WriterView) loadRootDir() {
	dirEntries, err := os.ReadDir(v.rootDir)
	if err != nil {
		v.statusMsg = "Error: " + err.Error()
		v.rootNodes = nil
		v.visibleNodes = nil
		return
	}

	v.rootNodes = v.rootNodes[:0]
	for _, e := range dirEntries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		v.rootNodes = append(v.rootNodes, &treeNode{
			name:  name,
			path:  filepath.Join(v.rootDir, name),
			isDir: e.IsDir(),
			depth: 0,
		})
	}
	sortNodes(v.rootNodes)
	v.rebuildVisible()
	v.treeCursor = 0
	v.treeScrollY = 0
}

func (v *WriterView) rebuildVisible() {
	v.visibleNodes = flattenVisible(v.rootNodes)
}

func (v *WriterView) openFile(path string) {
	info, err := os.Stat(path)
	if err != nil {
		v.statusMsg = "Error: " + err.Error()
		return
	}
	if info.Size() > maxFileSize {
		v.statusMsg = "File too large (>1MB)"
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		v.statusMsg = "Error: " + err.Error()
		return
	}

	// Detect encoding
	encIdx := detectEncoding(data)

	// Binary check: skip for UTF-16
	if encIdx != 5 && encIdx != 6 {
		checkLen := len(data)
		if checkLen > 512 {
			checkLen = 512
		}
		for _, b := range data[:checkLen] {
			if b == 0 {
				v.statusMsg = "Binary file (cannot open)"
				return
			}
		}
	}

	v.rawBytes = data
	v.fileEncoding = encIdx
	v.currentFile = path

	// Check for temp file (unsaved changes from previous session)
	if tempData := loadTempFile(v.tempDir, path); tempData != nil {
		v.editor.load(string(tempData))
		v.editor.dirty = true
		v.dirtyFiles[path] = true
		v.statusMsg = filepath.Base(path) + " (restored from temp)"
	} else {
		decoded := decodeBytes(data, encIdx)
		v.editor.load(decoded)
		v.statusMsg = filepath.Base(path)
	}

	v.mode = modeEditor
}

func (v *WriterView) openFileWithEncoding(encIdx int) {
	if v.currentFile == "" || v.rawBytes == nil {
		return
	}
	v.fileEncoding = encIdx
	decoded := decodeBytes(v.rawBytes, encIdx)
	v.editor.load(decoded)
	v.statusMsg = filepath.Base(v.currentFile) + " [" + supportedEncodings[encIdx].name + "]"
	// Remove temp file since we're reloading from disk
	removeTempFile(v.tempDir, v.currentFile)
	delete(v.dirtyFiles, v.currentFile)
}

func (v *WriterView) saveFile() {
	if v.currentFile == "" {
		v.statusMsg = "No file to save"
		return
	}
	content := v.editor.content()
	data := encodeString(content, v.fileEncoding)
	if err := os.WriteFile(v.currentFile, data, 0644); err != nil {
		v.statusMsg = "Save error: " + err.Error()
		return
	}
	v.rawBytes = data
	v.editor.dirty = false
	removeTempFile(v.tempDir, v.currentFile)
	delete(v.dirtyFiles, v.currentFile)
	v.statusMsg = "Saved: " + filepath.Base(v.currentFile)
}

func (v *WriterView) discardChanges() {
	if v.currentFile == "" {
		v.statusMsg = "No file open"
		return
	}
	// Remove temp file
	removeTempFile(v.tempDir, v.currentFile)
	delete(v.dirtyFiles, v.currentFile)

	// Reload from disk
	data, err := os.ReadFile(v.currentFile)
	if err != nil {
		v.statusMsg = "Error: " + err.Error()
		return
	}
	v.rawBytes = data
	decoded := decodeBytes(data, v.fileEncoding)
	v.editor.load(decoded)
	v.statusMsg = "Discarded: " + filepath.Base(v.currentFile)
}

func (v *WriterView) saveTempIfDirty() {
	if v.currentFile == "" || !v.editor.dirty {
		return
	}
	content := v.editor.content()
	if err := saveTempFile(v.tempDir, v.currentFile, []byte(content)); err == nil {
		v.dirtyFiles[v.currentFile] = true
	}
}

func (v *WriterView) createNewFile(name string) {
	if name == "" {
		v.statusMsg = "Empty file name"
		return
	}

	// Determine parent directory
	var parentDir string
	if v.treeCursor < len(v.visibleNodes) {
		n := v.visibleNodes[v.treeCursor]
		if n.isDir {
			parentDir = n.path
		} else {
			parentDir = filepath.Dir(n.path)
		}
	} else {
		parentDir = v.rootDir
	}

	newPath := filepath.Join(parentDir, name)

	// Check if file already exists
	if _, err := os.Stat(newPath); err == nil {
		v.statusMsg = "File already exists: " + name
		return
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(newPath), 0755); err != nil {
		v.statusMsg = "Error: " + err.Error()
		return
	}

	// Create the file
	if err := os.WriteFile(newPath, []byte{}, 0644); err != nil {
		v.statusMsg = "Error: " + err.Error()
		return
	}

	// Reload tree
	v.loadRootDir()

	// Open the new file
	v.currentFile = newPath
	v.rawBytes = []byte{}
	v.fileEncoding = 0 // UTF-8
	v.editor.load("")
	v.mode = modeEditor
	v.statusMsg = "Created: " + name
}

// Layout helpers

func (v *WriterView) sideWidth() int {
	sw := v.width / 4
	if sw < 20 {
		sw = 20
	}
	if sw > 50 {
		sw = 50
	}
	return sw
}

func (v *WriterView) mainWidth() int {
	return v.width - v.sideWidth()
}

func (v *WriterView) panelInnerHeight() int {
	h := v.height - 4
	if h < 1 {
		h = 1
	}
	return h
}

// truncate truncates a string to fit within maxWidth display columns.
func truncate(s string, maxWidth int) string {
	w := 0
	for i, r := range s {
		rw := runewidth.RuneWidth(r)
		if w+rw > maxWidth {
			return s[:i]
		}
		w += rw
	}
	return s
}

// padRight pads a string to exactly width display columns.
func padRight(s string, width int) string {
	sw := runewidth.StringWidth(s)
	if sw >= width {
		return truncate(s, width)
	}
	return s + strings.Repeat(" ", width-sw)
}

const tabWidth = 4

// expandTabs replaces each \t with spaces up to the next tab stop.
func expandTabs(s string, tw int) string {
	if !strings.Contains(s, "\t") {
		return s
	}
	var b strings.Builder
	col := 0
	for _, r := range s {
		if r == '\t' {
			n := tw - (col % tw)
			b.WriteString(strings.Repeat(" ", n))
			col += n
		} else {
			b.WriteRune(r)
			col += runewidth.RuneWidth(r)
		}
	}
	return b.String()
}

// buildCursorLine splits a line into (before, cursorChar, after) with tabs
// expanded to spaces. curCol is the rune-based cursor column.
func buildCursorLine(runes []rune, curCol int, tw int) (before, cursorCh, after string) {
	var bBefore, bCursor, bAfter strings.Builder
	col := 0
	for i, r := range runes {
		if r == '\t' {
			n := tw - (col % tw)
			if i < curCol {
				bBefore.WriteString(strings.Repeat(" ", n))
			} else if i == curCol {
				// Cursor on a tab: show 1 space reversed, rest goes to after
				bCursor.WriteByte(' ')
				if n > 1 {
					bAfter.WriteString(strings.Repeat(" ", n-1))
				}
			} else {
				bAfter.WriteString(strings.Repeat(" ", n))
			}
			col += n
		} else {
			if i < curCol {
				bBefore.WriteRune(r)
			} else if i == curCol {
				bCursor.WriteRune(r)
			} else {
				bAfter.WriteRune(r)
			}
			col += runewidth.RuneWidth(r)
		}
	}
	before = bBefore.String()
	cursorCh = bCursor.String()
	if cursorCh == "" {
		cursorCh = " " // cursor at end of line
	}
	after = bAfter.String()
	return
}

// ---------------- Render ----------------

func (v *WriterView) Render() []g.ComponentFunc {
	sw := v.sideWidth()
	mw := v.mainWidth()
	ih := v.panelInnerHeight()

	sidePanel := v.renderSidePanel(sw, ih)
	mainPanel := v.renderMainPanel(mw, ih)

	content := tui.HBox(sidePanel, mainPanel)
	statusBar := v.renderStatusBar()
	keyHelp := v.renderKeyHelp()

	return []g.ComponentFunc{content, statusBar, keyHelp}
}

func (v *WriterView) renderSidePanel(width, innerHeight int) g.ComponentFunc {
	// In encoding mode, render the encoding selection panel
	if v.mode == modeEncoding {
		return v.renderEncodingPanel(width, innerHeight)
	}

	nodes := v.visibleNodes
	if v.mode == modeSearch && v.searchQuery != "" {
		nodes = v.searchResults
	}

	contentWidth := width - 2
	if contentWidth < 1 {
		contentWidth = 1
	}

	lines := make([]g.ComponentFunc, 0, innerHeight)

	// Header
	var headerText string
	if v.mode == modeSearch {
		headerText = " / " + v.searchQuery + "▌"
	} else if v.mode == modeNewFile {
		headerText = " New: " + v.newFileName + "▌"
	} else {
		headerText = " " + truncate(filepath.Base(v.rootDir), contentWidth-1)
	}
	lines = append(lines, tui.Text(
		ts.Bold(true), ts.FgColor("81"),
		g.Literal(padRight(headerText, contentWidth)),
	))
	lines = append(lines, tui.Text(
		ts.Faint(true),
		g.Literal(strings.Repeat("─", contentWidth)),
	))

	listHeight := innerHeight - 2
	if listHeight < 0 {
		listHeight = 0
	}

	// Adjust scroll so cursor stays visible
	if v.treeCursor < v.treeScrollY {
		v.treeScrollY = v.treeCursor
	}
	if v.treeCursor >= v.treeScrollY+listHeight {
		v.treeScrollY = v.treeCursor - listHeight + 1
	}

	for i := 0; i < listHeight; i++ {
		idx := v.treeScrollY + i
		if idx >= len(nodes) {
			lines = append(lines, tui.Text(g.Literal(strings.Repeat(" ", contentWidth))))
			continue
		}
		n := nodes[idx]
		isCursor := (idx == v.treeCursor) && (v.mode == modeTree || v.mode == modeSearch || v.mode == modeNewFile)

		// Indentation
		indent := strings.Repeat("  ", n.depth)

		// Prefix: cursor marker or space
		prefix := " "
		if isCursor {
			prefix = ">"
		}

		// Directory indicator
		var indicator string
		if n.isDir {
			if n.expanded {
				indicator = "▾ "
			} else {
				indicator = "▸ "
			}
		} else {
			indicator = "  "
		}

		// Dirty marker for files
		dirtyMarker := ""
		if !n.isDir && v.dirtyFiles[n.path] {
			dirtyMarker = "[+]"
		}

		label := truncate(prefix+indent+indicator+n.name+dirtyMarker, contentWidth)
		label = padRight(label, contentWidth)

		var style []g.ComponentFunc
		if isCursor {
			style = append(style, ts.Reverse(true), ts.Bold(true))
		} else if n.isDir {
			style = append(style, ts.FgColor("220"))
		} else if v.dirtyFiles[n.path] {
			style = append(style, ts.FgColor("214"))
		}
		style = append(style, g.Literal(label))
		lines = append(lines, tui.Text(style...))
	}

	isActive := v.mode == modeTree || v.mode == modeSearch || v.mode == modeNewFile
	var borderColor string
	if isActive {
		borderColor = "81"
	} else {
		borderColor = "240"
	}

	return tui.Box(
		ts.Border("rounded"),
		ts.BorderColor(borderColor),
		ts.Width(contentWidth),
		tui.VBox(lines...),
	)
}

func (v *WriterView) renderEncodingPanel(width, innerHeight int) g.ComponentFunc {
	contentWidth := width - 2
	if contentWidth < 1 {
		contentWidth = 1
	}

	lines := make([]g.ComponentFunc, 0, innerHeight)

	// Header
	headerText := " Select Encoding"
	lines = append(lines, tui.Text(
		ts.Bold(true), ts.FgColor("81"),
		g.Literal(padRight(headerText, contentWidth)),
	))
	lines = append(lines, tui.Text(
		ts.Faint(true),
		g.Literal(strings.Repeat("─", contentWidth)),
	))

	listHeight := innerHeight - 2
	if listHeight < 0 {
		listHeight = 0
	}

	for i := 0; i < listHeight; i++ {
		if i >= len(supportedEncodings) {
			lines = append(lines, tui.Text(g.Literal(strings.Repeat(" ", contentWidth))))
			continue
		}

		enc := supportedEncodings[i]
		isCursor := i == v.encodingCursor
		isCurrent := i == v.fileEncoding

		prefix := " "
		if isCursor {
			prefix = ">"
		}

		label := prefix + " " + enc.name
		if isCurrent {
			label += " (current)"
		}
		label = padRight(truncate(label, contentWidth), contentWidth)

		var style []g.ComponentFunc
		if isCursor {
			style = append(style, ts.Reverse(true), ts.Bold(true))
		} else if isCurrent {
			style = append(style, ts.FgColor("118"))
		}
		style = append(style, g.Literal(label))
		lines = append(lines, tui.Text(style...))
	}

	return tui.Box(
		ts.Border("rounded"),
		ts.BorderColor("81"),
		ts.Width(contentWidth),
		tui.VBox(lines...),
	)
}

func (v *WriterView) renderMainPanel(width, innerHeight int) g.ComponentFunc {
	contentWidth := width - 2
	if contentWidth < 1 {
		contentWidth = 1
	}

	lines := make([]g.ComponentFunc, 0, innerHeight)

	// Header
	var headerText string
	if v.currentFile != "" {
		headerText = " " + filepath.Base(v.currentFile)
		if v.editor.dirty {
			headerText += " [+]"
		}
	} else {
		headerText = " No file open"
	}
	lines = append(lines, tui.Text(
		ts.Bold(true), ts.FgColor("212"),
		g.Literal(padRight(truncate(headerText, contentWidth), contentWidth)),
	))
	lines = append(lines, tui.Text(
		ts.Faint(true),
		g.Literal(strings.Repeat("─", contentWidth)),
	))

	editorHeight := innerHeight - 2
	if editorHeight < 0 {
		editorHeight = 0
	}

	if v.currentFile == "" {
		for i := 0; i < editorHeight; i++ {
			if i == editorHeight/2 {
				msg := "Select a file from the tree"
				lines = append(lines, tui.Text(
					ts.Faint(true), ts.Italic(true),
					g.Literal(padRight(msg, contentWidth)),
				))
			} else {
				lines = append(lines, tui.Text(g.Literal(strings.Repeat(" ", contentWidth))))
			}
		}
	} else {
		lines = append(lines, v.renderEditor(contentWidth, editorHeight)...)
	}

	isActive := v.mode == modeEditor
	var borderColor string
	if isActive {
		borderColor = "212"
	} else {
		borderColor = "240"
	}

	return tui.Box(
		ts.Border("rounded"),
		ts.BorderColor(borderColor),
		ts.Width(contentWidth),
		tui.VBox(lines...),
	)
}

func (v *WriterView) renderEditor(width, height int) []g.ComponentFunc {
	buf := v.editor
	totalLines := len(buf.lines)

	gutterWidth := len(strconv.Itoa(totalLines)) + 1
	if gutterWidth < 3 {
		gutterWidth = 3
	}
	textWidth := width - gutterWidth - 1 // 1 for "│" separator
	if textWidth < 1 {
		textWidth = 1
	}

	// Adjust scroll
	if buf.curRow < buf.scrollY {
		buf.scrollY = buf.curRow
	}
	if buf.curRow >= buf.scrollY+height {
		buf.scrollY = buf.curRow - height + 1
	}

	lines := make([]g.ComponentFunc, 0, height)
	for i := 0; i < height; i++ {
		lineIdx := buf.scrollY + i
		if lineIdx >= totalLines {
			gutter := strings.Repeat(" ", gutterWidth) + "│"
			lines = append(lines, tui.Text(
				ts.Faint(true),
				g.Literal(padRight(gutter, width)),
			))
			continue
		}

		lineNum := strconv.Itoa(lineIdx + 1)
		gutter := strings.Repeat(" ", gutterWidth-len(lineNum)) + lineNum + "│"

		lineRunes := []rune(buf.lines[lineIdx])
		isCursorLine := v.mode == modeEditor && lineIdx == buf.curRow

		if isCursorLine {
			col := buf.curCol
			if col > len(lineRunes) {
				col = len(lineRunes)
			}

			// Split line at cursor with tab expansion
			before, cursorChar, after := buildCursorLine(lineRunes, col, tabWidth)

			beforeText := truncate(gutter+before, gutterWidth+1+textWidth)
			usedWidth := runewidth.StringWidth(beforeText)
			remainWidth := width - usedWidth

			if remainWidth <= 0 {
				lines = append(lines, tui.Text(
					ts.BgColor("236"), ts.Bold(true),
					g.Literal(padRight(beforeText, width)),
				))
			} else {
				cursorChar = truncate(cursorChar, remainWidth)
				cursorWidth := runewidth.StringWidth(cursorChar)
				afterWidth := remainWidth - cursorWidth
				if afterWidth < 0 {
					afterWidth = 0
				}
				afterText := padRight(truncate(after, afterWidth), afterWidth)

				lines = append(lines, tui.Text(
					tui.Text(ts.BgColor("236"), g.Literal(beforeText)),
					tui.Text(ts.Reverse(true), ts.Bold(true), g.Literal(cursorChar)),
					tui.Text(ts.BgColor("236"), g.Literal(afterText)),
				))
			}
		} else {
			// Expand tabs for non-cursor lines too
			lineStr := expandTabs(string(lineRunes), tabWidth)
			text := truncate(lineStr, textWidth)
			fullLine := gutter + padRight(text, textWidth)
			lines = append(lines, tui.Text(g.Literal(fullLine)))
		}
	}

	return lines
}

func (v *WriterView) renderStatusBar() g.ComponentFunc {
	var left string
	switch v.mode {
	case modeTree:
		left = " TREE"
	case modeSearch:
		left = " SEARCH"
	case modeEditor:
		left = " EDIT"
	case modeNewFile:
		left = " NEW FILE"
	case modeEncoding:
		left = " ENCODING"
	}
	if v.statusMsg != "" {
		left += " | " + v.statusMsg
	}

	var right string
	if v.currentFile != "" {
		encName := supportedEncodings[v.fileEncoding].name
		if v.mode == modeEditor {
			right = fmt.Sprintf("%s | Ln %d, Col %d ", encName, v.editor.curRow+1, v.editor.curCol+1)
		} else {
			right = encName + " "
		}
	}

	return tc.StatusBar(left, "", right, v.width)
}

func (v *WriterView) renderKeyHelp() g.ComponentFunc {
	switch v.mode {
	case modeTree:
		return tc.KeyHelp([][2]string{
			{"j/k", "move "},
			{"Enter", "open "},
			{"h", "collapse "},
			{"/", "search "},
			{"+", "new file "},
			{"w", "save "},
			{"u", "discard "},
			{"e", "encoding "},
			{"Tab", "editor "},
			{"q", "quit"},
		})
	case modeSearch:
		return tc.KeyHelp([][2]string{
			{"type", "filter "},
			{"Enter", "open "},
			{"Esc", "cancel"},
		})
	case modeEditor:
		return tc.KeyHelp([][2]string{
			{"type", "edit "},
			{"Tab", "tree "},
			{"arrows", "move"},
		})
	case modeNewFile:
		return tc.KeyHelp([][2]string{
			{"type", "filename "},
			{"Enter", "create "},
			{"Esc", "cancel"},
		})
	case modeEncoding:
		return tc.KeyHelp([][2]string{
			{"j/k", "move "},
			{"Enter", "select "},
			{"Esc", "cancel"},
		})
	}
	return tui.Text(g.Literal(""))
}

// ---------------- Event Handling ----------------

func (v *WriterView) HandleEvent(event app.Event) error {
	if event.Type == app.EventResize {
		v.width = event.Width
		v.height = event.Height
		return nil
	}

	switch v.mode {
	case modeTree:
		return v.handleTreeMode(event)
	case modeSearch:
		return v.handleSearchMode(event)
	case modeEditor:
		return v.handleEditorMode(event)
	case modeNewFile:
		return v.handleNewFileMode(event)
	case modeEncoding:
		return v.handleEncodingMode(event)
	}
	return nil
}

func (v *WriterView) handleTreeMode(event app.Event) error {
	nodes := v.visibleNodes

	switch event.Key {
	case "q":
		cleanupTempDir(v.tempDir)
		return fmt.Errorf("quit")
	case "j", "down":
		if v.treeCursor < len(nodes)-1 {
			v.treeCursor++
		}
	case "k", "up":
		if v.treeCursor > 0 {
			v.treeCursor--
		}
	case "enter", "l":
		if v.treeCursor < len(nodes) {
			n := nodes[v.treeCursor]
			if n.isDir {
				// Toggle expand/collapse
				if !n.expanded {
					n.loadChildren()
					n.expanded = true
				} else {
					n.expanded = false
				}
				v.rebuildVisible()
			} else {
				v.openFile(n.path)
			}
		}
	case "h":
		if v.treeCursor < len(nodes) {
			n := nodes[v.treeCursor]
			if n.isDir && n.expanded {
				// Collapse this directory
				n.expanded = false
				v.rebuildVisible()
			} else {
				// Jump to parent directory node
				pi := findParentIndex(v.visibleNodes, v.treeCursor)
				if pi >= 0 {
					v.treeCursor = pi
				} else {
					// Go to parent rootDir
					parent := filepath.Dir(v.rootDir)
					if parent != v.rootDir {
						v.rootDir = parent
						v.loadRootDir()
					}
				}
			}
		}
	case "backspace":
		parent := filepath.Dir(v.rootDir)
		if parent != v.rootDir {
			v.rootDir = parent
			v.loadRootDir()
		}
	case "/":
		v.mode = modeSearch
		v.searchQuery = ""
		v.searchResults = append([]*treeNode(nil), v.visibleNodes...)
	case "tab":
		if v.currentFile != "" {
			v.mode = modeEditor
		}
	case "+":
		v.mode = modeNewFile
		v.newFileName = ""
	case "w":
		v.saveFile()
	case "u":
		v.discardChanges()
	case "e":
		if v.currentFile != "" {
			v.mode = modeEncoding
			v.encodingCursor = v.fileEncoding
		}
	}
	return nil
}

func (v *WriterView) handleSearchMode(event app.Event) error {
	switch event.Key {
	case "esc":
		v.mode = modeTree
		v.searchQuery = ""
	case "enter":
		results := v.searchResults
		if v.treeCursor < len(results) {
			n := results[v.treeCursor]
			if n.isDir {
				if !n.expanded {
					n.loadChildren()
					n.expanded = true
				} else {
					n.expanded = false
				}
				v.rebuildVisible()
				v.mode = modeTree
			} else {
				v.openFile(n.path)
			}
		}
		v.searchQuery = ""
	case "backspace":
		if len(v.searchQuery) > 0 {
			runes := []rune(v.searchQuery)
			v.searchQuery = string(runes[:len(runes)-1])
			v.filterSearch()
		}
	case "up":
		if v.treeCursor > 0 {
			v.treeCursor--
		}
	case "down":
		if v.treeCursor < len(v.searchResults)-1 {
			v.treeCursor++
		}
	default:
		if len(event.Runes) > 0 {
			v.searchQuery += string(event.Runes)
			v.filterSearch()
		}
	}
	return nil
}

func (v *WriterView) filterSearch() {
	query := strings.ToLower(v.searchQuery)
	v.searchResults = v.searchResults[:0]
	for _, n := range v.visibleNodes {
		if strings.Contains(strings.ToLower(n.name), query) {
			v.searchResults = append(v.searchResults, n)
		}
	}
	v.treeCursor = 0
	v.treeScrollY = 0
}

func (v *WriterView) handleEditorMode(event app.Event) error {
	switch event.Key {
	case "tab":
		// Save to temp if dirty before switching to tree
		v.saveTempIfDirty()
		v.mode = modeTree
	case "up":
		if v.editor.curRow > 0 {
			v.editor.curRow--
			v.clampCol()
		}
	case "down":
		if v.editor.curRow < len(v.editor.lines)-1 {
			v.editor.curRow++
			v.clampCol()
		}
	case "left":
		if v.editor.curCol > 0 {
			v.editor.curCol--
		}
	case "right":
		lineLen := len([]rune(v.editor.lines[v.editor.curRow]))
		if v.editor.curCol < lineLen {
			v.editor.curCol++
		}
	case "enter":
		v.editor.newLine()
	case "backspace":
		v.editor.backspace()
	case "home":
		v.editor.curCol = 0
	case "end":
		v.editor.curCol = len([]rune(v.editor.lines[v.editor.curRow]))
	default:
		if len(event.Runes) > 0 {
			for _, r := range event.Runes {
				v.editor.insertRune(r)
			}
		}
	}
	return nil
}

func (v *WriterView) handleNewFileMode(event app.Event) error {
	switch event.Key {
	case "esc":
		v.mode = modeTree
		v.newFileName = ""
	case "enter":
		v.createNewFile(v.newFileName)
		v.newFileName = ""
	case "backspace":
		if len(v.newFileName) > 0 {
			runes := []rune(v.newFileName)
			v.newFileName = string(runes[:len(runes)-1])
		}
	default:
		if len(event.Runes) > 0 {
			v.newFileName += string(event.Runes)
		}
	}
	return nil
}

func (v *WriterView) handleEncodingMode(event app.Event) error {
	switch event.Key {
	case "esc":
		v.mode = modeTree
	case "j", "down":
		if v.encodingCursor < len(supportedEncodings)-1 {
			v.encodingCursor++
		}
	case "k", "up":
		if v.encodingCursor > 0 {
			v.encodingCursor--
		}
	case "enter":
		v.openFileWithEncoding(v.encodingCursor)
		v.mode = modeEditor
	}
	return nil
}

func (v *WriterView) clampCol() {
	lineLen := len([]rune(v.editor.lines[v.editor.curRow]))
	if v.editor.curCol > lineLen {
		v.editor.curCol = lineLen
	}
}

func main() {
	if err := app.Run(func() app.View { return &WriterView{} }); err != nil {
		if err.Error() == "quit" {
			return
		}
		log.Fatal(err)
	}
	// Cleanup temp directory on exit
	if globalTempDir != "" {
		cleanupTempDir(globalTempDir)
	}
}
