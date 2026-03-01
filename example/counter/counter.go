package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	gp "github.com/tomo3110/gerbera/property"
	"github.com/tomo3110/gerbera/live"
)

type CounterView struct{ Count int }

func (v *CounterView) Mount(params live.Params) error {
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
				gd.Button(live.Click("dec"), gp.Value("-")),
				gd.Button(live.Click("inc"), gp.Value("+")),
			),
		),
	}
}

func (v *CounterView) HandleEvent(event string, payload live.Payload) error {
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
	flag.Parse()
	http.Handle("/", live.Handler(func() live.View { return &CounterView{} }))
	log.Printf("counter running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
