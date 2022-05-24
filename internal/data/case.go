package database

import (
	"database/sql"
	"errors"
	"github.com/lib/pq"
)

type Case struct {
	ID   int64    `json:"id"`
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type CaseDB struct {
	DB *sql.DB
}

//Add a new case to the database
func (s *CaseDB) Add(cs *Case, user *User) error {
	if cs.Name == "" {
		return errors.New("case name can't be empty")
	}
	if user.ID == 0 {
		return errors.New("user id can't be empty")
	}
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// first insert the case into the cases table and get the id
	var caseID int64
	err = tx.QueryRow(`INSERT INTO "cases" ("name", "tags") VALUES ($1, $2) RETURNING id;`, cs.Name, pq.Array(cs.Tags)).Scan(&caseID)
	if err != nil {
		return err
	}

	// then insert the case into the case_user table with the user id
	_, err = tx.Exec(`INSERT INTO "user_cases" ("user_id", "case_id") VALUES ($1, $2);`, user.ID, caseID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

//List all cases in the database or an error
func (s *CaseDB) List() ([]Case, error) {
	rows, err := s.DB.Query(`SELECT * FROM "cases"`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cases []Case
	for rows.Next() {
		var cs Case
		err := rows.Scan(&cs.ID, &cs.Name, pq.Array(&cs.Tags))
		if err != nil {
			return nil, err
		}
		cases = append(cases, cs)
	}

	return cases, nil
}

// GetByName returns a case by name from the database or an error
func (s *CaseDB) GetByName(name string) (*Case, error) {
	cs := &Case{}
	err := s.DB.QueryRow(`SELECT * FROM "cases" WHERE name = $1`, name).Scan(&cs.ID, &cs.Name, pq.Array(&cs.Tags))
	if err != nil {
		return nil, err
	}
	return cs, nil
}

// GetByID returns a case by id from the database or an error
func (s *CaseDB) GetByID(id int64) (*Case, error) {
	cs := &Case{}
	err := s.DB.QueryRow(`SELECT * FROM "cases" WHERE id = $1`, id).Scan(&cs.ID, &cs.Name, pq.Array(&cs.Tags))
	if err != nil {
		return nil, err
	}
	return cs, nil
}

//GetByUserID returns a case by id from the database or an error
func (s *CaseDB) GetByUserID(userID int64) ([]Case, error) {
	rows, err := s.DB.Query(`SELECT * FROM "cases" WHERE id IN (SELECT case_id FROM "user_cases" WHERE user_id = $1)`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cases []Case
	for rows.Next() {
		var cs Case
		err := rows.Scan(&cs.ID, &cs.Name, pq.Array(&cs.Tags))
		if err != nil {
			return nil, err
		}
		cases = append(cases, cs)
	}
	return cases, nil
}

// Remove removes a case from the database
func (s *CaseDB) Remove(cs *Case) error {
	// check if the case exists
	result, err := s.GetByID(cs.ID)
	if result == nil {
		return errors.New("case not found")
	}
	if err != nil {
		return err
	}
	tx, err := s.DB.Begin()
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

func (s *CaseDB) SearchByTags(tags []string) ([]Case, error) {
	sel := `SELECT * FROM "cases" WHERE $1 <@ tags`
	rows, err := s.DB.Query(sel, pq.Array(tags))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cases []Case
	for rows.Next() {
		var cs Case
		err := rows.Scan(&cs.ID, &cs.Name, pq.Array(&cs.Tags))
		if err != nil {
			return nil, err
		}
		cases = append(cases, cs)
	}
	return cases, nil
}
