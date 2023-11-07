//go:build integration

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/miloszizic/der/db"

	"github.com/go-chi/chi/v5"
	"github.com/miloszizic/der/service"
)

func TestCreateCaseHandler(t *testing.T) {
	tests := []struct {
		name            string
		requestBody     map[string]interface{}
		wantStatus      int
		withUser        bool
		withValidParams bool
	}{
		{
			name: "with valid case parameters and valid user",
			requestBody: map[string]interface{}{
				"case_year":     int32(2022),
				"case_type_id":  "", // Placeholder for UUID
				"case_number":   int32(12345),
				"case_court_id": "", // Placeholder for UUID
			},
			wantStatus:      http.StatusCreated,
			withUser:        true,
			withValidParams: true,
		},
		{
			name: "without user",
			requestBody: map[string]interface{}{
				"case_year":     int32(2022),
				"case_type_id":  "", // Placeholder for UUID
				"case_number":   int32(12345),
				"case_court_id": "", // Placeholder for UUID
			},
			wantStatus:      http.StatusInternalServerError,
			withUser:        false,
			withValidParams: true,
		},
		{
			name: "with invalid case parameters",
			requestBody: map[string]interface{}{
				"invalid_key": "invalid_value",
			},
			wantStatus:      http.StatusNotFound,
			withUser:        true,
			withValidParams: false,
		},
		{
			name: "with valid case parameters but failed",
			requestBody: map[string]interface{}{
				"case_year":     int32(2022),
				"case_type_id":  "", // Placeholder for UUID
				"case_number":   int32(12345),
				"case_court_id": "", // Placeholder for UUID
			},
			wantStatus:      http.StatusBadRequest,
			withUser:        true,
			withValidParams: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, user, _ := NewTestEvidenceServer(t)
			// getting valid caseTypeID and caseCourtID
			caseTypeID, err := app.stores.DBStore.GetCaseTypeIDByName(context.Background(), "KM")
			if err != nil {
				t.Fatal(err)
			}
			caseCourtID, err := app.stores.DBStore.GetCourtIDByShortName(context.Background(), "OSPG")
			if err != nil {
				t.Fatal(err)
			}

			// modifying the request body to use valid IDs
			if tt.withValidParams {
				tt.requestBody["case_type_id"] = caseTypeID.String()
				tt.requestBody["case_court_id"] = caseCourtID.String()
			}

			requestBody, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatal(err)
			}

			// Making a request
			request := httptest.NewRequest("POST", "/cases", bytes.NewBuffer(requestBody))

			// Only add the user to the context if it was created
			if tt.withUser {
				ctx := context.WithValue(request.Context(), userContextKey, user)
				request = request.WithContext(ctx)
			}

			// Recording the response
			response := httptest.NewRecorder()
			app.CreateCaseHandler(response, request)

			// Checking the status code
			if response.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
			}
		})
	}
}

