package api

import (
	"bytes"
	"context"
	"encoding/json"
	"evidence/internal/data"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreatingCases(t *testing.T) {
	tests := []struct {
		name        string
		requestBody map[string]interface{}
		wantStatus  int
	}{
		{
			name: "with a valid name without special characters succeeded",
			requestBody: map[string]interface{}{
				"name": "testcase",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "with a invalid name with special characters failed",
			requestBody: map[string]interface{}{
				"name": "OSPGK25/22",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "with invalid name with special characters failed",
			requestBody: map[string]interface{}{
				"name": "OSPG-K-25-22",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "with UPPERCASE characters failed",
			requestBody: map[string]interface{}{
				"name": "OSPGK2522",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "with LOWERCASE and allowed special characters succeeded",
			requestBody: map[string]interface{}{
				"name": "pspg-k-25-22",
			},
			wantStatus: http.StatusCreated,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestServer(t)
			requestBody, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatal(err)
			}
			user := &data.User{
				Username: "test",
			}
			err = user.Password.Set("test")
			if err != nil {
				t.Fatal(err)
			}
			err = app.stores.UserDB.Add(user)
			if err != nil {
				t.Fatal(err)
			}
			payload := &Payload{
				Username: "test",
			}
			//Making a request
			request := httptest.NewRequest("POST", "/cases", bytes.NewBuffer(requestBody))
			//Adding payload ctx to request
			ctx := context.WithValue(request.Context(), authorizationPayloadKey, payload)
			reqWithPayload := request.WithContext(ctx)
			//Recording the response
			response := httptest.NewRecorder()
			app.CreateCaseHandler(response, reqWithPayload)
			//Checking the status code
			if response.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
			}
		})
	}
}
func TestBadJSONRequestFailsOnPOST(t *testing.T) {
	app := newTestServer(t)
	cases := []struct {
		name        string
		requestBody []byte
		handler     func(w http.ResponseWriter, r *http.Request)
	}{
		{
			name:        "create case handler",
			requestBody: []byte("{}"),
			handler:     app.CreateCaseHandler,
		},
		{
			name:        "remove case handler",
			requestBody: []byte("{}"),
			handler:     app.RemoveCaseHandler,
		},
		{
			name:        "create evidence handler",
			requestBody: []byte("{}"),
			handler:     app.CreateEvidenceHandler,
		},
		{
			name:        "delete evidence handler",
			requestBody: []byte("{}"),
			handler:     app.DeleteEvidenceHandler,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			user := &data.User{
				Username: "test",
			}
			err := user.Password.Set("test")
			if err != nil {
				t.Fatal(err)
			}
			err = app.stores.UserDB.Add(user)
			if err != nil {
				t.Fatal(err)
			}
			payload := &Payload{
				Username: "test",
			}
			request := httptest.NewRequest("POST", "/cases", bytes.NewBuffer(tt.requestBody))
			ctx := context.WithValue(request.Context(), authorizationPayloadKey, payload)
			reqWithPayload := request.WithContext(ctx)
			response := httptest.NewRecorder()
			app.CreateCaseHandler(response, reqWithPayload)
			if response.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, response.Code)
			}
		})
	}
}
func TestCaseInFSShouldBeRemovedIfCreationInDBFails(t *testing.T) {
	app := newTestServer(t)
	requestBody := map[string]interface{}{
		"name": "testcase",
	}
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequest("POST", "/cases", bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		t.Fatal(err)
	}
	payload := &Payload{
		Username: "test",
	}
	//Adding payload ctx to request
	ctx := context.WithValue(request.Context(), authorizationPayloadKey, payload)
	reqWithPayload := request.WithContext(ctx)
	response := httptest.NewRecorder()
	app.CreateCaseHandler(response, reqWithPayload)
	err = app.stores.StoreFS.RemoveCase("testcase")
	if err.Error() != "case does not exist" {
		t.Errorf("expected case not found, got %v", err)
	}

}
func TestCaseDeletion(t *testing.T) {
	tests := []struct {
		name        string
		caseToAdd   string
		requestBody map[string]interface{}
		wantStatus  int
	}{
		{
			name:      "with existing case succeeded",
			caseToAdd: "ssss",
			requestBody: map[string]interface{}{
				"name": "ssss",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "with non existing case failed",
			caseToAdd: "ssss",
			requestBody: map[string]interface{}{
				"name": "OSPGK25-22",
			},
			wantStatus: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestServer(t)
			requestBody, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatal(err)
			}
			user := &data.User{
				Username: "test",
			}
			err = user.Password.Set("test")
			if err != nil {
				t.Fatal(err)
			}
			err = app.stores.UserDB.Add(user)
			if err != nil {
				t.Fatal(err)
			}
			user.ID = 1
			payload := &Payload{
				Username: "test",
			}
			//adding the cases to the database and storage
			cs := &data.Case{
				Name: tt.caseToAdd,
			}
			err = app.stores.CaseDB.Add(cs, user)
			if err != nil {
				t.Fatal(err)
			}
			err = app.stores.StoreFS.AddCase(cs)
			if err != nil {
				t.Fatal(err)
			}
			//Making a delete request
			request := httptest.NewRequest("DELETE", "/cases", bytes.NewBuffer(requestBody))
			//Adding payload ctx to request
			ctx := context.WithValue(request.Context(), authorizationPayloadKey, payload)
			reqWithPayload := request.WithContext(ctx)
			//Recording the response
			response := httptest.NewRecorder()
			app.RemoveCaseHandler(response, reqWithPayload)
			//Checking the status code
			if response.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
			}

		})
	}
}
func TestDeleteCaseThatDoesNotExistInDBFailed(t *testing.T) {
	app := newTestServer(t)
	user := &data.User{
		Username: "test",
	}
	err := user.Password.Set("test")
	if err != nil {
		t.Fatal(err)
	}
	err = app.stores.UserDB.Add(user)
	if err != nil {
		t.Fatal(err)
	}
	user.ID = 1
	payload := &Payload{
		Username: "test",
	}
	//adding the cases to the database and storage
	cs := &data.Case{
		Name: "ssss",
	}
	err = app.stores.StoreFS.AddCase(cs)
	if err != nil {
		t.Fatal(err)
	}
	requestBody := map[string]interface{}{
		"name": "ssss",
	}
	body, err := json.Marshal(requestBody)
	//Making a delete request
	request := httptest.NewRequest("DELETE", "/cases", bytes.NewBuffer(body))
	//Adding payload ctx to request
	ctx := context.WithValue(request.Context(), authorizationPayloadKey, payload)
	reqWithPayload := request.WithContext(ctx)
	//Recording the response
	response := httptest.NewRecorder()
	app.RemoveCaseHandler(response, reqWithPayload)
	//Checking the status code
	if response.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}
