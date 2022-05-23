package api

import (
	"database/sql"
	"evidence/internal/data"
	"fmt"
	"net/http"
)

type caseRequest struct {
	Name string `json:"name"`
	Tag  string `json:"tag"`
}

//CreateCaseHandler makes a new case for a user in the database and FS
func (app *Application) CreateCaseHandler(w http.ResponseWriter, r *http.Request) {
	//check payload for user
	user, err := app.payloadUserChecker(r)
	if err != nil {
		switch {
		case err == ErrInvalidToken:
			app.notPermittedResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}
	//read JSON request
	var req caseRequest
	err = app.readJSON(w, r, &req)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// declare the case
	cs := &data.Case{
		Name: req.Name,
	}
	//create case in minio storage
	err = app.stores.StoreFS.AddCase(cs)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// create a bucket in the database
	err = app.stores.CaseDB.Add(cs, user)
	if err != nil {
		// delete the bucket from FS of the case was not created in the database
		err = app.stores.StoreFS.RemoveCase(cs.Name)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		app.serverErrorResponse(w, r, fmt.Errorf("failed to create the case in db: %v", err))
		return
	}
	//write the response
	err = app.writeJSON(w, http.StatusCreated, envelope{"Case": "case successfully created"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// RemoveCaseHandler removes a case from the database and FS
func (app *Application) RemoveCaseHandler(w http.ResponseWriter, r *http.Request) {
	// check payload for user
	_, err := app.payloadUserChecker(r)
	if err != nil {
		app.notPermittedResponse(w, r)
		return
	}
	// read JSON request
	var req caseRequest
	err = app.readJSON(w, r, &req)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// remove case from FS
	cs, err := app.stores.CaseDB.GetByName(req.Name)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}
	err = app.stores.StoreFS.RemoveCase(req.Name)
	if err != nil {
		switch {
		case err.Error() == "case does not exist":
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}
	// remove the case in the database
	err = app.stores.CaseDB.Remove(cs)
	if err != nil {
		switch {
		case err.Error() == "case not found":
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}
	// write the response
	err = app.writeJSON(w, http.StatusOK, envelope{"CaseDB": "case successfully removed"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// ListCasesHandler returns a list of cases that exist in bought database and storage
func (app *Application) ListCasesHandler(w http.ResponseWriter, r *http.Request) {
	//	check the payload for user
	_, err := app.payloadUserChecker(r)
	if err != nil {
		app.notPermittedResponse(w, r)
	}
	// get all cases from the database
	dbCases, err := app.stores.CaseDB.List()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	//get all cases in the FS
	storageCases, err := app.stores.StoreFS.ListCases()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// remove cases that are not in the database
	var finalListOfCases []data.Case
	for _, dbCase := range dbCases {
		for _, storageCase := range storageCases {
			if dbCase.Name == storageCase.Name {
				finalListOfCases = append(finalListOfCases, dbCase)
			}
		}
	}
	// write the response
	err = app.writeJSON(w, http.StatusOK, envelope{"Cases": finalListOfCases}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
