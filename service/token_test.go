package service

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestTokenGeneratedAndVerifiedForUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		username string
		duration time.Duration
	}{
		{
			name:     "1",
			username: "Simba",
			duration: time.Duration(1) * time.Hour,
		},
		{
			name:     "2",
			username: "Pumba",
			duration: time.Duration(1) * time.Hour,
		},
		{
			name:     "3",
			username: "Timon",
			duration: time.Duration(1) * time.Hour,
		},
		{
			name:     "4",
			username: "Nala",
			duration: time.Duration(1) * time.Hour,
		},
	}
	for _, tt := range tests {
		pt := tt
		t.Run(pt.name, func(t *testing.T) {
			t.Parallel()

			maker, err := NewPasetoMaker("nigkjtvbrhugwpgaqbemmvnqbtywfrcq")
			if err != nil {
				t.Errorf("Failed to create a new paseto maker: %s", err)
			}
			token, payload, err := maker.CreateToken(pt.username, pt.duration)
			if err != nil {
				t.Errorf("Failed to create a new token: %s", err)
			}
			if payload == nil {
				t.Errorf("Failed to create a new token: payload is nil")
			}
			if token == "" {
				t.Errorf("Failed to create a new token: token is empty")
			}
			decryptedPayload, err := maker.VerifyToken(token)
			if err != nil {
				t.Errorf("Failed to verify token: %s", err)
			}
			if !cmp.Equal(pt.username, decryptedPayload.Username) {
				t.Errorf(cmp.Diff(pt.username, decryptedPayload.Username))
			}
		})
	}
}

func TestExpiredTokenDetectedSuccessfully(t *testing.T) {
	maker, err := NewPasetoMaker("nigkjtvbrhugwpgaqbemmvnqbtywfrcq")
	if err != nil {
		t.Errorf("Failed to create a new paseto maker: %s", err)
	}
	// default expectedUser
	username := "Simba"
	// default duration
	duration := time.Duration(1) * time.Microsecond

	token, payload, err := maker.CreateToken(username, duration)
	if err != nil {
		t.Errorf("Failed to create a new token: %s", err)
		return
	}

	if payload == nil {
		t.Errorf("Failed to create a new token: payload is nil")
		return
	}

	time.Sleep(time.Duration(1) * time.Millisecond)

	_, err = maker.VerifyToken(token)
	if err == nil {
		t.Errorf("Token should have expired but it didn't")
	}
}
