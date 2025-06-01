package jwt

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEncode(t *testing.T) {
	id := uuid.New()
	NewEncoder([]byte("some secret")).EncodeUUID(id, time.Hour)
}

func TestDecode(t *testing.T) {
	id := uuid.New()
	token, err := NewEncoder([]byte("some secret")).EncodeUUID(id, time.Hour)
	if err != nil {
		t.Errorf("Failed to encode token: %v", err)
	}
	decodedID, err := NewDecoder([]byte("some secret")).DecodeUUID(token)
	if err != nil {
		t.Errorf("Failed to decode token: %v", err)
	}
	if decodedID != id {
		t.Errorf("Decoded ID does not match original ID")
	}
}

func TestDecodeInvalidToken(t *testing.T) {
	_, err := NewDecoder([]byte("some secret")).DecodeUUID("invalid_token")
	if err == nil {
		t.Errorf("Expected error decoding invalid token")
	}
}

func TestDecodeInvalidUUID(t *testing.T) {
	id := uuid.New()
	token, err := NewEncoder([]byte("some secret")).EncodeUUID(id, time.Hour)
	if err != nil {
		t.Errorf("Failed to encode token: %v", err)
	}
	token = token[:len(token)-1] // Remove last character to make it invalid
	_, err = NewDecoder([]byte("some secret")).DecodeUUID(token)
	if err == nil {
		t.Errorf("Expected error decoding invalid UUID")
	}
}

func TestDecodeInvalidSignature(t *testing.T) {
	id := uuid.New()
	token, err := NewEncoder([]byte("some secret")).EncodeUUID(id, time.Hour)
	if err != nil {
		t.Errorf("Failed to encode token: %v", err)
	}
	token = token[:len(token)-1] // Remove last character to make it invalid
	_, err = NewDecoder([]byte("invalid secret")).DecodeUUID(token)
	if err == nil {
		t.Errorf("Expected error decoding invalid signature")
	}
}

func TestDecodeExpiredToken(t *testing.T) {
	id := uuid.New()
	token, err := NewEncoder([]byte("some secret")).EncodeUUID(id, -time.Hour)
	if err != nil {
		t.Errorf("Failed to encode token: %v", err)
	}
	_, err = NewDecoder([]byte("some secret")).DecodeUUID(token)
	if err == nil {
		t.Errorf("Expected error decoding expired token")
	}
}
