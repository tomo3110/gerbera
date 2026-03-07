package main

import (
	"fmt"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
	gs "github.com/tomo3110/gerbera/styles"
	gu "github.com/tomo3110/gerbera/ui"
)

type badgeCounts struct {
	Notifications int
	Messages      int
}

// snsPage returns the SSR shell layout with an embedded LiveMount.
func snsPage(title, activePage, liveEndpoint string, badges badgeCounts) g.Components {
	return g.Components{
		gd.Head(
			gd.Title(title+" — SNS"),
			gd.Meta(gp.Attr("name", "viewport"), gp.Attr("content", "width=device-width, initial-scale=1")),
			gu.Theme(),
			gs.CSS(snsCSS+drawerCSS),
		),
		gd.Body(
			// Mobile header
			gd.Div(
				gp.Class("sns-mobile-header"),
				gd.Button(
					gp.Attr("style", "background:none;border:none;cursor:pointer"),
					gp.Attr("onclick", "document.getElementById('sns-drawer').classList.toggle('open')"),
					gu.Icon("menu", ""),
				),
				gd.H1(gp.Value("SNS")),
			),
			// Main layout
			gd.Div(
				gp.Class("sns-layout"),
				// Desktop sidebar
				gd.Nav(
					gp.Class("sns-sidebar"),
					gd.Div(
						gp.Attr("style", "padding: 0 var(--g-space-md) var(--g-space-md); font-size: 1.3rem; font-weight: 700"),
						gp.Value("SNS"),
					),
					sidebarLinks(activePage, badges),
				),
				// Content area
				gd.Main(
					gp.Class("sns-main"),
					gl.LiveMount(liveEndpoint),
				),
			),
			// Mobile drawer overlay
			gd.Div(
				gp.ID("sns-drawer"),
				gp.Class("sns-drawer-overlay"),
				gd.Div(
					gp.Class("sns-drawer-panel"),
					gd.Div(
						gp.Class("sns-drawer-header"),
						gd.Span(gp.Attr("style", "font-weight:700;font-size:1.1rem"), gp.Value("SNS")),
						gd.Button(
							gp.Attr("style", "background:none;border:none;cursor:pointer;font-size:1.2rem"),
							gp.Attr("onclick", "document.getElementById('sns-drawer').classList.remove('open')"),
							gp.Value("\u00d7"),
						),
					),
					sidebarLinks(activePage, badges),
				),
			),
			g.Literal(`<script>document.getElementById('sns-drawer').addEventListener('click',function(e){if(e.target===this)this.classList.remove('open')})</script>`),
		),
	}
}

func sidebarLinks(activePage string, badges badgeCounts) g.ComponentFunc {
	return gd.Div(
		sidebarLink("home", "Home", "/", activePage == "home", 0),
		sidebarLink("search", "Search", "/search", activePage == "search", 0),
		sidebarLink("bell", "Notifications", "/", activePage == "", badges.Notifications),
		sidebarLink("mail", "Messages", "/messages", activePage == "messages", badges.Messages),
		sidebarLink("user", "Profile", "/profile", activePage == "profile", 0),
		sidebarLink("settings", "Settings", "/settings", activePage == "settings", 0),
		gd.Div(gp.Attr("style", "padding: var(--g-space-sm) var(--g-space-md)"),
			gd.A(gp.Attr("href", "/logout"),
				gu.Button("Logout", gu.ButtonOutline, gu.ButtonSmall, gp.Attr("style", "width:100%")),
			),
		),
	)
}

func sidebarLink(icon, label, href string, active bool, badgeCount int) g.ComponentFunc {
	cls := "sns-nav-link"
	if active {
		cls = "sns-nav-link active"
	}
	var children g.Components
	children = append(children,
		gp.Class(cls),
		gp.Attr("href", href),
		gu.Icon(icon, ""),
		gd.Span(gp.Value(label)),
	)
	if badgeCount > 0 {
		children = append(children,
			gd.Span(gp.Class("sns-nav-badge"), gp.Value(fmt.Sprintf("%d", badgeCount))),
		)
	}
	return gd.A(children...)
}

const drawerCSS = `
/* === Mobile Drawer Overlay === */
.sns-drawer-overlay {
	display: none;
	position: fixed;
	inset: 0;
	z-index: 100;
	background: rgba(0,0,0,0.4);
}
.sns-drawer-overlay.open {
	display: block;
}
.sns-drawer-panel {
	width: 260px;
	height: 100%;
	background: var(--g-bg-surface);
	overflow-y: auto;
}
.sns-drawer-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: var(--g-space-md);
	border-bottom: 1px solid var(--g-border);
}
@media (min-width: 769px) {
	.sns-drawer-overlay { display: none !important; }
}
a.sns-nav-link {
	text-decoration: none;
}
`
