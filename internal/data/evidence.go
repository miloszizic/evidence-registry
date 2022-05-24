package data

import (
	"database/sql"
	"io"
)

type Evidence struct {
	ID     int64     `json:"id"`
	CaseID int64     `json:"case_id,omitempty"`
	File   io.Reader `json:"file,omitempty"`
	Name   string    `json:"name,omitempty"`
	Hash   string    `json:"hash,omitempty"`
}

type Comment struct {
	ID         int64  `json:"id"`
	EvidenceID int64  `json:"evidence_id,omitempty"`
	Text       string `json:"text,omitempty"`
}

type EvidenceDB struct {
	DB *sql.DB
}

// Create is used to create a new evidence in specific case in the database
// It returns the new evidence ID
func (s *EvidenceDB) Create(evidence *Evidence) (int64, error) {
	// check if the evidence already exists

	err := s.DB.QueryRow(`INSERT INTO evidences (case_id, name, hash) VALUES ($1, $2, $3) RETURNING id;`, evidence.CaseID, evidence.Name, evidence.Hash).Scan(&evidence.ID)
	if err != nil {
		return 0, err
	}
	return evidence.ID, nil
}

// GetByID is used to get an evidence by its ID from specific case in the database
func (s *EvidenceDB) GetByID(id int64) (*Evidence, error) {
	var evidence Evidence
	err := s.DB.QueryRow("SELECT id, case_id, name, hash FROM evidences WHERE id = $1", id).Scan(&evidence.ID, &evidence.CaseID, &evidence.Name, &evidence.Hash)
	return &evidence, err
}

// GetByName is used to get an evidence by its name from specific case in the database
func (s *EvidenceDB) GetByName(cs *Case, name string) (*Evidence, error) {
	var object Evidence
	err := s.DB.QueryRow("SELECT id, case_id, name, hash FROM evidences WHERE case_id = $1 AND name = $2", cs.ID, name).Scan(&object.ID, &object.CaseID, &object.Name, &object.Hash)
	return &object, err
}

// Remove is used to delete an evidence from specific case in the database
func (s *EvidenceDB) Remove(evidence *Evidence) error {
	_, err := s.DB.Exec(`DELETE FROM "evidences" WHERE id = $1 AND case_id = $2;`, evidence.ID, evidence.CaseID)
	return err
}

// GetByCaseID is used to get all evidences from specific case in the database
func (s *EvidenceDB) GetByCaseID(CaseID int64) ([]Evidence, error) {
	rows, err := s.DB.Query(`SELECT id, case_id, name, hash FROM evidences WHERE case_id = $1;`, CaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evidences []Evidence
	for rows.Next() {
		var object Evidence
		err := rows.Scan(&object.ID, &object.CaseID, &object.Name, &object.Hash)
		if err != nil {
			return nil, err
		}
		evidences = append(evidences, object)
	}
	return evidences, nil
}

//AddComment is used to add a comment to an evidence in the database
func (s *EvidenceDB) AddComment(comment *Comment) error {
	_, err := s.DB.Exec(`INSERT INTO "comments" ("evidence_id", "content") VALUES ($1, $2 );`, comment.EvidenceID, comment.Text)
	return err
}

//GetCommentsByID is used to get all comments from an evidence in the database
func (s *EvidenceDB) GetCommentsByID(evidenceID int64) ([]Comment, error) {
	rows, err := s.DB.Query(`SELECT id, evidence_id, content FROM comments WHERE evidence_id = $1;`, evidenceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		err := rows.Scan(&comment.ID, &comment.EvidenceID, &comment.Text)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, nil
}
