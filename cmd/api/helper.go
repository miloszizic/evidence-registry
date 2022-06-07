package api

import (
	"encoding/json"
	"errors"
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
		return nil, data.WrapErrorf(err, data.ErrCodeInvalid, "invalid id parameter")
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
		return nil, data.WrapErrorf(err, data.ErrCodeInvalid, "invalid id parameter")
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

func (app *Application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
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
	w.Write(js)

	return nil
}

func (app *Application) readJSON(r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1_048_576))
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return data.NewErrorf(data.ErrCodeInvalid, "body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return data.NewErrorf(data.ErrCodeInvalid, "body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return data.NewErrorf(data.ErrCodeInvalid, "body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return data.NewErrorf(data.ErrCodeInvalid, "body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return data.NewErrorf(data.ErrCodeInvalid, "body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return data.NewErrorf(data.ErrCodeInvalid, "body contains unknown key %s", fieldName)

		case err.Error() == "http: request body too large":
			return data.NewErrorf(data.ErrCodeInvalid, "body must not be larger than %d bytes", 1_048_576)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return data.WrapErrorf(err, data.ErrCodeUnknown, "unexpected error")
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return data.NewErrorf(data.ErrCodeInvalid, "body must only contain a single JSON value")
	}

	return nil
}

// respondError writes an error response to all kinds of errors.
func (app *Application) respondError(w http.ResponseWriter, r *http.Request, err error) {
	var verr *data.Error
	if errors.As(err, &verr) {
		switch verr.Code() {
		case data.ErrCodeNotFound:
			app.notFoundResponse(w, r)
		case data.ErrCodeConflict:
			app.alreadyExists(w, r)
		case data.ErrCodeInvalidCredentials:
			app.invalidCredentialsResponse(w, r)
		case data.ErrCodeExists:
			app.alreadyExists(w, r)
		case data.ErrCodeUnknown:
			app.serverErrorResponse(w, r, err)
		case data.ErrCodeInvalid:
			app.badRequestResponse(w, r, err)
		default:
			app.serverErrorResponse(w, r, err)
		}
	}
	app.serverErrorResponse(w, r, err)
}
