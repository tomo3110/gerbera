package main

import (
	"flag"
	"log"
	"net/http"

	g "github.com/tomo3110/gerbera"
	gc "github.com/tomo3110/gerbera/components"
	gd "github.com/tomo3110/gerbera/dom"
	gp "github.com/tomo3110/gerbera/property"
)

func main() {
	addr := flag.String("addr", ":8800", "running address")
	mux := g.NewServeMux(
		gc.BootStrapCDNHead("Gerbera Template Engine !"),
		body(),
	)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func body() g.ComponentFunc {
	return gd.Body(
		gp.Class("container"),
		gd.H1(gp.Value("Gerbera Template Engine !")),
		gd.P(gp.Value("view html template Test")),
	)
}
