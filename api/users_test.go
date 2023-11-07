//go:build integration

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/google/uuid"

	"github.com/miloszizic/der/service"
)

func TestCreateUserHandler(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		wantStatus  int
	}{
		{
			name: "with valid user parameters",
			requestBody: `{
				"username": "john",
				"email":    "john@example.com",
				"password": "ComplexPass321**"
			}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:        "fails with invalid JSON body",
			requestBody: "sdfsf",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name: "where password is too simple",
			requestBody: `{
				"username": "john",
				"email":    "invalid-email",
				"password": "123"
			}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "where username is missing",
			requestBody: `{
				"email":    "john@example.com",
				"password": "ComplexPass321**"
			}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "where password is too common",
			requestBody: `{
				"expectedUser": "john",
				"password": "123456789"
			}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewTestServer(t)
			req, err := http.NewRequest("POST", "/register", strings.NewReader(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}

			// Create a response recorder to capture the response
			recorder := httptest.NewRecorder()

			// Call the handler function
			app.CreateUserHandler(recorder, req)

			// Check the response status code
			if recorder.Code != tt.wantStatus {
				t.Errorf("expected status code %d, got %d", tt.wantStatus, recorder.Code)
			}
		})
	}
}

func TestUpdateUserPasswordHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		userID     string
	}{
		{
			name:       "Valid password update",
			body:       `{"password": "ComplexPass321**"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid password update",
			body:       `{"password": "123"}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "Common password update",
			body:       fmt.Sprintf(`{"password": "%s"}`, service.CommonPasswords[0]),
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "Invalid userID",
			body:       `{"password": "ComplexPass321**"}`,
			wantStatus: http.StatusBadRequest,
			userID:     "invalid-uuid",
		},
		{
			name:       "Missing password in body",
			body:       `{}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, user, _ := NewTestEvidenceServer(t)

			reqURL := fmt.Sprintf("/users/%s/password", user.ID.String())
			if tt.userID != "" {
				reqURL = fmt.Sprintf("/users/%s/password", tt.userID)
			}

			req, err := http.NewRequest(http.MethodPut, reqURL, strings.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}

			// Set the user ID in chi's router context
			rctx := chi.NewRouteContext()
			// we need to add the userID to the URL params
			rctx.URLParams.Add("userID", user.ID.String())
			// if we have a custom userID, add it to the URL params instead
			if tt.userID != "" {
				rctx.URLParams.Add("userID", tt.userID)
			}
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create a response recorder to capture the response
			recorder := httptest.NewRecorder()

			// Call the handler function
			app.UpdateUserPasswordHandler(recorder, req)

			// Check the response status code
			if recorder.Code != tt.wantStatus {
				t.Errorf("expected status code %d, got %d", tt.wantStatus, recorder.Code)
			}
		})
	}
}

func TestLoginHandler(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		wantStatus  int
	}{
		{
			name: "with valid login parameters",
			requestBody: `{
				"username": "john",
				"password": "ComplexPass321**"
			}`,
			wantStatus: http.StatusOK,
		},
		{
			name: "where credential syntax is correct but there is no user",
			requestBody: `{
				"username": "johny",
				"password": "ComplexPass321**"
			}`,
			wantStatus: http.StatusNotFound,
		},
		{
			name: "where expectedUser is missing",
			requestBody: `{
				"password": "ComplexPass321**"
			}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "where password is missing",
			requestBody: `{
				"username": "john"
			}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "where user does not exist",
			requestBody: `{
				"username": "nonexistent",
				"password": "ComplexPass321**"
			}`,
			wantStatus: http.StatusNotFound,
		},
		{
			name: "where invalid credentials",
			requestBody: `{
				"username": "john",
				"password": "WrongPass123**"
			}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:        "with invalid JSON body",
			requestBody: `{`,
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewTestServer(t)
			// Create a test user
			user := service.CreateUserParams{
				Username: "john",
				Password: "ComplexPass321**",
			}
			_, err := app.stores.CreateUser(context.Background(), user)
			if err != nil {
				t.Fatal(err)
			}

			// Create a request with the test data
			req, err := http.NewRequest("POST", "/login", strings.NewReader(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}

			// Create a response recorder to capture the response
			recorder := httptest.NewRecorder()

			// Call the handler function
			app.UserLoginHandler(recorder, req)

			// Check the response status code
			if recorder.Code != tt.wantStatus {
				t.Errorf("expected status code %d, got %d", tt.wantStatus, recorder.Code)
			}
		})
	}
}

