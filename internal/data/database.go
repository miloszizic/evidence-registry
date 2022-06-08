package data

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"io"
)

type Case struct {
	ID   int64    `json:"id"`
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}
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

type DBStore interface {
	AddCase(cs *Case, user *User) error
	CaseExists(name string) (bool, error)
	ListCases() ([]Case, error)
	GetCaseByName(name string) (*Case, error)
	GetCaseByID(id int64) (*Case, error)
	GetCaseByUserID(userID int64) ([]Case, error)
	RemoveCase(cs *Case) error
	FindCaseByTags(tags []string) ([]Case, error)
	CreateEvidence(evidence *Evidence) (int64, error)
	GetEvidenceByID(id int64, caseID int64) (*Evidence, error)
	EvidenceExists(evidence *Evidence) (bool, error)
	GetEvidenceByName(cs *Case, name string) (*Evidence, error)
	RemoveEvidence(evidence *Evidence) error
	GetEvidenceByCaseID(CaseID int64) ([]Evidence, error)
	AddComment(comment *Comment) error
	GetCommentsByID(evidenceID int64) ([]Comment, error)
}

type DB struct {
	DB *sql.DB
}

func NewDBStore(db *sql.DB) DBStore {
	return &DB{
		DB: db,
	}
}

// AddCase a new case to the database or return an error
func (d *DB) AddCase(cs *Case, user *User) error {
	if cs.Name == "" {
		return fmt.Errorf("%w : case name cannot be empty", ErrInvalidRequest)
	}
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}

	// first insert the case into the cases table and get the id
	var caseID int64
	err = tx.QueryRow(`INSERT INTO "cases" ("name", "tags") VALUES ($1, $2) RETURNING id;`, cs.Name, pq.Array(cs.Tags)).Scan(&caseID)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec(`INSERT INTO "user_cases" ("user_id", "case_id") VALUES ($1, $2)`, user.ID, caseID)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}

// CaseExists returns true if the case exists in the database
func (d *DB) CaseExists(name string) (bool, error) {
	var count int
	_ = d.DB.QueryRow(`SELECT COUNT(*) FROM "cases" WHERE name = $1`, name).Scan(&count)
	return count > 0, nil
}

// ListCases all cases in the database or an error
func (d *DB) ListCases() ([]Case, error) {
	rows, err := d.DB.Query(`SELECT * FROM "cases"`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cases []Case
	for rows.Next() {
		var cs Case
		rErr := rows.Scan(&cs.ID, &cs.Name, pq.Array(&cs.Tags))
		if rErr != nil {
			return nil, err
		}
		cases = append(cases, cs)
	}

	return cases, nil
}

// GetCaseByName returns a case by name from the database or an error
func (d *DB) GetCaseByName(name string) (*Case, error) {
	cs := &Case{}
	err := d.DB.QueryRow(`SELECT * FROM "cases" WHERE name = $1`, name).Scan(&cs.ID, &cs.Name, pq.Array(&cs.Tags))
	if err != nil {
		return nil, err
	}
	return cs, nil
}

// GetCaseByID returns a case by id from the database or an error
func (d *DB) GetCaseByID(id int64) (*Case, error) {
	cs := &Case{}
	err := d.DB.QueryRow(`SELECT * FROM "cases" WHERE id = $1`, id).Scan(&cs.ID, &cs.Name, pq.Array(&cs.Tags))
	if err != nil {
		return nil, err
	}
	return cs, nil
}

//GetCaseByUserID returns a case by id from the database or an error
func (d *DB) GetCaseByUserID(userID int64) ([]Case, error) {
	rows, err := d.DB.Query(`SELECT * FROM "cases" WHERE id IN (SELECT case_id FROM "user_cases" WHERE user_id = $1)`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w : user id : %d", ErrNotFound, userID)
		}
		return nil, err
	}
	defer rows.Close()

	var cases []Case
	for rows.Next() {
		var cs Case
		rErr := rows.Scan(&cs.ID, &cs.Name, pq.Array(&cs.Tags))
		if rErr != nil {
			return nil, rErr
		}
		cases = append(cases, cs)
	}
	return cases, nil
}

