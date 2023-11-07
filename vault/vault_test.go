package vault_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/miloszizic/der/db"
	"github.com/miloszizic/der/service"
	"github.com/miloszizic/der/vault"
)

func TestCreateCaseInOBS(t *testing.T) {
	tests := []struct {
		name       string
		caseParams db.CreateCaseParams
		wantErr    bool
	}{
		{
			name: "successful with just letters",
			caseParams: db.CreateCaseParams{
				Name: "test",
			},
			wantErr: false,
		},
		{
			name: "successful with just letters and numbers",
			caseParams: db.CreateCaseParams{
				Name: "test23test",
			},
			wantErr: false,
		},
		{
			name: "successful with just letters and supported special characters",
			caseParams: db.CreateCaseParams{
				Name: "test-test",
			},
			wantErr: false,
		},
		{
			name: "failed with an unsupported special character ",
			caseParams: db.CreateCaseParams{
				Name: "test/test",
			},
			wantErr: true,
		},
		{
			name: "failed with space ",
			caseParams: db.CreateCaseParams{
				Name: "test test",
			},
			wantErr: true,
		},
		{
			name: "failed with no name ",
			caseParams: db.CreateCaseParams{
				Name: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := service.GetTestStores(t)
			if err != nil {
				t.Errorf("failed to get test stores: %v", err)
			}
			err = store.ObjectStore.CreateCase(context.Background(), tt.caseParams)
			got := err != nil
			if got != tt.wantErr {
				t.Errorf("CreateCase() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateCaseInOBSReturnedErrorBecauseCaseAlreadyExists(t *testing.T) {
	store, err := service.GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	err = store.ObjectStore.CreateCase(context.Background(), db.CreateCaseParams{
		Name: "test",
	})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	err = store.ObjectStore.CreateCase(context.Background(), db.CreateCaseParams{
		Name: "test",
	})
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestCreateEvidenceInOBS(t *testing.T) {
	tests := []struct {
		name     string
		evName   string
		caseName string
		File     io.Reader
		wantErr  bool
	}{
		{
			name:     "successful with just letters",
			evName:   "test",
			caseName: "testcase",
			File:     bytes.NewBufferString("s"),
			wantErr:  false,
		},
		{
			name:     "successful with just letters and numbers",
			evName:   "test23test",
			caseName: "testcase",
			File:     bytes.NewBufferString("s"),
			wantErr:  false,
		},
		{
			name:     "unsuccessful with slash",
			evName:   "test/test",
			caseName: "testcase",
			File:     bytes.NewBufferString("s"),
			wantErr:  true,
		},
		{
			name:     "unsuccessful with no file",
			evName:   "test",
			caseName: "testcase",
			File:     nil,
			wantErr:  true,
		},
		{
			name:     "unsuccessful with space",
			evName:   "test test",
			caseName: "testcase",
			File:     bytes.NewBufferString("s"),
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := service.GetTestStores(t)
			if err != nil {
				t.Errorf("failed to get test stores: %v", err)
			}
			err = store.ObjectStore.CreateCase(context.Background(), db.CreateCaseParams{
				Name: tt.caseName,
			})
			if err != nil {
				t.Errorf("failed to add case: %v", err)
			}
			_, err = store.ObjectStore.CreateEvidence(context.Background(), tt.evName, tt.caseName, tt.File)
			got := err != nil
			if got != tt.wantErr {
				t.Errorf("expected %v, got %v", tt.wantErr, got)
			}
		})
	}
}

func TestGetEvidenceInOBSSuccessful(t *testing.T) {
	store, err := service.GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	testCase := db.CreateCaseParams{
		Name: "test",
	}
	testEvidenceName := "test"
	err = store.ObjectStore.CreateCase(context.Background(), testCase)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}

	_, err = store.ObjectStore.CreateEvidence(context.Background(), "test", "test", bytes.NewBufferString("sample"))
	if err != nil {
		t.Errorf("failed to add evidence: %v", err)
	}
	got, err := store.ObjectStore.GetEvidence(context.Background(), testCase.Name, testEvidenceName)
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

func TestGetNonexistentEvidenceInOBS(t *testing.T) {
	store, err := service.GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}

	testCase := db.CreateCaseParams{
		Name: "test",
	}
	err = store.ObjectStore.CreateCase(context.Background(), testCase)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}

	_, err = store.ObjectStore.GetEvidence(context.Background(), testCase.Name, "nonexistentEvidence")
	if err == nil {
		t.Errorf("expected error when getting nonexistent evidence, got nil")
	} else if !errors.Is(err, vault.ErrNotFound) {
		t.Errorf("expected ErrNotFound when getting nonexistent evidence, got %v", err)
	}
}

func TestCreateEvidenceWithDifferentFilesGeneratedDifferentHashes(t *testing.T) {
	store, err := service.GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	err = store.ObjectStore.CreateCase(context.Background(), db.CreateCaseParams{
		Name: "test",
	})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	testEvidenceName1 := "test"
	testEvidenceName2 := "test"
	hash1, err := store.ObjectStore.CreateEvidence(context.Background(), testEvidenceName1, "test", bytes.NewBufferString("ss"))
	if err != nil {
		t.Errorf("failed to add evidence: %v", err)
	}
	hash2, err := store.ObjectStore.CreateEvidence(context.Background(), testEvidenceName2, "test", bytes.NewBufferString("ssss"))
	if err != nil {
		t.Errorf("failed to add evidence: %v", err)
	}
	if hash1 == hash2 {
		t.Errorf("expected different hashes, got same")
	}
}

func TestRemoveEvidenceInOBSSuccessful(t *testing.T) {
	store, err := service.GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	testCase := db.CreateCaseParams{
		Name: "test",
	}
	testEvidence := "test"

	err = store.ObjectStore.CreateCase(context.Background(), testCase)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}

	_, err = store.ObjectStore.CreateEvidence(context.Background(), testEvidence, testCase.Name, bytes.NewBufferString(testEvidence))
	if err != nil {
		t.Errorf("failed to add evidence: %v", err)
	}
	err = store.ObjectStore.RemoveEvidence(context.Background(), testEvidence, testCase.Name)
	if err != nil {
		t.Errorf("failed to remove evidence: %v", err)
	}
}

func TestListCasesInOBSReturnedAllCasesInOBS(t *testing.T) {
	store, err := service.GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	case1 := db.CreateCaseParams{Name: "test"}
	case2 := db.CreateCaseParams{Name: "test2"}

	err = store.ObjectStore.CreateCase(context.Background(), case1)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	err = store.ObjectStore.CreateCase(context.Background(), case2)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	cases, err := store.ObjectStore.ListCases(context.Background())
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
			store, err := service.GetTestStores(t)
			if err != nil {
				t.Errorf("failed to get test stores: %v", err)
			}
			err = store.ObjectStore.CreateCase(context.Background(), db.CreateCaseParams{
				Name: "test",
			})
			if err != nil {
				t.Errorf("failed to add case: %v", err)
			}
			evName := "test"
			_, err = store.ObjectStore.CreateEvidence(context.Background(), evName, tt.caseName, bytes.NewBufferString("s"))
			if err != nil {
				t.Errorf("failed to add evidence: %v", err)
			}
			got, err := store.ObjectStore.EvidenceExists(context.Background(), tt.caseName, tt.evidenceName)
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
	store, err := service.GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	testCase := db.CreateCaseParams{
		Name: "test",
	}
	want := []string{
		"test",
		"test2",
	}
	err = store.ObjectStore.CreateCase(context.Background(), testCase)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	for _, ev := range want {
		_, err = store.ObjectStore.CreateEvidence(context.Background(), ev, testCase.Name, bytes.NewBufferString(ev))
		if err != nil {
			t.Errorf("failed to add evidence: %v", err)
		}
	}
	evidences, err := store.ObjectStore.ListEvidences(context.Background(), testCase.Name)
	if err != nil {
		t.Errorf("failed to list all evidence: %v", err)
	}
	if len(evidences) != len(want) {
		t.Errorf("expected %v evidence, got %v", len(want), len(evidences))
	}
	for i, e := range evidences {
		if e.Name != want[i] {
			t.Errorf("expected %v, got %v", want[i], e.Name)
		}
	}
}

func TestRemoveCaseInOBS(t *testing.T) {
	store, err := service.GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	caseName := "test"

	err = store.ObjectStore.CreateCase(context.Background(), db.CreateCaseParams{Name: caseName})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}

	err = store.ObjectStore.RemoveCase(context.Background(), caseName)
	if err != nil {
		t.Errorf("failed to remove case: %v", err)
	}

	exists, err := store.ObjectStore.CaseExists(context.Background(), caseName)
	if err != nil {
		t.Errorf("failed to check case: %v", err)
	}
	if exists {
		t.Errorf("expected case to be removed, but it exists")
	}
}

func TestCaseExistsInOBS(t *testing.T) {
	store, err := service.GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get test stores: %v", err)
	}
	caseName := "test"

	err = store.ObjectStore.CreateCase(context.Background(), db.CreateCaseParams{Name: caseName})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}

	exists, err := store.ObjectStore.CaseExists(context.Background(), caseName)
	if err != nil {
		t.Errorf("failed to check case: %v", err)
	}
	if !exists {
		t.Errorf("expected case to exist, but it does not")
	}
}