func TestGetCaseHandler(t *testing.T) {
	tests := []struct {
		name       string
		caseID     string
		wantStatus int
	}{
		{
			name:       "get existing case by ID",
			caseID:     "",
			wantStatus: http.StatusOK,
		},
		{
			name:       "get non-existing case by ID",
			caseID:     "99999999-9999-9999-9999-999999999999",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "get case with invalid ID",
			caseID:     "invalid_id",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewTestServer(t)
			user := service.CreateUserParams{
				Username: "test",
				Password: "test",
			}
			// create user for testing
			createdUser, err := app.stores.CreateUser(context.Background(), user)
			if err != nil {
				t.Fatal(err)
			}
			caseTypeID, err := app.stores.DBStore.GetCaseTypeIDByName(context.Background(), "KM")
			if err != nil {
				t.Fatal(err)
			}
			caseCourtID, err := app.stores.DBStore.GetCourtIDByShortName(context.Background(), "OSPG")
			if err != nil {
				t.Fatal(err)
			}
			// create a case for testing
			if tt.wantStatus == http.StatusOK {
				createdCase, err := app.stores.CreateCase(context.Background(), createdUser.ID, service.CreateCaseParams{
					CaseYear:    int32(2022),
					CaseTypeID:  caseTypeID,
					CaseNumber:  12345,
					CaseCourtID: caseCourtID,
				})
				if err != nil {
					t.Fatal(err)
				}
				tt.caseID = createdCase.ID.String()
			}

			payload := &service.Payload{
				Username: "test",
			}
			// Making a request
			r := chi.NewRouter()
			r.Get("/cases/{caseID}", app.GetCaseHandler)
			request := httptest.NewRequest("GET", fmt.Sprintf("/cases/%s", tt.caseID), nil)
			// Adding payload ctx to request
			ctx := context.WithValue(request.Context(), authorizationPayloadKey, payload)
			reqWithPayload := request.WithContext(ctx)
			// Recording the response
			response := httptest.NewRecorder()
			// Serving the request through our chi router
			r.ServeHTTP(response, reqWithPayload)

			// Checking the status code
			if response.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
			}

			if tt.wantStatus == http.StatusOK {
				// Unmarshal the response
				var resp envelope
				err = json.Unmarshal(response.Body.Bytes(), &resp)
				if err != nil {
					t.Fatalf("could not unmarshal response: %v", err)
				}
				// Check if the case is correctly returned
				if resp["Case"] == nil {
					t.Errorf("expected a case in response, got nil")
				}
			}
		})
	}
}

func TestDeleteCaseHandler(t *testing.T) {
	tests := []struct {
		name       string
		caseID     string
		wantStatus int
	}{
		{
			name:       "delete existing case by ID",
			caseID:     "",
			wantStatus: http.StatusOK,
		},
		{
			name:       "delete non-existing case by ID",
			caseID:     "99999999-9999-9999-9999-999999999999",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "delete case with invalid ID",
			caseID:     "invalid_id",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewTestServer(t)
			user := service.CreateUserParams{
				Username: "test",
				Password: "test",
			}
			// create user for testing
			createdUser, err := app.stores.CreateUser(context.Background(), user)
			if err != nil {
				t.Fatal(err)
			}
			caseTypeID, err := app.stores.DBStore.GetCaseTypeIDByName(context.Background(), "KM")
			if err != nil {
				t.Fatal(err)
			}
			caseCourtID, err := app.stores.DBStore.GetCourtIDByShortName(context.Background(), "OSPG")
			if err != nil {
				t.Fatal(err)
			}
			// create a case for testing
			if tt.wantStatus == http.StatusOK {
				createdCase, err := app.stores.CreateCase(context.Background(), createdUser.ID, service.CreateCaseParams{
					CaseYear:    int32(2022),
					CaseTypeID:  caseTypeID,
					CaseNumber:  12345,
					CaseCourtID: caseCourtID,
				})
				if err != nil {
					t.Fatal(err)
				}
				tt.caseID = createdCase.ID.String()
			}

			payload := &service.Payload{
				Username: "test",
			}
			// Making a request
			r := chi.NewRouter()
			r.Delete("/cases/{caseID}", app.DeleteCaseHandler)
			request := httptest.NewRequest("DELETE", fmt.Sprintf("/cases/%s", tt.caseID), nil)
			// Adding payload ctx to request
			ctx := context.WithValue(request.Context(), authorizationPayloadKey, payload)
			reqWithPayload := request.WithContext(ctx)
			// Recording the response
			response := httptest.NewRecorder()
			// Serving the request through our chi router
			r.ServeHTTP(response, reqWithPayload)

			// Checking the status code
			if response.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
			}

			// If the status was NoContent, ensure the case no longer exists
			if tt.wantStatus == http.StatusOK {
				_, err := app.stores.DBStore.GetCase(context.Background(), uuid.MustParse(tt.caseID))
				if err == nil {
					t.Errorf("expected case to be deleted, but it still exists")
				}
			}
		})
	}
}

