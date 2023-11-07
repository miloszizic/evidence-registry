package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/go-chi/chi/v5"
	"github.com/miloszizic/der/service"

	"golang.org/x/exp/constraints"
)

var RgxEmail = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{1,}$`)

// NotBlank checks if the string has characters other than whitespaces.
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// MinRunes checks if the string has at least n runes.
func MinRunes(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

// MaxRunes checks if the string has at most n runes.
func MaxRunes(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// Between checks if the value lies between min and max, both inclusive.
func Between[T constraints.Ordered](value, min, max T) bool {
	return value >= min && value <= max
}

// Matches checks if the string value matches the given regular expression.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// In checks if the value exists in the provided safelist.
func In[T comparable](value T, safelist ...T) bool {
	for i := range safelist {
		if value == safelist[i] {
			return true
		}
	}

	return false
}

// AllIn checks if all the values exist in the provided safelist.
func AllIn[T comparable](values []T, safelist ...T) bool {
	for i := range values {
		if !In(values[i], safelist...) {
			return false
		}
	}

	return true
}

// NotIn checks if the value does not exist in the provided blocklist.
func NotIn[T comparable](value T, blocklist ...T) bool {
	for i := range blocklist {
		if value == blocklist[i] {
			return false
		}
	}

	return true
}

// NoDuplicates checks if the slice does not contain duplicate values.
func NoDuplicates[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}

	return len(values) == len(uniqueValues)
}

// IsEmail checks if the string is a valid email address.
func IsEmail(value string) bool {
	const maxEmailLength = 254

	if len(value) > maxEmailLength {
		return false
	}

	return RgxEmail.MatchString(value)
}

// IsURL checks if the string is a valid URL.
func IsURL(value string) bool {
	u, err := url.ParseRequestURI(value)
	if err != nil {
		return false
	}

	return u.Scheme != "" && u.Host != ""
}

// idParser is a helper function that parses the specified ID parameter from the URL.
func idParser(r *http.Request, paramName string) (uuid.UUID, error) {
	urlID := chi.URLParam(r, paramName)

	id, err := uuid.Parse(urlID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w : invalid id parameter", service.ErrInvalidRequest)
	}

	return id, nil
}

// caseIDParser is a helper function that extracts the 'caseID' parameter from the request URL.
// It delegates the parsing to a generic 'idParser' method, passing 'caseID' as the key.
// It returns the parsed ID as an uuid.UUID or an error if the parsing fails.
func (app *Application) caseIDParser(r *http.Request) (uuid.UUID, error) {
	return idParser(r, "caseID")
}

// caseTypeIDParser is a helper function that extracts the 'caseTypeID' parameter from the request URL.
// It delegates the parsing to a generic 'idParser' method, passing 'caseTypeID' as the key.
// It returns the parsed ID as an uuid.UUID or an error if the parsing fails.
func (app *Application) caseTypeIDParser(r *http.Request) (uuid.UUID, error) {
	return idParser(r, "caseTypeID")
}

// roleIDParser is a helper function that extracts the 'roleID' parameter from the request URL.
// It delegates the parsing to a generic 'idParser' method, passing 'roleID' as the key.
// It returns the parsed ID as an uuid.UUID or an error if the parsing fails.
func (app *Application) roleIDParser(r *http.Request) (uuid.UUID, error) {
	return idParser(r, "roleID")
}

// permissionIDParser is a helper function that extracts the 'permissionID' parameter from the request URL.
// It delegates the parsing to a generic 'idParser' method, passing 'permissionID' as the key.
// It returns the parsed ID as an uuid.UUID or an error if the parsing fails.
func permissionIDParser(r *http.Request) (uuid.UUID, error) {
	return idParser(r, "permissionID")
}

// evidenceIDParser is a helper function that extracts the 'evidenceID' parameter from the request URL.
// It delegates the parsing to a generic 'idParser' method, passing 'evidenceID' as the key.
// It returns the parsed ID as an uuid.UUID or an error if the parsing fails.
func evidenceIDParser(r *http.Request) (uuid.UUID, error) {
	return idParser(r, "evidenceID")
}

// userIDParser is a helper function that extracts the 'userID' parameter from the request URL.
// It delegates the parsing to a generic 'idParser' method, passing 'userID' as the key.
// It returns the parsed ID as an uuid.UUID or an error if the parsing fails.
func userIDParser(r *http.Request) (uuid.UUID, error) {
	return idParser(r, "userID")
}

// HealthCheck is an HTTP handler that checks the status of various components of the application and responds with a health status report.
// It verifies the connection to the database and file store, responding with 'online' if the connection is successful and 'offline' otherwise.
// A response is returned with HTTP status '200 OK' containing the health status of the application, database, and file store.
// If any component is 'offline', an HTTP '500 Internal Server Error' is returned.
func (app *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Create a response struct
	type healthCheckResponse struct {
		Application string `json:"application"`
		Database    string `json:"db"`
		FileStore   string `json:"file_store"`
	}

	response := healthCheckResponse{
		Application: "online",
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)

	defer cancel()
	// Check db connection
	if err := app.stores.DB.PingContext(ctx); err != nil {
		response.Database = "offline"

		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response.Database = "online"
	}
	// Check file store connection
	_, err := app.stores.ListCases(r.Context())
	if err != nil {
		response.FileStore = "offline"

		app.respondError(w, r, err)
	} else {
		response.FileStore = "online"
	}

	app.respond(w, r, http.StatusOK, envelope{"Health status": response})
}

// requestUserParser is a helper function that retrieves the authenticated user from the request context.
// It expects the context to include an 'authorizationPayloadKey' storing a Payload object, from which it retrieves the expectedUser.
// If the user is not found in the database, it returns an error.
func (app *Application) requestUserParser(r *http.Request) (*service.User, error) {
	authPayloadVal := r.Context().Value(authorizationPayloadKey)
	if authPayloadVal == nil {
		return nil, service.ErrMissingUser
	}

	authPayload, ok := authPayloadVal.(*service.Payload)
	if !ok {
		return nil, fmt.Errorf("unexpected value in request context")
	}

	user, err := app.stores.GetUserByUsername(r.Context(), authPayload.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("getting user : %w", service.ErrNotFound)
		}

		return nil, fmt.Errorf("getting user : %w", err)
	}

	return &user, nil
}

// respond is a helper function that sends a response with a provided HTTP status code and a payload.
// It uses the writeJSON method to write the data into the response.
// If an error occurs while writing the response, it triggers a server error response.
func (app *Application) respond(w http.ResponseWriter, r *http.Request, status int, data envelope) {
	err := app.writeJSON(w, status, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func paramsParser[T any](app *Application, r *http.Request) (T, error) {
	var params T

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		app.logger.Errorw("Error in decoding request body", "error", err)
		return params, service.ErrInvalidRequest
	}

	return params, nil
}

// fileParser is a helper function that extracts a file from the multipart form data in an HTTP request.
// It returns a Reader for reading the file contents, the filename, and any error that occurred.
// If the form does not include an 'upload_file' part or if the file is empty, it returns an error.
func (*Application) fileParser(r *http.Request) (io.Reader, string, error) {
	file, header, err := r.FormFile("upload_file")

	// If file is nil, return immediately
	if file == nil {
		return nil, "", fmt.Errorf("received nil file")
	}

	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return nil, "", fmt.Errorf("%w: no file to upload: %w", service.ErrInvalidRequest, err)
		}

		return nil, "", fmt.Errorf("parsing file: %w", err)
	}

	// Check if the file is empty
	pos, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, "", fmt.Errorf("seeking file: %w", err)
	}

	if pos == 0 {
		return nil, "", fmt.Errorf("%w: file is empty", service.ErrInvalidRequest)
	}

	// Seek back to the start of the file
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, "", fmt.Errorf("seeking file: %v", err)
	}

	return file, header.Filename, nil
}

// respondEvidence is a helper function that sends the contents of an evidence file in the HTTP response.
// After writing the file, it uses the writeJSON method to add a status message to the response.
// If an error occurs while writing the response, it triggers a server error response.

func (app *Application) respondEvidence(w http.ResponseWriter, r *http.Request, filename string, file io.Reader) {
	// Determine the MIME type based on a file extension
	ext := filepath.Ext(filename)
	mimeType := mime.TypeByExtension(ext)

	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Set headers
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", mimeType)

	// Respond with evidence content
	_, err := io.Copy(w, file)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("responding with evidence: %w", err))
		return
	}
}

type evidenceParams struct {
	Description    string    `json:"description"`
	EvidenceTypeID uuid.UUID `json:"evidence_type_id"`
}

// evidenceParamsParser is a helper function that extracts evidence parameters from the multipart form data in a HTTP request.
// It expects the form to include an 'evidence' field containing a JSON-encoded evidenceParams object.
// If the parsing is successful, it returns the parsed parameters. If not, it returns an error.
func evidenceParamsParser(r *http.Request) (evidenceParams, error) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return evidenceParams{}, fmt.Errorf("error parsing multipart form: %w", err)
	}

	evidenceJSON := r.FormValue("evidence") // Assuming "evidence" is the form field that contains the JSON.
	if evidenceJSON == "" {
		return evidenceParams{}, fmt.Errorf("%w: missing evidence field", service.ErrInvalidRequest)
	}

	request := evidenceParams{}

	if err := json.Unmarshal([]byte(evidenceJSON), &request); err != nil {
		return evidenceParams{}, fmt.Errorf("error unmarshalling evidence data: %w", err)
	}

	return request, nil
}
