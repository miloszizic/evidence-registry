package api

import (
	"bytes"
	"encoding/json"
	"github.com/miloszizic/der/internal/data"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPingHandlerReturnedAPong(t *testing.T) {
	app := newTestServer(t)

	ping, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Errorf("Failed to create ping request: %s", err)
	}
	response := httptest.NewRecorder()
	app.Ping(response, ping)
	if response.Body.String() != "pong" {
		t.Errorf("Expected pong, got %s", response.Body.String())
	}

}
func TestLogin(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name: "successful with correct credentials",
			requestBody: map[string]interface{}{
				"username": "Simba",
				"password": "opsAdmin",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "unsuccessful with incorrect credentials",
			requestBody: map[string]interface{}{
				"username": "Simba",
				"password": "opsAdmin1",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "unsuccessful with missing username",
			requestBody: map[string]interface{}{
				"password": "opsAdmin",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "unsuccessful with missing password",
			requestBody: map[string]interface{}{
				"username": "Simba",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "unsuccessful with missing username and password",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestServer(t)
			user := &data.User{Username: "Simba"}
			err := user.Password.Set("opsAdmin")
			if err != nil {
				t.Errorf("failed to set password: %v", err)
			}
			err = app.stores.User.Add(user)
			if err != nil {
				t.Errorf("Failed to add user: %s", err)
			}
			body, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Errorf("Failed to marshal request body: %s", err)
			}
			response := httptest.NewRecorder()
			request, err := http.NewRequest("POST", "/Login", bytes.NewReader(body))
			if err != nil {
				t.Errorf("Error creating a new request: %v", err)
			}
			app.Login(response, request)
			if status := response.Code; status != tt.expectedStatus {
				t.Errorf("Handler returned wrong status code. Expected: %d. Got: %d.", tt.expectedStatus, status)
			}
		})
	}
}
func TestLoginWithUserThatDoesNotExistFailed(t *testing.T) {
	app := newTestServer(t)
	//Making test user
	createUserBody := map[string]interface{}{
		"username": "Simba",
		"password": "opsAdmin1",
	}
	requestBody, err := json.Marshal(createUserBody)
	if err != nil {
		t.Errorf("Failed to marshal request body: %s", err)
	}
	request, err := http.NewRequest("POST", "/Login", bytes.NewReader(requestBody))
	if err != nil {
		t.Errorf("Error creating a new request: %v", err)
	}
	response := httptest.NewRecorder()
	app.Login(response, request)
	if status := response.Code; status != http.StatusUnauthorized {
		t.Errorf("Handler returned wrong status code. Expected: %d. Got: %d.", http.StatusUnauthorized, status)
	}
}
func TestLoginFailedWithBadJSONRequest(t *testing.T) {
	app := newTestServer(t)
	response := httptest.NewRecorder()
	request, err := http.NewRequest("POST", "/Login", bytes.NewReader([]byte("badJsonRequest")))
	if err != nil {
		t.Errorf("Error creating a new request: %v", err)
	}
	app.Login(response, request)
	if status := response.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code. Expected: %d. Got: %d.", http.StatusBadRequest, status)
	}
}
func TestCreateUserHandlerCreatedANewUserSuccessfully(t *testing.T) {
	app := newTestServer(t)
	//Making test user
	createUserBody := map[string]interface{}{
		"username": "Simba",
		"password": "opsAdmin1",
	}
	requestBody, err := json.Marshal(createUserBody)
	if err != nil {
		t.Errorf("Failed to marshal request body: %s", err)
	}
	request, err := http.NewRequest("POST", "/register", bytes.NewReader(requestBody))
	if err != nil {
		t.Errorf("Error creating a new request: %v", err)
	}
	response := httptest.NewRecorder()
	app.CreateUserHandler(response, request)
	if status := response.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code. Expected: %d. Got: %d.", http.StatusCreated, status)
	}
}
func TestCreateUserHandlerFailedWithBadJSONRequest(t *testing.T) {
	app := newTestServer(t)
	requestBody, err := json.Marshal("badJsonRequest")
	if err != nil {
		t.Errorf("Failed to marshal request body: %s", err)
	}
	request, err := http.NewRequest("POST", "/register", bytes.NewReader(requestBody))
	if err != nil {
		t.Errorf("Error creating a new request: %v", err)
	}
	response := httptest.NewRecorder()
	app.CreateUserHandler(response, request)
	if status := response.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code. Expected: %d. Got: %d.", http.StatusBadRequest, status)
	}
}
func TestCreateUserHandlerFailedWithEmptyPassword(t *testing.T) {
	app := newTestServer(t)
	//Making test user
	createUserBody := map[string]interface{}{
		"username": "Simba",
		"password": "",
	}
	requestBody, err := json.Marshal(createUserBody)
	if err != nil {
		t.Errorf("Failed to marshal request body: %s", err)
	}
	request, err := http.NewRequest("POST", "/register", bytes.NewReader(requestBody))
	if err != nil {
		t.Errorf("Error creating a new request: %v", err)
	}
	response := httptest.NewRecorder()
	app.CreateUserHandler(response, request)
	if status := response.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code. expected: %d. Got: %d.", http.StatusBadRequest, status)
	}
}
