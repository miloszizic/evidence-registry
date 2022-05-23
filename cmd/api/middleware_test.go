package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
	request.Header.Set(authorizationHeaderKey, authorizationHeader)
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
				request.Header.Set(authorizationHeaderKey, "Bearer invalid-token")
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
				request.Header.Del(authorizationHeaderKey)
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
				request.Header.Set(authorizationHeaderKey, "Basic invalid-token")
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
				request.Header.Set(authorizationHeaderKey, "Basic invalid-token")
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
				request.Header.Set(authorizationHeaderKey, "invalid-token")
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
