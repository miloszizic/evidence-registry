package api

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/miloszizic/der/service"

	"github.com/miloszizic/der/db"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func addAuthorization(
	t *testing.T,
	request *http.Request,
	tokenMaker service.Maker,
	duration time.Duration,
) {
	t.Helper()
	// add an authorization type and token to request header
	const authorizationType = "Bearer"
	// create a token for the user
	const username = "user"

	token, payload, err := tokenMaker.CreateToken(username, duration)
	if err != nil {
		t.Fatal(err)
	}

	if payload == nil {
		t.Fatal("payload is nil")
	}

	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, token)
	request.Header.Set(string(authorizationHeaderKey), authorizationHeader)
}

func TestMiddlewareAuthWithRequestHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker service.Maker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "should pass with valid token",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker service.Maker) {
				t.Helper()
				addAuthorization(t, request, tokenMaker, time.Hour)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				if recorder.Code != http.StatusOK {
					t.Fatalf("expected status code %d, got %d", http.StatusOK, recorder.Code)
				}
			},
		},
		{
			name: "should fail with invalid token",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker service.Maker) {
				t.Helper()
				addAuthorization(t, request, tokenMaker, time.Hour)
				request.Header.Set(string(authorizationHeaderKey), "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSJ9.c7SpmftjdwaJH6gNkoyxrjxgTrX9tXgWK3ZZ8mAvJIY")
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				if recorder.Code != http.StatusUnauthorized {
					t.Fatalf("expected status code %d, got %d", http.StatusUnauthorized, recorder.Code)
				}
			},
		},
		{
			name: "should fail with missing token",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker service.Maker) {
				t.Helper()
				addAuthorization(t, request, tokenMaker, time.Hour)
				request.Header.Del(string(authorizationHeaderKey))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				if recorder.Code != http.StatusUnauthorized {
					t.Fatalf("expected status code %d, got %d", http.StatusUnauthorized, recorder.Code)
				}
			},
		},
		{
			name: "should fail with invalid token payload",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker service.Maker) {
				t.Helper()
				addAuthorization(t, request, tokenMaker, time.Hour)
				request.Header.Set(string(authorizationHeaderKey), "Basic invalid-token")
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				if recorder.Code != http.StatusUnauthorized {
					t.Fatalf("expected status code %d, got %d", http.StatusUnauthorized, recorder.Code)
				}
			},
		},
		{
			name: "should fail with expired token",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker service.Maker) {
				t.Helper()
				addAuthorization(t, request, tokenMaker, time.Microsecond)
				time.Sleep(time.Second)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				if recorder.Code != http.StatusUnauthorized {
					t.Fatalf("expected status code %d, got %d", http.StatusUnauthorized, recorder.Code)
				}
			},
		},
		{
			name: "should fail with invalid token type",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker service.Maker) {
				t.Helper()
				addAuthorization(t, request, tokenMaker, time.Hour)
				request.Header.Set(string(authorizationHeaderKey), "Basic invalid-token")
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				if recorder.Code != http.StatusUnauthorized {
					t.Fatalf("expected status code %d, got %d", http.StatusUnauthorized, recorder.Code)
				}
			},
		},
		{
			name: "invalid authorization header",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker service.Maker) {
				t.Helper()
				request.Header.Set(string(authorizationHeaderKey), "invalid-token")
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				if recorder.Code != http.StatusUnauthorized {
					t.Fatalf("expected status code %d, got %d", http.StatusUnauthorized, recorder.Code)
				}
			},
		},
	}

	for _, tt := range tests {
		pt := tt
		t.Run(pt.name, func(t *testing.T) {
			t.Parallel()
			app := NewTestServer(t)
			tokenMaker, err := service.NewPasetoMaker("nigkjtvbrhugwpgaqbemmvnqbtywfrcq")
			if err != nil {
				t.Fatal(err)
			}
			request := httptest.NewRequest("GET", "/", nil)
			recorder := httptest.NewRecorder()
			pt.setupAuth(t, request, tokenMaker)
			app.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			})).ServeHTTP(recorder, request)
			pt.checkResponse(t, recorder)
		})
	}
}

func TestUserParserMiddleware(t *testing.T) {
	app, user, _ := NewTestEvidenceServer(t)

	validPayload := &service.Payload{
		Username: user.Username,
	}

	tests := []struct {
		name         string
		contextValue interface{}
		expectedCode int
	}{
		{
			name:         "Successful user parsing",
			contextValue: validPayload,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Missing payload in context",
			contextValue: nil,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Unexpected value in context",
			contextValue: "unexpectedValue",
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "Username not found",
			contextValue: &service.Payload{Username: "nonexistentUsername"},
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/test", nil)
			if err != nil {
				t.Fatal(err)
			}

			req = req.WithContext(context.WithValue(req.Context(), authorizationPayloadKey, tt.contextValue))

			recorder := httptest.NewRecorder()

			middleware := app.UserParserMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			middleware.ServeHTTP(recorder, req)

			if recorder.Code != tt.expectedCode {
				t.Fatalf("Expected status code %d, got %d", tt.expectedCode, recorder.Code)
			}
		})
	}
}

