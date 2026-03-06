package live

import "testing"

func TestMatchPath(t *testing.T) {
	tests := []struct {
		name      string
		pattern   string
		path      string
		wantMatch bool
		wantParams PathParams
	}{
		{
			name:      "root matches root",
			pattern:   "/",
			path:      "/",
			wantMatch: true,
			wantParams: PathParams{},
		},
		{
			name:      "static path matches",
			pattern:   "/settings",
			path:      "/settings",
			wantMatch: true,
			wantParams: PathParams{},
		},
		{
			name:      "single parameter",
			pattern:   "/users/:id",
			path:      "/users/42",
			wantMatch: true,
			wantParams: PathParams{"id": "42"},
		},
		{
			name:      "multiple parameters",
			pattern:   "/users/:uid/posts/:pid",
			path:      "/users/1/posts/99",
			wantMatch: true,
			wantParams: PathParams{"uid": "1", "pid": "99"},
		},
		{
			name:      "segment count mismatch - too few",
			pattern:   "/users/:id",
			path:      "/users",
			wantMatch: false,
		},
		{
			name:      "segment count mismatch - too many",
			pattern:   "/users/:id",
			path:      "/users/42/extra",
			wantMatch: false,
		},
		{
			name:      "static prefix mismatch",
			pattern:   "/users/:id",
			path:      "/posts/42",
			wantMatch: false,
		},
		{
			name:      "trailing slash normalization - pattern",
			pattern:   "/users/:id/",
			path:      "/users/42",
			wantMatch: true,
			wantParams: PathParams{"id": "42"},
		},
		{
			name:      "trailing slash normalization - path",
			pattern:   "/users/:id",
			path:      "/users/42/",
			wantMatch: true,
			wantParams: PathParams{"id": "42"},
		},
		{
			name:      "trailing slash normalization - both",
			pattern:   "/users/:id/",
			path:      "/users/42/",
			wantMatch: true,
			wantParams: PathParams{"id": "42"},
		},
		{
			name:      "root does not match non-root",
			pattern:   "/",
			path:      "/users",
			wantMatch: false,
		},
		{
			name:      "non-root does not match root",
			pattern:   "/users",
			path:      "/",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, ok := MatchPath(tt.pattern, tt.path)
			if ok != tt.wantMatch {
				t.Errorf("MatchPath(%q, %q) match = %v, want %v", tt.pattern, tt.path, ok, tt.wantMatch)
				return
			}
			if !tt.wantMatch {
				return
			}
			if len(params) != len(tt.wantParams) {
				t.Errorf("MatchPath(%q, %q) params len = %d, want %d", tt.pattern, tt.path, len(params), len(tt.wantParams))
				return
			}
			for k, want := range tt.wantParams {
				if got := params.Get(k); got != want {
					t.Errorf("params[%q] = %q, want %q", k, got, want)
				}
			}
		})
	}
}

func TestPathParams_Get_nil(t *testing.T) {
	var p PathParams
	if got := p.Get("anything"); got != "" {
		t.Errorf("nil PathParams.Get() = %q, want empty string", got)
	}
}
