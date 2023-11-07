package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/miloszizic/der/db"
)

// CreateEvidenceParams defines the parameters that are needed to create an evidence.
type CreateEvidenceParams struct {
	CaseID         uuid.UUID `json:"case_id"`
	AppUserID      uuid.UUID `json:"app_user_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	EvidenceTypeID uuid.UUID `json:"evidence_type_id"`
}

// Evidence holds the information about evidence
type Evidence struct {
	ID             uuid.UUID      `json:"id"`
	CaseID         uuid.UUID      `json:"case_id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	AppUserID      uuid.UUID      `json:"app_user_id"`
	Name           string         `json:"name"`
	Description    sql.NullString `json:"description"`
	Hash           string         `json:"hash"`
	EvidenceTypeID uuid.UUID      `json:"evidence_type_id"`
}

// ConvertDBEvidenceToEvidence converts a db evidence to a service evidence.
func ConvertDBEvidenceToEvidence(dbEvidence db.Evidence) Evidence {
	return Evidence{
		ID:             dbEvidence.ID,
		CaseID:         dbEvidence.CaseID,
		CreatedAt:      dbEvidence.CreatedAt,
		UpdatedAt:      dbEvidence.UpdatedAt,
		AppUserID:      dbEvidence.AppUserID,
		Name:           dbEvidence.Name,
		Description:    dbEvidence.Description,
		Hash:           dbEvidence.Hash,
		EvidenceTypeID: dbEvidence.EvidenceTypeID,
	}
}

// CreateEvidence creates evidence in the db and the FS
func (s *Stores) CreateEvidence(ctx context.Context, request CreateEvidenceParams, file io.Reader) (Evidence, error) {
	// Begin a db transaction
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return Evidence{}, fmt.Errorf("beginning transaction: %w", err)
	}

	// Defer a rollback in case anything fails.
	defer tx.Rollback()

	// Create a query object with the transaction
	q := s.DBStore.WithTx(tx)

	// Set current user in session_data
	err = q.SetCurrentUser(ctx, request.AppUserID)
	if err != nil {
		return Evidence{}, fmt.Errorf("setting current user in audit: %w", err)
	}

	// check if the evidence already exists in the db
	existsParams := db.EvidenceExistsParams{
		Name:   request.Name,
		CaseID: request.CaseID,
	}

	exist, err := q.EvidenceExists(ctx, existsParams)
	if err != nil {
		return Evidence{}, fmt.Errorf("error checking evidence in DB: %w, evidence name: %q", err, request.Name)
	}

	if exist {
		return Evidence{}, fmt.Errorf("%w in DB: evidence name: %q", ErrAlreadyExists, request.Name)
	}

	// get case from the db
	cs, err := q.GetCase(ctx, request.CaseID)
	if err != nil {
		return Evidence{}, fmt.Errorf("error getting case from DB: %w", err)
	}

	// convert db case name to minio case name
	minioCaseName, err := ConvertDBFormatToMinio(cs.Name)
	if err != nil {
		return Evidence{}, fmt.Errorf("converting db case name to minio: %w", err)
	}

	// check if the evidence already exists in the ObjectStore
	exist, err = s.ObjectStore.EvidenceExists(ctx, minioCaseName, request.Name)
	if err != nil {
		return Evidence{}, fmt.Errorf("error checking evidence in object store: %w, evidence name: %q", err, request.Name)
	}

	if exist {
		return Evidence{}, fmt.Errorf("%w in object storage: evidence name: %q", ErrAlreadyExists, request.Name)
	}

	// create the evidence in ObjectStore and generate hash
	hash, err := s.ObjectStore.CreateEvidence(ctx, request.Name, minioCaseName, file)
	if err != nil {
		return Evidence{}, fmt.Errorf("error creating evidence in object storage: %w", err)
	}

	Description := HandleNullableString(request.Description)
	createEV := db.CreateEvidenceParams{
		CaseID:         cs.ID,
		AppUserID:      request.AppUserID,
		Name:           request.Name,
		Description:    Description,
		Hash:           hash,
		EvidenceTypeID: request.EvidenceTypeID,
	}

	DBEvidence, err := q.CreateEvidence(ctx, createEV)
	if err != nil {
		errR := s.ObjectStore.RemoveEvidence(ctx, request.Name, minioCaseName)
		if errR != nil {
			return Evidence{}, fmt.Errorf("error creating evidence in DB: %w, removing evidence from object store: %w", err, errR)
		}

		return Evidence{}, fmt.Errorf("error creating evidence in DB: %w, evidence name: %q", err, request.Name)
	}

	evidence := ConvertDBEvidenceToEvidence(DBEvidence)

	// If all operations are successful, commit the transaction
	if err := tx.Commit(); err != nil {
		return Evidence{}, fmt.Errorf("committing transaction: %w", err)
	}

	return evidence, nil
}

