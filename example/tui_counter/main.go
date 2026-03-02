package main

import (
	"fmt"
	"log"

	g "github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/tui"
	"github.com/tomo3110/gerbera/tui/app"
	tc "github.com/tomo3110/gerbera/tui/components"
	ts "github.com/tomo3110/gerbera/tui/style"
)

type CounterView struct {
	Count int
}

func (v *CounterView) Mount(params app.Params) error {
	v.Count = 0
	return nil
}

func (v *CounterView) Render() []g.ComponentFunc {
	return []g.ComponentFunc{
		tui.Box(
			ts.Border("rounded"), ts.Padding(1, 2, 1, 2), ts.Width(40), ts.Align("center"),
			tui.Text(
				ts.Bold(true), ts.FgColor("212"),
				g.Literal(fmt.Sprintf("カウント: %d", v.Count)),
			),
			tui.Spacer(),
			tui.HBox(
				tui.Text(g.Literal("[j] -  [k] +  [q] 終了")),
			),
			tui.Spacer(),
			tc.ProgressBar(v.Count, 20, 30),
		),
	}
}

func (v *CounterView) HandleEvent(event app.Event) error {
	switch event.Key {
	case "k":
		v.Count++
	case "j":
		v.Count--
	case "q":
		// Return error to signal quit (will be caught by Run)
		return fmt.Errorf("quit")
	}
	return nil
}

func main() {
	if err := app.Run(func() app.View { return &CounterView{} }); err != nil {
		if err.Error() == "quit" {
			return
		}
		log.Fatal(err)
	}
}
