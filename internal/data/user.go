package data

import (
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int64    `json:"id"`
	Username string   `json:"username,omitempty"`
	Password password `json:"password,omitempty"`
	Role     string   `json:"role,omitempty"`
	Token    string   `json:"token,omitempty"`
	Cases    []Case   `json:"buckets,omitempty"`
}

type password struct {
	plaintext *string
	hash      []byte
}

// Set takes a plaintext password and hashes it.
func (p *password) Set(plaintextPassword string) error {
	if plaintextPassword == "" {
		return fmt.Errorf("%w: password cannot be empty", ErrInvalidRequest)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

// Matches returns true if the plaintext password matches the hash.
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

type UserStore interface {
	Add(user *User) error
	GetByID(id int64) (*User, error)
	GetByUsername(username string) (*User, error)
	Remove(id int64) error
}

func NewUserStore(db *sql.DB) UserStore {
	return &UserDB{DB: db}
}

type UserDB struct {
	DB *sql.DB
}

// Add adds a user to the database if the username and password are not empty.
func (u *UserDB) Add(user *User) error {
	if user.Username == "" || user.Password.plaintext == nil {
		return fmt.Errorf("%w: username and password cannot be empty", ErrInvalidRequest)
	}
	if user.Role == "" {
		user.Role = "admin"
	}
	_, err := u.DB.Exec(`INSERT INTO "users" ("username", "password",role) VALUES ($1,$2,$3);`, user.Username, user.Password.hash, user.Role)
	return err
}

// GetByID returns a user from the database by ID
func (u *UserDB) GetByID(id int64) (*User, error) {
	var user User
	err := u.DB.QueryRow("SELECT id, username, password, role FROM users WHERE id = $1", id).Scan(&user.ID, &user.Username, &user.Password.hash, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: user id: %d", ErrNotFound, id)
		}
		return nil, err
	}
	return &user, err
}

// GetByUsername returns a user from the database by username
func (u *UserDB) GetByUsername(username string) (*User, error) {
	var user User
	err := u.DB.QueryRow("SELECT id, username, password, role FROM users WHERE username = $1", username).Scan(&user.ID, &user.Username, &user.Password.hash, &user.Role)
	if err != nil {
		return nil, err
	}
	return &user, err
}

// Remove find the user by ID and removes it from the database
func (u *UserDB) Remove(id int64) error {
	result, err := u.DB.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("%w: user id: %d", ErrNotFound, id)
	}
	return nil
}
