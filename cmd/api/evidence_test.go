package api

import (
	"bytes"
	"context"
	"encoding/json"
	"evidence/internal/data"
	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

// seedForHandlerTesting seeds the database with one user and one case for testing
func seedForHandlerTesting(t *testing.T, app *Application) {
	// get new test server
	user := &data.User{
		Username: "test",
	}
	user.Password.Set("test")
	err := app.stores.UserDB.Add(user)
	if err != nil {
		t.Fatal(err)
	}
	// create a case
	cs := &data.Case{
		Name: "test",
	}
	user.ID = 1
	err = app.stores.CaseDB.Add(cs, user)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	err = app.stores.StoreFS.AddCase(cs)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
}

func TestCreateEvidenceHandler(t *testing.T) {
	tests := []struct {
		name                 string
		alreadyAddedEvidence *data.Evidence
		evidenceNameToAdd    string
		caseID               string
		want                 int
	}{
		{
			name: "successfully created with correct request",
			alreadyAddedEvidence: &data.Evidence{
				CaseID: 1,
				Name:   "video",
				File:   bytes.NewBufferString("test"),
			},
			caseID:            "1",
			evidenceNameToAdd: "picture",
			want:              http.StatusOK,
		},
		{
			name: "that already exists returns an error",
			alreadyAddedEvidence: &data.Evidence{
				CaseID: 1,
				Name:   "video",
				File:   bytes.NewBufferString("test"),
			},
			caseID:            "1",
			evidenceNameToAdd: "video",
			want:              http.StatusConflict,
		},
		{
			name: "with no file attached returns an error",
			alreadyAddedEvidence: &data.Evidence{
				CaseID: 1,
				Name:   "video",
				File:   bytes.NewBufferString("test"),
			},
			caseID:            "1",
			evidenceNameToAdd: "",
			want:              http.StatusBadRequest,
		},
		{
			name: "with wrong caseID format fails",
			alreadyAddedEvidence: &data.Evidence{
				CaseID: 1,
				Name:   "video",
				File:   bytes.NewBufferString("test"),
			},
			evidenceNameToAdd: "picture",
			want:              http.StatusBadRequest,
		},
		{
			name: "with not casID that doesn't exist fails",
			alreadyAddedEvidence: &data.Evidence{
				CaseID: 1,
				Name:   "video",
				File:   bytes.NewBufferString("test"),
			},
			caseID:            "2",
			evidenceNameToAdd: "picture",
			want:              http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create a new test server
			app := newTestServer(t)
			// seed the database with one user,case and evidence for testing
			seedForHandlerTesting(t, app)
			// add evidence to the database
			_, err := app.stores.EvidenceDB.Create(tt.alreadyAddedEvidence)
			if err != nil {
				t.Errorf("failed to create evidence: %v", err)
			}
			// create a body with multipart form data
			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)
			part, _ := writer.CreateFormFile("uploadfile", tt.evidenceNameToAdd)
			part.Write([]byte(`sample-content`))
			writer.Close()
			// create a request
			req, err := http.NewRequest("POST", "/", body)
			if err != nil {
				t.Error(err)
			}
			// set the content-type header
			req.Header.Set("Content-Type", writer.FormDataContentType())
			// record a response
			rec := httptest.NewRecorder()
			payload := &Payload{
				Username: "test",
			}
			// add payload, URL parms and context
			rct := chi.NewRouteContext()
			rct.URLParams.Add("caseID", tt.caseID)
			ctx := context.WithValue(req.Context(), authorizationPayloadKey, payload)
			ctx = context.WithValue(ctx, chi.RouteCtxKey, rct)
			req = req.WithContext(ctx)
			app.CreateEvidenceHandler(rec, req)
			if rec.Code != tt.want {
				t.Errorf("expected status code %d, got %d", tt.want, rec.Code)
			}
		})
	}
}
func TestRetrieveEvidence(t *testing.T) {
	tests := []struct {
		name                 string
		alreadyAddedEvidence *data.Evidence
		caseID               string
		evidenceIDToGet      string
		want                 int
	}{
		{
			name: "successfully with correct request",
			alreadyAddedEvidence: &data.Evidence{
				CaseID: 1,
				Name:   "video",
				File:   bytes.NewBufferString("test"),
			},
			caseID:          "1",
			evidenceIDToGet: "1",
			want:            http.StatusOK,
		},
		{
			name: "with wrong caseID format fails",
			alreadyAddedEvidence: &data.Evidence{
				CaseID: 1,
				Name:   "video",
				File:   bytes.NewBufferString("test"),
			},
			caseID:          "2b",
			evidenceIDToGet: "1",
			want:            http.StatusBadRequest,
		},
		{
			name: "with not caseID that doesn't exist fails",
			alreadyAddedEvidence: &data.Evidence{
				CaseID: 1,
				Name:   "video",
				File:   bytes.NewBufferString("test"),
			},
			caseID:          "2",
			evidenceIDToGet: "1",
			want:            http.StatusNotFound,
		},
		{
			name: "with invalid evidence ID format fails",
			alreadyAddedEvidence: &data.Evidence{
				CaseID: 1,
				Name:   "video",
				File:   bytes.NewBufferString("test"),
			},
			caseID:          "1",
			evidenceIDToGet: "4b",
			want:            http.StatusBadRequest,
		},
		{
			name: "with not evidence ID that doesn't exist fails",
			alreadyAddedEvidence: &data.Evidence{
				CaseID: 1,
				Name:   "video",
				File:   bytes.NewBufferString("test"),
			},
			caseID:          "1",
			evidenceIDToGet: "4",
			want:            http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create a new test server
			app := newTestServer(t)
			// seed the database with one user,case and evidenceIDToGet for testing
			seedForHandlerTesting(t, app)
			// add evidenceIDToGet to the database

			hash, err := app.stores.StoreFS.AddEvidence(tt.alreadyAddedEvidence, "test", tt.alreadyAddedEvidence.File)
			if err != nil {
				return
			}
			tt.alreadyAddedEvidence.Hash = hash
			_, err = app.stores.EvidenceDB.Create(tt.alreadyAddedEvidence)
			if err != nil {
				t.Errorf("failed to create evidence: %v", err)
			}
			// create a request
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Error(err)
			}
			// record a response
			rec := httptest.NewRecorder()
			payload := &Payload{
				Username: "test",
			}
			// add payload, URL parms and context
			rct := chi.NewRouteContext()
			rct.URLParams.Add("caseID", tt.caseID)
			rct.URLParams.Add("evidenceID", tt.evidenceIDToGet)
			ctx := context.WithValue(req.Context(), authorizationPayloadKey, payload)
			ctx = context.WithValue(ctx, chi.RouteCtxKey, rct)
			req = req.WithContext(ctx)
			app.GetEvidenceHandler(rec, req)
			if rec.Code != tt.want {
				t.Errorf("expected status code %d, got %d", tt.want, rec.Code)
			}
		})
	}

}
func TestDeleteEvidence(t *testing.T) {
	tests := []struct {
		name             string
		addEvidence      *data.Evidence
		caseID           string
		deleteEvidenceID string
		want             int
	}{
		{
			name: "successfully deleted",
			addEvidence: &data.Evidence{
				CaseID: 1,
				Name:   "video",
				File:   bytes.NewBufferString("test"),
			},
			caseID:           "1",
			deleteEvidenceID: "1",
			want:             http.StatusOK,
		},
		{
			name: "that doesn't exist returns an error",
			addEvidence: &data.Evidence{
				CaseID: 1,
				Name:   "video",
				File:   bytes.NewBufferString("test"),
			},
			caseID:           "1",
			deleteEvidenceID: "2",
			want:             http.StatusNotFound,
		},
		{
			name: "with wrong caseID format fails",
			addEvidence: &data.Evidence{
				CaseID: 1,
				Name:   "video",
				File:   bytes.NewBufferString("test"),
			},
			caseID:           "1",
			deleteEvidenceID: "1b",
			want:             http.StatusBadRequest,
		},
		{
			name: "with caseID that doesn't exist fails",
			addEvidence: &data.Evidence{
				CaseID: 1,
				Name:   "video",
				File:   bytes.NewBufferString("test"),
			},
			caseID:           "2",
			deleteEvidenceID: "1",
			want:             http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create a new test server
			app := newTestServer(t)
			// seed the database with one user,case and evidence for testing
			seedForHandlerTesting(t, app)
			// add evidence to the database
			hash, err := app.stores.StoreFS.AddEvidence(tt.addEvidence, "test", tt.addEvidence.File)
			if err != nil {
				return
			}
			tt.addEvidence.Hash = hash
			_, err = app.stores.EvidenceDB.Create(tt.addEvidence)
			if err != nil {
				t.Errorf("failed to create evidence: %v", err)
			}
			// create a request
			req, err := http.NewRequest("DELETE", "/", nil)
			if err != nil {
				t.Error(err)
			}
			// record a response
			rec := httptest.NewRecorder()
			payload := &Payload{
				Username: "test",
			}
			// add payload, URL parms and context
			rct := chi.NewRouteContext()
			rct.URLParams.Add("caseID", tt.caseID)
			rct.URLParams.Add("evidenceID", tt.deleteEvidenceID)
			ctx := context.WithValue(req.Context(), authorizationPayloadKey, payload)
			ctx = context.WithValue(ctx, chi.RouteCtxKey, rct)
			req = req.WithContext(ctx)
			app.DeleteEvidenceHandler(rec, req)
			if rec.Code != tt.want {
				t.Errorf("expected status code %d, got %d", tt.want, rec.Code)
			}
		})

	}
}
func TestListingEvidencesFromCaseReturnsCorrectNumberOfEvidences(t *testing.T) {
	// create a new test server
	app := newTestServer(t)
	// seed the database with one user and case  for testing
	seedForHandlerTesting(t, app)
	// create a request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}
	// record a response
	rec := httptest.NewRecorder()
	payload := &Payload{
		Username: "test",
	}
	// declare evidences to add
	want := []*data.Evidence{
		{
			CaseID: 1,
			Name:   "video",
			File:   bytes.NewBufferString("test"),
		},
		{
			CaseID: 1,
			Name:   "picture",
			File:   bytes.NewBufferString("test1"),
		},
		{
			CaseID: 1,
			Name:   "audio",
			File:   bytes.NewBufferString("test2"),
		},
	}
	// add evidence to the database and FS
	for _, evidence := range want {
		hash, err := app.stores.StoreFS.AddEvidence(evidence, "test", evidence.File)
		if err != nil {
			return
		}
		evidence.Hash = hash
		_, err = app.stores.EvidenceDB.Create(evidence)
		if err != nil {
			t.Errorf("failed to create evidence: %v", err)
		}
	}
	// add payload, URL params and context
	rct := chi.NewRouteContext()
	rct.URLParams.Add("caseID", "1")
	ctx := context.WithValue(req.Context(), authorizationPayloadKey, payload)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rct)
	req = req.WithContext(ctx)
	// call the handler
	app.ListEvidencesHandler(rec, req)
	// check the response
	type response struct {
		Evidences []*data.Evidence `json:"evidences"`
	}
	var got response
	err = json.NewDecoder(rec.Body).Decode(&got)
	if err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	// check for correct status code
	if rec.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rec.Code)
	}
	// check for the correct number of evidences
	if len(got.Evidences) != len(want) {
		t.Errorf("expected %d evidences, got %d", len(want), len(got.Evidences))
	}
	// check for the correct evidences
	if !cmp.Equal(got.Evidences, want, cmpopts.IgnoreFields(data.Evidence{}, "File")) {
		t.Errorf(cmp.Diff(want, got.Evidences))
	}
}
func TestListingEvidencesInTheCase(t *testing.T) {
	tests := []struct {
		name   string
		caseID string
		want   int
	}{
		{
			name:   "with caseID that doesn't exist fails",
			caseID: "2",
			want:   http.StatusNotFound,
		},
		{
			name:   "with invalid caseID format fails",
			caseID: "a",
			want:   http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		app := newTestServer(t)
		// seed the database with one user and case  for testing
		seedForHandlerTesting(t, app)
		// create a request
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Error(err)
		}
		// record a response
		rec := httptest.NewRecorder()
		payload := &Payload{
			Username: "test",
		}
		// add payload, URL params and context
		rct := chi.NewRouteContext()
		rct.URLParams.Add("caseID", tt.caseID)
		ctx := context.WithValue(req.Context(), authorizationPayloadKey, payload)
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rct)
		req = req.WithContext(ctx)
		// call the handler
		app.ListEvidencesHandler(rec, req)
		// check the response
		if rec.Code != tt.want {
			t.Errorf("expected status code %d, got %d", tt.want, rec.Code)
		}
	}

}

