//go:build integration

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/miloszizic/der/db"

	"github.com/google/uuid"

	"github.com/go-chi/chi/v5"
	"github.com/miloszizic/der/service"
)

func TestCreateEvidenceHandler(t *testing.T) {
	tests := []struct {
		name           string
		evidenceParams service.CreateEvidenceParams
		file           io.Reader
		want           int
	}{
		{
			name: "successfully created with correct request",
			evidenceParams: service.CreateEvidenceParams{
				Description: "Sample video evidence",
			},
			file: bytes.NewBuffer([]byte(`Sample video evidence`)),
			want: http.StatusCreated,
		},
		{
			name:           "fail when evidence attributes  is missing",
			evidenceParams: service.CreateEvidenceParams{},
			file:           bytes.NewBuffer([]byte(`Sample video evidence`)),
			want:           http.StatusNotFound, // expect 404 Bad Request
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create a new server that has one user and case for testing
			app, createdUser, createdCase := NewTestEvidenceServer(t)

			evidenceTypeID, err := app.stores.DBStore.GetEvidenceIDByType(context.Background(), "Initial Evidence")
			if err != nil {
				t.Fatalf("Error getting EvidenceTypeID: %v", err)
			}

			if tt.want == http.StatusCreated {
				tt.evidenceParams.CaseID = createdCase.ID
				tt.evidenceParams.EvidenceTypeID = evidenceTypeID
				tt.evidenceParams.AppUserID = createdUser.ID
			}

			// Create the JSON payload
			payloadBytes, err := json.Marshal(tt.evidenceParams)
			if err != nil {
				t.Fatalf("error creating JSON payload: %v", err)
			}

			// Convert the JSON payload to a string
			jsonString := string(payloadBytes)

			// Create a buffer to hold the multipart form data
			body := &bytes.Buffer{}

			// Create a multipart writer to write the form data to the buffer
			writer := multipart.NewWriter(body)

			// Write the JSON data as a form field
			err = writer.WriteField("evidence", jsonString)
			if err != nil {
				t.Fatalf("error writing JSON data: %v", err)
			}

			// Create a form file for the file data
			part, err := writer.CreateFormFile("upload_file", "evidence.txt")
			if err != nil {
				t.Fatalf("error creating form file: %v", err)
			}

			// Write the file data to the form file
			_, err = io.Copy(part, tt.file)
			if err != nil {
				t.Fatalf("error writing file data: %v", err)
			}

			// Close the multipart writer
			err = writer.Close()
			if err != nil {
				t.Fatalf("error closing multipart writer: %v", err)
			}

			// Create a request with the multipart form data
			url := fmt.Sprintf("/cases/%s/evidences", tt.evidenceParams.CaseID)
			req, err := http.NewRequest("POST", url, body)
			if err != nil {
				t.Fatalf("error creating request: %v", err)
			}

			// Set the content-type header to multipart/form-data with the boundary
			req.Header.Set("Content-Type", writer.FormDataContentType())

			// Record a response
			rec := httptest.NewRecorder()

			// Prepare the route context with necessary URL parameters
			rct := chi.NewRouteContext()
			rct.URLParams.Add("caseID", tt.evidenceParams.CaseID.String())

			// Add URL parameters to the request
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rct)

			// Add the user to the context
			ctx = context.WithValue(ctx, userContextKey, createdUser)

			// Use the updated context in the request
			req = req.WithContext(ctx)
			// call the handler
			app.CreateEvidenceHandler(rec, req)

			if rec.Code != tt.want {
				t.Errorf("expected status code %d, got %d. Response body: %s", tt.want, rec.Code, rec.Body.String())
			}
		})
	}
}

