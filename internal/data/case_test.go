////go:build integration

package data_test

import (
	"database/sql"
	"evidence/internal/data"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/google/go-cmp/cmp"
)

func TestNewCaseCreatedSuccessfullyInDB(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error creating case service: %v", err)
	}
	want := &data.Case{
		ID:   1,
		Name: "TestCase",
	}
	reqCase := &data.Case{
		Name: "TestCase",
	}
	reqUser := &data.User{
		ID:       1,
		Username: "TestUser",
	}
	err = reqUser.Password.Set("TestPassword")
	if err != nil {
		t.Error(err)
	}
	err = store.UserDB.Add(reqUser)
	if err != nil {
		t.Errorf("Error creating user: %v", err)
	}
	err = store.CaseDB.Add(reqCase, reqUser)
	if err != nil {
		t.Errorf("Error creating case: %v", err)
	}
	got, err := store.CaseDB.GetByName("TestCase")
	if err != nil {
		t.Errorf("Error getting cases: %v", err)
	}
	if !cmp.Equal(got, want, cmpopts.IgnoreFields(data.Case{}, "ID")) {
		t.Errorf(cmp.Diff(want, got))
	}
}
func TestCaseCreation(t *testing.T) {
	testCases := []struct {
		name     string
		caseData *data.Case
		userData *data.User
		password string
		wantErr  bool
	}{
		{
			name: "with valid database",
			caseData: &data.Case{
				Name: "TestCase",
				Tags: []string{"tag1", "tag2"},
			},
			userData: &data.User{
				ID:       1,
				Username: "TestUser",
			},
			password: "TestPassword",
			wantErr:  false,
		},
		{
			name: "with invalid database",
			caseData: &data.Case{
				Name: "",
				Tags: []string{"tag1", "tag2"},
			},
			userData: &data.User{
				ID:       1,
				Username: "TestUser",
			},
			password: "TestPassword",
			wantErr:  true,
		},
	}
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error creating case service: %v", err)
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.userData.Password.Set(tc.password)
			if err != nil {
				t.Error(err)
			}
			err = store.UserDB.Add(tc.userData)
			if err != nil {
				t.Errorf("Error creating user: %v", err)
			}
			err = store.CaseDB.Add(tc.caseData, tc.userData)
			if (err != nil) != tc.wantErr {
				t.Errorf("Expected error: %v, got: %v", tc.wantErr, err)
			}
		})
	}
}

func TestSearchingForCaseByIDReturnsCorrectCase(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error creating case service: %v", err)
	}
	want := &data.Case{
		ID:   1,
		Name: "TestCase",
	}

	reqCase := &data.Case{
		Name: "TestCase",
	}
	reqUser := &data.User{
		ID:       1,
		Username: "TestUser2",
	}
	err = reqUser.Password.Set("TestPassword")
	if err != nil {
		t.Error(err)
	}
	err = store.UserDB.Add(reqUser)
	if err != nil {
		t.Errorf("Error creating user: %v", err)
	}
	err = store.CaseDB.Add(reqCase, reqUser)
	if err != nil {
		t.Errorf("Error creating case: %v", err)
	}
	got, err := store.CaseDB.GetByID(1)
	if err != nil {
		t.Errorf("Error getting cases: %v", err)
	}
	if !cmp.Equal(got, want, cmpopts.IgnoreFields(data.Case{}, "ID")) {
		t.Errorf(cmp.Diff(want, got))
	}
}
func TestSearchingForCaseByInvalidIDReturnsError(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error creating case service: %v", err)
	}
	_, err = store.CaseDB.GetByID(10)
	if err != sql.ErrNoRows {
		t.Errorf("expected error no rows, got %v", err)
	}
}
func TestCaseFoundByNameWasSuccessfullyDeletedInDB(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error creating case service: %v", err)
	}
	testCase := &data.Case{
		ID:   1,
		Name: "TestCase2",
	}
	testUser := &data.User{
		ID:       1,
		Username: "TestUser3",
	}
	err = testUser.Password.Set("TestPassword")
	if err != nil {
		t.Error(err)
	}
	err = store.UserDB.Add(testUser)
	if err != nil {
		t.Errorf("Error creating user: %v", err)
	}
	err = store.CaseDB.Add(testCase, testUser)
	if err != nil {
		t.Errorf("Error creating case: %v", err)
	}
	err = store.CaseDB.Remove(testCase)
	if err != nil {
		t.Errorf("Error deleting case: %v", err)
	}
	_, err = store.CaseDB.GetByName("TestCase")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
