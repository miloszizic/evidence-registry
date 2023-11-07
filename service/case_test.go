//go:build integration

package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/miloszizic/der/service"
)

func TestCreateCaseInDBAndOBS(t *testing.T) {
	// get test stores
	stores, err := service.GetTestStores(t)
	if err != nil {
		t.Fatalf("Error getting test stores: %v", err)
	}

	// create user for testing
	userReq := service.CreateUserParams{
		Username: "test",
		Password: "test",
	}

	_, err = stores.CreateUser(context.Background(), userReq)
	if err != nil {
		t.Fatalf("Error adding user: %v", err)
	}

	caseTypeID, err := stores.DBStore.GetCaseTypeIDByName(context.Background(), "KM")
	if err != nil {
		t.Fatalf("Error getting CaseTypeID: %v", err)
	}

	caseCourtID, err := stores.DBStore.GetCourtIDByShortName(context.Background(), "OSPG")
	if err != nil {
		t.Fatalf("Error getting CaseCourtID: %v", err)
	}

	tests := []struct {
		name    string
		cs      service.CreateCaseParams
		wantErr bool
	}{
		{
			name: "with valid case parameters",
			cs: service.CreateCaseParams{
				CaseTypeID:  caseTypeID,
				CaseNumber:  144,
				CaseYear:    2023,
				CaseCourtID: caseCourtID,
			},
			wantErr: false,
		},
		{
			name: "with invalid CaseTypeID",
			cs: service.CreateCaseParams{
				CaseTypeID:  uuid.New(), // empty UUID would be considered "invalid"
				CaseNumber:  144,
				CaseYear:    2023,
				CaseCourtID: caseCourtID,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fetch created user
			user, err := stores.DBStore.GetUserByUsername(context.Background(), userReq.Username)
			if err != nil {
				t.Errorf("Error fetching created user: %v", err)
			}

			// test CreateCase
			_, err = stores.CreateCase(context.Background(), user.ID, tt.cs)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCase() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateCaseFailsIfCaseAlreadyExists(t *testing.T) {
	// get test stores
	stores, err := service.GetTestStores(t)
	if err != nil {
		t.Fatalf("Error getting test stores: %v", err)
	}
	// create user for testing
	userReq := service.CreateUserParams{
		Username: "test",
		Password: "test",
	}

	_, err = stores.CreateUser(context.Background(), userReq)
	if err != nil {
		t.Fatalf("Error adding user: %v", err)
	}

	// fetch created user
	user, err := stores.DBStore.GetUserByUsername(context.Background(), userReq.Username)
	if err != nil {
		t.Fatalf("Error fetching created user: %v", err)
	}

	caseTypeID, err := stores.DBStore.GetCaseTypeIDByName(context.Background(), "KM")
	if err != nil {
		t.Fatalf("Error getting CaseTypeID: %v", err)
	}

	caseCourtID, err := stores.DBStore.GetCourtIDByShortName(context.Background(), "OSPG")
	if err != nil {
		t.Fatalf("Error getting CaseCourtID: %v", err)
	}

	cs := service.CreateCaseParams{
		CaseTypeID:  caseTypeID, // value is 'KM', 'KRIVIČNI POSTUPAK PREMA MALOLJETNICIMA'
		CaseNumber:  1,
		CaseYear:    2023,
		CaseCourtID: caseCourtID, // value is APELACIONI SUD CG
	}

	// Create the initial case
	_, err = stores.CreateCase(context.Background(), user.ID, cs)
	if err != nil {
		t.Fatalf("Error creating initial case: %v", err)
	}

	// Attempt to create the same case again
	_, err = stores.CreateCase(context.Background(), user.ID, cs)
	if !errors.Is(err, service.ErrAlreadyExists) {
		t.Errorf("CreateCase() error = %v, wantErr %v", err, service.ErrAlreadyExists)
	}
}

func TestGetCaseByID(t *testing.T) {
	tests := []struct {
		name    string
		caseID  string
		wantErr bool
	}{
		{
			name:    "case exists",
			caseID:  "",
			wantErr: false,
		},
		{
			name:    "case does not exist",
			caseID:  "99999999-9999-9999-9999-999999999999", // ID that doesn't exist
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// get test stores
			stores, err := service.GetTestStores(t)
			if err != nil {
				t.Fatalf("Error getting test stores: %v", err)
			}
			// create user for testing
			userReq := service.CreateUserParams{
				Username: "test",
				Password: "test",
			}
			_, err = stores.CreateUser(context.Background(), userReq)
			if err != nil {
				t.Fatalf("Error adding user: %v", err)
			}

			// Fetch created user
			user, err := stores.DBStore.GetUserByUsername(context.Background(), userReq.Username)
			if err != nil {
				t.Fatalf("Error fetching created user: %v", err)
			}
			// create CaseTypeID and CaseCourtID
			caseTypeID, err := stores.DBStore.GetCaseTypeIDByName(context.Background(), "KM")
			if err != nil {
				t.Fatalf("Error getting CaseTypeID: %v", err)
			}

			caseCourtID, err := stores.DBStore.GetCourtIDByShortName(context.Background(), "OSPG")
			if err != nil {
				t.Fatalf("Error getting CaseCourtID: %v", err)
			}

			// Create a case for existing case test
			cs := service.CreateCaseParams{
				CaseTypeID:  caseTypeID,
				CaseNumber:  2,
				CaseYear:    2023,
				CaseCourtID: caseCourtID,
			}
			createdCase, err := stores.CreateCase(context.Background(), user.ID, cs)
			if err != nil {
				t.Fatalf("Error creating case: %v", err)
			}
			if tt.wantErr == false {
				tt.caseID = createdCase.ID.String()
			}

			// parse the case id
			caseID := uuid.MustParse(tt.caseID)
			// test GetCaseByID

			_, err = stores.GetCaseByID(context.Background(), caseID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCaseByID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListCasesRetrievedAllCasesThatExistInDBAndODB(t *testing.T) {
	// get test stores
	stores, err := service.GetTestStores(t)
	if err != nil {
		t.Fatalf("Error getting test stores: %v", err)
	}
	// create user for testing
	userReq := service.CreateUserParams{
		Username: "test",
		Password: "test",
	}
	_, err = stores.CreateUser(context.Background(), userReq)

	if err != nil {
		t.Fatalf("Error adding user: %v", err)
	}

	// Fetch created user
	user, err := stores.DBStore.GetUserByUsername(context.Background(), userReq.Username)
	if err != nil {
		t.Fatalf("Error fetching created user: %v", err)
	}

	// create CaseTypeID and CaseCourtID
	caseTypeID, err := stores.DBStore.GetCaseTypeIDByName(context.Background(), "KM")
	if err != nil {
		t.Fatalf("Error getting CaseTypeID: %v", err)
	}

	caseCourtID, err := stores.DBStore.GetCourtIDByShortName(context.Background(), "OSPG")
	if err != nil {
		t.Fatalf("Error getting CaseCourtID: %v", err)
	}

	// Create some cases
	caseCount := 5
	for i := 0; i < caseCount; i++ {
		cs := service.CreateCaseParams{
			CaseTypeID:  caseTypeID,
			CaseNumber:  int32(i),
			CaseYear:    2023,
			CaseCourtID: caseCourtID,
		}

		_, err = stores.CreateCase(context.Background(), user.ID, cs)
		if err != nil {
			t.Fatalf("Error creating case: %v", err)
		}
	}

	// test ListCases
	cases, err := stores.ListCases(context.Background())
	if err != nil {
		t.Fatalf("ListCases() error = %v", err)
	}

	if len(cases) != caseCount {
		t.Errorf("ListCases() returned %d cases, want %d", len(cases), caseCount)
	}
}
