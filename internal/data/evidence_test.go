////go:build integration
package database_test

import (
	"database/sql"
	database2 "evidence/internal/data/database"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestRetrieveAllEvidencesFromCase(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get store: %v", err)
	}
	// create a user
	user := &database2.User{
		ID:       1,
		Username: "test",
	}
	user.Password.Set("test")
	err = store.UserDB.Add(user)
	if err != nil {
		t.Errorf("failed to add user: %v", err)
	}
	// create a case
	caseToAdd := &database2.Case{
		Name: "test",
	}
	err = store.CaseDB.Add(caseToAdd, user)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	want := []database2.Evidence{
		{ID: 1, CaseID: 1, Name: "video"},
		{ID: 2, CaseID: 1, Name: "picture"},
	}
	for _, evidence := range want {
		_, err := store.EvidenceDB.Create(&evidence)
		if err != nil {
			t.Errorf("creating the evidence failed: %v", err)
		}
	}
	got, err := store.EvidenceDB.GetByCaseID(1)
	if err != nil {
		t.Errorf("failed to get evidences by case ID: %v", err)
	}
	if !cmp.Equal(want, got, cmpopts.IgnoreFields(database2.Evidence{}, "ID", "Hash")) {
		t.Errorf(cmp.Diff(want, got))
	}

}
func TestCreateOneEvidenceInCase(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get store: %v", err)
	}
	want := []database2.Evidence{
		{ID: 1, CaseID: 1, Name: "video"},
	}

	// create a user
	user := &database2.User{
		ID:       1,
		Username: "test",
	}
	user.Password.Set("test")
	err = store.UserDB.Add(user)
	if err != nil {
		t.Errorf("failed to add user: %v", err)
	}
	// create a case
	caseToAdd := &database2.Case{
		Name: "test",
	}
	err = store.CaseDB.Add(caseToAdd, user)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	testEvidence := &database2.Evidence{
		CaseID: 1,
		Name:   "video",
	}
	ID, err := store.EvidenceDB.Create(testEvidence)
	if err != nil {
		t.Errorf("failed to create evidence: %v", err)
	}
	got, err := store.EvidenceDB.GetByCaseID(ID)
	if err != nil {
		t.Errorf("failed to get evidence from case with error: %v", err)
	}
	if !cmp.Equal(want, got, cmpopts.IgnoreFields(database2.Evidence{}, "ID", "Hash")) {
		t.Errorf(cmp.Diff(want, got))
	}
}
func TestDeleteEvidenceByNameFromTheCase(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get store: %v", err)
	}
	evidencesToAdd := []database2.Evidence{
		{ID: 1, CaseID: 1, Name: "video"},
		{ID: 2, CaseID: 1, Name: "picture"},
	}
	// create a user
	user := &database2.User{
		ID:       1,
		Username: "test",
	}
	user.Password.Set("test")
	err = store.UserDB.Add(user)
	if err != nil {
		t.Errorf("failed to add user: %v", err)
	}
	// create a case
	caseToAdd := &database2.Case{
		Name: "test",
	}
	err = store.CaseDB.Add(caseToAdd, user)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}

	for _, ev := range evidencesToAdd {
		_, err = store.EvidenceDB.Create(&ev)
		if err != nil {
			t.Errorf("failed to create evidence: %v", err)
		}
	}
	evidenceToDelete := &database2.Evidence{ID: 1, CaseID: 1}
	err = store.EvidenceDB.Remove(evidenceToDelete)
	if err != nil {
		t.Errorf("failed to delete evidence: %v", err)
	}
	got, err := store.EvidenceDB.GetByID(1)
	if err != sql.ErrNoRows {
		t.Errorf("Expected no rows, got %v", got)
	}
}
func TestFindingTheEvidenceByItsName(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get store: %v", err)
	}
	want := &database2.Evidence{
		ID: 1, CaseID: 1, Name: "video",
	}
	testEvidences := []database2.Evidence{
		{ID: 1, CaseID: 1, Name: "video"},
		{ID: 2, CaseID: 1, Name: "picture"},
	}
	testCase := &database2.Case{ID: 1}
	// create a user
	user := &database2.User{
		ID:       1,
		Username: "test",
	}
	user.Password.Set("test")
	err = store.UserDB.Add(user)
	if err != nil {
		t.Errorf("failed to add user: %v", err)
	}
	// create a case
	caseToAdd := &database2.Case{
		ID:   1,
		Name: "test",
	}
	err = store.CaseDB.Add(caseToAdd, user)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	for _, evidence := range testEvidences {
		_, err := store.EvidenceDB.Create(&evidence)
		if err != nil {
			t.Errorf("creating the evidence failed: %v", err)
		}
	}
	got, err := store.EvidenceDB.GetByName(testCase, "video")
	if !cmp.Equal(want, got) {
		t.Errorf(cmp.Diff(want, got))
	}
}

func TestAddingCommentToTheEvidences(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Errorf("failed to get store: %v", err)
	}
	want := []database2.Comment{
		{ID: 1, EvidenceID: 1, Text: "something interesting"},
	}
	testComment := &database2.Comment{EvidenceID: 1, Text: "something interesting"}
	testEvidences := []database2.Evidence{
		{ID: 1, CaseID: 1, Name: "video"},
		{ID: 2, CaseID: 1, Name: "picture"},
	}
	// create a user
	user := &database2.User{
		ID:       1,
		Username: "test",
	}
	user.Password.Set("test")
	err = store.UserDB.Add(user)
	if err != nil {
		t.Errorf("failed to add user: %v", err)
	}
	// create a case
	caseToAdd := &database2.Case{
		Name: "test",
	}
	err = store.CaseDB.Add(caseToAdd, user)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	for _, evidence := range testEvidences {
		_, err := store.EvidenceDB.Create(&evidence)
		if err != nil {
			t.Errorf("creating the evidence failed: %v", err)
		}
	}
	err = store.EvidenceDB.AddComment(testComment)
	if err != nil {
		t.Errorf("failed to add the test comment: %v", err)
	}
	got, err := store.EvidenceDB.GetCommentsByID(1)
	if !cmp.Equal(want, got) {
		t.Errorf(cmp.Diff(want, got))
	}
}
