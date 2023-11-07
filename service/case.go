// Package service will be used for service related operations on minio and database.
package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/miloszizic/der/db"
)

// CreateCaseParams are the parameters that are used to create a new case in the database and minio.
type CreateCaseParams struct {
	CaseTypeID  uuid.UUID `json:"case_type_id"`
	CaseNumber  int32     `json:"case_number"`
	CaseYear    int32     `json:"case_year"`
	CaseCourtID uuid.UUID `json:"case_court_id"`
	Tags        []string  `json:"tags"`
}

// The Case holds the details of a case in the service layer.
type Case struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CaseTypeID  uuid.UUID `json:"case_type_id"`
	CaseNumber  int32     `json:"case_number"`
	CaseYear    int32     `json:"case_year"`
	CaseCourtID uuid.UUID `json:"case_court_id"`
	Tags        []string  `json:"tags"`
}

// ConvertDBCaseToCase converts a db case to a service case.
func ConvertDBCaseToCase(DBCase db.Case) Case {
	return Case{
		ID:          DBCase.ID,
		Name:        DBCase.Name,
		CreatedAt:   DBCase.CreatedAt,
		UpdatedAt:   DBCase.UpdatedAt,
		CaseTypeID:  DBCase.CaseTypeID,
		CaseNumber:  DBCase.CaseNumber,
		CaseYear:    DBCase.CaseYear,
		CaseCourtID: DBCase.CaseCourtID,
		Tags:        DBCase.Tags,
	}
}

// ConvertDBCaseTypeToCaseType converts a db case type to a service case type.
func ConvertDBCaseTypeToCaseType(dbCaseType db.CaseType) CaseType {
	return CaseType{
		ID:          dbCaseType.ID,
		Name:        dbCaseType.Name,
		Description: dbCaseType.Description,
	}
}

// CreateCase creates a new case in the database and minio.
func (s *Stores) CreateCase(ctx context.Context, userID uuid.UUID, request CreateCaseParams) (*Case, error) {
	// Begin a db transaction
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}

	// Defer a rollback in case anything fails.
	defer tx.Rollback()

	// Create a query object with the transaction
	q := s.DBStore.WithTx(tx)

	// Set current user in session_data
	err = q.SetCurrentUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("setting current userin audit: %w", err)
	}

	// Get CaseType and CourtType from ID
	caseType, err := q.GetCaseType(ctx, request.CaseTypeID)
	if err != nil {
		return nil, fmt.Errorf("error while getting case type name : %w", err)
	}

	courtType, err := q.GetCourtShortName(ctx, request.CaseCourtID)
	if err != nil {
		return nil, fmt.Errorf("error while getting court short name : %w", err)
	}

	caseName, err := GenerateCaseNameForDB(courtType.ShortName, caseType.Name, request.CaseNumber, request.CaseYear)
	if err != nil {
		return nil, fmt.Errorf("error generating case name : %w", err)
	}

	// Check if the case already exists
	exists, err := q.CaseExists(ctx, caseName)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, fmt.Errorf("%w : case : %q ", ErrAlreadyExists, caseName)
	}

	cs := db.CreateCaseParams{
		Name:        caseName,
		CaseTypeID:  request.CaseTypeID,
		CaseNumber:  request.CaseNumber,
		CaseYear:    request.CaseYear,
		CaseCourtID: request.CaseCourtID,
		Tags:        request.Tags,
	}

	// Create a case in the db
	createdCase, err := q.CreateCase(ctx, cs)
	if err != nil {
		return nil, fmt.Errorf("creating case in DB: %w", err)
	}

	// set the created case ID to user case ID for user_cases table
	userCase := db.CreateUserCaseParams{
		UserID: userID,
		CaseID: createdCase.ID,
	}

	// Add to user_cases record
	_, err = q.CreateUserCase(ctx, userCase)
	if err != nil {
		return nil, fmt.Errorf("creating record in user_cases: %w", err)
	}
	// Generate Case name for the ObjectStore
	caseNameMinio, err := GenerateCaseNameForMinio(courtType.ShortName, caseType.Name, cs.CaseNumber, cs.CaseYear)
	if err != nil {
		return nil, fmt.Errorf("error generating case name : %w", err)
	}

	cs.Name = caseNameMinio

	// Create a case in ObjectStore
	err = s.ObjectStore.CreateCase(ctx, cs)
	if err != nil {
		switch {
		case err.Error() == "Bucket name contains invalid characters":
			return nil, fmt.Errorf("%w : case contains invalid characters: %q ", ErrInvalidRequest, cs.Name)
		case err.Error() == "Bucket name cannot be empty":
			return nil, fmt.Errorf("%w : case name cannot be empty ", ErrInvalidRequest)
		default:
			return nil, fmt.Errorf("creating case in objects store: %w", err)
		}
	}

	// If we reach here, it means all operations are successful, so we commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	result := ConvertDBCaseToCase(createdCase)

	return &result, nil
}

