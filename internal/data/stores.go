package data

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
)

type Stores struct {
	User    UserStore
	DBStore DBStore
	OBStore OBStore
}

// NewStores creates a new Stores object
func NewStores(db *sql.DB, client *minio.Client) Stores {
	return Stores{
		User:    NewUserStore(db),
		DBStore: NewDBStore(db),
		OBStore: NewOBS(client),
	}
}
func (s *Stores) CreateCase(user *User, name string) error {
	exists, err := s.DBStore.CaseExists(name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%w : case : %q ", ErrAlreadyExists, name)
	}
	// create case struct
	cs := &Case{
		Name: name,
	}
	// create case in OBStore
	err = s.OBStore.CreateCase(cs)
	if err != nil {
		switch {
		case err.Error() == "Bucket name contains invalid characters":
			return fmt.Errorf("%w : case contains invalid characters: %q ", ErrInvalidRequest, cs.Name)
		case err.Error() == "Bucket name cannot be empty":
			return fmt.Errorf("%w : case name cannot be empty ", ErrInvalidRequest)
		default:
			return fmt.Errorf("creating case in objects store : %w", err)
		}
	}
	// create case in database
	err = s.DBStore.AddCase(cs, user)
	if err != nil {
		errR := s.OBStore.RemoveCase(name)
		if errR != nil {
			return fmt.Errorf("creating case in DB : %w, removing case from object store : %v ", err, errR)
		}
		return fmt.Errorf("creating case in DB store : %w", err)
	}
	return nil
}
func (s *Stores) GetCaseByID(id int64) (*Case, error) {
	cs, err := s.DBStore.GetCaseByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w : case id : %d ", ErrNotFound, id)
		}
		return nil, err
	}
	return cs, nil
}
func (s *Stores) RemoveCase(name string) error {
	// check if case exists in the database
	exist, err := s.DBStore.CaseExists(name)
	if err != nil {
		return err
	}
	if !exist {
		return fmt.Errorf("%w : case name : %q ", ErrNotFound, name)
	}
	// get the case
	cs, err := s.DBStore.GetCaseByName(name)
	if err != nil {
		return fmt.Errorf(" getting case in DB :%w, case name: %q  ", err, name)
	}
	// check if case exists in the OBStore
	exist, err = s.OBStore.CaseExists(cs.Name)
	if err != nil {
		return fmt.Errorf(" checking case in object store :%w, case name: %q  ", err, name)
	}
	if !exist {
		return fmt.Errorf(" %w: case name: %q ", ErrNotFound, name)
	}
	// remove case from OBStore
	err = s.OBStore.RemoveCase(cs.Name)
	if err != nil {
		return fmt.Errorf("%w : removing case from object store: %q ", err, cs.Name)
	}
	// remove the case in the database
	err = s.DBStore.RemoveCase(cs)
	if err != nil {
		return fmt.Errorf("%w : removing case from DB store: %q ", err, cs.Name)
	}
	return nil
}
func (s *Stores) ListCases() ([]Case, error) {
	// get all cases from the database
	casesDB, err := s.DBStore.ListCases()
	if err != nil {
		return nil, fmt.Errorf("list cases from DB: %w ", err)
	}
	//get all cases in the OBStore
	casesFS, err := s.OBStore.ListCases()
	if err != nil {
		return nil, fmt.Errorf("list cases from object storage: %w ", err)
	}
	// remove cases that are not in the database
	var List []Case
	for _, caseDB := range casesDB {
		for _, caseFS := range casesFS {
			if caseDB.Name == caseFS.Name {
				List = append(List, caseDB)
			}
		}
	}
	return List, nil
}

// CreateEvidence creates an evidence in the database and the FS
func (s *Stores) CreateEvidence(ev *Evidence, cs *Case) error {
	// check if the evidence already exists in the database
	exist, err := s.DBStore.EvidenceExists(ev)
	if err != nil {
		return fmt.Errorf("chaking evidence in DB: %w , evidence name: %q ", err, ev.Name)
	}
	if exist {
		return fmt.Errorf(" %w in DB: evidence name: %q ", ErrAlreadyExists, ev.Name)
	}
	//check if the evidence already exists in the OBStore
	exist, err = s.OBStore.EvidenceExists(cs.Name, ev.Name)
	if err != nil {
		return fmt.Errorf("chaking evidence in object store: %w , evidence name: %q ", err, ev.Name)
	}
	if exist {
		return fmt.Errorf(" %w in object storage: evidence name: %q ", ErrAlreadyExists, ev.Name)
	}
	// create the evidence in OBStore and generate hash
	hash, err := s.OBStore.CreateEvidence(ev, cs.Name, ev.File)
	if err != nil {
		return err
	}
	// create the evidence in DB
	ev.Hash = hash
	id, err := s.DBStore.CreateEvidence(ev)
	if err != nil {
		errR := s.OBStore.RemoveEvidence(ev, cs.Name)
		if errR != nil {
			return fmt.Errorf("creating evidence in DB : %w, removing evidence from object store : %v ", err, errR)
		}
		return fmt.Errorf("creating evidence in DB: %w , evidence name: %q ", err, ev.Name)
	}
	ev.ID = id
	return nil
}
func (s *Stores) GetEvidenceByID(id int64, csID int64) (*Evidence, error) {
	ev, err := s.DBStore.GetEvidenceByID(id, csID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w : evidence id : %d ", ErrNotFound, id)
		}
		return nil, fmt.Errorf("getting evidence from DB: %w , evidence id: %d ", err, id)
	}
	return ev, nil
}
func (s *Stores) DownloadEvidence(ev *Evidence) (*io.ReadCloser, error) {
	// check if the evidence exists in the database
	cs, err := s.DBStore.GetCaseByID(ev.CaseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w : case id : %d ", ErrNotFound, ev.CaseID)
		}
		return nil, fmt.Errorf("getting case by ID from DB: %w , evidence id: %d ", err, ev.CaseID)
	}
	// check if the evidence exists in the OBStore
	exist, err := s.OBStore.EvidenceExists(cs.Name, ev.Name)
	if err != nil {
		return nil, fmt.Errorf("chaking evidence in object store: %w , evidence name: %q ", err, ev.Name)
	}
	if !exist {
		return nil, fmt.Errorf(" %w in object storage: evidence name: %q ", ErrNotFound, ev.Name)
	}
	evidence, err := s.OBStore.GetEvidence(cs.Name, ev.Name)
	if err != nil {
		return nil, fmt.Errorf("getting evidence in object store: %w , evidence name: %q ", err, ev.Name)
	}
	return &evidence, nil
}

