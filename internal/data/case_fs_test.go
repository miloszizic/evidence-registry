package storage_test

import (
	"evidence/internal/data/database"
	"testing"
)

func TestMinioCaseCreation(t *testing.T) {
	tests := []struct {
		name    string
		addCase *database.Case
		wantErr bool
	}{
		{
			name: "successful with just letters",
			addCase: &database.Case{
				Name: "test",
			},
			wantErr: false,
		},
		{
			name: "successful with just letters and numbers",
			addCase: &database.Case{
				Name: "test23test",
			},
			wantErr: false,
		},
		{
			name: "successful with just letters and supported special characters",
			addCase: &database.Case{
				Name: "test-test",
			},
			wantErr: false,
		},
		{
			name: "failed with an unsupported special character ",
			addCase: &database.Case{
				Name: "test/test",
			},
			wantErr: true,
		},
		{
			name: "failed with space ",
			addCase: &database.Case{
				Name: "test/test",
			},
			wantErr: true,
		},
		{
			name: "failed with no name ",
			addCase: &database.Case{
				Name: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := data.GetTestStores(t)
			if err != nil {
				t.Errorf("failed to get test stores: %v", err)
			}
			err = store.CaseFS.Create(tt.addCase)
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
		addCase    *database.Case
		deleteCase string
		wantErr    bool
	}{
		{
			name: "successful for existing case",
			addCase: &database.Case{
				Name: "test",
			},
			deleteCase: "test",
			wantErr:    false,
		},
		{
			name: "failed for non-existing case",
			addCase: &database.Case{
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
			err = store.CaseFS.Create(tt.addCase)
			if err != nil {
				t.Errorf("failed to add case: %v", err)
			}
			err = store.CaseFS.Remove(tt.deleteCase)
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
	err = store.CaseFS.Create(&database.Case{
		Name: "test",
	})
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	err = store.CaseFS.Create(&database.Case{
		Name: "test",
	})
	got := err != nil
	if !got {
		t.Errorf("expected true, got false")
	}
}
