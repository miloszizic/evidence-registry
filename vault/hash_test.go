package vault_test

import (
	"testing"

	"github.com/miloszizic/der/vault"
	"golang.org/x/crypto/bcrypt"
)

func TestHash(t *testing.T) {
	t.Parallel()
	pass := "mysecretpassword"
	hashedPassword, err := vault.Hash(pass)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hashedPassword == pass || hashedPassword == "" {
		t.Fatalf("hashedPassword should not be the same as plaintext password or empty string")
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(pass))
	if err != nil {
		t.Fatalf("hashed password does not match the plaintext password: %v", err)
	}
}

func TestMatches(t *testing.T) {
	t.Parallel()
	pass := "mysecretpassword"
	wrongPassword := "wrongpassword"
	hashedPassword, err := vault.Hash(pass)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	matches, err := vault.Matches(pass, hashedPassword)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !matches {
		t.Fatalf("the function returned false, but we expected true as the passwords should match")
	}

	matches, err = vault.Matches(wrongPassword, hashedPassword)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if matches {
		t.Fatalf("the function returned true, but we expected false as the passwords should not match")
	}
}
