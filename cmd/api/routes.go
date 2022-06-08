package api

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (app *Application) routes() http.Handler {
	// make a new router
	r := chi.NewRouter()
	r.Use(Logger(app.logger))

	r.Group(func(r chi.Router) {
		// not protected routes
		r.Get("/ping", app.Ping)
		r.Post("/login", app.Login)
	})
	// protected routes
	r.Group(func(r chi.Router) {
		// Middlewares in use
		r.Use(app.AuthMiddleware)
		r.Use(app.MiddlewarePermissionChecker)

		// users routes
		r.Post("/register", app.CreateUserHandler)

		// cases
		r.Post("/cases", app.CreateCaseHandler)
		r.Get("/cases", app.ListCasesHandler)
		r.Get("/cases/{CaseID}", app.GetCaseHandler)
		r.Delete("/cases/{caseID}", app.RemoveCaseHandler)

		// evidences
		r.Get("/cases/{caseID}/evidences", app.ListEvidencesHandler)
		r.Post("/cases/{caseID}/evidences", app.CreateEvidenceHandler)
		r.Get("/cases/{caseID}/evidences/{evidenceID}", app.DownloadEvidenceHandler)
		r.Delete("/cases/{caseID}/evidences/{evidenceID}", app.DeleteEvidenceHandler)
		r.Post("/cases/{caseID}/evidences/{evidenceID}/comment", app.AddCommentHandler)
	})
	return r
}
