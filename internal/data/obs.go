package data

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/minio/minio-go/v7"
	"io"
	"strings"
)

// OBStore is object-base storage interface for storing and retrieving data from object storage
type OBStore interface {
	CreateCase(cs *Case) error
	RemoveCase(name string) error
	CaseExists(name string) (bool, error)
	ListCases() ([]Case, error)
	CreateEvidence(evidence *Evidence, caseName string, file io.Reader) (string, error)
	EvidenceExists(caseName string, evidenceName string) (bool, error)
	RemoveEvidence(evidence *Evidence, caseName string) error
	ListEvidences(caseName string) ([]Evidence, error)
	GetEvidence(caseName string, evidenceName string) (io.ReadCloser, error)
}

func NewOBS(minio *minio.Client) OBStore {
	return &FS{Minio: minio}
}

type FS struct {
	Minio *minio.Client
}

// CreateCase adds a new case to the storeFS, Case name should be unique and must within the
// following rules:
// Names must be between 3 and 63 characters long.
// Names can consist only of lowercase letters, numbers, dots (.), and hyphens (-).
// Names must begin and end with a letter or number.
func (f *FS) CreateCase(cs *Case) error {
	exists, err := f.Minio.BucketExists(context.Background(), cs.Name)
	if exists {
		return NewErrorf(ErrCodeConflict, "caseFS: case already exists")
	}
	if err != nil {
		return err
	}
	err = f.Minio.MakeBucket(context.Background(), cs.Name, minio.MakeBucketOptions{})
	if err != nil {
		return err
	}
	return nil
}

// RemoveCase removes a case from the storeFS
func (f *FS) RemoveCase(name string) error {
	err := f.Minio.RemoveBucket(context.Background(), name)
	if err != nil {
		return err
	}
	return nil
}
func (f *FS) CaseExists(name string) (bool, error) {
	exists, err := f.Minio.BucketExists(context.Background(), name)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// ListCases returns a list of cases in the FS
func (f *FS) ListCases() ([]Case, error) {
	var cases []Case
	buckets, err := f.Minio.ListBuckets(context.Background())
	if err != nil {
		return cases, err
	}
	for _, bucket := range buckets {
		cases = append(cases, Case{Name: bucket.Name})
	}
	return cases, nil
}

// CreateEvidence adds a new evidence to the storeFS and returns a SHA256 hash of that file, Evidence name should be unique within case and
// must not contain forward slash
func (f *FS) CreateEvidence(evidence *Evidence, caseName string, file io.Reader) (string, error) {
	// check if caseName contains forward slash
	if strings.Contains(evidence.Name, "/") || strings.Contains(evidence.Name, " ") {
		return "", NewErrorf(ErrCodeInvalid, "OBStore: evidence name cannot contain forward slash or space")
	}
	if file == nil {
		return "", NewErrorf(ErrCodeInvalid, "FS.CreateEvidence: file cannot be nil")
	}
	h := sha256.New()
	putFile := io.TeeReader(file, h)

	_, err := f.Minio.PutObject(context.Background(), caseName, evidence.Name, putFile, -1, minio.PutObjectOptions{})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil

}
func (f *FS) EvidenceExists(caseName string, evidenceName string) (bool, error) {
	_, err := f.Minio.StatObject(context.Background(), caseName, evidenceName, minio.StatObjectOptions{})
	if err != nil {
		if err.Error() == "The specified key does not exist." {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// RemoveEvidence removes an evidence from specific case and the storeFS
func (f *FS) RemoveEvidence(evidence *Evidence, caseName string) error {
	err := f.Minio.RemoveObject(context.Background(), caseName, evidence.Name, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

// ListEvidences returns a list of evidence in the FS
func (f *FS) ListEvidences(caseName string) ([]Evidence, error) {
	var evidence []Evidence
	objects := f.Minio.ListObjects(context.Background(), caseName, minio.ListObjectsOptions{})
	for object := range objects {
		if object.Err != nil {
			return evidence, object.Err
		}
		evidence = append(evidence, Evidence{Name: object.Key})
	}
	return evidence, nil
}
func (f *FS) GetEvidence(caseName string, evidenceName string) (io.ReadCloser, error) {
	object, err := f.Minio.GetObject(context.Background(), caseName, evidenceName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return object, nil
}