// GetCaseByID will return the case with the given ID.
func (s *Stores) GetCaseByID(ctx context.Context, id uuid.UUID) (Case, error) {
	dbCS, err := s.DBStore.GetCase(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Case{}, fmt.Errorf("%w : case id : %d ", ErrNotFound, id)
		}

		return Case{}, err
	}
	cs := Case{
		ID:         dbCS.ID,
		Name:       dbCS.Name,
		CaseTypeID: dbCS.CaseTypeID,
	}

	return cs, nil
}

// ListCases will return a list of all the cases that are both inside database and minio and will ignore
// the cases that are only in minio but not in the database.
func (s *Stores) ListCases(ctx context.Context) ([]Case, error) {
	casesDB, err := s.DBStore.ListCases(ctx)
	if err != nil {
		return nil, fmt.Errorf("list cases from DB: %w ", err)
	}

	caseMap := make(map[string]Case)
	for _, caseDB := range casesDB {
		minioName, err := ConvertDBFormatToMinio(caseDB.Name)
		if err != nil {
			return nil, fmt.Errorf("converting db case name to minio: %w", err)
		}
		serviceCase := ConvertDBCaseToCase(caseDB)
		caseMap[minioName] = serviceCase
	}

	var List []Case
	casesFS, err := s.ObjectStore.ListCases(ctx)
	if err != nil {
		return nil, fmt.Errorf("list cases from object storage: %w ", err)
	}

	for _, caseFS := range casesFS {
		if localCase, ok := caseMap[caseFS.Name]; ok {
			List = append(List, localCase)
		}
	}

	return List, nil
}

// DeleteCase will delete the case with the given ID.
func (s *Stores) DeleteCase(ctx context.Context, id uuid.UUID) error {
	// Begin a db transaction
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	defer tx.Rollback()

	q := s.DBStore.WithTx(tx)

	// Set current user in session_data
	err = q.SetCurrentUser(ctx, uuid.Nil)
	if err != nil {
		return fmt.Errorf("setting current userin audit: %w", err)
	}

	// Get a case from DB
	caseDB, err := q.GetCase(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w : case id : %d ", ErrNotFound, id)
		}

		return err
	}

	// Delete case from DB
	err = q.DeleteCase(ctx, id)
	if err != nil {
		return fmt.Errorf("deleting case from DB: %w", err)
	}

	minioName, err := ConvertDBFormatToMinio(caseDB.Name)
	if err != nil {
		return fmt.Errorf("converting db case name to minio: %w", err)
	}

	// Delete case from ObjectStore
	err = s.ObjectStore.RemoveCase(ctx, minioName)
	if err != nil {
		return fmt.Errorf("deleting case from object store: %w", err)
	}

	// If we reach here, it means all operations are successful, so we commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

