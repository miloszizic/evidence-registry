package api

import (
	"errors"
	"net/http"
)

var (
	permissionDenied = errors.New("permission denied")
)

func (app *Application) logError(r *http.Request, err error) {
	app.logger.PrintError(err, map[string]string{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})
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

func (app *Application) invalidAuthorisationHeader(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication header"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
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

func (app *Application) notPermittedResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account doesn't have the necessary permissions to access this resource"
	app.errorResponse(w, r, http.StatusForbidden, message)
}

func (app *Application) evidenceAlreadyExists(w http.ResponseWriter, r *http.Request) {
	message := "evidence already exists"
	app.errorResponse(w, r, http.StatusConflict, message)
}
func (app *Application) noEvidenceAttacked(w http.ResponseWriter, r *http.Request) {
	message := "no evidence in request body"
	app.errorResponse(w, r, http.StatusBadRequest, message)
}