func TestListCasesHandler(t *testing.T) {
	tests := []struct {
		name       string
		cases      []service.CreateCaseParams
		wantStatus int
	}{
		{
			name:       "returned no cases where there are non",
			cases:      nil,
			wantStatus: http.StatusOK,
		},
		{
			name: "returned two cases where there are two",
			cases: []service.CreateCaseParams{
				{CaseYear: 2022, CaseTypeID: uuid.UUID{}, CaseNumber: 12345, CaseCourtID: uuid.UUID{}},
				{CaseYear: 2023, CaseTypeID: uuid.UUID{}, CaseNumber: 12346, CaseCourtID: uuid.UUID{}},
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewTestServer(t)

			// Create a test user
			user := db.CreateUserParams{
				Username: "testuser",
				Password: "testpassword",
			}
			createdUser, err := app.stores.DBStore.CreateUser(context.Background(), user)
			if err != nil {
				t.Fatalf("failed to create mock user: %v", err)
			}
			caseTypeID, err := app.stores.DBStore.GetCaseTypeIDByName(context.Background(), "KM")
			if err != nil {
				t.Fatal(err)
			}
			caseCourtID, err := app.stores.DBStore.GetCourtIDByShortName(context.Background(), "OSPG")
			if err != nil {
				t.Fatal(err)
			}
			// Now we assign the IDs to the test case parameters
			for i := range tt.cases {
				tt.cases[i].CaseTypeID = caseTypeID
				tt.cases[i].CaseCourtID = caseCourtID
			}
			// Adding test cases
			var wantCases []service.Case
			for _, c := range tt.cases {
				caseObj, err := app.stores.CreateCase(context.Background(), createdUser.ID, c)
				if err != nil {
					t.Fatalf("failed to create case: %v", err)
				}
				wantCases = append(wantCases, *caseObj)
			}

			// Making a request
			request := httptest.NewRequest("GET", "/cases", nil)
			// Recording the response
			response := httptest.NewRecorder()

			app.ListCasesHandler(response, request)

			// Checking the status code
			if response.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
			}

			// Checking the body of the response
			var respEnvelope struct {
				Cases []service.Case `json:"Cases"`
			}
			bodyBytes, _ := io.ReadAll(response.Body)
			err = json.Unmarshal(bodyBytes, &respEnvelope)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}
			if diff := cmp.Diff(wantCases, respEnvelope.Cases); diff != "" {
				t.Errorf("cases mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRequestUserParser(t *testing.T) {
	tests := []struct {
		name          string
		expectedUser  string
		user          db.CreateUserParams
		wantError     bool
		wantUserMatch bool
	}{
		{
			name:         "successful parsed existing user",
			expectedUser: "existingUser",
			user: db.CreateUserParams{
				Username: "existingUser",
				Password: "password",
			},
			wantError:     false,
			wantUserMatch: true,
		},
		{
			name:         "returns error when user does not exist",
			expectedUser: "nonExistingUser",
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewTestServer(t)
			if tt.wantUserMatch {
				_, err := app.stores.DBStore.CreateUser(context.Background(), tt.user)
				if err != nil {
					t.Fatalf("failed to create mock user: %v", err)
				}
			}

			// Mock the payload and the context
			payload := &service.Payload{
				Username: tt.expectedUser,
			}
			ctx := context.WithValue(context.Background(), authorizationPayloadKey, payload)

			// Create a mock request with the context
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request = request.WithContext(ctx)

			gotUser, err := app.requestUserParser(request)
			if (err != nil) != tt.wantError {
				t.Errorf("requestUserParser() error = %v, wantErr %v", err, tt.wantError)
				return
			}
			if tt.wantUserMatch {
				if gotUser.Username != tt.user.Username {
					t.Errorf("requestUserParser() user = %v, want %v", gotUser.Username, tt.user.Username)
				}
			}
		})
	}
}
