package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/miloszizic/der/vault"

	"github.com/miloszizic/der/service"
	"go.uber.org/zap"
)

func (app *Application) logError(r *http.Request, err error) {
	if app.logger == nil {
		fmt.Println("Logger is not initialized")
		return
	}

	if r == nil || r.URL == nil {
		app.logger.Error("received nil request or nil URL",
			zap.Error(err))
		return
	}

	app.logger.Error("error processing request", zap.String("request_method", r.Method), zap.String("request_url", r.URL.String()),
		zap.Error(err))
}

func (app *Application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{"error": message}

	const internalServerError = 500

	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(internalServerError)
	}
}

// serverErrorResponse is a helper function that handles internal server errors.
// It logs the error, then sends a '500 Internal Server Error' status with a standard error message.
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
	app.logError(r, err)

	message := "the request could not be understood by the server due to malformed syntax"

	app.errorResponse(w, r, http.StatusBadRequest, message)
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

func (app *Application) failedValidation(w http.ResponseWriter, r *http.Request, v service.Validator) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, v)
}

// the respondError writes an error response to all kinds of errors.
func (app *Application) respondError(w http.ResponseWriter, r *http.Request, err error) {
	var syntaxError *json.SyntaxError

	var unmarshalTypeError *json.UnmarshalTypeError

	var invalidUnmarshalError *json.InvalidUnmarshalError

	switch {
	case errors.As(err, &syntaxError):
		app.badRequestResponse(w, r, err)

	case errors.As(err, &unmarshalTypeError):
		app.badRequestResponse(w, r, err)

	case errors.As(err, &invalidUnmarshalError):
		app.badRequestResponse(w, r, err)

	case errors.Is(err, io.ErrUnexpectedEOF):
		app.badRequestResponse(w, r, errors.New("body contains badly-formed JSON"))

	case errors.Is(err, io.EOF):
		app.badRequestResponse(w, r, errors.New("body must not be empty"))

	case strings.HasPrefix(err.Error(), "json: unknown field "):
		fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
		app.badRequestResponse(w, r, fmt.Errorf("body contains unknown key %s", fieldName))

	case err.Error() == "http: request body too large":
		app.badRequestResponse(w, r, err)

	case errors.Is(err, sql.ErrNoRows), errors.Is(err, service.ErrNotFound), errors.Is(err, vault.ErrNotFound):
		app.notFoundResponse(w, r)

	case errors.Is(err, service.ErrMissingUser):
		app.badRequestResponse(w, r, err)

	case errors.Is(err, service.ErrAlreadyExists), errors.Is(err, vault.ErrAlreadyExists):
		app.alreadyExists(w, r)

	case errors.Is(err, service.ErrInvalidRequest), errors.Is(err, vault.ErrInvalidRequest):
		app.badRequestResponse(w, r, err)

	case errors.Is(err, service.ErrUnauthorized):
		app.unauthorizedUser(w, r)

	case errors.Is(err, service.ErrInvalidCredentials):
		app.invalidCredentialsResponse(w, r)

	default:
		app.serverErrorResponse(w, r, err)
	}
}
