package api

import (
	"fmt"
	"net/http"

	"github.com/miloszizic/der/service"
)

// CreateEvidenceHandler is an HTTP handler function that creates a new evidence and associates it with a specific case.
// The request must include the case's ID as a parameter caseID in URL.
// The request should also contain a multipart/form-data body with fields:
// 'upload_file' - the evidence file to be uploaded, 'evidence' - a JSON-encoded object with evidence parameters.
// The logged-in user (obtained from the request context) is assigned as the author of the new evidence.
func (app *Application) CreateEvidenceHandler(w http.ResponseWriter, r *http.Request) {

	userCTX := r.Context().Value(userContextKey)

	user, ok := userCTX.(*service.User)
	if !ok {
		app.logger.Errorw("Error getting user from context", "error", fmt.Errorf("user is not of type *db.AppUser"))
		app.respondError(w, r, fmt.Errorf("*db.AppUser type assertion failed"))

		return
	}

	evParams, err := evidenceParamsParser(r)
	if err != nil {
		fmt.Printf("Handler error: %v\n", err)
		app.respondError(w, r, err)

		return
	}

	file, fileName, err := app.fileParser(r)
	if err != nil {
		app.respondError(w, r, err)
		fmt.Printf("Handler error: %v\n", err)

		return
	}
	// parse CaseID from URL
	caseID, err := app.caseIDParser(r)
	if err != nil {
		app.respondError(w, r, err)
		fmt.Printf("Handler error: %v\n", err)

		return
	}

	evidenceParams := service.CreateEvidenceParams{
		AppUserID:      user.ID,
		CaseID:         caseID,
		Name:           fileName,
		Description:    evParams.Description,
		EvidenceTypeID: evParams.EvidenceTypeID,
	}

	ev, err := app.stores.CreateEvidence(r.Context(), evidenceParams, file)
	if err != nil {
		app.respondError(w, r, err)
		return
	}

	app.respond(w, r, http.StatusCreated, envelope{"Evidence": ev})
}

// GetEvidenceHandler is an HTTP handler function that fetches and returns details of a specific evidence.
// The request must include the evidence's ID as a parameter evidenceID in URL.
func (app *Application) GetEvidenceHandler(w http.ResponseWriter, r *http.Request) {
	evID, err := app.evidenceIDParser(r)
	if err != nil {
		app.respondError(w, r, err)
		return
	}

	evidence, err := app.stores.GetEvidenceByID(r.Context(), evID)
	if err != nil {
		app.respondError(w, r, err)
		return
	}

	app.respond(w, r, http.StatusOK, envelope{"Evidence": evidence})
}

// ListEvidencesHandler is an HTTP handler function that fetches and returns a list of evidences for a specific case.
// The request must include the case's ID as a parameter caseID in URL.
func (app *Application) ListEvidencesHandler(w http.ResponseWriter, r *http.Request) {
	csID, err := app.caseIDParser(r)
	if err != nil {
		app.respondError(w, r, err)
		return
	}

	cs, err := app.stores.GetCaseByID(r.Context(), csID)
	if err != nil {
		app.respondError(w, r, err)
		return
	}

	evidences, err := app.stores.ListEvidences(r.Context(), &cs)
	if err != nil {
		app.respondError(w, r, err)
		return
	}

	app.respond(w, r, http.StatusOK, envelope{"evidences": evidences})
}

// DownloadEvidenceHandler is an HTTP handler function that fetches and serves a specific evidence file.
// The request must include the evidence's ID as a parameter evidenceID in URL.
func (app *Application) DownloadEvidenceHandler(w http.ResponseWriter, r *http.Request) {
	// Get evidence from the request
	ev, err := app.evidenceIDParser(r)
	if err != nil {
		app.respondError(w, r, err)
		return
	}

	evidence, err := app.stores.GetEvidenceByID(r.Context(), ev)
	if err != nil {
		app.respondError(w, r, err)
		return
	}

	// Get evidence from the ObjectStore
	file, filename, err := app.stores.DownloadEvidence(r.Context(), *evidence)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	defer file.Close()

	// Respond with evidence content and headers
	app.respondEvidence(w, r, filename, file)
}

// ListEvidenceTypesHandler is an HTTP handler function that fetches and returns a list of evidence types.
func (app *Application) ListEvidenceTypesHandler(w http.ResponseWriter, r *http.Request) {
	evidenceTypes, err := app.stores.ListEvidenceTypes(r.Context())
	if err != nil {
		app.respondError(w, r, err)
		return
	}

	app.respond(w, r, http.StatusOK, envelope{"EvidenceTypes": evidenceTypes})
}
