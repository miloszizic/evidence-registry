package vault

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/miloszizic/der/db"
	"github.com/minio/minio-go/v7"
)

type CreateEvidenceRequest struct {
	CaseID      int32
	AppUserID   int32
	Name        string
	Description string
}

// CreateEvidence adds a new evidence to the storeFS and returns a SHA256 hash of that file, Evidence name should be unique within case and
// must not contain forward slash
func (f *FS) CreateEvidence(ctx context.Context, evName string, caseName string, file io.Reader) (string, error) {
	// check if caseName contains forward slash
	if strings.Contains(evName, "/") || strings.Contains(evName, " ") {
		return "", fmt.Errorf("%w : evidence can't contain forward slash or space : %q ", ErrInvalidRequest, evName)
	}
	if file == nil {
		return "", fmt.Errorf("%w : file can't be nil ", ErrInvalidRequest)
	}
	h := sha256.New()
	putFile := io.TeeReader(file, h)

	_, err := f.Minio.PutObject(ctx, caseName, evName, putFile, -1, minio.PutObjectOptions{})
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// EvidenceExists checks if an evidence exists in the storeFS using Case Name and Evidence name
func (f *FS) EvidenceExists(ctx context.Context, caseName string, evidenceName string) (bool, error) {
	_, err := f.Minio.StatObject(ctx, caseName, evidenceName, minio.StatObjectOptions{})
	if err != nil {
		if err.Error() == "The specified key does not exist." {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// RemoveEvidence removes an evidence from specific case and the storeFS
func (f *FS) RemoveEvidence(ctx context.Context, evName string, caseName string) error {
	err := f.Minio.RemoveObject(ctx, caseName, evName, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

// ListEvidences returns a list of evidence in the FS
func (f *FS) ListEvidences(ctx context.Context, caseName string) ([]db.Evidence, error) {
	var evidence []db.Evidence
	objects := f.Minio.ListObjects(ctx, caseName, minio.ListObjectsOptions{})
	for object := range objects {
		if object.Err != nil {
			return evidence, object.Err
		}
		evidence = append(evidence, db.Evidence{Name: object.Key})
	}
	return evidence, nil
}

// GetEvidence returns an evidence in the FS using Case Name and Evidence name
func (f *FS) GetEvidence(ctx context.Context, caseName string, evidenceName string) (io.ReadCloser, error) {
	object, err := f.Minio.GetObject(ctx, caseName, evidenceName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	_, err = object.Stat()
	if err != nil {
		if err.Error() == "The specified key does not exist." {
			return nil, fmt.Errorf("%w : evidence : %q not found", ErrNotFound, evidenceName)
		}
		return nil, err
	}
	return object, nil
}
