package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/miloszizic/der/internal/data"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// readIDParam reads an id param from a request.
func (app *Application) caseParser(r *http.Request) (*data.Case, error) {
	urlID := chi.URLParam(r, "caseID")
	id, err := strconv.ParseInt(urlID, 10, 64)
	if err != nil || id < 1 {
		return nil, fmt.Errorf("%w : invalid id parameter", data.ErrInvalidRequest)
	}
	cs, err := app.stores.GetCaseByID(id)
	if err != nil {
		return nil, err
	}
	return cs, nil
}

// evidenceParser parses the request url and returns caseID and evidenceID.
func (app *Application) evidenceParser(r *http.Request) (*data.Evidence, error) {
	evID := chi.URLParam(r, "evidenceID")
	id, err := strconv.ParseInt(evID, 10, 64)
	if err != nil || id < 1 {
		return nil, fmt.Errorf("%w : invalid id parameter", data.ErrInvalidRequest)
	}
	cs, err := app.caseParser(r)
	if err != nil {
		return nil, err
	}
	ev, err := app.stores.GetEvidenceByID(id, cs.ID)
	if err != nil {
		return nil, err
	}
	return ev, nil
}

// Envelope type for better documentation, also it's to make sure that your JSON
// always returns its response as a non-array JSON object for security reasons.
type envelope map[string]interface{}

func (*Application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "Application/json")
	w.WriteHeader(status)
	_, err = w.Write(js)
	if err != nil {
		return err
	}

	return nil
}

func (*Application) readJSON(r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1_048_576))
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("%w: request body contains badly-formed JSON (at position %d)", data.ErrInvalidRequest, syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return fmt.Errorf("%w: request body contains badly-formed JSON", data.ErrInvalidRequest)

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("%w: request body contains an invalid value for the %q field (at position %d)", data.ErrInvalidRequest, unmarshalTypeError.Field, unmarshalTypeError.Offset)
			}
			return fmt.Errorf("%w: request body contains an invalid value (at position %d)", data.ErrInvalidRequest, unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return fmt.Errorf("%w: request body must not be empty", data.ErrInvalidRequest)

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("%w: request body contains unknown field %s", data.ErrInvalidRequest, fieldName)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("%w: request body is larger than %d ", data.ErrInvalidRequest, 1_048_576)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return fmt.Errorf("reading JSON : %w: ", err)
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return fmt.Errorf("%w: request body must only contain a single JSON object", data.ErrInvalidRequest)
	}

	return nil
}

// respondError writes an error response to all kinds of errors.
func (app *Application) respondError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		app.unauthorizedUser(w, r)
	case errors.Is(err, data.ErrNotFound):
		app.notFoundResponse(w, r)
	case errors.Is(err, data.ErrAlreadyExists):
		app.alreadyExists(w, r)
	case errors.Is(err, data.ErrInvalidRequest):
		app.badRequestResponse(w, r, err)
	case errors.Is(err, data.ErrUnauthorized):
		app.unauthorizedUser(w, r)
	case errors.Is(err, data.ErrInvalidCredentials):
		app.invalidCredentialsResponse(w, r)
	default:
		app.serverErrorResponse(w, r, err)
	}
}
