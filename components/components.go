package components

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
)

func BootStrapCDNHead(title string) gerbera.ComponentFunc {
	return dom.Head(
		dom.Meta(
			property.Attr("charset", "utf-8"),
		),
		dom.Meta(
			property.Attr("name", "description"),
			property.Attr("content", "Todo Sapropertyle App"),
		),
		dom.Meta(
			property.Attr("name", "viewport"),
			property.Attr("content", "width=device-width, initial-scale=1"),
		),
		dom.Title(title),
		dom.Link(
			property.Attr("rel", "stylesheet"),
			property.Attr("href", "https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/css/bootstrap.min.css"),
			property.Attr("integrity", "sha384-Gn5384xqQ1aoWXA+058RXPxPg6fy4IWvTNh0E263XmFcJlSAwiGgFAW/dAiS6JXm"),
			property.Attr("crossorigin", "anonymous"),
		),
		dom.Script(
			"https://code.jquery.com/jquery-3.2.1.slim.min.js",
			property.Attr("integrity", "sha384-KJ3o2DKtIkvYIK3UENzmM7KCkRr/rE9/Qpg6aAZGJwFDMVNA/GpGFF93hXpG5KkN"),
			property.Attr("crossorigin", "anonymous"),
		),
		dom.Script(
			"https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.12.9/udom/popper.min.js",
			property.Attr("integrity", "sha384-ApNbgh9B+Y1QKtv3Rn7W3mgPxhU9K/ScQsAP7hUibX39j7fakFPskvXusvfa0b4Q"),
			property.Attr("crossorigin", "anonymous"),
		),
		dom.Script(
			"https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/js/bootstrap.min.js",
			property.Attr("integrity", "sha384-JZR6Spejh4U02d8jOt6vLEHfe/JQGiRRSQQxSfFWpi1MquVdAyjUar5+76PVCmYl"),
			property.Attr("crossorigin", "anonymous"),
		),
	)
}

func MaterilalizecssCDNHead(title string) gerbera.ComponentFunc {
	return dom.Head(
		dom.Meta(
			property.Attr("charset", "utf-8"),
		),
		dom.Meta(
			property.Attr("name", "description"),
			property.Attr("content", "Todo Sapropertyle App"),
		),
		dom.Meta(
			property.Attr("name", "viewport"),
			property.Attr("content", "width=device-width, initial-scale=1"),
		),
		dom.Title(title),
		dom.Link(
			property.Attr("rel", "stylesheet"),
			property.Attr("href", "https://cdnjs.cloudflare.com/ajax/libs/materialize/1.0.0-beta/css/materialize.min.css"),
		),
		dom.Script("https://cdnjs.cloudflare.com/ajax/libs/materialize/1.0.0-beta/js/materialize.min.js"),
	)
}
