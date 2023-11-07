package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/o1egl/paseto"
	"golang.org/x/crypto/chacha20poly1305"
)

// The Payload represents the data that is encoded into a token. It contains:
// - ID: a unique identifier for the token.
// - Username: the name of the user for whom the token was issued.
// - IssuedAt: the timestamp indicating when the token was created.
// - ExpiresAt: the timestamp indicating when the token will expire.
type Payload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ValidTime checks if the tokenMaker has expired
func (payload *Payload) ValidTime() error {
	if time.Now().After(payload.ExpiresAt) {
		return errors.New("token has expired")
	}

	return nil
}

// NewPayload creates a new payload for specific user and duration
func NewPayload(username string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:        tokenID,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(duration),
	}

	return payload, nil
}

// Maker interface for creating and verifying all types of tokens
type Maker interface {
	// CreateToken creates a new tokenMaker
	CreateToken(username string, duration time.Duration) (string, *Payload, error)

	VerifyToken(token string) (*Payload, error)

	CreateRefreshToken(username string, duration time.Duration) (string, *Payload, error)

	VerifyRefreshToken(token string) (*Payload, error)
}

// PasetoMaker is a paseto implementation of maker interface
type PasetoMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

// NewPasetoMaker initializes a new instance of PasetoMaker, which implements the Maker interface
// using the PASETO token standard. The function requires a symmetric key for token encryption and decryption.
// The length of the symmetricKey must match the expected key size used by the chacha20poly1305 algorithm.
func NewPasetoMaker(symmetricKey string) (Maker, error) {
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: must be exactly %d characters", chacha20poly1305.KeySize)
	}

	maker := &PasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}

	return maker, nil
}

// CreateToken creates a new Paseto token for a specific user
func (maker *PasetoMaker) CreateToken(username string, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", payload, err
	}

	token, err := maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
	if err != nil {
		return "", nil, err
	}

	return token, payload, nil
}

// VerifyToken attempts to decrypt and validate the provided token using the PASETO standard.
func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}

	err := maker.paseto.Decrypt(token, maker.symmetricKey, payload, nil)
	if err != nil {
		return nil, err
	}

	err = payload.ValidTime()
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func (maker *PasetoMaker) CreateRefreshToken(username string, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", payload, err
	}

	token, err := maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
	if err != nil {
		return "", nil, err
	}

	return token, payload, nil
}

func (maker *PasetoMaker) VerifyRefreshToken(token string) (*Payload, error) {
	return maker.VerifyToken(token)
}
