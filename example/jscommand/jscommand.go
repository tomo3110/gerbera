package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	ge "github.com/tomo3110/gerbera/expr"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
	gs "github.com/tomo3110/gerbera/styles"
)

// logEntry implements ConvertToMap so it can be used with expr.Each.
type logEntry struct {
	Index   int
	Message string
}

func (e logEntry) ToMap() g.Map {
	return g.Map{"index": e.Index, "message": e.Message}
}

// JSCommandView demonstrates all available JS Commands.
type JSCommandView struct {
	gl.CommandQueue
	Log []logEntry
}

func (v *JSCommandView) Mount(_ gl.Params) error {
	return nil
}

func (v *JSCommandView) appendLog(msg string) {
	v.Log = append(v.Log, logEntry{Index: len(v.Log) + 1, Message: msg})
}

func (v *JSCommandView) HandleEvent(event string, _ gl.Payload) error {
	switch event {
	// Focus / Blur
	case "do_focus":
		v.Focus("#target-input")
		v.appendLog("Focus → #target-input")
	case "do_blur":
		v.Blur("#target-input")
		v.appendLog("Blur → #target-input")

	// Scroll
	case "scroll_top":
		v.ScrollTo("#scroll-box", "0", "0")
		v.appendLog("ScrollTo top → #scroll-box")
	case "scroll_bottom":
		v.ScrollTo("#scroll-box", "9999", "0")
		v.appendLog("ScrollTo bottom → #scroll-box")

	// Class manipulation
	case "add_class":
		v.AddClass("#style-box", "highlighted")
		v.appendLog("AddClass 'highlighted' → #style-box")
	case "remove_class":
		v.RemoveClass("#style-box", "highlighted")
		v.appendLog("RemoveClass 'highlighted' → #style-box")
	case "toggle_class":
		v.ToggleClass("#style-box", "highlighted")
		v.appendLog("ToggleClass 'highlighted' → #style-box")

	// Show / Hide / Toggle
	case "do_show":
		v.Show("#toggle-box")
		v.appendLog("Show → #toggle-box")
	case "do_hide":
		v.Hide("#toggle-box")
		v.appendLog("Hide → #toggle-box")
	case "do_toggle":
		v.Toggle("#toggle-box")
		v.appendLog("Toggle → #toggle-box")

	case "clear_log":
		v.Log = nil
	}
	return nil
}

