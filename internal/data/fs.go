package data

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/minio/minio-go/v7"
	"io"
	"strings"
)

var (
	caseAlreadyExists     = errors.New("case already exists")
	caseDoesNotExist      = errors.New("case does not exist")
	evidenceAlreadyExists = errors.New("evidence already exists")
	forwardSlash          = errors.New("forward slash is not allowed as case name")
)

type StoreFS struct {
	Minio *minio.Client
}

// AddCase adds a new case to the storeFS, Case name should be unique and must within the
// following rules:
// Names must be between 3 and 63 characters long.
// Names can consist only of lowercase letters, numbers, dots (.), and hyphens (-).
// Names must begin and end with a letter or number.
func (s *StoreFS) AddCase(cs *Case) error {
	exists, err := s.Minio.BucketExists(context.Background(), cs.Name)
	if exists {
		return caseAlreadyExists
	}
	if err != nil {
		return err
	}
	err = s.Minio.MakeBucket(context.Background(), cs.Name, minio.MakeBucketOptions{})
	if err != nil {
		return err
	}
	return nil
}

// RemoveCase removes a case from the storeFS
func (s *StoreFS) RemoveCase(name string) error {
	exists, err := s.Minio.BucketExists(context.Background(), name)
	if !exists {
		return caseDoesNotExist
	}
	if err != nil {
		return err
	}
	err = s.Minio.RemoveBucket(context.Background(), name)
	if err != nil {
		return err
	}
	return nil
}

// AddEvidence adds a new evidence to the storeFS and returns a SHA256 hash of that file, Evidence name should be unique within case and
// must not contain forward slash
func (s *StoreFS) AddEvidence(evidence *Evidence, caseName string, file io.Reader) (string, error) {
	if strings.Contains(caseName, "/") {
		return "", forwardSlash
	}
	evidenceCh := s.Minio.ListObjects(context.Background(), caseName, minio.ListObjectsOptions{})
	for object := range evidenceCh {
		if object.Key == evidence.Name {
			return "", evidenceAlreadyExists
		}
	}
	h := sha256.New()
	putFile := io.TeeReader(file, h)

	_, err := s.Minio.PutObject(context.Background(), caseName, evidence.Name, putFile, -1, minio.PutObjectOptions{})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// RemoveEvidence removes an evidence from specific case and the storeFS
func (s *StoreFS) RemoveEvidence(evidence *Evidence, caseName string) error {
	err := s.Minio.RemoveObject(context.Background(), caseName, evidence.Name, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

// ListCases returns a list of cases in the storeFS
func (s *StoreFS) ListCases() ([]Case, error) {
	var cases []Case
	buckets, err := s.Minio.ListBuckets(context.Background())
	if err != nil {
		return cases, err
	}
	for _, bucket := range buckets {
		cases = append(cases, Case{Name: bucket.Name})
	}
	return cases, nil
}

// ListEvidences returns a list of evidence in the FS
func (s *StoreFS) ListEvidences(caseName string) ([]Evidence, error) {
	var evidence []Evidence
	objects := s.Minio.ListObjects(context.Background(), caseName, minio.ListObjectsOptions{})
	for object := range objects {
		if object.Err != nil {
			return evidence, object.Err
		}
		evidence = append(evidence, Evidence{Name: object.Key})
	}
	return evidence, nil
}
func (s *StoreFS) GetEvidence(caseName string, evidenceName string) (io.ReadCloser, error) {
	object, err := s.Minio.GetObject(context.Background(), caseName, evidenceName, minio.GetObjectOptions{})
	stat, err := object.Stat()
	if err != nil {
		return nil, err
	}
	if stat.Size == 0 {
		return nil, errors.New("evidence does not exist")
	}
	if err != nil {
		return nil, err
	}
	return object, nil
}
