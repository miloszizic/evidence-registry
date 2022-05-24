package data_test

import (
	"bytes"
	"evidence/internal/data"
	"io"
	"testing"
)

func TestMinioMakeEvidence(t *testing.T) {
	tests := []struct {
		name          string
		addEvidence   *data.Evidence
		caseName      string
		fileForUpload io.Reader
		wantErr       bool
	}{
		{
			name: "successful with just letters",
			addEvidence: &data.Evidence{
				Name: "test",
			},
			caseName:      "test",
			fileForUpload: bytes.NewBufferString("s"),
			wantErr:       false,
		},
		{
			name: "successful with just letters and numbers",
			addEvidence: &data.Evidence{
				Name: "test23test",
			},
			caseName:      "test",
			fileForUpload: bytes.NewBufferString("s"),
			wantErr:       false,
		},
		{
			name: "successful with just letters and supported special characters",
			addEvidence: &data.Evidence{
				Name: "test-test",
			},
			caseName:      "test",
			fileForUpload: bytes.NewBufferString("s"),
			wantErr:       false,
		},
		{
			name: "failed with space ",
			addEvidence: &data.Evidence{
				Name: "test test",
			},
			caseName:      "test",
			fileForUpload: bytes.NewBufferString("s"),
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := GetTestStores(t)
			if err != nil {
				t.Errorf("failed to get test stores: %v", err)
			}
			err = store.CaseFS.Create(&data.Case{
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
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	err = store.CaseFS.Create(&data.Case{
		Name: "test",
	})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	hash1, err := store.EvidenceFS.Create(&data.Evidence{
		Name: "test1",
	}, "test", bytes.NewBufferString("ss"))
	if err != nil {
		t.Errorf("failed to add evidence: %v", err)
	}
	hash2, err := store.EvidenceFS.Create(&data.Evidence{
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
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	err = store.CaseFS.Create(&data.Case{
		Name: "test",
	})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	_, err = store.EvidenceFS.Create(&data.Evidence{
		Name: "test",
	}, "test", bytes.NewBufferString("s"))
	_, err = store.EvidenceFS.Create(&data.Evidence{
		Name: "test",
	}, "test", bytes.NewBufferString("s"))
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}
func TestDeleteMinioEvidence(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	testCase := &data.Case{
		Name: "test",
	}
	testEvidence := &data.Evidence{
		Name: "test",
		File: bytes.NewBufferString("s"),
	}
	err = store.CaseFS.Create(testCase)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}

	_, err = store.EvidenceFS.Create(testEvidence, testCase.Name, testEvidence.File)
	if err != nil {
		t.Errorf("failed to add evidence: %v", err)
	}
	err = store.EvidenceFS.Remove(testEvidence, testCase.Name)
	if err != nil {
		t.Errorf("failed to remove evidence: %v", err)
	}
}
func TestListAllCases(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	err = store.CaseFS.Create(&data.Case{
		Name: "test",
	})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	err = store.CaseFS.Create(&data.Case{
		Name: "test2",
	})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	cases, err := store.CaseFS.List()
	if err != nil {
		t.Errorf("failed to list all cases: %v", err)
	}
	if len(cases) != 2 {
		t.Errorf("expected 2 cases, got %v", len(cases))
	}
}
func TestListAllEvidence(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	testCase := &data.Case{
		Name: "test",
	}
	want := []data.Evidence{
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
		_, err = store.EvidenceFS.Create(&e, testCase.Name, e.File)
		if err != nil {
			t.Errorf("failed to add evidence: %v", err)
		}
	}
	evidence, err := store.EvidenceFS.List(testCase.Name)
	if err != nil {
		t.Errorf("failed to list all evidence: %v", err)
	}
	if len(evidence) != 2 {
		t.Errorf("expected 2 evidence, got %v", len(evidence))
	}
}
func TestDownloadingNonExistedEvidenceFailed(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	testCase := &data.Case{
		Name: "test",
	}
	err = store.CaseFS.Create(testCase)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	_, err = store.EvidenceFS.Get(testCase.Name, "something")
	if err.Error() != "The specified key does not exist." {
		t.Errorf("expected error, got %v", err)
	}
}
