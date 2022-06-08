////go:build integration

package data_test

import (
	"database/sql"
	"errors"
	"github.com/miloszizic/der/internal/data"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/google/go-cmp/cmp"
)

func TestAddCaseCreatedCaseInDBSuccessfully(t *testing.T) {
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
	err = store.User.Add(reqUser)
	if err != nil {
		t.Errorf("Error creating user: %v", err)
	}
	err = store.DBStore.AddCase(reqCase, reqUser)
	if err != nil {
		t.Errorf("Error creating case: %v", err)
	}
	got, err := store.DBStore.GetCaseByName("TestCase")
	if err != nil {
		t.Errorf("Error getting cases: %v", err)
	}
	if !cmp.Equal(got, want, cmpopts.IgnoreFields(data.Case{}, "ID")) {
		t.Errorf(cmp.Diff(want, got))
	}
}
func TestCaseExistsReturnedTrueForExistingCase(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error creating case service: %v", err)
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
	err = store.User.Add(reqUser)
	if err != nil {
		t.Errorf("Error creating user: %v", err)
	}
	err = store.DBStore.AddCase(reqCase, reqUser)
	if err != nil {
		t.Errorf("Error creating case: %v", err)
	}
	got, err := store.DBStore.CaseExists("TestCase")
	if err != nil {
		t.Errorf("Error getting cases: %v", err)
	}
	if got != true {
		t.Errorf("Expected: %v, got: %v", true, got)
	}
}
func TestAddCase(t *testing.T) {
	testCases := []struct {
		name     string
		caseData *data.Case
		userData *data.User
		password string
		wantErr  bool
	}{
		{
			name: "with valid case name successful",
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
			name: "with invalid case name failed",
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
		{
			name: "with invalid user failed",
			caseData: &data.Case{
				Name: "TestCase",
			},
			userData: &data.User{
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
			err = store.User.Add(tc.userData)
			if err != nil {
				t.Errorf("Error creating user: %v", err)
			}
			err = store.DBStore.AddCase(tc.caseData, tc.userData)
			if (err != nil) != tc.wantErr {
				t.Errorf("Expected error: %v, got: %v", tc.wantErr, err)
			}
		})
	}
}

func TestGetCaseByIDReturnedCorrectCase(t *testing.T) {
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
	err = store.User.Add(reqUser)
	if err != nil {
		t.Errorf("Error creating user: %v", err)
	}
	err = store.DBStore.AddCase(reqCase, reqUser)
	if err != nil {
		t.Errorf("Error creating case: %v", err)
	}
	got, err := store.DBStore.GetCaseByID(1)
	if err != nil {
		t.Errorf("Error getting cases: %v", err)
	}
	if !cmp.Equal(got, want, cmpopts.IgnoreFields(data.Case{}, "ID")) {
		t.Errorf(cmp.Diff(want, got))
	}
}
func TestGetCaseByIDReturnedFalseForMissingCase(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error creating case service: %v", err)
	}
	_, err = store.DBStore.GetCaseByID(10)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
func TestCaseExistsReturnedFalseForMissingCase(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error creating case service: %v", err)
	}
	got, err := store.DBStore.CaseExists("TestCase")
	if err != nil {
		t.Errorf("Error getting cases: %v", err)
	}
	if got != false {
		t.Errorf("Expected: false, got: %v", got)
	}

}
func TestRemoveCaseRemovedCaseFromDB(t *testing.T) {
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
	err = store.User.Add(testUser)
	if err != nil {
		t.Errorf("Error creating user: %v", err)
	}
	err = store.DBStore.AddCase(testCase, testUser)
	if err != nil {
		t.Errorf("Error creating case: %v", err)
	}
	err = store.DBStore.RemoveCase(testCase)
	if err != nil {
		t.Errorf("Error deleting case: %v", err)
	}
	_, err = store.DBStore.GetCaseByName("TestCase")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
func TestGetCaseByNameReturnedErrorForMissingCase(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error creating case service: %v", err)
	}
	_, err = store.DBStore.GetCaseByName("NonExistent")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
func TestGetCaseByUserIDReturnedAllCasesForSpecificUser(t *testing.T) {
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
	err = store.User.Add(testUser)
	if err != nil {
		t.Errorf("Error creating user: %v", err)
	}

	for _, testCase := range want {
		err = store.DBStore.AddCase(&testCase, testUser)
		if err != nil {
			t.Errorf("Error creating case: %v", err)
		}
	}
	got, err := store.DBStore.GetCaseByUserID(1)
	if err != nil {
		t.Errorf("Error retriving cases for specific user ID : %v", err)
	}
	if !cmp.Equal(want, got, cmpopts.IgnoreFields(data.Case{}, "ID")) {
		t.Errorf(cmp.Diff(want, got))
	}
}
func TestListCasesListedAllCasesInDB(t *testing.T) {
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
	err = store.User.Add(testUser)
	if err != nil {
		t.Errorf("Error creating user: %v", err)
	}

	for _, testCase := range want {
		err = store.DBStore.AddCase(&testCase, testUser)
		if err != nil {
			t.Errorf("Error creating case: %v", err)
		}
	}
	got, err := store.DBStore.ListCases()
	if err != nil {
		t.Errorf("failed to get all cases: %v", err)
	}
	if !cmp.Equal(want, got, cmpopts.IgnoreFields(data.Case{}, "ID")) {
		t.Errorf(cmp.Diff(want, got))
	}
}
func TestFindCaseByTagsFoundCorrectCases(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error creating case service: %v", err)
	}
	want := []data.Case{
		{Name: "TestCase", Tags: []string{"tag1", "tag2"}},
	}
	// CreateCase a user
	testUser := &data.User{
		ID:       1,
		Username: "TestUser3",
	}
	err = testUser.Password.Set("TestPassword")
	if err != nil {
		t.Error(err)
	}
	err = store.User.Add(testUser)
	if err != nil {
		t.Errorf("Error creating user: %v", err)
	}
	// CreateCase a case
	err = store.DBStore.AddCase(&want[0], testUser)
	if err != nil {
		t.Errorf("Error creating case: %v", err)
	}
	searchableTags := []string{"tag1", "tag2"}
	got, err := store.DBStore.FindCaseByTags(searchableTags)
	if err != nil {
		t.Errorf("failed to get cases by tags: %v", err)
	}
	if !cmp.Equal(want, got, cmpopts.IgnoreFields(data.Case{}, "ID")) {
		t.Errorf(cmp.Diff(want, got))
	}

}
func TestGetEvidenceByCaseIDReturnedAllEvidenceForSpecificCase(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get store: %v", err)
	}
	// create a user
	user := &data.User{
		ID:       1,
		Username: "test",
	}
	err = user.Password.Set("test")
	if err != nil {
		t.Errorf("Failed to set password: %v", err)
	}
	err = store.User.Add(user)
	if err != nil {
		t.Errorf("failed to add user: %v", err)
	}
	// create a case
	caseToAdd := &data.Case{
		Name: "test",
	}
	err = store.DBStore.AddCase(caseToAdd, user)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	want := []data.Evidence{
		{ID: 1, CaseID: 1, Name: "video"},
		{ID: 2, CaseID: 1, Name: "picture"},
	}
	for _, evidence := range want {
		_, err := store.DBStore.CreateEvidence(&evidence)
		if err != nil {
			t.Errorf("creating the evidence failed: %v", err)
		}
	}
	got, err := store.DBStore.GetEvidenceByCaseID(1)
	if err != nil {
		t.Errorf("failed to get evidences by case ID: %v", err)
	}
	if !cmp.Equal(want, got, cmpopts.IgnoreFields(data.Evidence{}, "ID", "Hash")) {
		t.Errorf(cmp.Diff(want, got))
	}

}
func TestCreateEvidenceCreatedNewEvidence(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get store: %v", err)
	}
	want := []data.Evidence{
		{ID: 1, CaseID: 1, Name: "video"},
	}

	// create a user
	user := &data.User{
		ID:       1,
		Username: "test",
	}
	err = user.Password.Set("test")
	if err != nil {
		t.Errorf("Failed to set password: %v", err)
	}
	err = store.User.Add(user)
	if err != nil {
		t.Errorf("failed to add user: %v", err)
	}
	// create a case
	caseToAdd := &data.Case{
		Name: "test",
	}
	err = store.DBStore.AddCase(caseToAdd, user)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	testEvidence := &data.Evidence{
		CaseID: 1,
		Name:   "video",
	}
	ID, err := store.DBStore.CreateEvidence(testEvidence)
	if err != nil {
		t.Errorf("failed to create evidence: %v", err)
	}
	got, err := store.DBStore.GetEvidenceByCaseID(ID)
	if err != nil {
		t.Errorf("failed to get evidence from case with error: %v", err)
	}
	if !cmp.Equal(want, got, cmpopts.IgnoreFields(data.Evidence{}, "ID", "Hash")) {
		t.Errorf(cmp.Diff(want, got))
	}
}
func TestRemoveEvidenceDeletedAEvidenceFromTheCase(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get store: %v", err)
	}
	evidencesToAdd := []data.Evidence{
		{ID: 1, CaseID: 1, Name: "video"},
		{ID: 2, CaseID: 1, Name: "picture"},
	}
	// create a user
	user := &data.User{
		ID:       1,
		Username: "test",
	}
	err = user.Password.Set("test")
	if err != nil {
		t.Errorf("failed to set password: %v", err)
	}
	err = store.User.Add(user)
	if err != nil {
		t.Errorf("failed to add user: %v", err)
	}
	// create a case
	caseToAdd := &data.Case{
		Name: "test",
	}
	err = store.DBStore.AddCase(caseToAdd, user)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}

	for _, ev := range evidencesToAdd {
		_, err = store.DBStore.CreateEvidence(&ev)
		if err != nil {
			t.Errorf("failed to create evidence: %v", err)
		}
	}
	evidenceToDelete := &data.Evidence{ID: 1, CaseID: 1}
	err = store.DBStore.RemoveEvidence(evidenceToDelete)
	if err != nil {
		t.Errorf("failed to delete evidence: %v", err)
	}
	_, err = store.DBStore.GetEvidenceByID(1, 1)
	if err != sql.ErrNoRows {
		t.Errorf("Expected ErrNoRows, got %v", err)
	}
}
func TestGetEvidenceByNameRetrievedCorrectEvidence(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get store: %v", err)
	}
	want := &data.Evidence{
		ID: 2, CaseID: 1, Name: "picture",
	}
	testEvidences := []data.Evidence{
		{ID: 1, CaseID: 1, Name: "video"},
		{ID: 2, CaseID: 1, Name: "picture"},
	}
	testCase := &data.Case{ID: 1}
	// create a user
	user := &data.User{
		ID:       1,
		Username: "test",
	}
	err = user.Password.Set("test")
	if err != nil {
		t.Errorf("failed to set password: %v", err)
	}
	err = store.User.Add(user)
	if err != nil {
		t.Errorf("failed to add user: %v", err)
	}
	// create a case
	caseToAdd := &data.Case{
		ID:   1,
		Name: "test",
	}
	err = store.DBStore.AddCase(caseToAdd, user)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	for _, evidence := range testEvidences {
		_, err := store.DBStore.CreateEvidence(&evidence)
		if err != nil {
			t.Errorf("creating the evidence failed: %v", err)
		}
	}
	got, err := store.DBStore.GetEvidenceByName(testCase, "picture")
	if err != nil {
		t.Errorf("failed to get evidences by case ID: %v", err)
	}
	if !cmp.Equal(want, got) {
		t.Errorf(cmp.Diff(want, got))
	}
}
func TestGetEvidenceByNameReturnedErrorForMissingEvidence(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get store: %v", err)
	}
	testEvidences := []data.Evidence{
		{ID: 1, CaseID: 1, Name: "video"},
		{ID: 2, CaseID: 1, Name: "picture"},
	}
	testCase := &data.Case{ID: 1}
	// create a user
	user := &data.User{
		ID:       1,
		Username: "test",
	}
	err = user.Password.Set("test")
	if err != nil {
		t.Errorf("failed to set password: %v", err)
	}
	err = store.User.Add(user)
	if err != nil {
		t.Errorf("failed to add user: %v", err)
	}
	// create a case
	caseToAdd := &data.Case{
		ID:   1,
		Name: "test",
	}
	err = store.DBStore.AddCase(caseToAdd, user)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	for _, evidence := range testEvidences {
		_, err = store.DBStore.CreateEvidence(&evidence)
		if err != nil {
			t.Errorf("creating the evidence failed: %v", err)
		}
	}
	_, err = store.DBStore.GetEvidenceByName(testCase, "dog")
	if errors.Is(err, data.ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}

}
func TestEvidenceExistsReturnedFalseForMissingEvidence(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get store: %v", err)
	}
	testEvidence := &data.Evidence{
		ID: 1,
	}
	got, err := store.DBStore.EvidenceExists(testEvidence)
	if err != nil {
		t.Errorf("failed to get evidence: %v", err)
	}
	if got {
		t.Errorf("Expected false, got %v", got)
	}
}
func TestAddCommentAddedCommentToTheEvidence(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get store: %v", err)
	}
	want := []data.Comment{
		{ID: 1, EvidenceID: 1, Text: "something interesting"},
	}
	testComment := &data.Comment{EvidenceID: 1, Text: "something interesting"}
	testEvidences := []data.Evidence{
		{ID: 1, CaseID: 1, Name: "video"},
		{ID: 2, CaseID: 1, Name: "picture"},
	}
	// create a user
	user := &data.User{
		ID:       1,
		Username: "test",
	}
	err = user.Password.Set("test")
	if err != nil {
		t.Errorf("failed to set password: %v", err)
	}
	err = store.User.Add(user)
	if err != nil {
		t.Errorf("failed to add user: %v", err)
	}
	// create a case
	caseToAdd := &data.Case{
		Name: "test",
	}
	err = store.DBStore.AddCase(caseToAdd, user)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	for _, evidence := range testEvidences {
		_, err := store.DBStore.CreateEvidence(&evidence)
		if err != nil {
			t.Errorf("creating the evidence failed: %v", err)
		}
	}
	err = store.DBStore.AddComment(testComment)
	if err != nil {
		t.Errorf("failed to add the test comment: %v", err)
	}
	got, err := store.DBStore.GetCommentsByID(1)
	if err != nil {
		t.Errorf("failed to get evidences by case ID: %v", err)
	}
	if !cmp.Equal(want, got) {
		t.Errorf(cmp.Diff(want, got))
	}
}
