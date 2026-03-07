package assets

import (
	"net/url"
	"sort"

	"github.com/tomo3110/gerbera"
)

const (
	MetaKeyScripts     = "gerbera.web.scripts"
	MetaKeyStyleSheets = "gerbera.web.stylesheets"
)

// RequireScript registers a script path for automatic injection before </body>.
// Duplicate paths are ignored — only one <script> tag is emitted per unique path.
func RequireScript(n gerbera.Node, path *url.URL) {
	scripts := getStringSet(n, MetaKeyScripts)
	scripts[path.String()] = true
	n.SetMeta(MetaKeyScripts, scripts)
}

// RequireStyleSheet registers a stylesheet path for automatic injection before </head>.
// Duplicate paths are ignored — only one <link> tag is emitted per unique path.
func RequireStyleSheet(n gerbera.Node, path *url.URL) {
	styles := getStringSet(n, MetaKeyStyleSheets)
	styles[path.String()] = true
	n.SetMeta(MetaKeyStyleSheets, styles)
}

// Scripts returns the registered script paths from the node's metadata.
func Scripts(n gerbera.Node) []string {
	return getStringSetKeys(n, MetaKeyScripts)
}

// StyleSheets returns the registered stylesheet paths from the node's metadata.
func StyleSheets(n gerbera.Node) []string {
	return getStringSetKeys(n, MetaKeyStyleSheets)
}

func getStringSet(n gerbera.Node, key string) map[string]bool {
	if v := n.Meta(key); v != nil {
		return v.(map[string]bool)
	}
	return map[string]bool{}
}

func getStringSetKeys(n gerbera.Node, key string) []string {
	set := getStringSet(n, key)
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
