package vault

import (
	"context"
	"fmt"

	"github.com/miloszizic/der/db"
	"github.com/minio/minio-go/v7"
)

// CreateCase adds a new case to the storeFS, Case name should be unique and must within the
// following rules:
// Names must be between 3 and 63 characters long.
// Names can consist only of lowercase letters, numbers, dots (.), and hyphens (-).
// Names must begin and end with a letter or number.
func (f *FS) CreateCase(ctx context.Context, cs db.CreateCaseParams) error {
	exists, err := f.Minio.BucketExists(ctx, cs.Name)
	if exists {
		return fmt.Errorf("%w : case : %q", ErrAlreadyExists, cs.Name)
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
func (f *FS) RemoveCase(ctx context.Context, name string) error {
	err := f.Minio.RemoveBucket(ctx, name)
	if err != nil {
		return err
	}

	return nil
}

func (f *FS) CaseExists(ctx context.Context, name string) (bool, error) {
	exists, err := f.Minio.BucketExists(ctx, name)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// ListCases returns a list of cases in the FS
func (f *FS) ListCases(ctx context.Context) ([]db.Case, error) {
	var cases []db.Case
	buckets, err := f.Minio.ListBuckets(ctx)
	if err != nil {
		return cases, err
	}
	for _, bucket := range buckets {
		cases = append(cases, db.Case{Name: bucket.Name})
	}

	return cases, nil
}

// RenameCase will copy all files from the old case to the new case and remove the old case
// this is useful when we want to rename a case and minio doesn't support renaming buckets
// WARNING: this is a very expensive and slow (if case is big )operation and should be used
// only when necessary because it will copy all files from the old case to the new case and
// remove the old case

func (f *FS) RenameCase(ctx context.Context, oldCase, newCase string) error {
	// check if the old case exists
	exists, err := f.Minio.BucketExists(ctx, oldCase)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%w : case : %q", ErrNotFound, oldCase)
	}
	// check if the new case exists
	exists, err = f.Minio.BucketExists(ctx, newCase)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%w : case : %q", ErrAlreadyExists, newCase)
	}
	// create the new case
	err = f.Minio.MakeBucket(ctx, newCase, minio.MakeBucketOptions{})
	if err != nil {
		return err
	}
	// copy all files from the old case to the new case
	objects := f.Minio.ListObjects(ctx, oldCase, minio.ListObjectsOptions{Recursive: true})
	for object := range objects {
		// create source options
		srcOpts := minio.CopySrcOptions{
			Bucket: oldCase,
			Object: object.Key,
		}

		// create destination options
		dstOpts := minio.CopyDestOptions{
			Bucket: newCase,
			Object: object.Key,
		}

		// copy object to new case
		_, err := f.Minio.CopyObject(ctx, dstOpts, srcOpts)
		if err != nil {
			return fmt.Errorf("failed to copy object %q to new case %q: %w", object.Key, newCase, err)
		}
	}

	// remove the old case after all files have been copied
	err = f.Minio.RemoveBucket(ctx, oldCase)
	if err != nil {
		return fmt.Errorf("failed to remove old case %q: %w", oldCase, err)
	}
	return nil
}
