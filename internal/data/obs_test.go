package data_test

import (
	"bytes"
	"github.com/miloszizic/der/internal/data"
	"io"
	"strings"
	"testing"
)

func TestCreateCaseInOBS(t *testing.T) {
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
			err = store.ObjectStore.CreateCase(tt.addCase)
			got := err != nil
			if got != tt.wantErr {
				t.Errorf("expected %v, got %v", tt.wantErr, got)
			}
		})
	}
}
func TestRemoveCaseInOBS(t *testing.T) {
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
			err = store.ObjectStore.CreateCase(tt.addCase)
			if err != nil {
				t.Errorf("failed to add case: %v", err)
			}
			err = store.ObjectStore.RemoveCase(tt.deleteCase)
			got := err != nil
			if got != tt.wantErr {
				t.Errorf("expected %v, got %v", tt.wantErr, got)
			}
		})
	}
}
func TestCreateCaseInOBSReturnedErrorBecauseCaseAlreadyExists(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	err = store.ObjectStore.CreateCase(&data.Case{
		Name: "test",
	})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	err = store.ObjectStore.CreateCase(&data.Case{
		Name: "test",
	})
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}
func TestCreateEvidenceInOBS(t *testing.T) {
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
			err = store.ObjectStore.CreateCase(&data.Case{
				Name: tt.caseName,
			})
			if err != nil {
				t.Errorf("failed to add case: %v", err)
			}
			_, err = store.ObjectStore.CreateEvidence(tt.addEvidence, tt.caseName, tt.File)
			got := err != nil
			if got != tt.wantErr {
				t.Errorf("expected %v, got %v", tt.wantErr, got)
			}
		})
	}
}
func TestGetEvidenceInOBSSuccessful(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	testCase := &data.Case{
		Name: "test",
	}
	testEvidence := &data.Evidence{
		Name: "test",
		File: bytes.NewBufferString("sample"),
	}
	err = store.ObjectStore.CreateCase(testCase)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}

	_, err = store.ObjectStore.CreateEvidence(testEvidence, testCase.Name, testEvidence.File)
	if err != nil {
		t.Errorf("failed to add evidence: %v", err)
	}
	got, err := store.ObjectStore.GetEvidence(testCase.Name, testEvidence.Name)
	if err != nil {
		t.Errorf("failed to get evidence: %v", err)
	}
	buf := new(strings.Builder)
	_, err = io.Copy(buf, got)
	if err != nil {
		t.Errorf("failed to copy evidence: %v", err)
	}
	if buf.String() != "sample" {
		t.Errorf("expected %v, got %v", "sample", buf.String())
	}

}
func TestCreateEvidenceWithDifferentFilesGeneratedDifferentHashes(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	err = store.ObjectStore.CreateCase(&data.Case{
		Name: "test",
	})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	hash1, err := store.ObjectStore.CreateEvidence(&data.Evidence{
		Name: "test1",
	}, "test", bytes.NewBufferString("ss"))
	if err != nil {
		t.Errorf("failed to add evidence: %v", err)
	}
	hash2, err := store.ObjectStore.CreateEvidence(&data.Evidence{
		Name: "test2",
	}, "test", bytes.NewBufferString("ssss"))
	if err != nil {
		t.Errorf("failed to add evidence: %v", err)
	}
	if hash1 == hash2 {
		t.Errorf("expected different hashes, got same")
	}

}
func TestRemoveEvidenceInOBSSuccessful(t *testing.T) {
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
	err = store.ObjectStore.CreateCase(testCase)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}

	_, err = store.ObjectStore.CreateEvidence(testEvidence, testCase.Name, testEvidence.File)
	if err != nil {
		t.Errorf("failed to add evidence: %v", err)
	}
	err = store.ObjectStore.RemoveEvidence(testEvidence, testCase.Name)
	if err != nil {
		t.Errorf("failed to remove evidence: %v", err)
	}
}
func TestListCasesInOBSReturnedAllCasesInOBS(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	err = store.ObjectStore.CreateCase(&data.Case{
		Name: "test",
	})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	err = store.ObjectStore.CreateCase(&data.Case{
		Name: "test2",
	})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	cases, err := store.ObjectStore.ListCases()
	if err != nil {
		t.Errorf("failed to list all cases: %v", err)
	}
	if len(cases) != 2 {
		t.Errorf("expected 2 cases, got %v", len(cases))
	}
}
func TestEvidenceExistsInOBSReturns(t *testing.T) {
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
			err = store.ObjectStore.CreateCase(&data.Case{
				Name: "test",
			})
			if err != nil {
				t.Errorf("failed to add case: %v", err)
			}
			_, err = store.ObjectStore.CreateEvidence(&data.Evidence{
				Name: "test",
			}, tt.caseName, bytes.NewBufferString("s"))
			if err != nil {
				t.Errorf("failed to add evidence: %v", err)
			}
			got, err := store.ObjectStore.EvidenceExists(tt.caseName, tt.evidenceName)
			if err != nil {
				t.Errorf("failed to check evidence exists: %v", err)
			}
			if got != tt.wantExists {
				t.Errorf("expected %v, got %v", tt.wantExists, got)
			}
		})
	}
}
func TestListEvidencesReturnedAllEvidencesInOBS(t *testing.T) {
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
	err = store.ObjectStore.CreateCase(testCase)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	for _, e := range want {
		_, err = store.ObjectStore.CreateEvidence(&e, testCase.Name, e.File)
		if err != nil {
			t.Errorf("failed to add evidence: %v", err)
		}
	}
	evidence, err := store.ObjectStore.ListEvidences(testCase.Name)
	if err != nil {
		t.Errorf("failed to list all evidence: %v", err)
	}
	if len(evidence) != 2 {
		t.Errorf("expected 2 evidence, got %v", len(evidence))
	}
}
