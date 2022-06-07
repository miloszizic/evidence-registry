package data

import (
	"database/sql"
	"errors"
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
	exist, err := s.DBStore.CaseExists(name)
	if err != nil {
		return WrapErrorf(err, ErrCodeUnknown, "DBStore.CaseExists")
	}
	if exist {
		return NewErrorf(ErrCodeExists, "CreateCase: case already exists")
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
			return NewErrorf(ErrCodeInvalid, "CreateCase: invalid case name")
		case err.Error() == "Bucket name cannot be empty":
			return NewErrorf(ErrCodeInvalid, "CreateCase: invalid case name")
		default:
			return WrapErrorf(err, ErrCodeUnknown, "OBStore.CreateCase")
		}
	}
	// create case in database
	err = s.DBStore.AddCase(cs, user)
	if err != nil {
		rErr := s.OBStore.RemoveCase(name)
		if rErr != nil {
			return WrapErrorf(rErr, ErrCodeUnknown, "OBStore.RemoveCase")
		}
		return WrapErrorf(err, ErrCodeUnknown, "DBStore.AddCase")
	}
	return nil
}
func (s *Stores) GetCaseByID(id int64) (*Case, error) {
	cs, err := s.DBStore.GetCaseByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewErrorf(ErrCodeNotFound, "GetCaseByID: case not found ")
		}
		return nil, WrapErrorf(err, ErrCodeUnknown, "DBStore.GetCaseByID")
	}
	return cs, nil
}
func (s *Stores) RemoveCase(name string) error {
	// check if case exists in the database
	exist, err := s.DBStore.CaseExists(name)
	if err != nil {
		return WrapErrorf(err, ErrCodeUnknown, "DBStore.CaseExists")
	}
	if !exist {
		return NewErrorf(ErrCodeNotFound, "RemoveCase: case not found")
	}
	// get the case
	cs, err := s.DBStore.GetCaseByName(name)
	if err != nil {
		return WrapErrorf(err, ErrCodeUnknown, "DBStore.GetCaseByName")
	}
	// check if case exists in the OBStore
	exist, err = s.OBStore.CaseExists(cs.Name)
	if err != nil {
		return WrapErrorf(err, ErrCodeUnknown, "OBStore.CaseExists")
	}
	if !exist {
		return NewErrorf(ErrCodeNotFound, "RemoveCase: case not found in OBStore")
	}
	// remove case from OBStore
	err = s.OBStore.RemoveCase(cs.Name)
	if err != nil {
		return WrapErrorf(err, ErrCodeUnknown, "OBStore.RemoveCase")
	}
	// remove the case in the database
	err = s.DBStore.RemoveCase(cs)
	if err != nil {
		return WrapErrorf(err, ErrCodeUnknown, "DBStore.RemoveCase")
	}
	return nil
}
func (s *Stores) ListCases() ([]Case, error) {
	// get all cases from the database
	casesDB, err := s.DBStore.ListCases()
	if err != nil {
		return nil, WrapErrorf(err, ErrCodeUnknown, "DBStore.ListCases")
	}
	//get all cases in the OBStore
	casesFS, err := s.OBStore.ListCases()
	if err != nil {
		return nil, WrapErrorf(err, ErrCodeUnknown, "OBStore.ListCases")
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
		return WrapErrorf(err, ErrCodeUnknown, "DBStore.EvidenceExists")
	}
	if exist {
		return NewErrorf(ErrCodeExists, "CreateEvidence: evidence already exists in DB")
	}
	//check if the evidence already exists in the OBStore
	exist, err = s.OBStore.EvidenceExists(cs.Name, ev.Name)
	if err != nil {
		return WrapErrorf(err, ErrCodeUnknown, "OBStore.EvidenceExists")
	}
	if exist {
		return NewErrorf(ErrCodeExists, "CreateEvidence: evidence already exists in OBStore")
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
			return WrapErrorf(errR, ErrCodeUnknown, "OBStore.RemoveEvidence")
		}
		return WrapErrorf(err, ErrCodeUnknown, "DBStore.CreateEvidence")
	}
	ev.ID = id
	return nil
}
func (s *Stores) GetEvidenceByID(id int64, csID int64) (*Evidence, error) {
	ev, err := s.DBStore.GetEvidenceByID(id, csID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewErrorf(ErrCodeNotFound, "GetEvidenceByID: evidence not found in DB")
		}
		return nil, WrapErrorf(err, ErrCodeUnknown, "GetEvidenceByID")
	}
	return ev, nil
}
func (s *Stores) DownloadEvidence(ev *Evidence) (*io.ReadCloser, error) {
	// check if the evidence exists in the database
	cs, err := s.DBStore.GetCaseByID(ev.CaseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewErrorf(ErrCodeNotFound, "DownloadEvidence: evidence not found in DB")
		}
		return nil, WrapErrorf(err, ErrCodeUnknown, "DownloadEvidence")
	}
	// check if the evidence exists in the OBStore
	exist, err := s.OBStore.EvidenceExists(cs.Name, ev.Name)
	if err != nil {
		return nil, WrapErrorf(err, ErrCodeUnknown, "OBStore.EvidenceExists")
	}
	if !exist {
		return nil, NewErrorf(ErrCodeNotFound, "DownloadEvidence: evidence not found in OBStore")
	}
	evidence, err := s.OBStore.GetEvidence(cs.Name, ev.Name)
	if err != nil {
		return nil, WrapErrorf(err, ErrCodeUnknown, "OBStore.GetEvidence")
	}
	return &evidence, nil
}

