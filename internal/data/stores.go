package data

import (
	"database/sql"
	"errors"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Stores struct {
	UserDB     UserModel
	CaseDB     CaseModel
	EvidenceDB EvidenceModel
	StoreFS    StoreFS
}

// NewStores creates a new Stores object
func NewStores(db *sql.DB, client *minio.Client) Stores {
	return Stores{
		UserDB:     UserModel{DB: db},
		CaseDB:     CaseModel{DB: db},
		EvidenceDB: EvidenceModel{DB: db},
		StoreFS:    StoreFS{Minio: client},
	}
}

// FromPostgresDB opens a connection to a Postgres database.
func FromPostgresDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	return db, nil
}

// FromMinio creates a new Minio client.
func FromMinio(endpoint, accessKeyID, secretAccessKey string) (*minio.Client, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
	})
	if err != nil {
		return nil, err
	}
	return minioClient, nil
}
