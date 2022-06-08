package api

import (
	"errors"
	"fmt"
	"github.com/miloszizic/der/internal/data"
	"io"
	"net/http"
)

// CreateEvidenceHandler creates an evidence in a specific case
func (app *Application) CreateEvidenceHandler(w http.ResponseWriter, r *http.Request) {
	cs, err := app.caseParser(r)
	if err != nil {
		fmt.Println(err)
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
	app.respond(w, r, http.StatusCreated, envelope{"Evidence": ev})
}

// ListEvidencesHandler returns all evidences for a case by comparing evidences in the
// database with the ones in the ObjectStore
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
	app.respond(w, r, http.StatusOK, envelope{"evidences": evidences})
}

// DownloadEvidenceHandler returns an evidence from the database and the ObjectStore
func (app *Application) DownloadEvidenceHandler(w http.ResponseWriter, r *http.Request) {
	// get evidence from the request
	ev, err := app.evidenceParser(r)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	// get evidence from the ObjectStore
	file, err := app.stores.DownloadEvidence(ev)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	//respond with evidence content
	err = app.respondEvidence(w, r, *file)
	if err != nil {
		app.respondError(w, r, err)
		return
	}

}

// DeleteEvidenceHandler deletes an evidence from the database and the ObjectStore
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

	app.respond(w, r, http.StatusOK, envelope{"evidence": "successfully deleted"})
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
	app.respond(w, r, http.StatusCreated, envelope{"comment": "successfully added comment"})
}

// fileParser parses the evidence from the request body and returns it
func (*Application) fileParser(r *http.Request, cs *data.Case) (*data.Evidence, error) {
	file, handler, err := r.FormFile("upload_file")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return nil, fmt.Errorf("%w : no file to uploaded : %v", data.ErrInvalidRequest, err)
		}
		return nil, fmt.Errorf("parsing file : %w", err)
	}
	if file == nil {
		return nil, fmt.Errorf("%w: file is empty", data.ErrInvalidRequest)
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
		return nil, fmt.Errorf("parsing JSON comment : %w", err)
	}
	comment := data.Comment{
		Text:       cm.Text,
		EvidenceID: ev.ID,
	}
	return &comment, nil
}

// respondEvidence returns a response with the evidence content with status code 200
func (app *Application) respondEvidence(w http.ResponseWriter, r *http.Request, file io.ReadCloser) error {
	// respond with evidence content
	_, err := io.Copy(w, file)
	if err != nil {
		return fmt.Errorf("responding with evidence : %w", err)
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"evidence": "downloaded"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	return nil
}
