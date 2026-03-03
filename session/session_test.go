package session

import "testing"

func TestSession_GetSet(t *testing.T) {
	s := &Session{Values: make(map[string]any)}
	s.Set("name", "alice")
	if got := s.Get("name"); got != "alice" {
		t.Errorf("Get(name) = %v, want alice", got)
	}
}

func TestSession_GetString(t *testing.T) {
	s := &Session{Values: make(map[string]any)}
	s.Set("name", "bob")
	s.Set("count", 42)

	if got := s.GetString("name"); got != "bob" {
		t.Errorf("GetString(name) = %q, want bob", got)
	}
	if got := s.GetString("count"); got != "" {
		t.Errorf("GetString(count) = %q, want empty", got)
	}
	if got := s.GetString("missing"); got != "" {
		t.Errorf("GetString(missing) = %q, want empty", got)
	}
}

func TestSession_Delete(t *testing.T) {
	s := &Session{Values: make(map[string]any)}
	s.Set("key", "value")
	s.Delete("key")
	if got := s.Get("key"); got != nil {
		t.Errorf("Get(key) after Delete = %v, want nil", got)
	}
}

func TestSession_Clear(t *testing.T) {
	s := &Session{Values: make(map[string]any)}
	s.Set("a", 1)
	s.Set("b", 2)
	s.Clear()
	if got := s.Get("a"); got != nil {
		t.Errorf("Get(a) after Clear = %v, want nil", got)
	}
	if len(s.Values) != 0 {
		t.Errorf("Values length after Clear = %d, want 0", len(s.Values))
	}
}

func TestSession_Modified(t *testing.T) {
	s := &Session{Values: make(map[string]any)}
	if s.Modified() {
		t.Error("new session should not be modified")
	}
	s.Set("key", "value")
	if !s.Modified() {
		t.Error("session should be modified after Set")
	}
}
