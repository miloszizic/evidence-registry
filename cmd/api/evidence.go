package api

import (
	"database/sql"
	"encoding/json"
	"evidence/internal/data"
	"fmt"
	"io"
	"net/http"
)

// ListEvidencesHandler returns all evidences for a case by comparing evidences in the
// database with the ones in the FS
func (app *Application) ListEvidencesHandler(w http.ResponseWriter, r *http.Request) {

	cs, err := app.getCaseFromUrl(r)
	if err != nil {
		switch {
		case err.Error() == "invalid id parameter":
			app.badRequestResponse(w, r, err)
			return
		case err.Error() == "sql: no rows in result set":
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}
	// list evidences in the database
	evidencesDB, err := app.stores.EvidenceDB.GetByCaseID(cs.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// list evidences in FS
	evidencesFS, err := app.stores.EvidenceFS.List(cs.Name)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// filter out evidences that are not in the FS
	var result []data.Evidence
	for _, evDB := range evidencesDB {
		for _, evFS := range evidencesFS {
			if evDB.Name == evFS.Name {
				result = append(result, evDB)
			}
		}
	}
	// return response
	err = app.writeJSON(w, http.StatusOK, envelope{"evidences": result}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// CreateEvidenceHandler creates a new evidence in the database and in the ostorage bucket (CaseDB)
func (app *Application) CreateEvidenceHandler(w http.ResponseWriter, r *http.Request) {
	// get the case  from the URL
	cs, err := app.getCaseFromUrl(r)
	if err != nil {
		fmt.Println(err)
		switch {
		case err.Error() == "invalid id parameter":
			app.badRequestResponse(w, r, err)
			return
		case err.Error() == "sql: no rows in result set":
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}
	// get the evidence file from the request body
	file, handler, err := r.FormFile("uploadfile")
	if file == nil {
		app.noEvidenceAttacked(w, r)
		return
	}
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer file.Close()
	evidence := &data.Evidence{
		Name:   handler.Filename,
		CaseID: cs.ID,
	}
	// check if the evidence already exists
	_, err = app.stores.EvidenceDB.GetByName(cs, evidence.Name)
	if err == sql.ErrNoRows {
		// create the evidence in FS and generate hash
		hash, err := app.stores.EvidenceFS.Create(evidence, cs.Name, file)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		evidence.Hash = hash
		// create the evidence in DB
		id, err := app.stores.EvidenceDB.Create(evidence)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		evidence.ID = id
		// write response
		err = app.writeJSON(w, http.StatusOK, envelope{"evidence": evidence}, nil)
		if err != nil {
			app.serverErrorResponse(w, r, err)
		}
	} else {
		app.evidenceAlreadyExists(w, r)
		return
	}
}

// GetEvidenceHandler gets an evidence form a specific case (CaseDB)
func (app *Application) GetEvidenceHandler(w http.ResponseWriter, r *http.Request) {
	// get evidence from the URL
	ev, err := app.getEvidenceFromUrl(r)
	if err != nil {
		switch {
		case err.Error() == "invalid id parameter":
			app.badRequestResponse(w, r, err)
			return
		case err.Error() == "sql: no rows in result set":
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}
	// get the case  from the URL
	cs, err := app.getCaseFromUrl(r)
	if err != nil {
		switch {
		case err.Error() == "invalid id parameter":
			app.badRequestResponse(w, r, err)
			return
		case err.Error() == "sql: no rows in result set":
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}
	// get evidence from FS
	object, err := app.stores.EvidenceFS.Get(cs.Name, ev.Name)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return

	}
	// respond with evidence content
	_, err = io.Copy(w, object)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// DeleteEvidenceHandler deletes an evidnece from DB and FS
func (app *Application) DeleteEvidenceHandler(w http.ResponseWriter, r *http.Request) {
	//get evidence from URL
	ev, err := app.getEvidenceFromUrl(r)
	if err != nil {
		switch {
		case err.Error() == "invalid id parameter":
			app.badRequestResponse(w, r, err)
			return
		case err.Error() == "sql: no rows in result set":
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}
	// get case from URL
	cs, err := app.getCaseFromUrl(r)
	if err != nil {
		switch {
		case err.Error() == "invalid id parameter":
			app.badRequestResponse(w, r, err)
			return
		case err.Error() == "sql: no rows in result set":
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}
	// remove evidnece from FS
	err = app.stores.EvidenceFS.Remove(ev, cs.Name)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// remove evidence from db
	err = app.stores.EvidenceDB.Remove(ev)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// write response
	err = app.writeJSON(w, http.StatusOK, envelope{"evidence": "evidence successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

//AddCommentToEvidenceHandler adds a comment to an evidence
func (app *Application) AddCommentToEvidenceHandler(w http.ResponseWriter, r *http.Request) {
	// get evidence from URL
	ev, err := app.getEvidenceFromUrl(r)
	if err != nil {
		switch {
		case err.Error() == "invalid id parameter":
			app.badRequestResponse(w, r, err)
			return
		case err.Error() == "sql: no rows in result set":
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}
	// get the comment from the json body
	var input struct {
		Text string `json:"text"`
	}
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	comment := data.Comment{
		Text:       input.Text,
		EvidenceID: ev.ID,
	}
	// add comment to db
	err = app.stores.EvidenceDB.AddComment(&comment)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	// write response
	err = app.writeJSON(w, http.StatusOK, envelope{"comment": "comment successfully added"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
