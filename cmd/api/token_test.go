package api_test

import (
	"evidence/cmd/api"
	"github.com/google/go-cmp/cmp"
	"testing"
	"time"
)

func TestTokenGeneratedAndVerifiedForUser(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		duration time.Duration
	}{
		{
			name:     "U1",
			username: "Simba",
			duration: time.Duration(1) * time.Hour,
		},
		{
			name:     "U2",
			username: "Pumba",
			duration: time.Duration(1) * time.Hour,
		},
		{
			name:     "U3",
			username: "Timon",
			duration: time.Duration(1) * time.Hour,
		},
		{
			name:     "U4",
			username: "Nala",
			duration: time.Duration(1) * time.Hour,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			maker, err := api.NewPasetoMaker("nigkjtvbrhugwpgaqbemmvnqbtywfrcq")
			if err != nil {
				t.Errorf("Failed to create a new paseto maker: %s", err)
			}
			token, payload, err := maker.CreateToken(tc.username, tc.duration)
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
			if !cmp.Equal(tc.username, decryptedPayload.Username) {
				t.Errorf(cmp.Diff(tc.username, decryptedPayload.Username))
			}
		})
	}
}
func TestExpiredTokenDetectedSuccessfully(t *testing.T) {
	maker, err := api.NewPasetoMaker("nigkjtvbrhugwpgaqbemmvnqbtywfrcq")
	if err != nil {
		t.Errorf("Failed to create a new paseto maker: %s", err)
	}
	username := "Simba"
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
	if token == "" {
		t.Errorf("Failed to create a new token: token is empty")
	}
	time.Sleep(time.Duration(1) * time.Millisecond)
	_, err = maker.VerifyToken(token)
	if err == nil {
		t.Errorf("Token should have expired but it didn't")
	}
}
