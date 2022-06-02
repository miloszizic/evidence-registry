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
				Name: "test test",
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
			store, err := GetTestStores(t)
			if err != nil {
				t.Errorf("failed to get test stores: %v", err)
			}
			err = store.OBStore.CreateCase(tt.addCase)
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
			store, err := GetTestStores(t)
			if err != nil {
				t.Errorf("failed to get test stores: %v", err)
			}
			err = store.OBStore.CreateCase(tt.addCase)
			if err != nil {
				t.Errorf("failed to add case: %v", err)
			}
			err = store.OBStore.RemoveCase(tt.deleteCase)
			got := err != nil
			if got != tt.wantErr {
				t.Errorf("expected %v, got %v", tt.wantErr, got)
			}
		})
	}
}
func TestMinioFailedToAddCaseWithExistingName(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	err = store.OBStore.CreateCase(&data.Case{
		Name: "test",
	})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	err = store.OBStore.CreateCase(&data.Case{
		Name: "test",
	})
	got := err != nil
	if !got {
		t.Errorf("expected true, got false")
	}
}
func TestOBSCreateEvidence(t *testing.T) {
	tests := []struct {
		name        string
		addEvidence *data.Evidence
		caseName    string
		File        io.Reader
		wantErr     bool
	}{
		{
			name: "successful with just letters",
			addEvidence: &data.Evidence{
				Name: "test",
			},
			caseName: "testcase",
			File:     bytes.NewBufferString("s"),
			wantErr:  false,
		},
		{
			name: "successful with just letters and numbers",
			addEvidence: &data.Evidence{
				Name: "test23test",
			},
			caseName: "testcase",
			File:     bytes.NewBufferString("s"),
			wantErr:  false,
		},
		{
			name: "unsuccessful with slash",
			addEvidence: &data.Evidence{
				Name: "test/test",
			},
			caseName: "testcase",
			File:     bytes.NewBufferString("s"),
			wantErr:  true,
		},
		{
			name: "unsuccessful with no file",
			addEvidence: &data.Evidence{
				Name: "test/test",
			},
			caseName: "testcase",
			File:     nil,
			wantErr:  true,
		},
		{
			name: "unsuccessful with space",
			addEvidence: &data.Evidence{
				Name: "test test",
			},
			caseName: "testcase",
			File:     bytes.NewBufferString("s"),
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := GetTestStores(t)
			if err != nil {
				t.Errorf("failed to get test stores: %v", err)
			}
			err = store.OBStore.CreateCase(&data.Case{
				Name: tt.caseName,
			})
			if err != nil {
				t.Errorf("failed to add case: %v", err)
			}
			_, err = store.OBStore.CreateEvidence(tt.addEvidence, tt.caseName, tt.File)
			got := err != nil
			if got != tt.wantErr {
				t.Errorf("expected %v, got %v", tt.wantErr, got)
			}
		})
	}
}
func TestRetrieveExistingEvidenceSucceeds(t *testing.T) {
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
	err = store.OBStore.CreateCase(testCase)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}

	_, err = store.OBStore.CreateEvidence(testEvidence, testCase.Name, testEvidence.File)
	if err != nil {
		t.Errorf("failed to add evidence: %v", err)
	}
	_, err = store.OBStore.GetEvidence(testCase.Name, testEvidence.Name)
	if err != nil {
		t.Errorf("failed to retrieve evidence: %v", err)
	}

}
func TestCreatingDifferentEvidenceShouldDifferentiatedHash(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	err = store.OBStore.CreateCase(&data.Case{
		Name: "test",
	})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	hash1, err := store.OBStore.CreateEvidence(&data.Evidence{
		Name: "test1",
	}, "test", bytes.NewBufferString("ss"))
	if err != nil {
		t.Errorf("failed to add evidence: %v", err)
	}
	hash2, err := store.OBStore.CreateEvidence(&data.Evidence{
		Name: "test2",
	}, "test", bytes.NewBufferString("ssss"))
	if err != nil {
		t.Errorf("failed to add evidence: %v", err)
	}
	if hash1 == hash2 {
		t.Errorf("expected different hashes, got same")
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
	err = store.OBStore.CreateCase(testCase)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}

	_, err = store.OBStore.CreateEvidence(testEvidence, testCase.Name, testEvidence.File)
	if err != nil {
		t.Errorf("failed to add evidence: %v", err)
	}
	err = store.OBStore.RemoveEvidence(testEvidence, testCase.Name)
	if err != nil {
		t.Errorf("failed to remove evidence: %v", err)
	}
}
func TestListAllCases(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	err = store.OBStore.CreateCase(&data.Case{
		Name: "test",
	})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	err = store.OBStore.CreateCase(&data.Case{
		Name: "test2",
	})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	cases, err := store.OBStore.ListCases()
	if err != nil {
		t.Errorf("failed to list all cases: %v", err)
	}
	if len(cases) != 2 {
		t.Errorf("expected 2 cases, got %v", len(cases))
	}
}
func TestEvidenceExistsReturns(t *testing.T) {
	tests := []struct {
		name         string
		caseName     string
		evidenceName string
		wantExists   bool
	}{
		{
			name:         "true when evidence exists",
			caseName:     "test",
			evidenceName: "test",
			wantExists:   true,
		},
		{
			name:         "false when evidence does not exist",
			caseName:     "test",
			evidenceName: "test1",
			wantExists:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := GetTestStores(t)
			if err != nil {
				t.Errorf("failed to get test stores: %v", err)
			}
			err = store.OBStore.CreateCase(&data.Case{
				Name: "test",
			})
			if err != nil {
				t.Errorf("failed to add case: %v", err)
			}
			_, err = store.OBStore.CreateEvidence(&data.Evidence{
				Name: "test",
			}, tt.caseName, bytes.NewBufferString("s"))
			if err != nil {
				t.Errorf("failed to add evidence: %v", err)
			}
			got, err := store.OBStore.EvidenceExists(tt.caseName, tt.evidenceName)
			if err != nil {
				t.Errorf("failed to check evidence exists: %v", err)
			}
			if got != tt.wantExists {
				t.Errorf("expected %v, got %v", tt.wantExists, got)
			}
		})
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
	err = store.OBStore.CreateCase(testCase)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	for _, e := range want {
		_, err = store.OBStore.CreateEvidence(&e, testCase.Name, e.File)
		if err != nil {
			t.Errorf("failed to add evidence: %v", err)
		}
	}
	evidence, err := store.OBStore.ListEvidences(testCase.Name)
	if err != nil {
		t.Errorf("failed to list all evidence: %v", err)
	}
	if len(evidence) != 2 {
		t.Errorf("expected 2 evidence, got %v", len(evidence))
	}
}
