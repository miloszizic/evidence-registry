package api

import (
	"fmt"
	"net/http"

	"github.com/miloszizic/der/service"
)

// CreateCaseHandler is an HTTP handler that creates a new case in the system.
// The request must include the authenticated user's details in its context payload and
// case parameters in JSON format in its body. The parameters must include CaseTypeID, CaseNumber,
// CaseYear, CaseCourtID, and an optional array of tags. Upon successful creation, it returns a '201 Created'
func (app *Application) CreateCaseHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from context payload set by UserParserMiddleware
	userCTX := r.Context().Value(userContextKey)

	user, ok := userCTX.(*service.User)
	if !ok {
		app.logger.Errorw("Error getting user from context", "error", fmt.Errorf("user is not of type *db.AppUser"))
		app.respondError(w, r, fmt.Errorf("*service.User type assertion failed"))

		return
	}

	params, err := paramsParser[service.CreateCaseParams](app, r)
	if err != nil {
		app.logger.Errorw("Error parsing params from request", "error", err)
		app.respondError(w, r, err)

		return
	}
	// Create the case
	createdCase, err := app.stores.CreateCase(r.Context(), user.ID, params)
	if err != nil {
		app.logger.Errorw("Error creating case", "error", err)
		app.respondError(w, r, err)

		return
	}

	app.respond(w, r, http.StatusCreated, envelope{"Case": createdCase})
}

// GetCaseHandler is an HTTP handler that retrieves a specific case from the system.
// The request must include the case's ID as a parameter. If the case is found,
// it responds with a '200 OK' status and the case's details. Otherwise, an error response is returned.
func (app *Application) GetCaseHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.caseIDParser(r)
	if err != nil {
		app.logger.Errorw("Error parsing case ID from request", "error", err)
		app.respondError(w, r, err)

		return
	}

	cs, err := app.stores.GetCaseByID(r.Context(), id)
	if err != nil {
		app.logger.Errorw("Error getting case by ID", "error", err)
		app.respondError(w, r, err)

		return
	}
	// respond with a case
	app.respond(w, r, http.StatusOK, envelope{"Case": cs})
}

// DeleteCaseHandler is an HTTP handler that deletes a specific case from the system.
// The request must include the case's ID in the 'caseID' parameter in the URL.
// If the deletion is successful, it responds with a '200 OK' status and a success message.
// In case of an error, it responds with the corresponding error message.
func (app *Application) DeleteCaseHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.caseIDParser(r)
	if err != nil {
		app.logger.Errorw("Error parsing case ID from request", "error", err)
		app.respondError(w, r, err)

		return
	}

	err = app.stores.DeleteCase(r.Context(), id)
	if err != nil {
		app.logger.Errorw("Error deleting case", "error", err)
		app.respondError(w, r, err)

		return
	}

	app.respond(w, r, http.StatusOK, envelope{"Case": "case deleted successfully"})
}

// ListCasesHandler is an HTTP handler that retrieves a list of all cases in the system.
// It doesn't require any specific parameter in the request.
// If successful, it responds with a '200 OK' status and a list of cases. In case of an error, it responds with the corresponding error message.
func (app *Application) ListCasesHandler(w http.ResponseWriter, r *http.Request) {
	// get cases
	cases, err := app.stores.ListCases(r.Context())
	if err != nil {
		app.logger.Errorw("Error listing cases", "error", err)
		app.respondError(w, r, err)

		return
	}
	// respond with cases
	app.respond(w, r, http.StatusOK, envelope{"Cases": cases})
}

// CreateCaseTypeHandler is an HTTP handler that creates a new case type in the system.
// The request must include the authenticated user's details in its context payload and
// case type parameters in JSON format in its body. The parameters must include Name, Description,
func (app *Application) CreateCaseTypeHandler(w http.ResponseWriter, r *http.Request) {
	// parse params from request
	params, err := paramsParser[service.CaseType](app, r)
	if err != nil {
		app.logger.Errorw("Error parsing params from request", "error", err)
		app.respondError(w, r, err)

		return
	}
	// create a case type
	_, err = app.stores.CreateCaseType(r.Context(), params)
	if err != nil {
		app.logger.Errorw("Error creating case type", "error", err)
		app.respondError(w, r, err)

		return
	}

	app.respond(w, r, http.StatusCreated, envelope{"Case Type": "successfully created"})
}

// GetCaseTypeHandler is an HTTP handler that retrieves a specific case type from the system.
// The request must include the case type's ID as a parameter. If the case type is found,
// it responds with a '200 OK' status and the case type's details. Otherwise, an error response is returned.
func (app *Application) GetCaseTypeHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.caseTypeIDParser(r)
	if err != nil {
		app.logger.Errorw("Error parsing case type ID from request", "error", err)
		app.respondError(w, r, err)

		return
	}

	caseType, err := app.stores.GetCaseTypeByID(r.Context(), id)
	if err != nil {
		app.logger.Errorw("Error getting case type by ID", "error", err)
		app.respondError(w, r, err)
		return
	}
	// respond with a case type
	app.respond(w, r, http.StatusOK, envelope{"CaseType": caseType})
}

// ListCaseTypesHandler is an HTTP handler that retrieves a list of all case types in the system.
// It doesn't require any specific parameter in the request.
func (app *Application) ListCaseTypesHandler(w http.ResponseWriter, r *http.Request) {
	// get case types
	caseTypes, err := app.stores.ListCaseTypes(r.Context())
	if err != nil {
		app.logger.Errorw("Error getting case types", "error", err)
		app.respondError(w, r, err)

		return
	}
	// respond with case types
	app.respond(w, r, http.StatusOK, envelope{"CaseTypes": caseTypes})
}

// UpdateCaseTypeHandler is an HTTP handler that updates a specific case type in the system.
// The request must include the case type's ID as a parameter and case type parameters in JSON format in its body.
func (app *Application) UpdateCaseTypeHandler(w http.ResponseWriter, r *http.Request) {
	// parse case type ID from request
	id, err := app.caseTypeIDParser(r)
	if err != nil {
		app.logger.Errorw("Error parsing case type ID from request", "error", err)
		app.respondError(w, r, err)

		return
	}
	// parse params from request
	params, err := paramsParser[service.CaseType](app, r)
	if err != nil {
		app.logger.Errorw("Error parsing params from request", "error", err)
		app.respondError(w, r, err)

		return
	}
	// set case type ID
	params.ID = id
	// update case type
	_, err = app.stores.UpdateCaseType(r.Context(), params)
	if err != nil {
		app.logger.Errorw("Error updating case type", "error", err)
		app.respondError(w, r, err)

		return
	}

	app.respond(w, r, http.StatusOK, envelope{"CaseType": "case type updated successfully"})
}

// ListCourtsHandler is an HTTP handler that retrieves a list of all courts in the system.
// It doesn't require any specific parameter in the request.
func (app *Application) ListCourtsHandler(w http.ResponseWriter, r *http.Request) {
	// get courts
	courts, err := app.stores.GetCourts(r.Context())
	if err != nil {
		app.logger.Errorw("Error getting courts", "error", err)
		app.respondError(w, r, err)
		return
	}
	// respond with courts
	app.respond(w, r, http.StatusOK, envelope{"Courts": courts})
}