func TestAddCommentsToEvidences(t *testing.T) {
	tests := []struct {
		name        string
		requestBody map[string]interface{}
		evidenceID  string
		want        int
	}{
		{
			name: "with valid evidenceID and comment",
			requestBody: map[string]interface{}{
				"text": "something to say",
			},
			evidenceID: "1",
			want:       http.StatusOK,
		},
		{
			name: "with invalid evidenceID fails",
			requestBody: map[string]interface{}{
				"text": "something to say",
			},
			evidenceID: "3",
			want:       http.StatusNotFound,
		},
		{
			name: "with invalid evidenceID format fails",
			requestBody: map[string]interface{}{
				"text": "something to say",
			},
			evidenceID: "a",
			want:       http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create a new test server
			app := newTestServer(t)
			// seed the database with user and case for testing
			seedForHandlerTesting(t, app)
			// add evidences
			evidences := []*data.Evidence{
				{
					CaseID: 1,
					Name:   "video",
					File:   bytes.NewBufferString("test1"),
				},
				{
					CaseID: 1,
					Name:   "picture",
					File:   bytes.NewBufferString("test2"),
				},
			}
			for _, evidence := range evidences {
				hash, err := app.stores.StoreFS.AddEvidence(evidence, "test", evidence.File)
				if err != nil {
					return
				}
				evidence.Hash = hash
				_, err = app.stores.EvidenceDB.Create(evidence)
				if err != nil {
					t.Errorf("failed to create evidence: %v", err)
				}
			}
			// marshal the contents of the request body
			requestBody, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatal(err)
			}
			// create a request
			req, err := http.NewRequest("POST", "/", bytes.NewBuffer(requestBody))
			if err != nil {
				t.Error(err)
			}
			// record a response
			rec := httptest.NewRecorder()
			payload := &Payload{
				Username: "test",
			}
			// add payload, URL params and context
			rct := chi.NewRouteContext()
			rct.URLParams.Add("evidenceID", tt.evidenceID)
			ctx := context.WithValue(req.Context(), authorizationPayloadKey, payload)
			ctx = context.WithValue(ctx, chi.RouteCtxKey, rct)
			req = req.WithContext(ctx)
			// call the handler
			app.AddCommentToEvidenceHandler(rec, req)
			// check for correct status code
			if rec.Code != tt.want {
				t.Errorf("expected status code %d, got %d", tt.want, rec.Code)
			}
		})
	}
}