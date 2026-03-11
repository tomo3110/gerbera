package live

import "github.com/tomo3110/gerbera/assets"

// gerberaJS returns the client-side JavaScript content.
// The single source of truth is assets/files/gerbera.js.
func gerberaJSContent() string {
	return assets.JSString()
}