func (v *JSCommandView) Render() []g.ComponentFunc {
	logItems := make([]g.ConvertToMap, len(v.Log))
	for i, e := range v.Log {
		logItems[i] = e
	}

	return []g.ComponentFunc{
		gd.Head(
			gd.Title("JS Commands — Gerbera Demo"),
			gs.CSS(`
				body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #f5f5f5; margin: 0; padding: 24px; color: #333; }
				h1 { color: #2c3e50; }
				h3 { color: #34495e; margin-top: 28px; margin-bottom: 8px; }
				.section { background: #fff; border-radius: 8px; padding: 16px 20px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
				.buttons { display: flex; gap: 8px; margin-top: 8px; flex-wrap: wrap; }
				button { padding: 8px 18px; border: none; border-radius: 4px; cursor: pointer; font-size: 0.9rem; font-weight: 500; background: #3498db; color: #fff; }
				button:hover { opacity: 0.85; }
				.btn-danger { background: #e74c3c; }
				input[type="text"] { padding: 8px 12px; border: 1px solid #ccc; border-radius: 4px; font-size: 1rem; width: 300px; }
				#scroll-box { height: 120px; overflow-y: scroll; border: 1px solid #ccc; border-radius: 4px; padding: 8px; background: #fafafa; }
				#style-box { padding: 20px; border-radius: 4px; background: #ecf0f1; text-align: center; font-weight: 600; transition: all 0.3s; }
				#style-box.highlighted { background: #f1c40f; color: #2c3e50; }
				#toggle-box { padding: 20px; border-radius: 4px; background: #2ecc71; color: #fff; text-align: center; font-weight: 600; }
				.log { max-height: 200px; overflow-y: auto; font-family: "Courier New", monospace; font-size: 0.85rem; background: #2c3e50; color: #2ecc71; border-radius: 4px; padding: 12px; }
				.log-entry { padding: 2px 0; }
			`),
		),
		gd.Body(
			gd.H1(gp.Value("JS Commands Demo")),

			// Focus / Blur
			gd.Div(
				gp.Class("section"),
				gd.H3(gp.Value("Focus / Blur")),
				gd.Input(gp.Type("text"), gp.ID("target-input"), gp.Placeholder("Click Focus to focus this input")),
				gd.Div(
					gp.Class("buttons"),
					gd.Button(gl.Click("do_focus"), gp.Value("Focus")),
					gd.Button(gl.Click("do_blur"), gp.Value("Blur")),
				),
			),

			// ScrollTo
			gd.Div(
				gp.Class("section"),
				gd.H3(gp.Value("ScrollTo")),
				gd.Div(
					gp.ID("scroll-box"),
					scrollContent(),
				),
				gd.Div(
					gp.Class("buttons"),
					gd.Button(gl.Click("scroll_top"), gp.Value("Scroll to Top")),
					gd.Button(gl.Click("scroll_bottom"), gp.Value("Scroll to Bottom")),
				),
			),

			// AddClass / RemoveClass / ToggleClass
			gd.Div(
				gp.Class("section"),
				gd.H3(gp.Value("AddClass / RemoveClass / ToggleClass")),
				gd.Div(gp.ID("style-box"), gp.Value("Target Box")),
				gd.Div(
					gp.Class("buttons"),
					gd.Button(gl.Click("add_class"), gp.Value("AddClass")),
					gd.Button(gl.Click("remove_class"), gp.Value("RemoveClass")),
					gd.Button(gl.Click("toggle_class"), gp.Value("ToggleClass")),
				),
			),

			// Show / Hide / Toggle
			gd.Div(
				gp.Class("section"),
				gd.H3(gp.Value("Show / Hide / Toggle")),
				gd.Div(gp.ID("toggle-box"), gp.Value("Toggle Me")),
				gd.Div(
					gp.Class("buttons"),
					gd.Button(gl.Click("do_show"), gp.Value("Show")),
					gd.Button(gl.Click("do_hide"), gp.Value("Hide")),
					gd.Button(gl.Click("do_toggle"), gp.Value("Toggle")),
				),
			),

			// Command Log
			gd.Div(
				gp.Class("section"),
				gd.H3(gp.Value("Command Log")),
				ge.If(len(v.Log) > 0,
					gd.Div(
						gp.Class("log"),
						ge.Each(logItems, func(item g.ConvertToMap) g.ComponentFunc {
							m := item.ToMap()
							idx := m.Get("index").(int)
							msg := m.Get("message").(string)
							return gd.Div(gp.Class("log-entry"), gp.Value(fmt.Sprintf("#%d  %s", idx, msg)))
						}),
					),
					gd.P(gp.Value("No commands executed yet.")),
				),
				gd.Div(
					gp.Class("buttons"),
					gd.Button(gp.Class("btn-danger"), gl.Click("clear_log"), gp.Value("Clear Log")),
				),
			),
		),
	}
}

// scrollContent generates enough lines to make the scroll box scrollable.
func scrollContent() g.ComponentFunc {
	var children []g.ComponentFunc
	for i := 1; i <= 30; i++ {
		children = append(children, gd.P(gp.Value(fmt.Sprintf("Line %d — scroll content", i))))
	}
	return gd.Div(children...)
}

func main() {
	addr := flag.String("addr", ":8875", "listen address")
	debug := flag.Bool("debug", false, "enable debug panel")
	flag.Parse()

	var opts []gl.Option
	if *debug {
		opts = append(opts, gl.WithDebug())
	}

	http.Handle("/", gl.Handler(func(_ *http.Request) gl.View { return &JSCommandView{} }, opts...))
	log.Printf("jscommand running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
