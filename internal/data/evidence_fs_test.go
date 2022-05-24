package storage_test

import (
	"bytes"
	database2 "evidence/internal/data/database"
	"evidence/internal/database"
	"io"
	"testing"
)

func TestMinioMakeEvidence(t *testing.T) {
	tests := []struct {
		name          string
		addEvidence   *database2.Evidence
		caseName      string
		fileForUpload io.Reader
		wantErr       bool
	}{
		{
			name: "successful with just letters",
			addEvidence: &database2.Evidence{
				Name: "test",
			},
			caseName:      "test",
			fileForUpload: bytes.NewBufferString("s"),
			wantErr:       false,
		},
		{
			name: "successful with just letters and numbers",
			addEvidence: &database2.Evidence{
				Name: "test23test",
			},
			caseName:      "test",
			fileForUpload: bytes.NewBufferString("s"),
			wantErr:       false,
		},
		{
			name: "successful with just letters and supported special characters",
			addEvidence: &database2.Evidence{
				Name: "test-test",
			},
			caseName:      "test",
			fileForUpload: bytes.NewBufferString("s"),
			wantErr:       false,
		},
		{
			name: "failed with space ",
			addEvidence: &database2.Evidence{
				Name: "test test",
			},
			caseName:      "test",
			fileForUpload: bytes.NewBufferString("s"),
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := getTestStores(t)
			if err != nil {
				t.Errorf("failed to get test stores: %v", err)
			}
			err = store.CaseFS.Create(&database2.Case{
				Name: tt.caseName,
			})
			if err != nil {
				t.Errorf("failed to add case: %v", err)
			}
			_, err = store.EvidenceFS.Create(tt.addEvidence, tt.caseName, tt.fileForUpload)
			got := err != nil
			if got != tt.wantErr {
				t.Errorf("expected %v, got %v", tt.wantErr, got)
			}
		})
	}
}
func TestCreatingDifferentEvidenceShouldDifferentiatedHash(t *testing.T) {
	store, err := getTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	err = store.CaseFS.Create(&database2.Case{
		Name: "test",
	})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	hash1, err := store.EvidenceFS.Create(&database2.Evidence{
		Name: "test1",
	}, "test", bytes.NewBufferString("ss"))
	if err != nil {
		t.Errorf("failed to add evidence: %v", err)
	}
	hash2, err := store.EvidenceFS.Create(&database2.Evidence{
		Name: "test2",
	}, "test", bytes.NewBufferString("ssss"))
	if err != nil {
		t.Errorf("failed to add evidence: %v", err)
	}
	if hash1 == hash2 {
		t.Errorf("expected different hashes, got same")
	}

}
func TestCreatingNewEvidenceWithExistingName(t *testing.T) {
	store, err := database.getTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	err = store.CaseFS.Create(&database2.Case{
		Name: "test",
	})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	_, err = store.StoreFS.CreateEvidence(&database2.Evidence{
		Name: "test",
	}, "test", bytes.NewBufferString("s"))
	_, err = store.StoreFS.CreateEvidence(&database2.Evidence{
		Name: "test",
	}, "test", bytes.NewBufferString("s"))
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}
func TestDeleteMinioEvidence(t *testing.T) {
	store, err := database.getTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	testCase := &database2.Case{
		Name: "test",
	}
	testEvidence := &database2.Evidence{
		Name: "test",
		File: bytes.NewBufferString("s"),
	}
	err = store.CaseFS.Create(testCase)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}

	_, err = store.StoreFS.CreateEvidence(testEvidence, testCase.Name, testEvidence.File)
	if err != nil {
		t.Errorf("failed to add evidence: %v", err)
	}
	err = store.StoreFS.RemoveEvidence(testEvidence, testCase.Name)
	if err != nil {
		t.Errorf("failed to remove evidence: %v", err)
	}
}
func TestListAllCases(t *testing.T) {
	store, err := database.getTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	err = store.CaseFS.Create(&database2.Case{
		Name: "test",
	})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	err = store.CaseFS.Create(&database2.Case{
		Name: "test2",
	})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	cases, err := store.StoreFS.ListCases()
	if err != nil {
		t.Errorf("failed to list all cases: %v", err)
	}
	if len(cases) != 2 {
		t.Errorf("expected 2 cases, got %v", len(cases))
	}
}
func TestListAllEvidence(t *testing.T) {
	store, err := database.getTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	testCase := &database2.Case{
		Name: "test",
	}
	want := []database2.Evidence{
		{
			Name: "test",
			File: bytes.NewBufferString("file1"),
		},
		{
			Name: "test2",
			File: bytes.NewBufferString("file2"),
		},
	}
	err = store.CaseFS.Create(testCase)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	for _, e := range want {
		_, err = store.StoreFS.CreateEvidence(&e, testCase.Name, e.File)
		if err != nil {
			t.Errorf("failed to add evidence: %v", err)
		}
	}
	evidence, err := store.StoreFS.ListEvidences(testCase.Name)
	if err != nil {
		t.Errorf("failed to list all evidence: %v", err)
	}
	if len(evidence) != 2 {
		t.Errorf("expected 2 evidence, got %v", len(evidence))
	}
}
func TestDownloadingNonExistedEvidenceFailed(t *testing.T) {
	store, err := database.getTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	testCase := &database2.Case{
		Name: "test",
	}
	err = store.CaseFS.Create(testCase)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	_, err = store.StoreFS.GetEvidence(testCase.Name, "something")
	if err.Error() != "The specified key does not exist." {
		t.Errorf("expected error, got %v", err)
	}
}
