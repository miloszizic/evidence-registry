package service

import (
	"database/sql"
	"errors"

	"github.com/miloszizic/der/db"
	"github.com/miloszizic/der/vault"
	"github.com/minio/minio-go/v7"
)

// Errors returned by the service layer
var (
	// ErrNotFound returns when a resource is not found
	ErrNotFound = errors.New("resource not found")
	// ErrAlreadyExists returns when a resource already exists
	ErrAlreadyExists = errors.New("resource already exists")
	// ErrInvalidRequest returns when a request is invalid
	ErrInvalidRequest = errors.New("invalid request")
	// ErrUnauthorized returns when a user is not authorized
	ErrUnauthorized = errors.New("unauthorized")
	// ErrInvalidCredentials returns when a user has invalid credentials
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrMissingUser returns when a user is missing from the request context
	ErrMissingUser = errors.New("no user in request context")
)

// Stores is a collection of stores that can be used to access the database or object storage (minio)
type Stores struct {
	DB          *sql.DB
	DBStore     *db.Queries
	ObjectStore vault.ObjectStore
}

// NewStores creates a new Stores collection
func NewStores(dbs *sql.DB, client *minio.Client) Stores {
	return Stores{
		DB:          dbs,
		DBStore:     db.New(dbs),
		ObjectStore: vault.NewObjectStore(client),
	}
}