// CreateCaseType creates a new case type in the database. It verifies that the case type doesn't already exist.
// So the verification is not needed in the handler.
func (s *Stores) CreateCaseType(ctx context.Context, request CaseType) (CaseType, error) {
	// Check if the case type already exists
	exists, err := s.DBStore.CaseTypeExists(ctx, request.Name)
	if err != nil {
		return CaseType{}, err
	}

	if exists {
		return CaseType{}, fmt.Errorf("%w : case type : %q ", ErrAlreadyExists, request.Name)
	}

	cs := db.CreateCaseTypeParams{
		Name:        request.Name,
		Description: request.Description,
	}

	// Create a case type in the db
	createdCaseType, err := s.DBStore.CreateCaseType(ctx, cs)
	if err != nil {
		return CaseType{}, fmt.Errorf("creating case type in DB: %w", err)
	}

	return ConvertDBCaseTypeToCaseType(createdCaseType), nil
}

// UpdateCaseType updates a case type in the database.
func (s *Stores) UpdateCaseType(ctx context.Context, request CaseType) (CaseType, error) {
	// Check if the case type already exists
	exists, err := s.DBStore.CaseTypeExists(ctx, request.Name)
	if err != nil {
		return CaseType{}, err
	}

	if exists {
		return CaseType{}, fmt.Errorf("%w : case type : %q ", ErrAlreadyExists, request.Name)
	}

	cs := db.UpdateCaseTypeParams{
		ID:          request.ID,
		Name:        request.Name,
		Description: request.Description,
	}

	// Update a case type in the db
	updatedCaseType, err := s.DBStore.UpdateCaseType(ctx, cs)
	if err != nil {
		return CaseType{}, fmt.Errorf("updating case type in DB: %w", err)
	}

	return ConvertDBCaseTypeToCaseType(updatedCaseType), nil
}

// CaseType holds the details of a case type in the service layer.
type CaseType struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

// ListCaseTypes will return a list of all the case types.
func (s *Stores) ListCaseTypes(ctx context.Context) ([]CaseType, error) {
	DBCaseTypes, err := s.DBStore.GetCaseIDTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting all case types: %w", err)
	}

	caseTypes := make([]CaseType, 0, len(DBCaseTypes))

	for _, DBCaseType := range DBCaseTypes {
		caseType := CaseType{
			ID:          DBCaseType.ID,
			Name:        DBCaseType.Name,
			Description: DBCaseType.Description,
		}
		caseTypes = append(caseTypes, caseType)
	}
	return caseTypes, nil
}

// GetCaseTypeByID will return the case type with the given ID.
func (s *Stores) GetCaseTypeByID(ctx context.Context, id uuid.UUID) (CaseType, error) {
	// Check if the case type already exists
	exists, err := s.DBStore.CaseTypeExistsByID(ctx, id)
	if err != nil {
		return CaseType{}, err
	}

	if !exists {
		return CaseType{}, fmt.Errorf("%w : case type id : %d ", ErrNotFound, id)
	}

	DBCaseType, err := s.DBStore.GetCaseType(ctx, id)
	if err != nil {
		return CaseType{}, fmt.Errorf("getting case type from DB: %w", err)
	}

	caseType := ConvertDBCaseTypeToCaseType(DBCaseType)

	return caseType, nil
}

// The Court holds the details of a court in the service layer.
type Court struct {
	ID        uuid.UUID `json:"id"`
	Code      int32     `json:"code"`
	Name      string    `json:"name"`
	ShortName string    `json:"short_name"`
}

// GetCourts will return a list of all the courts.
func (s *Stores) GetCourts(ctx context.Context) ([]Court, error) {
	DBCourts, err := s.DBStore.ListCourts(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting all courts: %w", err)
	}

	courts := make([]Court, 0, len(DBCourts))

	for _, DBCourt := range DBCourts {
		c := Court{
			ID:        DBCourt.ID,
			Code:      DBCourt.Code,
			Name:      DBCourt.Name,
			ShortName: DBCourt.ShortName,
		}
		courts = append(courts, c)
	}

	return courts, nil
}
