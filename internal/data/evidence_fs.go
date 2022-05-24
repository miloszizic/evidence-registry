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

type EvidenceFS struct {
	Minio *minio.Client
}

// Create adds a new evidence to the storeFS and returns a SHA256 hash of that file, Evidence name should be unique within case and
// must not contain forward slash
func (e *EvidenceFS) Create(evidence *Evidence, caseName string, file io.Reader) (string, error) {
	if strings.Contains(caseName, "/") {
		return "", errors.New("forward slash is not allowed as case name")
	}
	evidenceCh := e.Minio.ListObjects(context.Background(), caseName, minio.ListObjectsOptions{})
	for object := range evidenceCh {
		if object.Key == evidence.Name {
			return "", errors.New("evidence already exists")
		}
	}
	h := sha256.New()
	putFile := io.TeeReader(file, h)

	_, err := e.Minio.PutObject(context.Background(), caseName, evidence.Name, putFile, -1, minio.PutObjectOptions{})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Remove removes an evidence from specific case and the storeFS
func (e *EvidenceFS) Remove(evidence *Evidence, caseName string) error {
	err := e.Minio.RemoveObject(context.Background(), caseName, evidence.Name, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

// List returns a list of evidence in the FS
func (e *EvidenceFS) List(caseName string) ([]Evidence, error) {
	var evidence []Evidence
	objects := e.Minio.ListObjects(context.Background(), caseName, minio.ListObjectsOptions{})
	for object := range objects {
		if object.Err != nil {
			return evidence, object.Err
		}
		evidence = append(evidence, Evidence{Name: object.Key})
	}
	return evidence, nil
}
func (e *EvidenceFS) Get(caseName string, evidenceName string) (io.ReadCloser, error) {
	object, err := e.Minio.GetObject(context.Background(), caseName, evidenceName, minio.GetObjectOptions{})
	_, err = object.Stat()
	if err != nil {
		return nil, err
	}
	return object, nil
}
