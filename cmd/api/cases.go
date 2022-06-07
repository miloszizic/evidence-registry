package api

import (
	"github.com/miloszizic/der/internal/data"
	"net/http"
	"strings"
)

// caseRequest is the request body for the case API
type caseRequest struct {
	Name string `json:"name"`
	Tag  string `json:"tag"`
}

// CreateCaseHandler creates a new case in the database and OBStore
func (app *Application) CreateCaseHandler(w http.ResponseWriter, r *http.Request) {
	user, name, err := app.requestParser(r)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	err = app.stores.CreateCase(user, name)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	app.respond(w, r, http.StatusCreated, envelope{"Case": "case successfully created"})
}

// RemoveCaseHandler removes a case from the database and OBStore
func (app *Application) RemoveCaseHandler(w http.ResponseWriter, r *http.Request) {
	_, name, err := app.requestParser(r)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	// delete case
	err = app.stores.RemoveCase(name)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	app.respond(w, r, http.StatusOK, envelope{"Case": "case successfully deleted"})
}
func (app *Application) GetCaseHandler(w http.ResponseWriter, r *http.Request) {
	cs, err := app.caseParser(r)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	// respond with case
	app.respond(w, r, http.StatusOK, envelope{"Case": cs})
}

// ListCasesHandler returns a list of cases that exist in bought database and storage
func (app *Application) ListCasesHandler(w http.ResponseWriter, r *http.Request) {
	// get cases
	cases, err := app.stores.ListCases()
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	// respond with cases
	app.respond(w, r, http.StatusOK, envelope{"Cases": cases})
}

// requestParser takes a request and returns a user and the case name
// if the request is not valid it returns an error
func (app *Application) requestParser(r *http.Request) (*data.User, string, error) {
	authPayload := r.Context().Value(authorizationPayloadKey).(*Payload)
	user, err := app.stores.User.GetByUsername(authPayload.Username)
	if err != nil {
		return nil, "", err
	}
	//read JSON request
	var req caseRequest
	err = app.readJSON(r, &req)
	if err != nil {
		if strings.HasPrefix(err.Error(), "body") {
			return nil, "", data.NewErrorf(data.ErrCodeInvalid, "readJSON")
		}
		return nil, "", data.WrapErrorf(err, data.ErrCodeUnknown, "requestParser")
	}
	return user, req.Name, nil
}

func (app *Application) respond(w http.ResponseWriter, r *http.Request, status int, data envelope) {
	err := app.writeJSON(w, status, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