func TestRefreshTokenHandler(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*testing.T, *Application, *service.User) string
		wantStatus int
	}{
		{
			name: "Valid Refresh Token",
			setup: func(t *testing.T, app *Application, user *service.User) string {
				t.Helper()
				body := fmt.Sprintf(`{"username": "%s", "password": "test"}`, user.Username) // assuming you know the correct password for the test user
				req, _ := http.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
				recorder := httptest.NewRecorder()
				app.UserLoginHandler(recorder, req)

				if recorder.Code != http.StatusOK {
					t.Fatalf("Login failed: %s", recorder.Body.String())
				}

				var response map[string]LoginUserResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %s", err)
				}

				return response["UserLogin"].RefreshToken
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Invalid Refresh Token",
			setup: func(t *testing.T, app *Application, user *service.User) string {
				t.Helper()
				// return an invalid refresh token
				return "invalid-token"
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "valid refresh token with blocked session",
			setup: func(t *testing.T, app *Application, user *service.User) string {
				t.Helper()
				body := fmt.Sprintf(`{"username": "%s", "password": "test"}`, user.Username) // assuming you know the correct password for the test user
				req, _ := http.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
				recorder := httptest.NewRecorder()
				app.UserLoginHandler(recorder, req)

				if recorder.Code != http.StatusOK {
					t.Fatalf("Login failed: %s", recorder.Body.String())
				}

				var response map[string]LoginUserResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %s", err)
				}

				refToken := response["UserLogin"].RefreshToken
				// verify the refresh token and get the refresh payload
				refreshPayload, err := app.tokenMaker.VerifyRefreshToken(refToken)
				if err != nil {
					t.Fatalf("Failed to verify refresh token: %s", err)
				}
				// get the session from the store using the refresh token ID
				session, err := app.stores.GetSession(context.Background(), refreshPayload.ID)
				if err != nil {
					t.Fatalf("Failed to get session: %s", err)
				}
				// block the session
				err = app.stores.InvalidateSession(context.Background(), session.ID)
				if err != nil {
					t.Fatalf("Failed to invalidate session: %s", err)
				}
				// return the refresh token with the blocked session
				return refToken
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "valid refresh token with non existent user",
			setup: func(t *testing.T, app *Application, user *service.User) string {
				t.Helper()
				body := fmt.Sprintf(`{"username": "%s", "password": "test"}`, user.Username) // assuming you know the correct password for the test user
				req, _ := http.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
				recorder := httptest.NewRecorder()
				app.UserLoginHandler(recorder, req)

				if recorder.Code != http.StatusOK {
					t.Fatalf("Login failed: %s", recorder.Body.String())
				}
				var response map[string]LoginUserResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %s", err)
				}
				// verify the refresh token and get the refresh payload
				refreshPayload, err := app.tokenMaker.VerifyRefreshToken(response["UserLogin"].RefreshToken)
				if err != nil {
					t.Fatalf("Failed to verify refresh token: %s", err)
				}
				// get the session from the store using the refresh token ID
				session, err := app.stores.GetSession(context.Background(), refreshPayload.ID)
				if err != nil {
					t.Fatalf("Failed to get session: %s", err)
				}
				// delete the user from the store
				err = app.stores.DeleteUser(context.Background(), session.UserID)
				if err != nil {
					t.Fatalf("Failed to delete user: %s", err)
				}
				// return the refresh token
				return response["UserLogin"].RefreshToken
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, user, _ := NewTestEvidenceServer(t)
			refreshToken := tt.setup(t, app, user)

			req, err := http.NewRequest(http.MethodPost, "/api/v1/refresh-token", strings.NewReader(fmt.Sprintf(`{"refresh_token": "%s"}`, refreshToken)))
			if err != nil {
				t.Fatal(err)
			}

			// Create a response recorder to capture the response
			recorder := httptest.NewRecorder()

			// Call the handler function
			app.RefreshTokenHandler(recorder, req)

			// Check the response status code
			if recorder.Code != tt.wantStatus {
				t.Errorf("expected status code %d, got %d", tt.wantStatus, recorder.Code)
			}
		})
	}
}

