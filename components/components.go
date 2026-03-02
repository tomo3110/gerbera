package components

import (
	"github.com/tomo3110/gerbera"
)

// ensureDefaults inspects parent.Children for charset and viewport meta tags,
// adding any that are missing. charset is prepended; viewport is appended.
func ensureDefaults(parent *gerbera.Element) {
	hasCharset := false
	hasViewport := false
	for _, child := range parent.Children {
		if child.TagName != "meta" {
			continue
		}
		if _, ok := child.Attr["charset"]; ok {
			hasCharset = true
		}
		if child.Attr["name"] == "viewport" {
			hasViewport = true
		}
	}
	if !hasCharset {
		meta := &gerbera.Element{
			TagName: "meta",
			Attr:    gerbera.AttrMap{"charset": "utf-8"},
		}
		parent.Children = append([]*gerbera.Element{meta}, parent.Children...)
	}
	if !hasViewport {
		meta := &gerbera.Element{
			TagName: "meta",
			Attr: gerbera.AttrMap{
				"name":    "viewport",
				"content": "width=device-width, initial-scale=1",
			},
		}
		parent.Children = append(parent.Children, meta)
	}
}

// BootstrapCSS adds Bootstrap 5.3.3 CSS and JS (jsdelivr CDN with SRI) to the
// parent element. It also ensures charset and viewport meta tags are present.
func BootstrapCSS() gerbera.ComponentFunc {
	return func(parent *gerbera.Element) error {
		ensureDefaults(parent)
		parent.Children = append(parent.Children,
			&gerbera.Element{
				TagName: "link",
				Attr: gerbera.AttrMap{
					"rel":         "stylesheet",
					"href":        "https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css",
					"integrity":   "sha384-QWTKZyjpPEjISv5WaRU9OFeRpok6YctnYmDr5pNlyT2bRjXh0JMhjY6hW+ALEwIH",
					"crossorigin": "anonymous",
				},
			},
			&gerbera.Element{
				TagName: "script",
				Attr: gerbera.AttrMap{
					"src":         "https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/js/bootstrap.bundle.min.js",
					"integrity":   "sha384-YvpcrYf0tY3lHB60NNkmXc5s9fDVZLESaAA55NDzOxhy9GkcIdslK1eN7N6jIeHz",
					"crossorigin": "anonymous",
				},
			},
		)
		return nil
	}
}

// MaterializeCSS adds Materialize CSS 1.0.0 CSS and JS (cdnjs CDN) to the
// parent element. It also ensures charset and viewport meta tags are present.
func MaterializeCSS() gerbera.ComponentFunc {
	return func(parent *gerbera.Element) error {
		ensureDefaults(parent)
		parent.Children = append(parent.Children,
			&gerbera.Element{
				TagName: "link",
				Attr: gerbera.AttrMap{
					"rel":  "stylesheet",
					"href": "https://cdnjs.cloudflare.com/ajax/libs/materialize/1.0.0/css/materialize.min.css",
				},
			},
			&gerbera.Element{
				TagName: "script",
				Attr: gerbera.AttrMap{
					"src": "https://cdnjs.cloudflare.com/ajax/libs/materialize/1.0.0/js/materialize.min.js",
				},
			},
		)
		return nil
	}
}

// TailwindCSS adds Tailwind CSS 4 browser script (jsdelivr CDN) to the parent
// element. It also ensures charset and viewport meta tags are present.
func TailwindCSS() gerbera.ComponentFunc {
	return func(parent *gerbera.Element) error {
		ensureDefaults(parent)
		parent.Children = append(parent.Children,
			&gerbera.Element{
				TagName: "script",
				Attr: gerbera.AttrMap{
					"src": "https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4",
				},
			},
		)
		return nil
	}
}