func TestSearchingForNonExistentCaseNameReturnsError(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error creating case service: %v", err)
	}
	_, err = store.CaseDB.GetByName("NonExistent")
	if err != sql.ErrNoRows {
		t.Errorf("expected error no rows, got %v", err)
	}
}
func TestReturnedAllCasesForSpecificUser(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error creating case service: %v", err)
	}
	want := []data.Case{
		{Name: "TestCase"},
		{Name: "TestCase2"},
		{Name: "TestCase3"},
	}

	testUser := &data.User{
		ID:       1,
		Username: "TestUser3",
	}
	err = testUser.Password.Set("TestPassword")
	if err != nil {
		t.Error(err)
	}
	err = store.UserDB.Add(testUser)
	if err != nil {
		t.Errorf("Error creating user: %v", err)
	}

	for _, testCase := range want {
		err = store.CaseDB.Add(&testCase, testUser)
		if err != nil {
			t.Errorf("Error creating case: %v", err)
		}
	}
	got, err := store.CaseDB.GetByUserID(1)
	if err != nil {
		t.Errorf("Error retriving cases for specific user ID : %v", err)
	}
	if !cmp.Equal(want, got, cmpopts.IgnoreFields(data.Case{}, "ID")) {
		t.Errorf(cmp.Diff(want, got))
	}
}
func TestRetrieveAllCasesInDB(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error creating case service: %v", err)
	}
	want := []data.Case{
		{Name: "TestCase"},
		{Name: "TestCase2"},
		{Name: "TestCase3"},
	}

	testUser := &data.User{
		ID:       1,
		Username: "TestUser3",
	}
	err = testUser.Password.Set("TestPassword")
	if err != nil {
		t.Error(err)
	}
	err = store.UserDB.Add(testUser)
	if err != nil {
		t.Errorf("Error creating user: %v", err)
	}

	for _, testCase := range want {
		err = store.CaseDB.Add(&testCase, testUser)
		if err != nil {
			t.Errorf("Error creating case: %v", err)
		}
	}
	got, err := store.CaseDB.List()
	if err != nil {
		t.Errorf("failed to get all cases: %v", err)
	}
	if !cmp.Equal(want, got, cmpopts.IgnoreFields(data.Case{}, "ID")) {
		t.Errorf(cmp.Diff(want, got))
	}
}
func TestRetireCasesByTags(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error creating case service: %v", err)
	}
	want := []data.Case{
		{Name: "TestCase", Tags: []string{"tag1", "tag2"}},
	}
	// Create a user
	testUser := &data.User{
		ID:       1,
		Username: "TestUser3",
	}
	err = testUser.Password.Set("TestPassword")
	if err != nil {
		t.Error(err)
	}
	err = store.UserDB.Add(testUser)
	if err != nil {
		t.Errorf("Error creating user: %v", err)
	}
	// Create a case
	err = store.CaseDB.Add(&want[0], testUser)
	if err != nil {
		t.Errorf("Error creating case: %v", err)
	}
	searchableTags := []string{"tag1", "tag2"}
	got, err := store.CaseDB.SearchByTags(searchableTags)
	if err != nil {
		t.Errorf("failed to get cases by tags: %v", err)
	}
	if !cmp.Equal(want, got, cmpopts.IgnoreFields(data.Case{}, "ID")) {
		t.Errorf(cmp.Diff(want, got))
	}

}