// GetEvidenceByID takes an id and returns the evidence with that id
func (s *Stores) GetEvidenceByID(ctx context.Context, id uuid.UUID) (*Evidence, error) {
	DBEvidence, err := s.DBStore.GetEvidence(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w : evidence id : %d ", ErrNotFound, id)
		}

		return nil, fmt.Errorf("getting evidence from DB: %w , evidence id: %d ", err, id)
	}
	evidence := ConvertDBEvidenceToEvidence(DBEvidence)

	return &evidence, nil
}

// DownloadEvidence takes evidence name and evidence case ID and returns the evidence for download
func (s *Stores) DownloadEvidence(ctx context.Context, ev Evidence) (io.ReadCloser, string, error) {
	// check if the evidence exists in the db
	existsParams := db.EvidenceExistsParams{
		Name:   ev.Name,
		CaseID: ev.CaseID,
	}

	exists, err := s.DBStore.EvidenceExists(ctx, existsParams)
	if err != nil {
		return nil, "", fmt.Errorf("error checking evidence in DB: %w, evidence name: %q", err, ev.Name)
	}

	if !exists {
		return nil, "", fmt.Errorf("%w in DB: evidence name: %q", ErrNotFound, ev.Name)
	}
	// check for the case in the db
	cs, err := s.DBStore.GetCase(ctx, ev.CaseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", fmt.Errorf("%w : case id : %d ", ErrNotFound, ev.CaseID)
		}

		return nil, "", fmt.Errorf("getting case by ID from DB: %w, evidence id: %d ", err, ev.CaseID)
	}
	// convert DB case name to minio case name
	minioCaseName, err := ConvertDBFormatToMinio(cs.Name)
	if err != nil {
		return nil, "", fmt.Errorf("converting db case name to minio: %w", err)
	}
	// check if the evidence exists in the ObjectStore
	exist, err := s.ObjectStore.EvidenceExists(ctx, minioCaseName, ev.Name)
	if err != nil {
		return nil, "", fmt.Errorf("chaking evidence in object store: %w , evidence name: %q ", err, ev.Name)
	}

	if !exist {
		return nil, "", fmt.Errorf(" %w in object storage: evidence name: %q ", ErrNotFound, ev.Name)
	}

	file, err := s.ObjectStore.GetEvidence(ctx, minioCaseName, ev.Name)
	if err != nil {
		return nil, "", fmt.Errorf("getting evidence in object store: %w , evidence name: %q ", err, ev.Name)
	}

	return file, ev.Name, nil
}

// ListEvidences returns all evidences for a case that are present in bought
// db and FS
func (s *Stores) ListEvidences(ctx context.Context, cs *Case) ([]Evidence, error) {
	minioCaseName, err := ConvertDBFormatToMinio(cs.Name)
	if err != nil {
		return nil, fmt.Errorf("converting db case name to minio: %w", err)
	}

	evidencesFS, err := s.ObjectStore.ListEvidences(ctx, minioCaseName)
	if err != nil {
		return nil, fmt.Errorf("getting evidences from object store: %w , case ID: %d ", err, cs.ID)
	}

	evidencesFSMap := make(map[string]struct{})
	for _, evFS := range evidencesFS {
		evidencesFSMap[evFS.Name] = struct{}{}
	}
	var result []Evidence

	DBEvidences, err := s.DBStore.GetEvidencesByCaseID(ctx, cs.ID)
	if err != nil {
		return nil, fmt.Errorf("getting evidences from DB: %w , case ID: %d ", err, cs.ID)
	}

	for _, DBEvidence := range DBEvidences {
		if _, exists := evidencesFSMap[DBEvidence.Name]; exists {
			serviceEvidence := ConvertDBEvidenceToEvidence(DBEvidence)
			result = append(result, serviceEvidence)
		}
	}

	return result, nil
}

// EvidenceType holds the information about an evidence type
type EvidenceType struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// ListEvidenceTypes returns all evidence types
func (s *Stores) ListEvidenceTypes(ctx context.Context) ([]EvidenceType, error) {
	DBEvidenceTypes, err := s.DBStore.ListEvidenceTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting evidence types from DB: %w ", err)
	}

	SVCEvidenceTypes := make([]EvidenceType, 0, len(DBEvidenceTypes))
	for _, DBEvidenceType := range DBEvidenceTypes {
		SVCEvidenceType := EvidenceType{
			ID:   DBEvidenceType.ID,
			Name: DBEvidenceType.Name,
		}
		SVCEvidenceTypes = append(SVCEvidenceTypes, SVCEvidenceType)
	}

	return SVCEvidenceTypes, nil
}