// TestCreateEvidenceHandlerNegative tests the negative scenarios for CreateEvidenceHandler
func TestCreateEvidenceHandlerNegative(t *testing.T) {
	// create a new server that has one user and case for testing
	app, createdUser, _ := NewTestEvidenceServer(t)

	_, err := app.stores.DBStore.GetEvidenceIDByType(context.Background(), "Initial Evidence")
	if err != nil {
		t.Fatalf("Error getting EvidenceTypeID: %v", err)
	}

	evidenceParams := service.CreateEvidenceParams{
		Description: "Sample video evidence",
	}

	// Create the JSON payload
	payloadBytes, err := json.Marshal(evidenceParams)
	if err != nil {
		t.Fatalf("error creating JSON payload: %v", err)
	}

	// Convert the JSON payload to a string
	jsonString := string(payloadBytes)

	// Create a buffer to hold the multipart form data
	body := &bytes.Buffer{}

	// Create a multipart writer to write the form data to the buffer
	writer := multipart.NewWriter(body)

	// Write the JSON data as a form field
	err = writer.WriteField("evidence", jsonString)
	if err != nil {
		t.Fatalf("error writing JSON data: %v", err)
	}

	// Create a form file for the file data
	_, err = writer.CreateFormFile("upload_file", "evidence.txt")
	if err != nil {
		t.Fatalf("error creating form file: %v", err)
	}

	// Close the multipart writer
	err = writer.Close()
	if err != nil {
		t.Fatalf("error closing multipart writer: %v", err)
	}

	// Create a request with the multipart form data
	url := fmt.Sprintf("/cases/%s/evidences", evidenceParams.CaseID)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}

	// Set the content-type header to multipart/form-data with the boundary
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Record a response
	rec := httptest.NewRecorder()

	// Prepare the payload
	payload := &service.Payload{
		Username: "test",
	}
	// Prepare the route context with necessary URL parameters
	rct := chi.NewRouteContext()
	rct.URLParams.Add("caseID", evidenceParams.CaseID.String())

	// Add payload, URL parameters, and context to the request
	ctx := context.WithValue(req.Context(), authorizationPayloadKey, payload)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rct)
	ctx = context.WithValue(ctx, userContextKey, createdUser)
	req = req.WithContext(ctx)

	// call the handler
	app.CreateEvidenceHandler(rec, req)

	want := http.StatusBadRequest
	if rec.Code != want {
		t.Errorf("expected status code %d, got %d. Response body: %s", want, rec.Code, rec.Body.String())
	}
}

func TestCreateEvidenceHandlerWithInvalidCaseID(t *testing.T) {
	app, createdUser, _ := NewTestEvidenceServer(t)

	evidenceParams := service.CreateEvidenceParams{
		Description: "Sample video evidence",
	}

	// Create the JSON payload
	payloadBytes, err := json.Marshal(evidenceParams)
	if err != nil {
		t.Fatalf("error creating JSON payload: %v", err)
	}

	// Convert the JSON payload to a string
	jsonString := string(payloadBytes)

	// Create a buffer to hold the multipart form data
	body := &bytes.Buffer{}

	// Create a multipart writer to write the form data to the buffer
	writer := multipart.NewWriter(body)

	// Write the JSON data as a form field
	err = writer.WriteField("evidence", jsonString)
	if err != nil {
		t.Fatalf("error writing JSON data: %v", err)
	}

	// Create a form file for the file data
	part, err := writer.CreateFormFile("upload_file", "evidence.txt")
	if err != nil {
		t.Fatalf("error creating form file: %v", err)
	}

	// Write the file data to the form file
	fileContent := "Sample file content for evidence."
	_, err = part.Write([]byte(fileContent))
	if err != nil {
		t.Fatalf("error writing file data: %v", err)
	}

	// Close the multipart writer
	err = writer.Close()
	if err != nil {
		t.Fatalf("error closing multipart writer: %v", err)
	}

	// Create a request with an invalid CaseID
	url := "/cases/invalid_case_id/evidences"
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}

	// Prepare the payload with the created user's details
	payload := &service.Payload{
		Username: createdUser.Username, // Assuming the user has a Username field
	}
	req = req.WithContext(context.WithValue(req.Context(), authorizationPayloadKey, payload))
	req = req.WithContext(context.WithValue(req.Context(), userContextKey, createdUser))

	// Set the content-type header to multipart/form-data with the boundary
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Record a response
	rec := httptest.NewRecorder()

	// call the handler
	app.CreateEvidenceHandler(rec, req)

	want := http.StatusBadRequest
	if rec.Code != want {
		t.Errorf("expected status code %d, got %d. Response body: %s", want, rec.Code, rec.Body.String())
	}
}

