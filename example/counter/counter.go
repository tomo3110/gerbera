package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	gp "github.com/tomo3110/gerbera/property"
	gl "github.com/tomo3110/gerbera/live"
)

type CounterView struct{ Count int }

func (v *CounterView) Mount(params gl.Params) error {
	v.Count = 0
	return nil
}

func (v *CounterView) Render() []g.ComponentFunc {
	return []g.ComponentFunc{
		gd.Head(gd.Title("カウンター")),
		gd.Body(
			gp.Class("container"),
			gd.H1(gp.Value(fmt.Sprintf("カウント: %d", v.Count))),
			gd.Div(
				gd.Button(gl.Click("dec"), gp.Value("-")),
				gd.Button(gl.Click("inc"), gp.Value("+")),
			),
		),
	}
}

func (v *CounterView) HandleEvent(event string, payload gl.Payload) error {
	switch event {
	case "inc":
		v.Count++
	case "dec":
		v.Count--
	}
	return nil
}

func main() {
	addr := flag.String("addr", ":8840", "listen address")
	debug := flag.Bool("debug", false, "enable debug panel")
	flag.Parse()

	var opts []gl.Option
	if *debug {
		opts = append(opts, gl.WithDebug())
	}

	http.Handle("/", gl.Handler(func(_ context.Context) gl.View { return &CounterView{} }, opts...))
	log.Printf("counter running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