func TestGetUserHandler(t *testing.T) {
	app, user, _ := NewTestEvidenceServer(t)

	tests := []struct {
		name         string
		userID       string
		expectedCode int
		expectedUser *service.User
		deleteUser   bool
	}{
		{
			name:         "Successful retrieval of user",
			userID:       user.ID.String(),
			expectedCode: http.StatusOK,
			expectedUser: user,
		},
		{
			name:         "Error while parsing user ID (invalid UUID)",
			userID:       "invalidUUID",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Error while fetching user from the store (user not found)",
			userID:       user.ID.String(),
			deleteUser:   true,
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/api/v1/users", nil)
			if err != nil {
				t.Fatal(err)
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("userID", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			// even if userID is correct, if the user is deleted, we expect a 404
			if tt.deleteUser {
				err = app.stores.DeleteUser(context.Background(), user.ID)
				if err != nil {
					t.Fatal(err)
				}
			}
			recorder := httptest.NewRecorder()
			app.GetUserHandler(recorder, req)

			if recorder.Code != tt.expectedCode {
				t.Fatalf("Expected status code %d, got %d", tt.expectedCode, recorder.Code)
			}

			if tt.expectedUser != nil {
				var response map[string]service.User
				err = json.Unmarshal(recorder.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %s", err)
				}
				if response["User"].ID.String() != tt.expectedUser.ID.String() {
					t.Fatalf("Expected user ID %s, got %s", tt.expectedUser.ID.String(), response["User"].ID.String())
				}
			}
		})
	}
}

func TestUpdateUserHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		setupUser      service.CreateUserParams // used to set up a user in the store before the test
		expectedStatus int
	}{
		{
			name: "successful user update",
			requestBody: `{
				"username": "john",
				"first_name": "John",
				"last_name": "Doe",
				"email": "johndoe@example.com"
			}`,
			setupUser: service.CreateUserParams{
				Username: "john",
				Email:    "john@example.com",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "unsuccessful with missing username",
			requestBody: `{
				"first_name": "John",
				"last_name": "Doe",
				"email": "johndoe@example.com"
			}`,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "unsuccessful with missing email",
			requestBody: `{
				"username": "john",
				"first_name": "John",
				"last_name": "Doe"
			}`,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "unsuccessful with non-existing user",
			requestBody: `{
				"username": "nonexistentuser",
				"first_name": "John",
				"last_name": "Doe",
				"email": "johndoe@example.com"
			}`,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "unsuccessful with invalid JSON body",
			requestBody:    `{`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewTestServer(t)

			// Set up the user in the store if provided
			var userID uuid.UUID
			if tt.setupUser.Username != "" {
				user, err := app.stores.CreateUser(context.Background(), tt.setupUser)
				if err != nil {
					t.Fatal(err)
				}
				userID = user.ID
			}

			url := fmt.Sprintf("/update-user/%s", userID)
			req, err := http.NewRequest("POST", url, strings.NewReader(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}

			// Prepare the route context with necessary URL parameters
			rct := chi.NewRouteContext()
			rct.URLParams.Add("userID", userID.String())
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rct)
			req = req.WithContext(ctx)

			recorder := httptest.NewRecorder()
			app.UpdateUserHandler(recorder, req)

			if recorder.Code != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, recorder.Code)
			}
		})
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
			name: "unsuccessful with missing expectedUser",
			requestBody: map[string]interface{}{
				"password": "opsAdmin",
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "unsuccessful with missing password",
			requestBody: map[string]interface{}{
				"username": "Simba",
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "unsuccessful with missing expectedUser and password",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewTestServer(t)

			userParams := service.CreateUserParams{
				Username: "Simba",
				Password: "opsAdmin",
			}
			_, err := app.stores.CreateUser(context.Background(), userParams)
			if err != nil {
				t.Errorf("Failed to create user: %s", err)
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
			app.UserLoginHandler(response, request)
			if status := response.Code; status != tt.expectedStatus {
				t.Errorf("Handler returned wrong status code. Expected: %d. Got: %d.", tt.expectedStatus, status)
			}
		})
	}
}

func TestLoginFailedWithBadJSONRequest(t *testing.T) {
	app := NewTestServer(t)
	response := httptest.NewRecorder()

	request, err := http.NewRequest("POST", "/Login", bytes.NewReader([]byte("badJsonRequest")))
	if err != nil {
		t.Errorf("Error creating a new request: %v", err)
	}

	app.UserLoginHandler(response, request)

	if status := response.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code. Expected: %d. Got: %d.", http.StatusBadRequest, status)
	}
}

func TestGetUsersWithRolesHandler(t *testing.T) {
	tests := []struct {
		name           string
		setupUsers     []service.CreateUserParams // used to set up multiple users in the store before the test
		expectedStatus int
	}{
		{
			name: "successful retrieval of users with roles",
			setupUsers: []service.CreateUserParams{
				{
					Username: "john",
					Email:    "john@example.com",
				},
				{
					Username: "jane",
					Email:    "jane@example.com",
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "retrieval with no users in database",
			setupUsers:     []service.CreateUserParams{},
			expectedStatus: http.StatusOK, // Assuming you return 200 even if no users exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewTestServer(t)

			// Set up the users in the store if provided
			for _, user := range tt.setupUsers {
				_, err := app.stores.CreateUser(context.Background(), user)
				if err != nil {
					t.Fatal(err)
				}
			}

			req, err := http.NewRequest("GET", "/usersWithRoles", nil)
			if err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()
			app.GetUsersWithRoleHandler(recorder, req)

			if recorder.Code != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, recorder.Code)
			}
		})
	}
}