func TestCreateEvidenceHandlerWithMissingEvidenceData(t *testing.T) {
	app, createdUser, _ := NewTestEvidenceServer(t)

	// Create a buffer to hold the multipart form data
	body := &bytes.Buffer{}

	// Create a multipart writer to write the form data to the buffer
	writer := multipart.NewWriter(body)

	// No evidence data is added to the request body, simulating missing evidence data.

	// Create a form file for the file data
	_, err := writer.CreateFormFile("upload_file", "evidence.txt")
	if err != nil {
		t.Fatalf("error creating form file: %v", err)
	}

	// Close the multipart writer
	err = writer.Close()
	if err != nil {
		t.Fatalf("error closing multipart writer: %v", err)
	}

	// Create a request with the multipart form data
	url := fmt.Sprintf("/cases/%s/evidences", "some_case_id") // Change "some_case_id" to a valid one if needed
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}

	// Prepare the payload with the created user's details
	payload := &service.Payload{
		Username: createdUser.Username, // Assuming the user has a Username field
	}
	req = req.WithContext(context.WithValue(req.Context(), authorizationPayloadKey, payload))
	req = req.WithContext(context.WithValue(req.Context(), userContextKey, createdUser))

	// Set the content-type header to multipart/form-data with the boundary
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Record a response
	rec := httptest.NewRecorder()

	// call the handler
	app.CreateEvidenceHandler(rec, req)

	want := http.StatusBadRequest
	if rec.Code != want {
		t.Errorf("expected status code %d, got %d. Response body: %s", want, rec.Code, rec.Body.String())
	}
}

