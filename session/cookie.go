package session

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

// CookieConfig holds cookie parameters.
type CookieConfig struct {
	Name     string
	Path     string
	MaxAge   int
	Secure   bool
	HttpOnly bool
	SameSite http.SameSite
}

func defaultCookieConfig() CookieConfig {
	return CookieConfig{
		Name:     "gerbera_session",
		Path:     "/",
		MaxAge:   86400, // 24 hours
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

// sign produces a signed cookie value: base64(id).base64(hmac).
func sign(sessionID string, key []byte) string {
	idEnc := base64.RawURLEncoding.EncodeToString([]byte(sessionID))
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(sessionID))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return idEnc + "." + sig
}

// verify validates a signed cookie value and returns the session ID.
func verify(cookieValue string, key []byte) (string, error) {
	parts := strings.SplitN(cookieValue, ".", 2)
	if len(parts) != 2 {
		return "", errors.New("session: invalid cookie format")
	}
	idBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", errors.New("session: invalid cookie encoding")
	}
	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", errors.New("session: invalid signature encoding")
	}
	mac := hmac.New(sha256.New, key)
	mac.Write(idBytes)
	expected := mac.Sum(nil)
	if !hmac.Equal(sigBytes, expected) {
		return "", errors.New("session: invalid signature")
	}
	return string(idBytes), nil
}

// writeCookie writes the session cookie to the response.
func writeCookie(w http.ResponseWriter, cfg CookieConfig, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.Name,
		Value:    value,
		Path:     cfg.Path,
		MaxAge:   cfg.MaxAge,
		Secure:   cfg.Secure,
		HttpOnly: cfg.HttpOnly,
		SameSite: cfg.SameSite,
	})
}

// clearCookie removes the session cookie from the response.
func clearCookie(w http.ResponseWriter, cfg CookieConfig) {
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.Name,
		Value:    "",
		Path:     cfg.Path,
		MaxAge:   -1,
		Secure:   cfg.Secure,
		HttpOnly: cfg.HttpOnly,
		SameSite: cfg.SameSite,
	})
}
