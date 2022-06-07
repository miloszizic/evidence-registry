package api

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/miloszizic/der/internal/data"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func addAuthorization(
	t *testing.T,
	request *http.Request,
	tokenMaker Maker,
	authorizationType string,
	username string,
	duration time.Duration,
) {
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
	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker Maker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "should pass with valid token",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker Maker) {
				addAuthorization(t, request, tokenMaker, "Bearer", "user", time.Hour)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusOK {
					t.Fatalf("expected status code %d, got %d", http.StatusOK, recorder.Code)
				}
			},
		},
		{
			name: "should fail with invalid token",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker Maker) {
				addAuthorization(t, request, tokenMaker, "Bearer", "user", time.Hour)
				request.Header.Set(string(authorizationHeaderKey), "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSJ9.c7SpmftjdwaJH6gNkoyxrjxgTrX9tXgWK3ZZ8mAvJIY")
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusUnauthorized {
					t.Fatalf("expected status code %d, got %d", http.StatusUnauthorized, recorder.Code)
				}
			},
		},
		{
			name: "should fail with missing token",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker Maker) {
				addAuthorization(t, request, tokenMaker, "Bearer", "user", time.Hour)
				request.Header.Del(string(authorizationHeaderKey))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusUnauthorized {
					t.Fatalf("expected status code %d, got %d", http.StatusUnauthorized, recorder.Code)
				}
			},
		},
		{
			name: "should fail with invalid token payload",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker Maker) {
				addAuthorization(t, request, tokenMaker, "Bearer", "user", time.Hour)
				request.Header.Set(string(authorizationHeaderKey), "Basic invalid-token")
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusUnauthorized {
					t.Fatalf("expected status code %d, got %d", http.StatusUnauthorized, recorder.Code)
				}
			},
		},
		{
			name: "should fail with expired token",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker Maker) {
				addAuthorization(t, request, tokenMaker, "Bearer", "user", time.Microsecond)
				time.Sleep(time.Second)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusUnauthorized {
					t.Fatalf("expected status code %d, got %d", http.StatusUnauthorized, recorder.Code)
				}
			},
		},
		{
			name: "should fail with invalid token type",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker Maker) {
				addAuthorization(t, request, tokenMaker, "Bearer", "user", time.Hour)
				request.Header.Set(string(authorizationHeaderKey), "Basic invalid-token")
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusUnauthorized {
					t.Fatalf("expected status code %d, got %d", http.StatusUnauthorized, recorder.Code)
				}
			},
		},
		{
			name: "invalid authorization header",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker Maker) {
				request.Header.Set(string(authorizationHeaderKey), "invalid-token")
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusUnauthorized {
					t.Fatalf("expected status code %d, got %d", http.StatusUnauthorized, recorder.Code)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := newTestServer(t)
			tokenMaker, err := NewPasetoMaker("nigkjtvbrhugwpgaqbemmvnqbtywfrcq")
			if err != nil {
				t.Fatal(err)
			}
			request := httptest.NewRequest("GET", "/", nil)
			recorder := httptest.NewRecorder()
			tc.setupAuth(t, request, tokenMaker)
			app.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			})).ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestMiddlewarePermissions(t *testing.T) {
	testCases := []struct {
		name          string
		request       func() *http.Request
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "denied with the user does not exist",
			request: func() *http.Request {
				request := httptest.NewRequest("GET", "/", nil)
				payload := &Payload{
					Username: "user",
				}
				ctx := context.WithValue(request.Context(), authorizationPayloadKey, payload)
				reqWithPayload := request.WithContext(ctx)
				return reqWithPayload
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusUnauthorized {
					t.Fatalf("expected status code %d, got %d", http.StatusUnauthorized, recorder.Code)
				}
			},
		},
		{
			name: "denied if payload is empty",
			request: func() *http.Request {
				request := httptest.NewRequest("GET", "/", nil)
				payload := &Payload{}
				ctx := context.WithValue(request.Context(), authorizationPayloadKey, payload)
				reqWithPayload := request.WithContext(ctx)
				return reqWithPayload
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusUnauthorized {
					t.Fatalf("expected status code %d, got %d", http.StatusUnauthorized, recorder.Code)
				}
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := newTestServer(t)
			recorder := httptest.NewRecorder()
			app.MiddlewarePermissionChecker(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			})).ServeHTTP(recorder, tc.request())
			tc.checkResponse(t, recorder)
		})
	}
}
func TestMiddlewarePermissionsAllowsOnlyAdminUsers(t *testing.T) {
	app := newTestServer(t)
	// seed the database with one user and a case
	seedForHandlerTesting(t, app)
	user := &data.User{
		Username: "testuser",
		Role:     "user",
	}
	err := user.Password.Set("testpassword")
	if err != nil {
		t.Fatal(err)
	}
	err = app.stores.User.Add(user)
	if err != nil {
		t.Fatal(err)
	}
	payload := &Payload{
		Username: "testuser",
	}
	request := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(request.Context(), authorizationPayloadKey, payload)
	reqWithPayload := request.WithContext(ctx)
	recorder := httptest.NewRecorder()
	app.MiddlewarePermissionChecker(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})).ServeHTTP(recorder, reqWithPayload)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status code %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}
func TestLogger(t *testing.T) {
	tests := []struct {
		name     string
		route    string
		handler  http.HandlerFunc
		expected string
	}{
		{
			name:  "should log request as OK",
			route: "/",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("hello"))
			}),
			expected: "OK",
		},
		{
			name:  "should log request as server error",
			route: "/",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}),
			expected: "Server Error",
		}, {
			name:  "should log request as client error",
			route: "/",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			}),
			expected: "Client Error",
		}, {
			name:  "should log request as redirect",
			route: "/",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/", http.StatusFound)
			}),
			expected: "Redirect",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := testLogger(&buf)
			r := chi.NewRouter()
			r.Use(Logger(logger))
			r.Get(tt.route, tt.handler)
			request := httptest.NewRequest("GET", tt.route, nil)
			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, request)
			if !strings.Contains(buf.String(), tt.expected) {
				t.Fatalf("expected to log request, got %s", buf.String())
			}
		})
	}
}