func TestMiddlewarePermission(t *testing.T) {
	tests := []struct {
		name       string
		permission string
		userRole   string
		wantStatus int
	}{
		{
			name:       "admin with view_case permissions can view_case",
			permission: "view_case",
			userRole:   "admin",
			wantStatus: http.StatusOK,
		},
		{
			name:       "editor without create_case permissions cannot create_case",
			permission: "create_case",
			userRole:   "editor",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "admin with create_case permissions can create_case",
			permission: "create_case",
			userRole:   "admin",
			wantStatus: http.StatusOK,
		},
		{
			name:       "viewer without create_case permissions cannot create_case",
			permission: "create_case",
			userRole:   "viewer",
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		pt := tt
		t.Run(pt.name, func(t *testing.T) {
			// Initialize the test server and db
			app := NewTestServer(t)

			// Set up the test user and their role
			user := db.CreateUserParams{
				Username: "test",
				Password: "test",
			}
			createdUser, err := app.stores.DBStore.CreateUser(context.Background(), user)
			if err != nil {
				t.Fatal(err)
			}

			// Get the role
			role, err := app.stores.DBStore.GetRoleByName(context.Background(), pt.userRole)
			if err != nil {
				t.Fatal(err)
			}
			roleID := service.HandleNullableUUID(role.ID)
			// Assign the role to the user
			err = app.stores.DBStore.AssignRoleToUser(context.Background(), db.AssignRoleToUserParams{
				RoleID: roleID,
				ID:     createdUser.ID,
			})
			if err != nil {
				t.Fatal(err)
			}

			// Set up the test request
			request := httptest.NewRequest("GET", "/test", nil)
			ctx := context.WithValue(request.Context(), authorizationPayloadKey, &service.Payload{
				Username: createdUser.Username,
			})
			request = request.WithContext(ctx)

			// Set up the test response recorder
			response := httptest.NewRecorder()

			// Set up the next handler to just return StatusOK
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Pass the request through the MiddlewarePermissionChecker
			permissionCheckerMiddleware := app.MiddlewarePermissionChecker(pt.permission)
			permissionCheckerMiddleware(nextHandler).ServeHTTP(response, request)

			// Check the response status code
			if response.Code != pt.wantStatus {
				t.Errorf("expected status %d, got %d", pt.wantStatus, response.Code)
			}
		})
	}
}

func TestLogger(t *testing.T) {
	// Prepare the buffer for logger
	var buf bytes.Buffer

	// Create a basic logger and direct its output to the buffer
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := config.Build()
	logger = logger.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core, zapcore.NewCore(
			zapcore.NewConsoleEncoder(config.EncoderConfig),
			zapcore.Lock(zapcore.AddSync(&buf)),
			zapcore.DebugLevel,
		))
	}))
	sugar := logger.Sugar()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/testpath":
			w.WriteHeader(http.StatusOK)
		case "/anotherpath":
			w.WriteHeader(http.StatusMovedPermanently)
		case "/errorpath":
			w.WriteHeader(http.StatusBadRequest)
		case "/servererror":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusNotFound)
		}

		_, err := w.Write([]byte("OK"))
		if err != nil {
			t.Fatal(err)
		}
	})

	loggedHandler := Logger(sugar)(handler)

	testCases := []struct {
		method string
		path   string
		code   int
	}{
		{"GET", "/testpath", http.StatusOK},
		{"POST", "/anotherpath", http.StatusMovedPermanently},
		{"PUT", "/errorpath", http.StatusBadRequest},
		{"DELETE", "/servererror", http.StatusInternalServerError},
		{"PATCH", "/unknownpath", http.StatusNotFound},
	}

	for _, tc := range testCases {
		req, err := http.NewRequest(tc.method, tc.path, nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		// Act
		loggedHandler.ServeHTTP(rr, req)

		// Let's wait a little bit to ensure logging has been completed
		time.Sleep(time.Millisecond * 50)

		// Check the response for specific strings that need to be in the log
		output := buf.String()
		if !strings.Contains(output, tc.path) {
			t.Errorf("Logger() log = %v, want %v", output, tc.path)
		}

		if !strings.Contains(output, tc.method) {
			t.Errorf("Logger() log = %v, want %v", output, tc.method)
		}

		if !strings.Contains(output, statusLabel(tc.code)) {
			t.Errorf("Logger() log = %v, want %v", output, statusLabel(tc.code))
		}
	}
}

func TestRecoverPanic(t *testing.T) {
	t.Parallel()
	// we are deliberately causing a panic
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("deliberate panic for testing!")
	})
	app := NewTestServer(t)
	recoveringHandler := app.recoverPanic(panicHandler)

	// Create a request and a response recorder
	req, err := http.NewRequest("GET", "/panic", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rec := httptest.NewRecorder()

	recoveringHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	// Check that the middleware does not affect non-panicking handlers
	nonPanicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("all good"))
		if err != nil {
			t.Fatalf("could not write response: %v", err)
		}
	})

	recoveringHandler = app.recoverPanic(nonPanicHandler)
	rec = httptest.NewRecorder()

	recoveringHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Body.String() != "all good" {
		t.Errorf("expected response body 'all good', got %s", rec.Body.String())
	}
}
