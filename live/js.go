package live

import "github.com/tomo3110/gerbera/assets"

// gerberaJS returns the client-side JavaScript content.
// The single source of truth is assets/files/gerbera.js.
func gerberaJSContent() string {
	return assets.JSString()
}

// gerberaDebugJSContent returns the debug panel JavaScript content.
// The single source of truth is assets/files/gerbera_debug.js.
func gerberaDebugJSContent() string {
	return assets.DebugJSString()
}