func TestGetEvidenceHandler(t *testing.T) {
	tests := []struct {
		name   string
		caseID uuid.UUID // Modified to UUID to match your other functions
		evID   uuid.UUID // Evidence ID should be UUID
		want   int
	}{
		{
			name: "successful retried the evidence",
			want: http.StatusOK,
		},
		{
			name: "failed when evidence does not exist",
			evID: uuid.New(), // Non-existent ID
			want: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create a new server that has one user and case for testing
			app, createdUser, createdCase := NewTestEvidenceServer(t)

			// Use the caseID from the created case in our test case
			tt.caseID = createdCase.ID

			evidenceTypeID, err := app.stores.DBStore.GetEvidenceIDByType(context.Background(), "Initial Evidence")
			if err != nil {
				t.Fatalf("Error getting EvidenceTypeID: %v", err)
			}

			evParams := service.CreateEvidenceParams{
				CaseID:         tt.caseID,
				AppUserID:      createdUser.ID,
				Name:           "video",
				Description:    "Sample video evidence",
				EvidenceTypeID: evidenceTypeID,
			}

			// Use the returned ID from CreateEvidence
			createdEvidence, err := app.stores.CreateEvidence(context.Background(), evParams, bytes.NewBuffer([]byte(`Sample video evidence`)))
			if err != nil {
				t.Fatalf("error creating evidence: %v", err)
			}

			// Check if we are in the non-existent ID scenario
			if tt.want != http.StatusNotFound {
				tt.evID = createdEvidence.ID
			}

			// Create a request
			url := fmt.Sprintf("/cases/%s/evidences/%s", tt.caseID, tt.evID)
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				t.Fatalf("error creating request: %v", err)
			}

			// Prepare the route context with necessary URL parameters
			rct := chi.NewRouteContext()
			rct.URLParams.Add("caseID", tt.caseID.String())
			rct.URLParams.Add("evidenceID", tt.evID.String())

			// Add URL parameters and context to the request
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rct)
			req = req.WithContext(ctx)

			// Record a response
			rec := httptest.NewRecorder()

			// Call the handler
			app.GetEvidenceHandler(rec, req)
			if rec.Code != tt.want {
				t.Errorf("expected status code %d, got %d. Response body: %s", tt.want, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestGetEvidenceHandlerWithInvalidEvidenceID(t *testing.T) {
	// create a new server that has one user and case for testing
	app, _, createdCase := NewTestEvidenceServer(t)

	// Create a request with an invalid evidenceID in the URL
	url := fmt.Sprintf("/api/v1/authenticated/cases/%s/evidences/invalid_evidence_id", createdCase.ID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}

	// Record a response
	rec := httptest.NewRecorder()

	// Prepare the payload
	payload := &service.Payload{
		Username: "test", // you mentioned that NewTestEvidenceServer returns a user, if you need to set more details about the user, do it here.
	}

	// Add payload to the request
	ctx := context.WithValue(req.Context(), authorizationPayloadKey, payload)
	req = req.WithContext(ctx)

	// call the handler
	app.GetEvidenceHandler(rec, req)

	want := http.StatusBadRequest
	if rec.Code != want {
		t.Errorf("expected status code %d, got %d. Response body: %s", want, rec.Code, rec.Body.String())
	}
}

func TestNoDuplicates(t *testing.T) {
	t.Parallel()
	tests := []struct {
		values []int
		want   bool
	}{
		{[]int{1, 2, 3, 4, 5}, true},
		{[]int{1, 2, 3, 4, 5, 5}, false},
		{[]int{}, true},
		{[]int{1, 1}, false},
	}

	for _, tt := range tests {
		if got := NoDuplicates(tt.values); got != tt.want {
			t.Errorf("NoDuplicates(%v) = %v; want %v", tt.values, got, tt.want)
		}
	}
}

func TestIsEmail(t *testing.T) {
	t.Parallel()
	tests := []struct {
		value string
		want  bool
	}{
		{"test@example.com", true},
		{"test@", false},
		{"", false},
		{"test@.com", false},
		{"test@domain", false},
		{"test@domain.c", true},
		{strings.Repeat("a", 244) + "@test.com", true},
		{strings.Repeat("a", 245) + "@testlongerdomain.com", false},
	}
	for _, tt := range tests {
		if got := IsEmail(tt.value); got != tt.want {
			t.Errorf("IsEmail(%q) = %v; want %v", tt.value, got, tt.want)
		}
	}
}

func TestIsURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		value string
		want  bool
	}{
		{"https://example.com", true},
		{"https://example.com", true},
		{"https://", false},
		{"", false},
		{"ftp://example.com", true},
		{"example.com", false},
	}

	for _, tt := range tests {
		if got := IsURL(tt.value); got != tt.want {
			t.Errorf("IsURL(%q) = %v; want %v", tt.value, got, tt.want)
		}
	}
}

func TestGetEvidenceContentHandler(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        int
	}{
		{
			name:        "successful retrieval of the evidence",
			description: "Sample video evidence",
			want:        http.StatusOK,
		},
		{
			name:        "failed when evidence does not exist",
			description: "Non-existent evidence",
			want:        http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create a new server that has one user and case for testing
			app, createdUser, createdCase := NewTestEvidenceServer(t)

			evidenceTypeID, err := app.stores.DBStore.GetEvidenceIDByType(context.Background(), "Initial Evidence")
			if err != nil {
				t.Fatalf("Error getting EvidenceTypeID: %v", err)
			}

			evParams := service.CreateEvidenceParams{
				CaseID:         createdCase.ID,
				AppUserID:      createdUser.ID,
				Name:           "video",
				Description:    tt.description,
				EvidenceTypeID: evidenceTypeID,
			}

			// Use the returned ID from CreateEvidence
			createdEvidence, err := app.stores.CreateEvidence(context.Background(), evParams, bytes.NewBuffer([]byte(tt.description)))
			if err != nil {
				t.Fatalf("error creating evidence: %v", err)
			}

			// Check if we are in the non-existent ID scenario
			evidenceID := createdEvidence.ID
			if tt.want == http.StatusNotFound {
				evidenceID = uuid.New() // Non-existent ID
			}

			// Create a request
			url := fmt.Sprintf("/cases/%s/evidences/%s/content", createdCase.ID, evidenceID)
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				t.Fatalf("error creating request: %v", err)
			}

			// Prepare the route context with necessary URL parameters
			rct := chi.NewRouteContext()
			rct.URLParams.Add("caseID", createdCase.ID.String())
			rct.URLParams.Add("evidenceID", evidenceID.String())
			// Add URL parameters and context to the request
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rct)
			req = req.WithContext(ctx)

			// Record a response
			rec := httptest.NewRecorder()

			// Call the handler
			app.GetEvidenceHandler(rec, req)

			if rec.Code != tt.want {
				t.Errorf("expected status code %d, got %d. Response body: %s", tt.want, rec.Code, rec.Body.String())
			}

			if tt.want == http.StatusOK {
				var result struct {
					Evidence db.Evidence `json:"Evidence"`
				}
				err = json.NewDecoder(rec.Body).Decode(&result)
				if err != nil {
					t.Fatalf("error decoding response: %v", err)
				}
				if result.Evidence.Description.String != tt.description {
					t.Errorf("expected content '%s', got '%s'", tt.description, result.Evidence.Description.String)
				}
				if result.Evidence.Name != evParams.Name {
					t.Errorf("expected name '%s', got '%s'", evParams.Name, result.Evidence.Name)
				}
			}
		})
	}
}

