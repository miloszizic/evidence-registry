package data

import (
	"database/sql"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Stores struct {
	UserDB     UserDB
	CaseDB     CaseDB
	EvidenceDB EvidenceDB
	CaseFS     CaseFS
	EvidenceFS EvidenceFS
}

// NewStores creates a new Stores object
func NewStores(db *sql.DB, client *minio.Client) Stores {
	return Stores{
		UserDB:     UserDB{DB: db},
		CaseDB:     CaseDB{DB: db},
		EvidenceDB: EvidenceDB{DB: db},
		CaseFS:     CaseFS{Minio: client},
		EvidenceFS: EvidenceFS{Minio: client},
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
