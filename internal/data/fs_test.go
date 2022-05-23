package data_test

import (
	"bytes"
	"evidence/internal/data"
	"io"
	"testing"
)

func TestMinioCaseCreation(t *testing.T) {
	tests := []struct {
		name    string
		addCase *data.Case
		wantErr bool
	}{
		{
			name: "successful with just letters",
			addCase: &data.Case{
				Name: "test",
			},
			wantErr: false,
		},
		{
			name: "successful with just letters and numbers",
			addCase: &data.Case{
				Name: "test23test",
			},
			wantErr: false,
		},
		{
			name: "successful with just letters and supported special characters",
			addCase: &data.Case{
				Name: "test-test",
			},
			wantErr: false,
		},
		{
			name: "failed with an unsupported special character ",
			addCase: &data.Case{
				Name: "test/test",
			},
			wantErr: true,
		},
		{
			name: "failed with space ",
			addCase: &data.Case{
				Name: "test/test",
			},
			wantErr: true,
		},
		{
			name: "failed with no name ",
			addCase: &data.Case{
				Name: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := getTestStores(t)
			if err != nil {
				t.Errorf("failed to get test stores: %v", err)
			}
			err = store.StoreFS.AddCase(tt.addCase)
			got := err != nil
			if got != tt.wantErr {
				t.Errorf("expected %v, got %v", tt.wantErr, got)
			}
		})
	}
}
func TestMinioCaseDeletion(t *testing.T) {
	tests := []struct {
		name       string
		addCase    *data.Case
		deleteCase string
		wantErr    bool
	}{
		{
			name: "successful for existing case",
			addCase: &data.Case{
				Name: "test",
			},
			deleteCase: "test",
			wantErr:    false,
		},
		{
			name: "failed for non-existing case",
			addCase: &data.Case{
				Name: "test",
			},
			deleteCase: "test2",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := getTestStores(t)
			if err != nil {
				t.Errorf("failed to get test stores: %v", err)
			}
			err = store.StoreFS.AddCase(tt.addCase)
			if err != nil {
				t.Errorf("failed to add case: %v", err)
			}
			err = store.StoreFS.RemoveCase(tt.deleteCase)
			got := err != nil
			if got != tt.wantErr {
				t.Errorf("expected %v, got %v", tt.wantErr, got)
			}
		})
	}
}
func TestMinioFailedToAddCaseWithExistingName(t *testing.T) {
	store, err := getTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	err = store.StoreFS.AddCase(&data.Case{
		Name: "test",
	})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	err = store.StoreFS.AddCase(&data.Case{
		Name: "test",
	})
	got := err != nil
	if !got {
		t.Errorf("expected true, got false")
	}
}

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
			store, err := getTestStores(t)
			if err != nil {
				t.Errorf("failed to get test stores: %v", err)
			}
			err = store.StoreFS.AddCase(&data.Case{
				Name: tt.caseName,
			})
			if err != nil {
				t.Errorf("failed to add case: %v", err)
			}
			_, err = store.StoreFS.AddEvidence(tt.addEvidence, tt.caseName, tt.fileForUpload)
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
	err = store.StoreFS.AddCase(&data.Case{
		Name: "test",
	})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	hash1, err := store.StoreFS.AddEvidence(&data.Evidence{
		Name: "test1",
	}, "test", bytes.NewBufferString("ss"))
	if err != nil {
		t.Errorf("failed to add evidence: %v", err)
	}
	hash2, err := store.StoreFS.AddEvidence(&data.Evidence{
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
	store, err := getTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	err = store.StoreFS.AddCase(&data.Case{
		Name: "test",
	})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	_, err = store.StoreFS.AddEvidence(&data.Evidence{
		Name: "test",
	}, "test", bytes.NewBufferString("s"))
	_, err = store.StoreFS.AddEvidence(&data.Evidence{
		Name: "test",
	}, "test", bytes.NewBufferString("s"))
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}
func TestDeleteMinioEvidence(t *testing.T) {
	store, err := getTestStores(t)
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
	err = store.StoreFS.AddCase(testCase)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}

	_, err = store.StoreFS.AddEvidence(testEvidence, testCase.Name, testEvidence.File)
	if err != nil {
		t.Errorf("failed to add evidence: %v", err)
	}
	err = store.StoreFS.RemoveEvidence(testEvidence, testCase.Name)
	if err != nil {
		t.Errorf("failed to remove evidence: %v", err)
	}
}
func TestListAllCases(t *testing.T) {
	store, err := getTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	err = store.StoreFS.AddCase(&data.Case{
		Name: "test",
	})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	err = store.StoreFS.AddCase(&data.Case{
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
	store, err := getTestStores(t)
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
	err = store.StoreFS.AddCase(testCase)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	for _, e := range want {
		_, err = store.StoreFS.AddEvidence(&e, testCase.Name, e.File)
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
	store, err := getTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	testCase := &data.Case{
		Name: "test",
	}
	err = store.StoreFS.AddCase(testCase)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	ev, err := store.StoreFS.GetEvidence(testCase.Name, "something")
	if ev != nil {
		t.Errorf("expected nil, got %v", ev)
	}
}