// DeleteEvidence deletes the evidence from the database and the FS
func (s *Stores) DeleteEvidence(ev *Evidence) error {
	// check if the evidence exists in the database
	exist, err := s.DBStore.EvidenceExists(ev)
	if err != nil {
		return fmt.Errorf("chaking evidence in DB store: %w , evidence name: %q ", err, ev.Name)
	}
	if !exist {
		return fmt.Errorf(" %w in DB: evidence name: %q ", ErrNotFound, ev.Name)
	}
	// get case from DB
	cs, err := s.DBStore.GetCaseByID(ev.CaseID)
	if err != nil {
		return fmt.Errorf("getting case by ID in DB store: %w , case ID: %d and evidence name : %q ", err, ev.CaseID, ev.Name)
	}
	// check if the evidence exists in the OBStore
	exist, err = s.OBStore.EvidenceExists(cs.Name, ev.Name)
	if err != nil {
		return fmt.Errorf("chaking evidence in object store: %w , evidence name: %q ", err, ev.Name)
	}
	if !exist {
		return fmt.Errorf(" %w in object storage: evidence name: %q ", ErrNotFound, ev.Name)
	}
	// delete evidence from the database
	err = s.DBStore.RemoveEvidence(ev)
	if err != nil {
		return fmt.Errorf("removing evidence from DB: %w , evidence name: %q ", err, ev.Name)
	}
	// delete evidence from the OBStore
	err = s.OBStore.RemoveEvidence(ev, cs.Name)
	if err != nil {
		return fmt.Errorf("removing evidence from object store: %w , evidence name: %q ", err, ev.Name)
	}
	return nil
}

// ListEvidences returns all evidences for a case that are present in bought
// database and FS
func (s *Stores) ListEvidences(cs *Case) ([]Evidence, error) {
	// list evidences in the database
	evidencesDB, err := s.DBStore.GetEvidenceByCaseID(cs.ID)
	if err != nil {
		return nil, fmt.Errorf("getting evidences from DB: %w , case ID: %d ", err, cs.ID)
	}
	// list evidences in OBStore
	evidencesFS, err := s.OBStore.ListEvidences(cs.Name)
	if err != nil {
		return nil, fmt.Errorf("getting evidences from object store: %w , case ID: %d ", err, cs.ID)
	}
	// filter out evidences that are not in the OBStore
	var result []Evidence
	for _, evDB := range evidencesDB {
		for _, evFS := range evidencesFS {
			if evDB.Name == evFS.Name {
				result = append(result, evDB)
			}
		}
	}
	return result, nil
}

type UserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

// CreateUser creates a new user in the database.
func (s *Stores) CreateUser(request *UserRequest) error {
	usr := &User{
		Username: request.Username,
	}
	err := usr.Password.Set(request.Password)
	if err != nil {
		return fmt.Errorf("setting password: %w", err)
	}
	// create usr
	err = s.User.Add(usr)
	if err != nil {
		return fmt.Errorf("creating user in DB: %w", err)
	}
	return nil
}

// AddEvidenceComment adds comment to existing evidence
func (s *Stores) AddEvidenceComment(comment *Comment) error {
	err := s.DBStore.AddComment(comment)
	if err != nil {
		return fmt.Errorf("adding comment to DB: %w", err)
	}
	return nil
}

// FromPostgresDB opens a connection to a Postgres database.
func FromPostgresDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening DB connection: %w", err)
	}
	return db, nil
}

// FromMinio creates a new Minio client.
func FromMinio(endpoint, accessKeyID, secretAccessKey string) (*minio.Client, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("creating minio client: %w", err)
	}
	return minioClient, nil
}
