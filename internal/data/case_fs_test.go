package data_test

import (
	"evidence/internal/data"
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
			store, err := GetTestStores(t)
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
		Name: "test",
	})
	got := err != nil
	if !got {
		t.Errorf("expected true, got false")
	}
}
