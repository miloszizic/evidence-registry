package vault

import (
	"context"
	"errors"
	"io"

	"github.com/miloszizic/der/db"
	"github.com/minio/minio-go/v7"
)

var (
	ErrNotFound       = errors.New("resource not found")
	ErrAlreadyExists  = errors.New("resource already exists")
	ErrInvalidRequest = errors.New("invalid request")
)

// ObjectStore is object-base storage interface for storing and retrieving data from object storage
type ObjectStore interface {
	CreateCase(ctx context.Context, cs db.CreateCaseParams) error
	RemoveCase(ctx context.Context, name string) error
	CaseExists(ctx context.Context, name string) (bool, error)
	ListCases(ctx context.Context) ([]db.Case, error)
	CreateEvidence(ctx context.Context, evName string, caseName string, file io.Reader) (string, error)
	EvidenceExists(ctx context.Context, caseName string, evidenceName string) (bool, error)
	RemoveEvidence(ctx context.Context, evName string, caseName string) error
	ListEvidences(ctx context.Context, caseName string) ([]db.Evidence, error)
	GetEvidence(ctx context.Context, caseName string, evidenceName string) (io.ReadCloser, error)
}

func NewObjectStore(minio *minio.Client) ObjectStore {
	return &FS{Minio: minio}
}

type FS struct {
	Minio *minio.Client
}
