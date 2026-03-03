package ui

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
)

// Icon renders an inline SVG icon.
// name selects a built-in icon. size: "sm", "md" (default), "lg".
func Icon(name string, size ...string) gerbera.ComponentFunc {
	s := "md"
	if len(size) > 0 && size[0] != "" {
		s = size[0]
	}
	svg, ok := icons[name]
	if !ok {
		svg = icons["circle"]
	}
	return dom.Span(
		property.Class("g-icon", "g-icon-"+s),
		property.AriaHidden(true),
		gerbera.Literal(svg),
	)
}

// IconNames returns all available icon names.
func IconNames() []string {
	names := make([]string, 0, len(icons))
	for k := range icons {
		names = append(names, k)
	}
	return names
}

// All SVG icons are 18x18 viewBox, stroke-based for crisp rendering.
var icons = map[string]string{
	"home": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M3 7.5L9 2.5l6 5v7a1 1 0 01-1 1H4a1 1 0 01-1-1z"/><path d="M7 15.5v-5h4v5"/></svg>`,

	"user": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="9" cy="6" r="3"/><path d="M3 16.5v-1a4 4 0 014-4h4a4 4 0 014 4v1"/></svg>`,

	"users": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="7" cy="6" r="2.5"/><path d="M2 16v-1a3.5 3.5 0 013.5-3.5h3A3.5 3.5 0 0112 15v1"/><circle cx="13" cy="6.5" r="2"/><path d="M13 11.5a3.5 3.5 0 013 3.5v1"/></svg>`,

	"settings": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="9" cy="9" r="2.5"/><path d="M14.7 11.1a1 1 0 00.2 1.1l.04.04a1.2 1.2 0 11-1.7 1.7l-.04-.04a1 1 0 00-1.1-.2 1 1 0 00-.6.9v.12a1.2 1.2 0 11-2.4 0v-.06a1 1 0 00-.66-.93 1 1 0 00-1.1.2l-.04.04a1.2 1.2 0 11-1.7-1.7l.04-.04a1 1 0 00.2-1.1 1 1 0 00-.9-.6H4.7a1.2 1.2 0 110-2.4h.06a1 1 0 00.93-.66 1 1 0 00-.2-1.1l-.04-.04a1.2 1.2 0 111.7-1.7l.04.04a1 1 0 001.1.2h.04a1 1 0 00.6-.9V4.7a1.2 1.2 0 112.4 0v.06a1 1 0 00.6.93 1 1 0 001.1-.2l.04-.04a1.2 1.2 0 111.7 1.7l-.04.04a1 1 0 00-.2 1.1v.04a1 1 0 00.9.6h.12a1.2 1.2 0 110 2.4h-.06a1 1 0 00-.93.6z"/></svg>`,

	"search": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="7.5" cy="7.5" r="4.5"/><path d="M16 16l-3.5-3.5"/></svg>`,

	"plus": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M9 3v12M3 9h12"/></svg>`,

	"minus": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9h12"/></svg>`,

	"x": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M4 4l10 10M14 4L4 14"/></svg>`,

	"check": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M3.5 9.5l3.5 3.5 7.5-8"/></svg>`,

	"edit": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M10.5 3.5l4 4L6 16H2v-4z"/></svg>`,

	"trash": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M3 5h12M6 5V3.5a1 1 0 011-1h4a1 1 0 011 1V5"/><path d="M4.5 5l.7 10a1 1 0 001 .9h5.6a1 1 0 001-.9l.7-10"/></svg>`,

	"chevron-right": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M7 4l5 5-5 5"/></svg>`,

	"chevron-down": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M4 7l5 5 5-5"/></svg>`,

	"chevron-left": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4l-5 5 5 5"/></svg>`,

	"chevron-up": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M4 11l5-5 5 5"/></svg>`,

	"menu": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M3 4.5h12M3 9h12M3 13.5h12"/></svg>`,

	"folder": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M2 5.5V14a1 1 0 001 1h12a1 1 0 001-1V7a1 1 0 00-1-1H9L7.5 4H3a1 1 0 00-1 1.5z"/></svg>`,

	"file": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M10 2H5a1 1 0 00-1 1v12a1 1 0 001 1h8a1 1 0 001-1V6z"/><path d="M10 2v4h4"/></svg>`,

	"mail": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="4" width="14" height="10" rx="1"/><path d="M2 5l7 5 7-5"/></svg>`,

	"bell": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M7.5 15.5a1.5 1.5 0 003 0"/><path d="M4 11V8a5 5 0 0110 0v3l1.5 2H2.5z"/></svg>`,

	"calendar": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><rect x="2.5" y="3.5" width="13" height="12" rx="1"/><path d="M2.5 7.5h13M6 2v3M12 2v3"/></svg>`,

	"chart": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M5 13V9M9 13V5M13 13V8"/></svg>`,

	"download": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M9 2v10M5 8l4 4 4-4M3 14h12"/></svg>`,

	"upload": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M9 12V2M5 6l4-4 4 4M3 14h12"/></svg>`,

	"link": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M7.5 10.5a3 3 0 004.24 0l2.12-2.12a3 3 0 00-4.24-4.24L8.5 5.25"/><path d="M10.5 7.5a3 3 0 00-4.24 0L4.14 9.62a3 3 0 004.24 4.24L9.5 12.75"/></svg>`,

	"info": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="9" cy="9" r="7"/><path d="M9 8v4M9 6v.01"/></svg>`,

	"warning": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M9 2L1.5 15h15z"/><path d="M9 7v3M9 12v.01"/></svg>`,

	"circle": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="9" cy="9" r="4"/></svg>`,

	"lock": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><rect x="4" y="8" width="10" height="7" rx="1"/><path d="M6 8V5.5a3 3 0 016 0V8"/></svg>`,

	"logout": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M6 15H4a1 1 0 01-1-1V4a1 1 0 011-1h2M12 12l3-3-3-3M7 9h8"/></svg>`,

	"filter": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M2 3h14l-5.5 6.5V14l-3 2v-6.5z"/></svg>`,

	"sort-asc": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M9 3v12M5 7l4-4 4 4"/></svg>`,

	"sort-desc": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M9 3v12M5 11l4 4 4-4"/></svg>`,

	"eye": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M1.5 9s3-5.5 7.5-5.5S16.5 9 16.5 9s-3 5.5-7.5 5.5S1.5 9 1.5 9z"/><circle cx="9" cy="9" r="2.5"/></svg>`,

	"copy": `<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><rect x="6" y="6" width="9" height="9" rx="1"/><path d="M3 12V3a1 1 0 011-1h9"/></svg>`,
}
