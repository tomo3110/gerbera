package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	g "github.com/tomo3110/gerbera"
	gc "github.com/tomo3110/gerbera/components"
	gd "github.com/tomo3110/gerbera/dom"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
)

type TabDemoView struct {
	ActiveTab int
}

func (v *TabDemoView) Mount(params gl.Params) error {
	v.ActiveTab = 0
	return nil
}

func (v *TabDemoView) Render() g.Components {
	tabs := []gc.Tab{
		{
			Label:       "Home",
			Content:     homeContent(),
			ButtonAttrs: tabClickAttrs("0"),
		},
		{
			Label:       "Profile",
			Content:     profileContent(),
			ButtonAttrs: tabClickAttrs("1"),
		},
		{
			Label:       "Settings",
			Content:     settingsContent(),
			ButtonAttrs: tabClickAttrs("2"),
		},
	}

	return g.Components{
		gd.Head(
			gd.Title("Tab Demo"),
			gc.TabsDefaultCSS(),
			gd.Style(gp.Value(pageCSS)),
		),
		gd.Body(
			gp.Class("container"),
			gd.H1(gp.Value("Gerbera Tab Component Demo")),
			gd.P(gp.Value(fmt.Sprintf("Active tab: %d", v.ActiveTab))),
			gc.Tabs("demo", v.ActiveTab, tabs),
		),
	}
}

func (v *TabDemoView) HandleEvent(event string, payload gl.Payload) error {
	switch event {
	case "switch-tab":
		if idx, err := strconv.Atoi(payload["value"]); err == nil {
			v.ActiveTab = idx
		}
	}
	return nil
}

func tabClickAttrs(index string) g.Components {
	return g.Components{
		gl.Click("switch-tab"),
		gl.ClickValue(index),
	}
}

func homeContent() g.ComponentFunc {
	return gd.Div(
		gd.H2(gp.Value("Welcome Home")),
		gd.P(gp.Value("This is the home tab content. Click the tabs above to switch between panels.")),
		gd.P(gp.Value("The tab component is fully accessible with ARIA attributes.")),
	)
}

func profileContent() g.ComponentFunc {
	return gd.Div(
		gd.H2(gp.Value("Profile")),
		gd.P(gp.Value("User: Gerbera Developer")),
		gd.P(gp.Value("Role: Full-stack Engineer")),
	)
}

func settingsContent() g.ComponentFunc {
	return gd.Div(
		gd.H2(gp.Value("Settings")),
		gd.P(gp.Value("Theme: Light")),
		gd.P(gp.Value("Language: English / Japanese")),
	)
}

const pageCSS = `body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; max-width: 720px; margin: 40px auto; padding: 0 16px; color: #333; } h1 { margin-bottom: 8px; } h1 + p { color: #666; margin-bottom: 24px; }`

func main() {
	addr := flag.String("addr", ":8890", "listen address")
	debug := flag.Bool("debug", false, "enable debug panel")
	flag.Parse()

	var opts []gl.Option
	if *debug {
		opts = append(opts, gl.WithDebug())
	}

	http.Handle("/", gl.Handler(func(_ context.Context) gl.View { return &TabDemoView{} }, opts...))
	log.Printf("tabview running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
