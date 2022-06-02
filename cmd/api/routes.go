package api

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (app *Application) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(Logger(app.logger))

	r.Group(func(r chi.Router) {
		// Not protected routes
		r.Get("/ping", app.Ping)    // GET /ping - Returns "pong"
		r.Post("/login", app.Login) // POST /Login - Returns JWT r
	})

	// protected routes
	r.Group(func(r chi.Router) {
		// Middlewares in use
		r.Use(app.AuthMiddleware)
		r.Use(app.MiddlewarePermissionChecker)
		// Routes

		// Users routes
		r.Post("/register", app.CreateUserHandler)
		// cases
		r.Post("/cases", app.CreateCaseHandler)
		r.Get("/cases", app.ListCasesHandler)
		r.Get("/cases/{CaseID}", app.GetCaseHandler)
		r.Delete("/cases/{caseID}", app.RemoveCaseHandler)
		// evidences
		r.Get("/cases/{caseID}/evidences", app.ListEvidencesHandler)
		r.Post("/cases/{caseID}/evidences", app.CreateEvidenceHandler)
		r.Get("/cases/{caseID}/evidences/{evidenceID}", app.GetEvidenceHandler)
		r.Delete("/cases/{caseID}/evidences/{evidenceID}", app.DeleteEvidenceHandler)
		r.Post("/cases/{caseID}/evidences/{evidenceID}/comment", app.AddCommentHandler)
	})
	return r
}
