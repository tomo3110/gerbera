package session

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
)

const csrfKey = "_csrf_token"

// GenerateCSRFToken creates a new CSRF token, stores it in the session,
// and returns it. The token is a 32-character hex string.
func GenerateCSRFToken(sess *Session) string {
	b := make([]byte, 16)
	rand.Read(b)
	token := hex.EncodeToString(b)
	sess.Set(csrfKey, token)
	return token
}

// CSRFToken returns the CSRF token stored in the session, or empty string.
func CSRFToken(sess *Session) string {
	return sess.GetString(csrfKey)
}

// ValidCSRFToken reports whether the given token matches the one in the session.
// Uses constant-time comparison to prevent timing attacks.
func ValidCSRFToken(sess *Session, token string) bool {
	expected := CSRFToken(sess)
	if expected == "" || token == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(token)) == 1
}
