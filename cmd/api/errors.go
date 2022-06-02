package api

import (
	"go.uber.org/zap"
	"net/http"
)

func (app *Application) logError(r *http.Request, err error) {
	app.logger.Error("error processing request",
		zap.String("request_method", r.Method),
		zap.String("request_url", r.URL.String()),
		zap.Error(err))
}

func (app *Application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{"error": message}
	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

func (app *Application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)
	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (app *Application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

func (app *Application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (app *Application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (app *Application) invalidAuthorisationHeaderFormat(w http.ResponseWriter, r *http.Request) {
	message := "invalid authorization header format"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}
func (app *Application) tokenExpired(w http.ResponseWriter, r *http.Request) {
	message := "Your token has expired"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (app *Application) unauthorizedUser(w http.ResponseWriter, r *http.Request) {
	message := "user is not authorized to access this resource"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (app *Application) alreadyExists(w http.ResponseWriter, r *http.Request) {
	message := "resource already exists"
	app.errorResponse(w, r, http.StatusConflict, message)
}
