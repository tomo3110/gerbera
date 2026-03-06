package live

import "strings"

// PathParams holds named parameters extracted from a URL path.
type PathParams map[string]string

// Get returns the value for the given parameter name.
// Returns an empty string if the parameter does not exist.
func (p PathParams) Get(key string) string {
	if p == nil {
		return ""
	}
	return p[key]
}

// MatchPath matches a URL path against a pattern with ":param" segments.
// Returns the extracted parameters and true if the path matches.
//
//	MatchPath("/users/:id", "/users/42")           => {"id": "42"}, true
//	MatchPath("/users/:uid/posts/:pid", "/u/1/p/2") => nil, false
//	MatchPath("/users/:id", "/posts/42")            => nil, false
func MatchPath(pattern, path string) (PathParams, bool) {
	pattern = trimSlash(pattern)
	path = trimSlash(path)

	// Handle root path
	if pattern == "" && path == "" {
		return PathParams{}, true
	}

	patternSegs := strings.Split(pattern, "/")
	pathSegs := strings.Split(path, "/")

	if len(patternSegs) != len(pathSegs) {
		return nil, false
	}

	params := PathParams{}
	for i, seg := range patternSegs {
		if strings.HasPrefix(seg, ":") {
			params[seg[1:]] = pathSegs[i]
		} else if seg != pathSegs[i] {
			return nil, false
		}
	}
	return params, true
}

// trimSlash removes leading and trailing slashes from a path.
func trimSlash(s string) string {
	return strings.Trim(s, "/")
}
