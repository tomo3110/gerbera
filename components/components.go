package components

import (
	"fmt"

	"github.com/tomo3110/gerbera"
)

// Default CDN versions.
const (
	DefaultBootstrapVersion   = "5.3.3"
	DefaultMaterializeVersion = "1.0.0"
	DefaultTailwindVersion    = "4"
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

// BootstrapCSS adds Bootstrap CSS and JS (jsdelivr CDN) to the parent element.
// It also ensures charset and viewport meta tags are present.
// When called without arguments, uses the default version (5.3.3) with SRI hashes.
// When a version is specified, SRI hashes are omitted.
func BootstrapCSS(version ...string) gerbera.ComponentFunc {
	ver := DefaultBootstrapVersion
	if len(version) > 0 {
		ver = version[0]
	}
	return func(parent *gerbera.Element) error {
		ensureDefaults(parent)
		cssURL := fmt.Sprintf("https://cdn.jsdelivr.net/npm/bootstrap@%s/dist/css/bootstrap.min.css", ver)
		jsURL := fmt.Sprintf("https://cdn.jsdelivr.net/npm/bootstrap@%s/dist/js/bootstrap.bundle.min.js", ver)
		linkAttr := gerbera.AttrMap{
			"rel":  "stylesheet",
			"href": cssURL,
		}
		scriptAttr := gerbera.AttrMap{
			"src": jsURL,
		}
		if ver == DefaultBootstrapVersion {
			linkAttr["integrity"] = "sha384-QWTKZyjpPEjISv5WaRU9OFeRpok6YctnYmDr5pNlyT2bRjXh0JMhjY6hW+ALEwIH"
			linkAttr["crossorigin"] = "anonymous"
			scriptAttr["integrity"] = "sha384-YvpcrYf0tY3lHB60NNkmXc5s9fDVZLESaAA55NDzOxhy9GkcIdslK1eN7N6jIeHz"
			scriptAttr["crossorigin"] = "anonymous"
		}
		parent.Children = append(parent.Children,
			&gerbera.Element{TagName: "link", Attr: linkAttr},
			&gerbera.Element{TagName: "script", Attr: scriptAttr},
		)
		return nil
	}
}

// MaterializeCSS adds Materialize CSS and JS (cdnjs CDN) to the parent element.
// It also ensures charset and viewport meta tags are present.
// When called without arguments, uses the default version (1.0.0).
func MaterializeCSS(version ...string) gerbera.ComponentFunc {
	ver := DefaultMaterializeVersion
	if len(version) > 0 {
		ver = version[0]
	}
	return func(parent *gerbera.Element) error {
		ensureDefaults(parent)
		parent.Children = append(parent.Children,
			&gerbera.Element{
				TagName: "link",
				Attr: gerbera.AttrMap{
					"rel":  "stylesheet",
					"href": fmt.Sprintf("https://cdnjs.cloudflare.com/ajax/libs/materialize/%s/css/materialize.min.css", ver),
				},
			},
			&gerbera.Element{
				TagName: "script",
				Attr: gerbera.AttrMap{
					"src": fmt.Sprintf("https://cdnjs.cloudflare.com/ajax/libs/materialize/%s/js/materialize.min.js", ver),
				},
			},
		)
		return nil
	}
}

// TailwindCSS adds Tailwind CSS browser script (jsdelivr CDN) to the parent
// element. It also ensures charset and viewport meta tags are present.
// When called without arguments, uses the default version (4).
func TailwindCSS(version ...string) gerbera.ComponentFunc {
	ver := DefaultTailwindVersion
	if len(version) > 0 {
		ver = version[0]
	}
	return func(parent *gerbera.Element) error {
		ensureDefaults(parent)
		parent.Children = append(parent.Children,
			&gerbera.Element{
				TagName: "script",
				Attr: gerbera.AttrMap{
					"src": fmt.Sprintf("https://cdn.jsdelivr.net/npm/@tailwindcss/browser@%s", ver),
				},
			},
		)
		return nil
	}
}