// RemoveCase removes a case from the database or returns an error
func (d *DB) RemoveCase(cs *Case) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// first remove from user_cases table
	_, err = tx.Exec(`DELETE FROM "user_cases" WHERE case_id = $1`, cs.ID)
	if err != nil {
		return err
	}
	// then remove from cases table
	_, err = tx.Exec(`DELETE FROM "cases" WHERE id = $1`, cs.ID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// FindCaseByTags returns cases with matching tags
func (d *DB) FindCaseByTags(tags []string) ([]Case, error) {
	sel := `SELECT * FROM "cases" WHERE $1 <@ tags`
	rows, err := d.DB.Query(sel, pq.Array(tags))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cases []Case
	for rows.Next() {
		var cs Case
		rErr := rows.Scan(&cs.ID, &cs.Name, pq.Array(&cs.Tags))
		if rErr != nil {
			return nil, err
		}
		cases = append(cases, cs)
	}
	return cases, nil
}

// CreateEvidence is used to create a new evidence in specific case in the database
// It returns the new evidence ID
func (d *DB) CreateEvidence(evidence *Evidence) (int64, error) {
	err := d.DB.QueryRow(`INSERT INTO evidences (case_id, name, hash) VALUES ($1, $2, $3) RETURNING id;`, evidence.CaseID, evidence.Name, evidence.Hash).Scan(&evidence.ID)
	if err != nil {
		return 0, err
	}
	return evidence.ID, nil
}

// GetEvidenceByID is used to get an evidence by its ID from specific case in the database
func (d *DB) GetEvidenceByID(id int64, caseID int64) (*Evidence, error) {
	var evidence Evidence
	err := d.DB.QueryRow("SELECT id, case_id, name, hash FROM evidences WHERE id = $1 AND case_id = $2", id, caseID).Scan(&evidence.ID, &evidence.CaseID, &evidence.Name, &evidence.Hash)
	if err != nil {
		return nil, err
	}
	return &evidence, err
}

// EvidenceExists is used to check if an evidence exists in the database,
// evidence must contain a valid case ID in it
func (d *DB) EvidenceExists(evidence *Evidence) (bool, error) {
	var count int
	_ = d.DB.QueryRow("SELECT id FROM evidences WHERE case_id = $1 AND name = $2", evidence.CaseID, evidence.Name).Scan(&count)
	return count > 0, nil
}

// GetEvidenceByName is used to get an evidence by its name from specific case in the database
func (d *DB) GetEvidenceByName(cs *Case, name string) (*Evidence, error) {
	var object Evidence
	err := d.DB.QueryRow("SELECT id, case_id, name, hash FROM evidences WHERE case_id = $1 AND name = $2", cs.ID, name).Scan(&object.ID, &object.CaseID, &object.Name, &object.Hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w :evidence not found: %q", ErrInvalidRequest, name)
		}
		return nil, err
	}
	return &object, err
}

// RemoveEvidence is used to delete an evidence from specific case in the database
func (d *DB) RemoveEvidence(evidence *Evidence) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM comments WHERE evidence_id = $1`, evidence.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM "evidences" WHERE id = $1 AND case_id = $2;`, evidence.ID, evidence.CaseID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetEvidenceByCaseID is used to get all evidences from specific case in the database
func (d *DB) GetEvidenceByCaseID(CaseID int64) ([]Evidence, error) {
	rows, err := d.DB.Query(`SELECT id, case_id, name, hash FROM evidences WHERE case_id = $1;`, CaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evidences []Evidence
	for rows.Next() {
		var object Evidence
		rErr := rows.Scan(&object.ID, &object.CaseID, &object.Name, &object.Hash)
		if rErr != nil {
			return nil, rErr
		}
		evidences = append(evidences, object)
	}
	return evidences, nil
}

//AddComment is used to add a comment to an evidence in the database
func (d *DB) AddComment(comment *Comment) error {
	_, err := d.DB.Exec(`INSERT INTO "comments" ("evidence_id", "content") VALUES ($1, $2 );`, comment.EvidenceID, comment.Text)
	return err
}

//GetCommentsByID is used to get all comments from an evidence in the database
func (d *DB) GetCommentsByID(evidenceID int64) ([]Comment, error) {
	rows, err := d.DB.Query(`SELECT id, evidence_id, content FROM comments WHERE evidence_id = $1;`, evidenceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		rErr := rows.Scan(&comment.ID, &comment.EvidenceID, &comment.Text)
		if rErr != nil {
			return nil, rErr
		}
		comments = append(comments, comment)
	}
	return comments, nil
}
