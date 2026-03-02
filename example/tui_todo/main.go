package main

import (
	"fmt"
	"log"

	g "github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/expr"
	"github.com/tomo3110/gerbera/tui"
	"github.com/tomo3110/gerbera/tui/app"
	tc "github.com/tomo3110/gerbera/tui/components"
	ts "github.com/tomo3110/gerbera/tui/style"
)

type mode int

const (
	modeNormal mode = iota
	modeInput
)

type task struct {
	title string
	done  bool
}

type TodoView struct {
	tasks  []task
	cursor int
	mode   mode
	input  string
}

func (v *TodoView) Mount(params app.Params) error {
	v.tasks = []task{
		{title: "gerbera の TUI サポートを実装する", done: true},
		{title: "TODO サンプルアプリを作成する", done: false},
		{title: "ドキュメントを書く", done: false},
		{title: "テストを追加する", done: false},
	}
	return nil
}

func (v *TodoView) Render() []g.ComponentFunc {
	doneCount := 0
	for _, t := range v.tasks {
		if t.done {
			doneCount++
		}
	}
	total := len(v.tasks)

	// Title
	title := tui.Text(
		ts.Bold(true), ts.FgColor("212"),
		g.Literal("📋 TODO リスト"),
	)

	// Progress
	var progress g.ComponentFunc
	if total > 0 {
		progress = tc.ProgressBar(doneCount, total, 30)
	} else {
		progress = tui.Text(ts.Faint(true), g.Literal("タスクがありません"))
	}

	// Task list
	items := make([]g.ComponentFunc, len(v.tasks))
	for i, t := range v.tasks {
		items[i] = v.renderTask(i, t)
	}

	var listContent g.ComponentFunc
	if len(items) > 0 {
		listContent = tui.VBox(items...)
	} else {
		listContent = tui.Text(
			ts.Faint(true), ts.Italic(true),
			g.Literal("  (空) a キーでタスクを追加"),
		)
	}

	// Input area (shown in input mode)
	inputArea := expr.If(v.mode == modeInput,
		tui.VBox(
			tui.Divider(ts.Width(38)),
			tui.HBox(
				tui.Text(ts.Bold(true), ts.FgColor("220"), g.Literal("新規: ")),
				tui.Text(g.Literal(v.input+"▌")),
			),
		),
	)

	// Status bar
	status := tc.StatusBar(
		fmt.Sprintf(" %d/%d 完了", doneCount, total),
		"",
		fmt.Sprintf("タスク数: %d ", total),
		40,
	)

	// Key help
	var help g.ComponentFunc
	if v.mode == modeInput {
		help = tc.KeyHelp([][2]string{
			{"Enter", "追加  "},
			{"Esc", "キャンセル"},
		})
	} else {
		help = tc.KeyHelp([][2]string{
			{"j/k", "移動  "},
			{"Space", "切替  "},
			{"a", "追加  "},
			{"d", "削除  "},
			{"q", "終了"},
		})
	}

	return []g.ComponentFunc{
		tui.Box(
			ts.Border("rounded"), ts.BorderColor("63"),
			ts.Padding(1, 2, 0, 2), ts.Width(44),
			title,
			tui.Spacer(),
			progress,
			tui.Spacer(),
			listContent,
			inputArea,
		),
		status,
		help,
	}
}

func (v *TodoView) renderTask(index int, t task) g.ComponentFunc {
	isCursor := index == v.cursor
	check := "[ ] "
	if t.done {
		check = "[x] "
	}

	cursor := "  "
	if isCursor {
		cursor = "▸ "
	}

	var textStyle []g.ComponentFunc
	if t.done {
		textStyle = append(textStyle, ts.Faint(true), ts.Strikethrough(true))
	}
	if isCursor {
		textStyle = append(textStyle, ts.Bold(true), ts.FgColor("81"))
	}
	textStyle = append(textStyle, g.Literal(cursor+check+t.title))

	return tui.Text(textStyle...)
}

func (v *TodoView) HandleEvent(event app.Event) error {
	if v.mode == modeInput {
		return v.handleInputMode(event)
	}
	return v.handleNormalMode(event)
}

func (v *TodoView) handleNormalMode(event app.Event) error {
	switch event.Key {
	case "q":
		return fmt.Errorf("quit")
	case "j", "down":
		if v.cursor < len(v.tasks)-1 {
			v.cursor++
		}
	case "k", "up":
		if v.cursor > 0 {
			v.cursor--
		}
	case " ":
		if len(v.tasks) > 0 {
			v.tasks[v.cursor].done = !v.tasks[v.cursor].done
		}
	case "d":
		if len(v.tasks) > 0 {
			v.tasks = append(v.tasks[:v.cursor], v.tasks[v.cursor+1:]...)
			if v.cursor >= len(v.tasks) && v.cursor > 0 {
				v.cursor--
			}
		}
	case "a":
		v.mode = modeInput
		v.input = ""
	}
	return nil
}

func (v *TodoView) handleInputMode(event app.Event) error {
	switch event.Key {
	case "enter":
		if v.input != "" {
			v.tasks = append(v.tasks, task{title: v.input})
			v.cursor = len(v.tasks) - 1
		}
		v.mode = modeNormal
		v.input = ""
	case "esc":
		v.mode = modeNormal
		v.input = ""
	case "backspace":
		if len(v.input) > 0 {
			runes := []rune(v.input)
			v.input = string(runes[:len(runes)-1])
		}
	default:
		if len(event.Runes) > 0 {
			v.input += string(event.Runes)
		}
	}
	return nil
}

func main() {
	if err := app.Run(func() app.View { return &TodoView{} }); err != nil {
		if err.Error() == "quit" {
			return
		}
		log.Fatal(err)
	}
}
