//go:build integration

package service_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/miloszizic/der/service"
)

func TestCreateEvidenceWasSuccessful(t *testing.T) {
	// get test stores with an existing case and user
	stores, createdUser, createdCase, err := service.NewEvidenceTestingStores(t)
	if err != nil {
		t.Fatalf("Error getting evidence test server: %v", err)
	}

	evidenceTypeID, err := stores.DBStore.GetEvidenceIDByType(context.Background(), "Initial Evidence")
	if err != nil {
		t.Fatalf("Error getting EvidenceTypeID: %v", err)
	}

	ev := service.CreateEvidenceParams{
		// Populate your evidence fields here, make sure to not set the ID, CreatedAt, and UpdatedAt as they will be set by the DB.
		Name:           "TestEvidence",
		Description:    "This is a test",
		CaseID:         createdCase.ID,
		AppUserID:      createdUser.ID,
		EvidenceTypeID: evidenceTypeID,
	}

	buffer := bytes.NewBufferString("test")

	_, err = stores.CreateEvidence(context.Background(), ev, buffer)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateEvidenceDBError(t *testing.T) {
	// get test stores with existing case and user
	stores, _, createdCase, err := service.NewEvidenceTestingStores(t)
	if err != nil {
		t.Fatalf("Error getting evidence test server: %v", err)
	}
	fakeID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	ev := service.CreateEvidenceParams{
		Name:        "TestEvidence",
		Description: "This is a test",
		CaseID:      createdCase.ID,
		AppUserID:   fakeID, // This user will fail CreateEvidence because he doesn't exist in the DB'
	}

	buffer := bytes.NewBufferString("test")

	_, err = stores.CreateEvidence(context.Background(), ev, buffer)
	if err == nil {
		t.Fatal("Expected an error but got none")
	}
	// Check the error message to ensure it's the one you expect
	if !strings.Contains(err.Error(), "error creating evidence in DB") {
		t.Errorf("Expected DB error, but got: %v", err)
	}
}

func TestGetEvidenceByID(t *testing.T) {
	// get test stores with existing case and user
	stores, createdUser, createdCase, err := service.NewEvidenceTestingStores(t)
	if err != nil {
		t.Fatalf("Error getting evidence test server: %v", err)
	}
	evidenceTypeID, err := stores.DBStore.GetEvidenceIDByType(context.Background(), "Initial Evidence")
	if err != nil {
		t.Fatalf("Error getting EvidenceTypeID: %v", err)
	}
	ev := service.CreateEvidenceParams{
		Name:           "TestEvidence",
		Description:    "This is a test",
		CaseID:         createdCase.ID,
		AppUserID:      createdUser.ID,
		EvidenceTypeID: evidenceTypeID,
	}
	buffer := bytes.NewBufferString("test")
	createdEV, err := stores.CreateEvidence(context.Background(), ev, buffer)
	if err != nil {
		t.Fatal(err)
	}
	got, err := stores.DBStore.GetEvidence(context.Background(), createdEV.ID)
	if err != nil {
		t.Fatalf("Error getting EvidenceID: %v", err)
	}
	if got.ID != createdEV.ID {
		t.Fatalf("Expected ID %v, got %v", createdEV.ID, got.ID)
	}
}

func TestGetEvidenceByIDFailedWithFakeID(t *testing.T) {
	// get test stores with existing case and user
	stores, createdUser, createdCase, err := service.NewEvidenceTestingStores(t)
	if err != nil {
		t.Fatalf("Error getting evidence test server: %v", err)
	}
	fakeID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	evidenceTypeID, err := stores.DBStore.GetEvidenceIDByType(context.Background(), "Initial Evidence")
	if err != nil {
		t.Fatalf("Error getting EvidenceTypeID: %v", err)
	}
	ev := service.CreateEvidenceParams{
		Name:           "TestEvidence",
		Description:    "This is a test",
		CaseID:         createdCase.ID,
		AppUserID:      createdUser.ID,
		EvidenceTypeID: evidenceTypeID,
	}
	buffer := bytes.NewBufferString("test")
	_, err = stores.CreateEvidence(context.Background(), ev, buffer)
	if err != nil {
		t.Fatal(err)
	}
	_, err = stores.DBStore.GetEvidence(context.Background(), fakeID)
	if err == nil {
		t.Fatal("Expected an error but got none")
	}
}

func TestDownloadEvidence(t *testing.T) {
	// get test stores with existing a case and user
	stores, createdUser, createdCase, err := service.NewEvidenceTestingStores(t)
	if err != nil {
		t.Fatalf("Error getting evidence test server: %v", err)
	}
	// get specific evidence type ID to create evidence
	evidenceTypeID, err := stores.DBStore.GetEvidenceIDByType(context.Background(), "Initial Evidence")
	if err != nil {
		t.Fatalf("Error getting EvidenceTypeID: %v", err)
	}

	ev := service.CreateEvidenceParams{
		// Populate your evidence fields here, make sure to not set the ID, CreatedAt, and UpdatedAt as they will be set by the DB.
		Name:           "TestEvidence",
		Description:    "This is a test",
		CaseID:         createdCase.ID,
		AppUserID:      createdUser.ID,
		EvidenceTypeID: evidenceTypeID,
	}

	buffer := bytes.NewBufferString("test")

	createdEV, err := stores.CreateEvidence(context.Background(), ev, buffer)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		ev      service.Evidence
		wantErr bool
	}{
		{
			name: "successful download",
			ev: service.Evidence{
				ID:     createdEV.ID,
				CaseID: createdCase.ID,
				Name:   "TestEvidence", //
			},
			wantErr: false,
		},
		{
			name: "failure with non-existing evidence",
			ev: service.Evidence{
				ID:     uuid.MustParse("99999999-9999-9999-9999-999999999999"), // non-existing evidence
				CaseID: createdCase.ID,
				Name:   "TestEvidence1", // non-existing evidence name
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := stores.DownloadEvidence(context.Background(), tt.ev)
			if (err != nil) != tt.wantErr {
				t.Errorf("DownloadEvidence() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListEvidencesRetrievedAllEvidencesThatExistInDBAndODB(t *testing.T) {
	// get test stores with existing case and user
	stores, createdUser, createdCase, err := service.NewEvidenceTestingStores(t)
	if err != nil {
		t.Fatalf("Error getting evidence test server: %v", err)
	}
	// get specific evidence type ID to create an evidence
	evidenceTypeID, err := stores.DBStore.GetEvidenceIDByType(context.Background(), "Initial Evidence")
	if err != nil {
		t.Fatalf("Error getting EvidenceTypeID: %v", err)
	}

	// Create some evidences
	evidenceCount := 5
	for i := 0; i < evidenceCount; i++ {
		ev := service.CreateEvidenceParams{
			Name:           fmt.Sprintf("TestEvidence%d", i),
			Description:    fmt.Sprintf("This is a test %d", i),
			CaseID:         createdCase.ID,
			AppUserID:      createdUser.ID,
			EvidenceTypeID: evidenceTypeID,
		}

		buffer := bytes.NewBufferString("test")

		_, err = stores.CreateEvidence(context.Background(), ev, buffer)
		if err != nil {
			t.Fatalf("Error creating evidence: %v", err)
		}
	}

	// test ListEvidences
	evidences, err := stores.ListEvidences(context.Background(), createdCase)
	if err != nil {
		t.Fatalf("ListEvidences() error = %v", err)
	}

	if len(evidences) != evidenceCount {
		t.Errorf("ListEvidences() returned %d evidences, want %d", len(evidences), evidenceCount)
	}
}
