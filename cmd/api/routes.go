package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func (app *Application) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Group(func(r chi.Router) {
		// Not protected routes
		r.Get("/ping", app.Ping)    // GET /ping - Returns "pong"
		r.Post("/Login", app.Login) // POST /Login - Returns JWT r
	})

	// protected routes
	r.Group(func(r chi.Router) {
		// using auth middleware
		r.Use(app.AuthMiddleware)
		// Routes

		// users
		r.Post("/register", app.CreateUserHandler)
		// cases
		r.Post("/cases", app.CreateCaseHandler)
		r.Get("/cases", app.ListCasesHandler)
		r.Delete("/cases/{caseID}", app.RemoveCaseHandler)
		// evidences
		r.Get("/cases/{caseID}/evidences", app.ListEvidencesHandler)
		r.Post("/cases/{caseID}/evidences", app.CreateEvidenceHandler)
		r.Get("/cases/{caseID}/evidences/{evidenceID}", app.GetEvidenceHandler)
		r.Delete("/cases/{caseID}/evidences/{evidenceID}", app.DeleteEvidenceHandler)
		r.Post("/cases/{caseID}/evidences/{evidenceID}/comment", app.AddCommentToEvidenceHandler)
	})
	return r
}
