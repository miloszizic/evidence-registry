package storage

import (
	"context"
	"errors"
	"evidence/internal/data/database"
	"github.com/minio/minio-go/v7"
)

type CaseFS struct {
	Minio *minio.Client
}

// Create adds a new case to the storeFS, Case name should be unique and must within the
// following rules:
// Names must be between 3 and 63 characters long.
// Names can consist only of lowercase letters, numbers, dots (.), and hyphens (-).
// Names must begin and end with a letter or number.
func (c *CaseFS) Create(cs *database.Case) error {
	exists, err := c.Minio.BucketExists(context.Background(), cs.Name)
	if exists {
		return errors.New("case already exists")
	}
	if err != nil {
		return err
	}
	err = c.Minio.MakeBucket(context.Background(), cs.Name, minio.MakeBucketOptions{})
	if err != nil {
		return err
	}
	return nil
}

// Remove removes a case from the storeFS
func (c *CaseFS) Remove(name string) error {
	exists, err := c.Minio.BucketExists(context.Background(), name)
	if !exists {
		return errors.New("case does not exist")
	}
	if err != nil {
		return err
	}
	err = c.Minio.RemoveBucket(context.Background(), name)
	if err != nil {
		return err
	}
	return nil
}

// List returns a list of cases in the storeFS
func (c *CaseFS) List() ([]database.Case, error) {
	var cases []database.Case
	buckets, err := c.Minio.ListBuckets(context.Background())
	if err != nil {
		return cases, err
	}
	for _, bucket := range buckets {
		cases = append(cases, database.Case{Name: bucket.Name})
	}
	return cases, nil
}
