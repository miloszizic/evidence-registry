package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DecodeJSON is a helper function that decodes a JSON request body into a given destination.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	return decodeJSON(w, r, dst, false)
}

// DecodeJSONStrict is a helper function that decodes a JSON request body into a given destination.
func DecodeJSONStrict(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	return decodeJSON(w, r, dst, true)
}

// decodeJSON is a helper function that decodes a JSON request body into a given destination with the option to disallow unknown fields.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst interface{}, disallowUnknownFields bool) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)

	if disallowUnknownFields {
		dec.DisallowUnknownFields()
	}

	err := dec.Decode(dst)
	if err != nil {

		fmt.Println("DEBUG ERROR:", err.Error()) // Added this line temporarily
		var syntaxError *json.SyntaxError

		var unmarshalTypeError *json.UnmarshalTypeError

		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d): %w", syntaxError.Offset, err)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return fmt.Errorf("body contains badly-formed JSON: %w", err)

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q: %w", unmarshalTypeError.Field, err)
			}

			return fmt.Errorf("body contains incorrect JSON type (at character %d): %w", unmarshalTypeError.Offset, err)

		case errors.Is(err, io.EOF):
			return fmt.Errorf("body must not be empty: %w", err)

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s: %w", fieldName, err)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes: %w", maxBytes, err)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return fmt.Errorf("body must only contain a single JSON value: %w", err)
	}

	return nil
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
		fmt.Println("Error writing to ResponseWriter:", err)
	}

	return err
}
