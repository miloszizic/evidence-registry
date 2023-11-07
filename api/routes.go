// Package api provides the implementation for HTTP handlers
package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// route function sets the routes for the HTTP server
// it returns http.Handler
func (app *Application) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(app.recoverPanic)
	r.Route("/api/v1", func(r chi.Router) {
		app.notProtectedRoutes(r)
		app.protectedRoutes(r)
	})

	return r
}

// notProtectedRoutes function sets the routes that does not require protection
func (app *Application) notProtectedRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(Logger(app.logger))
		r.Get("/health", app.HealthCheck)
		r.Post("/login", app.UserLoginHandler)
		r.Post("/refresh-token", app.RefreshTokenHandler)
	})
}

// protectedRoutes function sets the routes that require protection
func (app *Application) protectedRoutes(r chi.Router) {
	r.Route("/authenticated", func(r chi.Router) {
		r.Use(app.AuthMiddleware)
		r.Use(app.UserParserMiddleware)
		r.Use(Logger(app.logger))
		app.adminRoutes(r)
		app.userRoutes(r)
		app.casesRoutes(r)
	})
}

// adminRoutes function sets the routes that require admin permissions
func (app *Application) adminRoutes(r chi.Router) {
	r.Route("/admin", func(r chi.Router) {
		// Create
		r.Group(func(r chi.Router) {
			r.Use(app.MiddlewarePermissionChecker("create_role"))
			// Roles and permissions
			r.Post("/roles", app.CreateRoleHandler)
			r.Post("/roles/{roleID}/permissions/{permissionID}", app.AddPermissionToRoleHandler)
			// CaseTypes
			r.Post("/caseTypes", app.CreateCaseTypeHandler)
			r.Put("/caseTypes/{caseTypeID}", app.UpdateCaseTypeHandler)
		})
		// View
		r.Group(func(r chi.Router) {
			r.Use(app.MiddlewarePermissionChecker("view_role"))
			// Roles and permissions
			r.Get("/roles", app.ListRolesHandler)
			r.Get("/permissions", app.ListPermissionsHandler)
			r.Get("/roles/{roleID},", app.GetRoleHandler)
			r.Get("/roles/{roleID}/permissions", app.GetRolePermissionsHandler)
			// CaseTypes
			r.Get("/caseTypes/{caseTypeID}", app.GetCaseTypeHandler)
			r.Get("/caseTypes", app.ListCaseTypesHandler)
		})
		// Delete
		r.Group(func(r chi.Router) {
			r.Use(app.MiddlewarePermissionChecker("delete_role"))
			r.Delete("/roles/{roleID}/permissions/{permissionID}", app.RemovePermissionFromRoleHandler)
			r.Delete("/roles/{roleID}", app.DeleteRoleHandler)
		})
	})
}

// userRoutes function sets the routes related to users
func (app *Application) userRoutes(r chi.Router) {
	r.Route("/users", func(r chi.Router) {
		// Create
		r.Group(func(r chi.Router) {
			r.Use(app.MiddlewarePermissionChecker("create_user"))
			r.Put("/{userID}", app.UpdateUserHandler)
			r.Post("/register", app.CreateUserHandler)
			r.Patch("/{userID}/password", app.UpdateUserPasswordHandler)
			r.Post("/{userID}/roles/{roleID}", app.AddRoleToUserHandler)
		})
		// View
		r.Group(func(r chi.Router) {
			r.Use(app.MiddlewarePermissionChecker("view_user"))
			r.Get("/", app.ListUsersHandler)
			r.Get("/{userID}", app.GetUserHandler)
			r.Get("/withRoles", app.GetUsersWithRoleHandler)
		})
		// Delete
		r.Group(func(r chi.Router) {
			r.Use(app.MiddlewarePermissionChecker("delete_user"))
			r.Delete("/{userID}", app.DeleteUserHandler)
		})
	})
}

// casesRoutes function sets the routes related to cases
func (app *Application) casesRoutes(r chi.Router) {
	r.Route("/cases", func(r chi.Router) {
		app.casesSubRoutes(r)
		app.evidencesRoutes(r)
	})
}

// casesSubRoutes function sets the sub-routes related to cases
func (app *Application) casesSubRoutes(r chi.Router) {
	// Create
	r.Group(func(r chi.Router) {
		r.Use(app.MiddlewarePermissionChecker("create_case"))
		r.Post("/", app.CreateCaseHandler)
		// r.Post("/courts", app.CreateCourtHandler)
	})
	// View
	r.Group(func(r chi.Router) {
		r.Use(app.MiddlewarePermissionChecker("view_case"))
		r.Get("/", app.ListCasesHandler)
		r.Get("/{caseID}", app.GetCaseHandler)
		r.Get("/courts", app.ListCourtsHandler)
		r.Get("/evidenceTypes", app.ListEvidenceTypesHandler)
	})
	// Delete
	r.Group(func(r chi.Router) {
		r.Use(app.MiddlewarePermissionChecker("delete_case"))
		r.Delete("/{caseID}", app.DeleteCaseHandler)
	})
}

// evidencesRoutes function sets the routes related to evidences
func (app *Application) evidencesRoutes(r chi.Router) {
	r.Route("/{caseID}/evidences", func(r chi.Router) {
		// Create
		r.Group(func(r chi.Router) {
			r.Use(app.MiddlewarePermissionChecker("create_evidence"))
			r.Post("/", app.CreateEvidenceHandler)
		})
		// View
		r.Group(func(r chi.Router) {
			r.Use(app.MiddlewarePermissionChecker("view_evidence"))
			r.Get("/", app.ListEvidencesHandler)
			r.Get("/{evidenceID}/download", app.DownloadEvidenceHandler)
			r.Get("/{evidenceID}", app.GetEvidenceHandler)
		})
		// Delete
		r.Group(func(r chi.Router) {
			r.Use(app.MiddlewarePermissionChecker("delete_evidence"))
			// r.Delete("/{evidenceID}", app.DeleteEvidenceHandler)
		})
	})
}
