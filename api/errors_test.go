//go:build integration

package api

import (
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/miloszizic/der/service"
)

func TestRespondErrorHandlerReturned(t *testing.T) {
	// Mocked http.Request
	r := &http.Request{}

	// Initialized app instance using NewTestEvidenceServer
	app, _, _ := NewTestEvidenceServer(t)

	// Define test cases based on the error_to_status_dict mappings
	tests := []struct {
		name       string
		inputError error
		wantStatus int
	}{
		{"ErrUnexpectedEOF", io.ErrUnexpectedEOF, http.StatusBadRequest},
		{"EOF", io.EOF, http.StatusBadRequest},
		{"ErrNoRows", sql.ErrNoRows, http.StatusNotFound},
		{"ErrNotFound", service.ErrNotFound, http.StatusNotFound},
		{"ErrAlreadyExists", service.ErrAlreadyExists, http.StatusConflict},
		{"ErrInvalidRequest", service.ErrInvalidRequest, http.StatusBadRequest},
		{"ErrUnauthorized", service.ErrUnauthorized, http.StatusUnauthorized},
		{"ErrInvalidCredentials", service.ErrInvalidCredentials, http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			// Call the app.respondError function with the test case's inputError
			app.respondError(w, r, tt.inputError)

			// Extract the returned status from the mocked ResponseWriter
			gotStatus := w.Result().StatusCode

			// Close response body
			defer w.Result().Body.Close()

			// Compare the returned status with the expected status
			if gotStatus != tt.wantStatus {
				t.Errorf("got status %d, want %d", gotStatus, tt.wantStatus)
			}
		})
	}
}