func TestDeleteCaseThatDoesNotExistInFSFailed(t *testing.T) {
	app := newTestServer(t)
	user := &data.User{
		Username: "test",
	}
	err := user.Password.Set("test")
	if err != nil {
		t.Fatal(err)
	}
	err = app.stores.UserDB.Add(user)
	if err != nil {
		t.Fatal(err)
	}
	user.ID = 1
	payload := &Payload{
		Username: "test",
	}
	//adding the cases to the database and storage
	cs := &data.Case{
		Name: "ssss",
	}
	err = app.stores.CaseDB.Add(cs, user)
	if err != nil {
		t.Fatal(err)
	}
	requestBody := map[string]interface{}{
		"name": "ssss",
	}
	body, err := json.Marshal(requestBody)
	//Making a delete request
	request := httptest.NewRequest("DELETE", "/cases", bytes.NewBuffer(body))
	//Adding payload ctx to request
	ctx := context.WithValue(request.Context(), authorizationPayloadKey, payload)
	reqWithPayload := request.WithContext(ctx)
	//Recording the response
	response := httptest.NewRecorder()
	app.RemoveCaseHandler(response, reqWithPayload)
	//Checking the status code
	if response.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}
func TestListedAllCasesThatExistInBoughtDBAndStorage(t *testing.T) {
	//Creating a test server
	app := newTestServer(t)
	user := &data.User{
		Username: "test",
	}
	//Adding a user to the database
	err := user.Password.Set("test")
	if err != nil {
		t.Fatal(err)
	}
	err = app.stores.UserDB.Add(user)
	if err != nil {
		t.Fatal(err)
	}
	user.ID = 1
	//adding the cases to the database and storage
	type CaseListResponse struct {
		Cases []data.Case `json:"cases"`
	}
	want := CaseListResponse{
		Cases: []data.Case{{Name: "pspg-k-25-22"}, {Name: "pspg-k-25-23"}, {Name: "pspg-k-25-24"}}}

	for _, cs := range want.Cases {
		err = app.stores.CaseDB.Add(&cs, user)
		if err != nil {
			t.Fatal(err)
		}
		err = app.stores.StoreFS.AddCase(&cs)
		if err != nil {
			t.Fatal(err)
		}
	}
	//Add additional case to storage to test filtering
	err = app.stores.StoreFS.AddCase(&data.Case{
		Name: "pspg-k-25-25",
	})
	payload := &Payload{
		Username: "test",
	}
	//Making a get request
	request := httptest.NewRequest("GET", "/cases", nil)
	//Adding payload ctx to request
	ctx := context.WithValue(request.Context(), authorizationPayloadKey, payload)
	reqWithPayload := request.WithContext(ctx)
	//Recording the response
	response := httptest.NewRecorder()
	app.ListCasesHandler(response, reqWithPayload)

	var got CaseListResponse
	err = json.NewDecoder(response.Body).Decode(&got)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(got, want, cmpopts.IgnoreFields(data.Case{}, "ID")) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestListedAllCasesFailedForNonAdminUser(t *testing.T) {
	//Creating a test server
	app := newTestServer(t)
	user := &data.User{
		Username: "test",
	}
	//Adding a user to the database
	err := user.Password.Set("test")
	if err != nil {
		t.Fatal(err)
	}
	user.Role = "normal"
	err = app.stores.UserDB.Add(user)
	if err != nil {
		t.Fatal(err)
	}
	user.ID = 1
	//adding the cases to the database and storage
	type CaseListResponse struct {
		Cases []data.Case `json:"cases"`
	}
	want := CaseListResponse{
		Cases: []data.Case{{Name: "pspg-k-25-22"}, {Name: "pspg-k-25-23"}, {Name: "pspg-k-25-24"}}}

	for _, cs := range want.Cases {
		err = app.stores.CaseDB.Add(&cs, user)
		if err != nil {
			t.Fatal(err)
		}
		err = app.stores.StoreFS.AddCase(&cs)
		if err != nil {
			t.Fatal(err)
		}
	}
	payload := &Payload{
		Username: "test",
	}
	//Making a get request
	request := httptest.NewRequest("GET", "/cases", nil)
	//Adding payload ctx to request
	ctx := context.WithValue(request.Context(), authorizationPayloadKey, payload)
	reqWithPayload := request.WithContext(ctx)
	//Recording the response
	response := httptest.NewRecorder()
	app.ListCasesHandler(response, reqWithPayload)
	if response.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, response.Code)
	}

}
