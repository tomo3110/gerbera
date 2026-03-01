package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	ge "github.com/tomo3110/gerbera/expr"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
	gs "github.com/tomo3110/gerbera/styles"
)

// ClockView displays a real-time clock and a stopwatch.
// It implements TickerView so the server pushes updates every 500ms.
type ClockView struct {
	Now       time.Time
	Running   bool
	Elapsed   time.Duration
	StartedAt time.Time
}

func (v *ClockView) Mount(_ gl.Params) error {
	v.Now = time.Now()
	return nil
}

func (v *ClockView) TickInterval() time.Duration {
	return 100 * time.Millisecond
}

func (v *ClockView) HandleTick() error {
	v.Now = time.Now()
	if v.Running {
		v.Elapsed = v.Now.Sub(v.StartedAt)
	}
	return nil
}

func (v *ClockView) HandleEvent(event string, _ gl.Payload) error {
	switch event {
	case "start":
		if !v.Running {
			v.Running = true
			v.StartedAt = time.Now().Add(-v.Elapsed)
		}
	case "stop":
		if v.Running {
			v.Elapsed = time.Now().Sub(v.StartedAt)
			v.Running = false
		}
	case "reset":
		v.Running = false
		v.Elapsed = 0
	}
	return nil
}

func (v *ClockView) Render() []g.ComponentFunc {
	clock := v.Now.Format("15:04:05")
	mins := int(v.Elapsed.Minutes())
	secs := int(v.Elapsed.Seconds()) % 60
	tenths := int(v.Elapsed.Milliseconds()/100) % 10
	stopwatch := fmt.Sprintf("%02d:%02d.%d", mins, secs, tenths)

	return []g.ComponentFunc{
		gd.Head(
			gd.Title("Clock — Gerbera TickerView Demo"),
			gs.CSS(`
				body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #1a1a2e; color: #e0e0e0; display: flex; justify-content: center; padding-top: 60px; margin: 0; }
				.container { text-align: center; }
				h1 { color: #8be9fd; margin-bottom: 8px; }
				.clock { font-size: 5rem; font-family: "Courier New", monospace; color: #f8f8f2; margin: 16px 0; }
				.stopwatch { font-size: 3rem; font-family: "Courier New", monospace; color: #50fa7b; margin: 16px 0; }
				.label { font-size: 0.9rem; color: #6272a4; text-transform: uppercase; letter-spacing: 2px; }
				.buttons { margin-top: 24px; display: flex; gap: 12px; justify-content: center; }
				.buttons button { padding: 10px 28px; font-size: 1rem; border: none; border-radius: 6px; cursor: pointer; font-weight: 600; }
				.btn-start { background: #50fa7b; color: #1a1a2e; }
				.btn-stop  { background: #ff5555; color: #fff; }
				.btn-reset { background: #6272a4; color: #fff; }
			`),
		),
		gd.Body(
			gd.Div(
				gp.Class("container"),
				gd.H1(gp.Value("Gerbera Clock")),

				gd.Div(gp.Class("label"), gp.Value("Current Time")),
				gd.Div(gp.Class("clock"), gp.Value(clock)),

				gd.Div(gp.Class("label"), gp.Value("Stopwatch")),
				gd.Div(gp.Class("stopwatch"), gp.Value(stopwatch)),

				gd.Div(
					gp.Class("buttons"),
					ge.If(!v.Running,
						gd.Button(gp.Class("btn-start"), gl.Click("start"), gp.Value("Start")),
						gd.Button(gp.Class("btn-stop"), gl.Click("stop"), gp.Value("Stop")),
					),
					gd.Button(gp.Class("btn-reset"), gl.Click("reset"), gp.Value("Reset")),
				),
			),
		),
	}
}

func main() {
	addr := flag.String("addr", ":8870", "listen address")
	debug := flag.Bool("debug", false, "enable debug panel")
	flag.Parse()

	var opts []gl.Option
	if *debug {
		opts = append(opts, gl.WithDebug())
	}

	http.Handle("/", gl.Handler(func() gl.View { return &ClockView{} }, opts...))
	log.Printf("clock running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
