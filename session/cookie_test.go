package session

import "testing"

func TestSignVerify_RoundTrip(t *testing.T) {
	key := []byte("test-secret-key")
	id := "session-123"
	signed := sign(id, key)
	got, err := verify(signed, key)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if got != id {
		t.Errorf("verify returned %q, want %q", got, id)
	}
}

func TestVerify_TamperedSignature(t *testing.T) {
	key := []byte("test-secret-key")
	signed := sign("session-123", key)
	// Tamper with the last character
	tampered := signed[:len(signed)-1] + "X"
	_, err := verify(tampered, key)
	if err == nil {
		t.Error("verify should fail for tampered signature")
	}
}

func TestVerify_WrongKey(t *testing.T) {
	key1 := []byte("key-one")
	key2 := []byte("key-two")
	signed := sign("session-123", key1)
	_, err := verify(signed, key2)
	if err == nil {
		t.Error("verify should fail for wrong key")
	}
}

func TestVerify_InvalidFormat(t *testing.T) {
	key := []byte("key")
	_, err := verify("no-dot-here", key)
	if err == nil {
		t.Error("verify should fail for invalid format")
	}
}

func TestVerify_InvalidBase64(t *testing.T) {
	key := []byte("key")
	_, err := verify("!!!.!!!", key)
	if err == nil {
		t.Error("verify should fail for invalid base64")
	}
}