func TestDeleteUserHandler(t *testing.T) {
	tests := []struct {
		name           string
		setupUser      *service.CreateUserParams // used to set up a user in the store before the test
		deleteUserID   uuid.UUID                 // ID of the user to try and delete
		expectedStatus int
	}{
		{
			name: "successfully delete a user",
			setupUser: &service.CreateUserParams{
				Username: "john",
				Email:    "john@example.com",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "try to delete a non-existent user",
			setupUser:      nil, // don't set up any user
			deleteUserID:   uuid.New(),
			expectedStatus: http.StatusNotFound, // or whatever your store returns when a user isn't found
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewTestServer(t)

			var userID uuid.UUID
			if tt.setupUser != nil {
				user, err := app.stores.CreateUser(context.Background(), *tt.setupUser)
				if err != nil {
					t.Fatal(err)
				}
				userID = user.ID
			} else {
				userID = tt.deleteUserID
			}

			// Create the request body
			body, err := json.Marshal(DeleteUserParams{ID: userID})
			if err != nil {
				t.Fatal(err)
			}

			req, err := http.NewRequest("DELETE", "/user", bytes.NewReader(body))
			if err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()
			app.DeleteUserHandler(recorder, req)

			if recorder.Code != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, recorder.Code)
			}
		})
	}
}

func TestGetUsersHandler(t *testing.T) {
	tests := []struct {
		name              string
		setupFunc         func(t *testing.T, app *Application) // function to populate the database
		expectedUserCount int
	}{
		{
			name: "successfully retrieve all users when there are 2 users",
			setupFunc: func(t *testing.T, app *Application) {
				t.Helper()
				// Create two users
				_, err := app.stores.CreateUser(context.Background(), service.CreateUserParams{
					Username:  "john1",
					FirstName: "John",
					LastName:  "Doe1",
					Email:     "john1@example.com",
					Password:  "ComplexPass321**",
				})
				if err != nil {
					t.Fatal(err)
				}
				_, err = app.stores.CreateUser(context.Background(), service.CreateUserParams{
					Username:  "john2",
					FirstName: "John",
					LastName:  "Doe2",
					Email:     "john2@example.com",
					Password:  "ComplexPass321**",
				})
				if err != nil {
					t.Fatal(err)
				}
			},
			expectedUserCount: 2,
		},
		{
			name: "successfully retrieve all users when there are 4 users",
			setupFunc: func(t *testing.T, app *Application) {
				t.Helper()
				// Create four users
				for i := 0; i < 4; i++ {
					_, err := app.stores.CreateUser(context.Background(), service.CreateUserParams{
						Username:  fmt.Sprintf("john%d", i+3),
						FirstName: "John",
						LastName:  fmt.Sprintf("Doe%d", i+3),
						Email:     fmt.Sprintf("john%d@example.com", i+3),
						Password:  "ComplexPass321**",
					})
					if err != nil {
						t.Fatal(err)
					}
				}
			},
			expectedUserCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewTestServer(t)
			tt.setupFunc(t, app)

			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, "/users", nil)
			if err != nil {
				t.Fatal(err)
			}
			app.ListUsersHandler(recorder, req)

			if recorder.Code != http.StatusOK {
				t.Errorf("expected status code %d, got %d", http.StatusOK, recorder.Code)
			}

			var response envelope
			err = json.NewDecoder(recorder.Body).Decode(&response)
			if err != nil {
				t.Fatal(err)
			}

			users, ok := response["Users"].([]interface{})
			if !ok {
				t.Fatalf("unexpected type for Users: %T", response["Users"])
			}

			if len(users) != tt.expectedUserCount {
				t.Errorf("expected %d users, got %d", tt.expectedUserCount, len(users))
			}
		})
	}
}

func TestAddRoleToUser(t *testing.T) {
	app, user, _ := NewTestEvidenceServer(t)
	role, err := app.stores.DBStore.GetRoleByName(context.Background(), "admin")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name           string
		userID         string
		roleID         string
		expectedStatus int
	}{
		{
			name:           "failed when missing the userID",
			userID:         "",
			roleID:         role.ID.String(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "failed when missing the roleID",
			userID:         user.ID.String(),
			roleID:         "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "was successful with valid parameters",
			userID:         user.ID.String(),
			roleID:         role.ID.String(),
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/v1/authenticated/users/%s/roles/%s", tt.userID, tt.roleID)
			req, err := http.NewRequest(http.MethodPost, url, nil)
			if err != nil {
				t.Fatal(err)
			}

			// Create and set up the router context
			rctx := chi.NewRouteContext()
			if tt.userID != "" {
				rctx.URLParams.Add("userID", tt.userID)
			}
			if tt.roleID != "" {
				rctx.URLParams.Add("roleID", tt.roleID)
			}
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create a response recorder to capture the response
			recorder := httptest.NewRecorder()

			// Call the handler function
			app.AddRoleToUserHandler(recorder, req)

			// Check the response status code
			if recorder.Code != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, recorder.Code)
			}
		})
	}
}
