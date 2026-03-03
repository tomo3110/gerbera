package session

import "testing"

func TestCSRFToken_GenerateAndValidate(t *testing.T) {
	sess := &Session{Values: make(map[string]any)}
	token := GenerateCSRFToken(sess)
	if token == "" {
		t.Fatal("GenerateCSRFToken returned empty token")
	}
	if !ValidCSRFToken(sess, token) {
		t.Error("ValidCSRFToken should return true for valid token")
	}
}

func TestCSRFToken_InvalidToken(t *testing.T) {
	sess := &Session{Values: make(map[string]any)}
	GenerateCSRFToken(sess)
	if ValidCSRFToken(sess, "wrong-token") {
		t.Error("ValidCSRFToken should return false for wrong token")
	}
}

func TestCSRFToken_EmptySession(t *testing.T) {
	sess := &Session{Values: make(map[string]any)}
	if ValidCSRFToken(sess, "any-token") {
		t.Error("ValidCSRFToken should return false when no token in session")
	}
}

func TestCSRFToken_EmptyInput(t *testing.T) {
	sess := &Session{Values: make(map[string]any)}
	GenerateCSRFToken(sess)
	if ValidCSRFToken(sess, "") {
		t.Error("ValidCSRFToken should return false for empty token")
	}
}

func TestCSRFToken_Get(t *testing.T) {
	sess := &Session{Values: make(map[string]any)}
	if got := CSRFToken(sess); got != "" {
		t.Errorf("CSRFToken on new session = %q, want empty", got)
	}
	token := GenerateCSRFToken(sess)
	if got := CSRFToken(sess); got != token {
		t.Errorf("CSRFToken = %q, want %q", got, token)
	}
}
