package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestProtectedRoutes tests the protected routes in the API. It makes sure that the routes are protected by the authentication middleware in case
// we change the routes in the future and forget to protect them.
func TestProtectedRoutes(t *testing.T) {
	t.Parallel()

	app := NewTestServer(t)
	r := app.routes()

	tests := []struct {
		method string
		path   string
	}{
		// Admin Routes
		// Create
		{"POST", "/api/v1/authenticated/admin/roles"},
		{"POST", "/api/v1/authenticated/admin/roles/{roleID}/permissions/{permissionID}"},
		{"POST", "/api/v1/authenticated/admin/caseTypes"},
		{"PUT", "/api/v1/authenticated/admin/caseTypes/{caseTypeID}"},
		// View
		{"GET", "/api/v1/authenticated/admin/roles"},
		{"GET", "/api/v1/authenticated/admin/permissions"},
		{"GET", "/api/v1/authenticated/admin/roles/{roleID}"},
		{"GET", "/api/v1/authenticated/admin/roles/{roleID}/permissions"},
		{"GET", "/api/v1/authenticated/admin/caseTypes/{caseTypeID}"},
		{"GET", "/api/v1/authenticated/admin/caseTypes"},
		// Delete
		{"DELETE", "/api/v1/authenticated/admin/roles/{roleID}/permissions/{permissionID}"},
		{"DELETE", "/api/v1/authenticated/admin/roles/{roleID}"},

		// User Routes
		// Create
		{"PUT", "/api/v1/authenticated/users/{userID}"},
		{"POST", "/api/v1/authenticated/users/register"},
		{"PATCH", "/api/v1/authenticated/users/{userID}/password"},
		{"POST", "/api/v1/authenticated/users/{userID}/roles/{roleID}"},
		// View
		{"GET", "/api/v1/authenticated/users/"},
		{"GET", "/api/v1/authenticated/users/{userID}"},
		{"GET", "/api/v1/authenticated/users/withRoles"},
		// Delete
		{"DELETE", "/api/v1/authenticated/users/{userID}"},

		// Cases Routes
		// Create
		{"POST", "/api/v1/authenticated/cases"},
		// View
		{"GET", "/api/v1/authenticated/cases/"},
		{"GET", "/api/v1/authenticated/cases/{caseID}"},
		{"GET", "/api/v1/authenticated/cases/courts"},
		{"GET", "/api/v1/authenticated/cases/evidenceTypes"},
		// Delete
		{"DELETE", "/api/v1/authenticated/cases/{caseID}"},

		// Evidences Routes
		// Create
		{"POST", "/api/v1/authenticated/cases/{caseID}/evidences"},
		// View
		{"GET", "/api/v1/authenticated/cases/{caseID}/evidences/"},
		{"GET", "/api/v1/authenticated/cases/{caseID}/evidences/{evidenceID}"},
		{"GET", "/api/v1/authenticated/cases/{caseID}/evidences/{evidenceID}/download"},
	}

	for _, tt := range tests {
		pt := tt
		t.Run(pt.path, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(pt.method, pt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusUnauthorized {
				t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, recorder.Code)
			}
		})
	}
}