// DeleteEvidence deletes the evidence from the database and the FS
func (s *Stores) DeleteEvidence(ev *Evidence) error {
	// check if the evidence exists in the database
	exist, err := s.DBStore.EvidenceExists(ev)
	if err != nil {
		return WrapErrorf(err, ErrCodeUnknown, "DBStore.EvidenceExists")
	}
	if !exist {
		return NewErrorf(ErrCodeNotFound, "DeleteEvidence: evidence not found in DB")
	}
	// get case from DB
	cs, err := s.DBStore.GetCaseByID(ev.CaseID)
	if err != nil {
		return WrapErrorf(err, ErrCodeUnknown, "DBStore.GetCaseByID ")
	}
	// check if the evidence exists in the OBStore
	exist, err = s.OBStore.EvidenceExists(cs.Name, ev.Name)
	if err != nil {
		return WrapErrorf(err, ErrCodeUnknown, "OBStore.EvidenceExists")
	}
	if !exist {
		return NewErrorf(ErrCodeNotFound, "DeleteEvidence: evidence not found in OBStore")
	}
	// delete evidence from the database
	err = s.DBStore.RemoveEvidence(ev)
	if err != nil {
		return WrapErrorf(err, ErrCodeUnknown, "DBStore.RemoveEvidence")
	}
	// delete evidence from the OBStore
	err = s.OBStore.RemoveEvidence(ev, cs.Name)
	if err != nil {
		return WrapErrorf(err, ErrCodeUnknown, "OBStore.RemoveEvidence")
	}
	return nil
}

// ListEvidences returns all evidences for a case that are present in bought
// database and FS
func (s *Stores) ListEvidences(cs *Case) ([]Evidence, error) {
	// list evidences in the database
	evidencesDB, err := s.DBStore.GetEvidenceByCaseID(cs.ID)
	if err != nil {
		return nil, WrapErrorf(err, ErrCodeUnknown, "DBStore.GetEvidenceByCaseID")
	}
	// list evidences in OBStore
	evidencesFS, err := s.OBStore.ListEvidences(cs.Name)
	if err != nil {
		return nil, WrapErrorf(err, ErrCodeUnknown, "OBStore.ListEvidences")
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
		return WrapErrorf(err, ErrCodeUnknown, "Password.Set")
	}
	// create usr
	err = s.User.Add(usr)
	if err != nil {
		return WrapErrorf(err, ErrCodeUnknown, "User.Add")
	}
	return nil
}

// AddEvidenceComment adds comment to existing evidence
func (s *Stores) AddEvidenceComment(comment *Comment) error {
	err := s.DBStore.AddComment(comment)
	if err != nil {
		return WrapErrorf(err, ErrCodeUnknown, "DBStore.AddComment")
	}
	return nil
}

// FromPostgresDB opens a connection to a Postgres database.
func FromPostgresDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, WrapErrorf(err, ErrCodeUnknown, "sql.Open")
	}
	return db, nil
}

// FromMinio creates a new Minio client.
func FromMinio(endpoint, accessKeyID, secretAccessKey string) (*minio.Client, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
	})
	if err != nil {
		return nil, WrapErrorf(err, ErrCodeUnknown, "minio.New")
	}
	return minioClient, nil
}