func TestListEvidencesHandler(t *testing.T) {
	// Create a new server with one user and case for testing
	app, createdUser, createdCase := NewTestEvidenceServer(t)

	evidenceTypeID, err := app.stores.DBStore.GetEvidenceIDByType(context.Background(), "Initial Evidence")
	if err != nil {
		t.Fatalf("Error getting EvidenceTypeID: %v", err)
	}

	// Create two evidences
	evidenceNames := []string{"video1", "video2"}
	for _, name := range evidenceNames {
		evParams := service.CreateEvidenceParams{
			CaseID:         createdCase.ID,
			AppUserID:      createdUser.ID,
			Name:           name,
			Description:    "Sample video evidence",
			EvidenceTypeID: evidenceTypeID,
		}

		_, err = app.stores.CreateEvidence(context.Background(), evParams, bytes.NewBuffer([]byte("Sample video evidence")))
		if err != nil {
			t.Fatalf("error creating evidence: %v", err)
		}
	}

	// Create a request
	url := fmt.Sprintf("/cases/%s/evidences", createdCase.ID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}

	// Prepare the route context with necessary URL parameters
	rct := chi.NewRouteContext()
	rct.URLParams.Add("caseID", createdCase.ID.String())
	// Add URL parameters and context to the request
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rct)
	req = req.WithContext(ctx)

	// Record a response
	rec := httptest.NewRecorder()

	// Call the handler
	app.ListEvidencesHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d. Response body: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	// Decode response body
	var result struct {
		Evidences []db.Evidence `json:"evidences"`
	}

	err = json.NewDecoder(rec.Body).Decode(&result)
	if err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	// Check the number of returned evidences
	if len(result.Evidences) != len(evidenceNames) {
		t.Errorf("expected %d evidences, got %d", len(evidenceNames), len(result.Evidences))
	}

	// Check the content of the returned evidences
	for i, evidence := range result.Evidences {
		if evidence.Name != evidenceNames[i] {
			t.Errorf("expected evidence name '%s', got '%s'", evidenceNames[i], evidence.Name)
		}
	}
}

func TestDownloadEvidenceHandler(t *testing.T) {
	wantContent := "Sample video evidence" // expected content of the evidence

	// Create a new server with one user and case for testing
	app, createdUser, createdCase := NewTestEvidenceServer(t)

	evidenceTypeID, err := app.stores.DBStore.GetEvidenceIDByType(context.Background(), "Initial Evidence")
	if err != nil {
		t.Fatalf("Error getting EvidenceTypeID: %v", err)
	}

	// Create an evidence
	evParams := service.CreateEvidenceParams{
		CaseID:         createdCase.ID,
		AppUserID:      createdUser.ID,
		Name:           "video",
		Description:    "Sample video evidence",
		EvidenceTypeID: evidenceTypeID,
	}

	createdEvidence, err := app.stores.CreateEvidence(context.Background(), evParams, bytes.NewBuffer([]byte(wantContent)))
	if err != nil {
		t.Fatalf("error creating evidence: %v", err)
	}

	// Create a request
	url := fmt.Sprintf("/cases/%s/evidences/%s/download", createdCase.ID, createdEvidence.ID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}

	// Prepare the route context with necessary URL parameters
	rct := chi.NewRouteContext()
	rct.URLParams.Add("caseID", createdCase.ID.String())
	rct.URLParams.Add("evidenceID", createdEvidence.ID.String())
	// Add URL parameters and context to the request
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rct)
	req = req.WithContext(ctx)

	// Record a response
	rec := httptest.NewRecorder()

	// Call the handler
	app.DownloadEvidenceHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d. Response body: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	// Check the content of the downloaded evidence
	receivedEvidence := rec.Body.String()
	if receivedEvidence != wantContent {
		t.Errorf("expected evidence content '%s', got '%s'", wantContent, receivedEvidence)
	}
}

func TestDownloadEvidenceHandlerReturns404(t *testing.T) {
	// Create a new server with one user and case for testing
	app, _, createdCase := NewTestEvidenceServer(t)

	// Non-existing evidence ID
	nonExistentEvID := uuid.New() // Choose an ID that certainly doesn't exist in your test setup

	// Create a request
	url := fmt.Sprintf("/cases/%s/evidences/%s/download", createdCase.ID, nonExistentEvID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}

	// Prepare the route context with necessary URL parameters
	rct := chi.NewRouteContext()
	rct.URLParams.Add("caseID", createdCase.ID.String())
	rct.URLParams.Add("evidenceID", nonExistentEvID.String())
	// Add URL parameters and context to the request
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rct)
	req = req.WithContext(ctx)

	// Record a response
	rec := httptest.NewRecorder()

	// Call the handler
	app.DownloadEvidenceHandler(rec, req)

	// Check if error status code is returned
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status code %d, got %d. Response body: %s", http.StatusNotFound, rec.Code, rec.Body.String())
	}
}
