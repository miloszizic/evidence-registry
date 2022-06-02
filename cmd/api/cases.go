package api

import (
	"evidence/internal/data"
	"net/http"
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
	app.respondCaseCreated(w, r)
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
	app.respondCaseDeleted(w, r)
}
func (app *Application) GetCaseHandler(w http.ResponseWriter, r *http.Request) {
	cs, err := app.caseParser(r)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	// respond with case
	app.respondCaseRetrieved(w, r, cs)
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
	app.respondCasesList(w, r, cases)
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
		return nil, "", err
	}
	return user, req.Name, nil
}

// respondCaseCreated writes a JSON response with the case created
func (app *Application) respondCaseRetrieved(w http.ResponseWriter, r *http.Request, cs *data.Case) {
	err := app.writeJSON(w, http.StatusOK, envelope{"case": cs}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// respondCaseCreated with status for created case
func (app *Application) respondCaseCreated(w http.ResponseWriter, r *http.Request) {
	err := app.writeJSON(w, http.StatusCreated, envelope{"Case": "case successfully created"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// respondCaseDeleted responds with status for deleted case
func (app *Application) respondCaseDeleted(w http.ResponseWriter, r *http.Request) {
	err := app.writeJSON(w, http.StatusOK, envelope{"Case": "case successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// respondCasesList responds with a list of cases in JSON
func (app *Application) respondCasesList(w http.ResponseWriter, r *http.Request, cases []data.Case) {
	err := app.writeJSON(w, http.StatusOK, envelope{"Cases": cases}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
