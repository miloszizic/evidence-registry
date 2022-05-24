package api

import (
	"encoding/json"
	"errors"
	"evidence/internal/data"
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strconv"
	"strings"
)

//func (app *Application) contextUserChecker(r *http.Request) (*data.User, error) {
//	authPayload := r.Context().Value(authorizationPayloadKey).(*Payload)
//	if authPayload == nil {
//		return nil, errors.New("no user found in context")
//	}
//	user, err := app.stores.UserDB.GetByUsername(authPayload.Username)
//	if err != nil {
//		switch {
//		case err == sql.ErrNoRows:
//			return nil, errors.New("user not found")
//		default:
//			return nil, err
//		}
//	}
//	if user.Role != "admin" {
//		return nil, errors.New("user is not an admin")
//	}
//	return user, nil
//}

// readIDParam reads an id param from a request.
func (app *Application) getCaseFromUrl(r *http.Request) (*data.Case, error) {
	urlID := chi.URLParam(r, "caseID")
	id, err := strconv.ParseInt(urlID, 10, 64)
	if err != nil || id < 1 {
		return nil, errors.New("invalid id parameter")
	}
	cs, err := app.stores.CaseDB.GetByID(id)
	if err != nil {
		return nil, err
	}
	return cs, nil
}
func (app *Application) getEvidenceFromUrl(r *http.Request) (*data.Evidence, error) {
	urlID := chi.URLParam(r, "evidenceID")
	id, err := strconv.ParseInt(urlID, 10, 64)
	if err != nil || id < 1 {
		return nil, errors.New("invalid id parameter")
	}
	ev, err := app.stores.EvidenceDB.GetByID(id)
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

func (app *Application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}
