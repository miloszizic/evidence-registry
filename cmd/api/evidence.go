package api

import (
	"evidence/internal/data"
	"io"
	"net/http"
)

// CreateEvidenceHandler creates an evidence in a specific case
func (app *Application) CreateEvidenceHandler(w http.ResponseWriter, r *http.Request) {
	cs, err := app.caseParser(r)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	// get the evidence file from the request body
	ev, err := app.fileParser(r, cs)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	err = app.stores.CreateEvidence(ev, cs)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	app.respondEvidenceCreated(w, r, ev)
}

// ListEvidencesHandler returns all evidences for a case by comparing evidences in the
// database with the ones in the OBStore
func (app *Application) ListEvidencesHandler(w http.ResponseWriter, r *http.Request) {
	cs, err := app.caseParser(r)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	evidences, err := app.stores.ListEvidences(cs)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	app.respondEvList(w, r, evidences)
}

// GetEvidenceHandler returns an evidence from the database and the OBStore
func (app *Application) GetEvidenceHandler(w http.ResponseWriter, r *http.Request) {
	// get evidence from the request
	ev, err := app.evidenceParser(r)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	// get evidence from the OBStore
	file, err := app.stores.DownloadEvidence(ev)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	// respond with evidence content
	err = app.respondEvidence(w, r, *file)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
}

// DeleteEvidenceHandler deletes an evidence from the database and the OBStore
func (app *Application) DeleteEvidenceHandler(w http.ResponseWriter, r *http.Request) {
	// get evidence from the request
	ev, err := app.evidenceParser(r)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	// delete evidence from the request
	err = app.stores.DeleteEvidence(ev)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	app.respondEvidenceDeleted(w, r, ev)
}

// AddCommentHandler adds a comment to an evidence in the database
func (app *Application) AddCommentHandler(w http.ResponseWriter, r *http.Request) {
	// get comment from the request
	cm, err := app.commentParser(r)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	// add comment to the database
	err = app.stores.AddEvidenceComment(cm)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	// respond with comment
	app.respondComment(w, r, cm)
}

// fileParser parses the evidence from the request body and returns it
func (app *Application) fileParser(r *http.Request, cs *data.Case) (*data.Evidence, error) {
	file, handler, err := r.FormFile("upload_file")
	if file == nil {
		return nil, data.NewErrorf(data.ErrCodeInvalid, "api: no file in request")
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	evidence := &data.Evidence{
		Name:   handler.Filename,
		CaseID: cs.ID,
		File:   file,
	}
	return evidence, nil
}

// commentParser parses the comment from the request and returns it
func (app *Application) commentParser(r *http.Request) (*data.Comment, error) {
	// get evidence from the request
	ev, err := app.evidenceParser(r)
	if err != nil {
		return nil, err
	}
	// get comment from the request
	var cm data.Comment
	err = app.readJSON(r, &cm)
	if err != nil {
		return nil, err
	}
	comment := data.Comment{
		Text:       cm.Text,
		EvidenceID: ev.ID,
	}
	return &comment, nil
}

// respondComment writes the comment to the response and sets the status code to 200
func (app *Application) respondComment(w http.ResponseWriter, r *http.Request, cm *data.Comment) {
	err := app.writeJSON(w, http.StatusCreated, envelope{"comment": cm}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// respondEvidence returns a response with the evidence content with status code 200
func (app *Application) respondEvidence(w http.ResponseWriter, r *http.Request, file io.ReadCloser) error {
	// respond with evidence content
	_, err := io.Copy(w, file)
	if err != nil {
		return err
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"evidence": "downloaded"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	return nil
}

// respondEvList writes the evidence list to the response and sets the status code to 200
func (app *Application) respondEvList(w http.ResponseWriter, r *http.Request, evidences []data.Evidence) {
	err := app.writeJSON(w, http.StatusOK, envelope{"evidences": evidences}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// respondEvidenceDeleted returns a response with the evidence deleted with status code 200
func (app *Application) respondEvidenceDeleted(w http.ResponseWriter, r *http.Request, ev *data.Evidence) {
	err := app.writeJSON(w, http.StatusOK, envelope{"Evidence was successfully deleted": ev}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// respondEvidenceCreated returns a response with the evidence created with status code 201
func (app *Application) respondEvidenceCreated(w http.ResponseWriter, r *http.Request, ev *data.Evidence) {
	err := app.writeJSON(w, http.StatusCreated, envelope{"evidence": ev}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
