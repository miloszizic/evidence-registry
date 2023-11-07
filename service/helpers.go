package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// getProjectRoot finds the project root by looking for the directory containing go.mod file.
func getProjectRoot() (string, error) {
	// Get the current file directory
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("unable to get the current file directory")
	}
	currentDir := filepath.Dir(currentFile)

	// Walk up from the current file path
	for {
		// Check does go.mod exists in the current directory
		if _, err := os.Stat(filepath.Join(currentDir, "go.mod")); err == nil {
			return currentDir, nil
		}

		// If reached the root directory, return with error
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			return "", fmt.Errorf("unable to find the project root directory ")
		}

		// Else, move to the parent directory
		currentDir = parentDir
	}
}

// FromPostgresDB opens a connection to a Postgres db.
func FromPostgresDB(dsn string, automigrate bool, env string) (*sql.DB, error) {
	const (
		// maxOpenConns is the maximum number of open connections to the database.
		maxOpenConns = 25
		// maxIdleConns is the maximum number of connections in the idle connection pool.
		maxIdleConns = 25
		// maxConnLifetime is the maximum amount of time a connection may be reused.
		maxConnIdleTime = 5 * time.Minute
		// maxConnIdleTime is the maximum amount of time a connection may be reused.
		maxConnLifetime = 2 * time.Hour
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening DB connection: %w", err)
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxIdleTime(maxConnIdleTime)
	db.SetConnMaxLifetime(maxConnLifetime)

	if automigrate {
		log.Println("Migration is enabled")

		projectRoot, err := getProjectRoot()
		if err != nil {
			log.Fatalf("Failed to get project root: %v", err)
		}

		migrationsDir := fmt.Sprintf("file://%s/db/migration", projectRoot)

		migrator, err := migrate.New(migrationsDir, dsn)
		if err != nil {
			return nil, fmt.Errorf("creating migrator: %w", err)
		}

		log.Println("Applying migrations")

		err = migrator.Up()

		switch {
		case errors.Is(err, migrate.ErrNoChange):
			if env != "test" {
				log.Println("No changes to migrate")
			}
		case err != nil:
			return nil, fmt.Errorf("applying migrations: %w", err)
		}
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

// GenerateCaseNameForMinio generates a unique case name for minio.
func GenerateCaseNameForMinio(courtShortName string, caseTypeName string, caseNumber int32, caseYear int32) (string, error) {
	// lowerYerLimit in a minimum age allowed for the case
	const lowerYearLimit = 1000

	const yearModulus = 100

	// Ensure the court name is not empty
	if courtShortName == "" {
		return "", errors.New("court short name must not be empty")
	}
	// Ensure the case type name is not empty
	if caseTypeName == "" {
		return "", errors.New("case type name must not be empty")
	}
	// Ensure the case number is not negative
	if caseNumber < 0 {
		return "", errors.New("case number must not be negative")
	}
	// Ensure the case year is a reasonable year (say, not less than 1000)
	if caseYear < lowerYearLimit {
		return "", errors.New("case year must be a valid year (not less than 1000)")
	}
	// Format case name
	caseName := fmt.Sprintf("%s-%s-%d-%d", strings.ToLower(courtShortName), strings.ToLower(caseTypeName), caseNumber, caseYear%yearModulus)

	return caseName, nil
}

// GenerateCaseNameForDB generates a unique case name for the database.
func GenerateCaseNameForDB(courtShortName string, caseTypeName string, caseNumber int32, caseYear int32) (string, error) {
	// lowerYerLimit in a minimum age allowed for the case
	const lowerYearLimit = 1000

	const yearModulus = 100

	// Ensure the court name is not empty
	if courtShortName == "" {
		return "", errors.New("court short name must not be empty")
	}

	// Ensure the case type name is not empty
	if caseTypeName == "" {
		return "", errors.New("case type name must not be empty")
	}

	// Ensure the case number is not negative
	if caseNumber < 0 {
		return "", errors.New("case number must not be negative")
	}

	// Ensure the case year is a reasonable year (say, not less than 1000)
	if caseYear < lowerYearLimit {
		return "", errors.New("case year must be a valid year (not less than 1000)")
	}

	// Format case name for db
	caseNameForDB := fmt.Sprintf("%s %s %d/%d", courtShortName, caseTypeName, caseNumber, caseYear%yearModulus)

	return caseNameForDB, nil
}

// ConvertDBFormatToMinio converts a case name from the database format to the minio format.
func ConvertDBFormatToMinio(dbFormat string) (string, error) {
	// Replace all white space with dashes and convert to a lower case
	minioFormat := strings.ReplaceAll(strings.ToLower(dbFormat), " ", "-")

	// Replace "/" with "-"
	minioFormat = strings.ReplaceAll(minioFormat, "/", "-")

	return minioFormat, nil
}

// HandleNullableString converts a string into a sql.NullString
func HandleNullableString(input string) sql.NullString {
	if input != "" {
		return sql.NullString{String: input, Valid: true}
	}

	return sql.NullString{Valid: false}
}

// HandleNullableBool converts a bool into a sql.NullBool
func HandleNullableBool(input bool) sql.NullBool {
	if input {
		return sql.NullBool{Bool: input, Valid: true}
	}

	return sql.NullBool{Valid: false}
}

// HandleNullableInt32 converts an int32 into a sql.NullInt32
func HandleNullableInt32(input int32) sql.NullInt32 {
	if input != 0 {
		return sql.NullInt32{Int32: input, Valid: true}
	}

	return sql.NullInt32{Valid: false}
}

// HandleNullableUUID converts a UUID into a sql.NullUUID
func HandleNullableUUID(input uuid.UUID) uuid.NullUUID {
	if input != uuid.Nil {
		return uuid.NullUUID{UUID: input, Valid: true}
	}

	return uuid.NullUUID{Valid: false}
}

// GetTestStores generates test stores for testing purposes
func GetTestStores(t *testing.T) (Stores, error) {
	t.Helper()
	// load test config
	config := LoadDefaultConfig()

	db, err := FromPostgresDB(config.Database.ConnectionInfo(), config.Database.Automigrate, config.Env)
	if err != nil {
		t.Errorf("Error connecting to db: %v", err)
	}
	// Restarting the database
	ResetTestPostgresDB(t, db)

	minioCfg := config.Minio

	minioClient, err := FromMinio(
		minioCfg.Endpoint,
		minioCfg.AccessKey,
		minioCfg.SecretKey,
	)
	if err != nil {
		t.Errorf("Error connecting to minio: %v", err)
	}
	// Restarting the minio
	RestartTestMinio(context.Background(), t, minioClient)

	newStores := NewStores(db, minioClient)

	return newStores, nil
}

// ResetTestPostgresDB resets the test postgres database.
func ResetTestPostgresDB(t *testing.T, sqlDB *sql.DB) {
	t.Helper()
	// Truncate tables
	tables := []string{
		"app_users",
		"user_cases",
		"cases",
		"evidence",
		"audit_logs",
	}
	for _, table := range tables {
		if _, err := sqlDB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE;", table)); err != nil {
			t.Fatalf("Failed to truncate table %s: %v", table, err)
		}
	}
}

// RestartTestMinio removes all content from minio test server
func RestartTestMinio(ctx context.Context, t *testing.T, minioClient *minio.Client) {
	t.Helper()

	buckets, err := minioClient.ListBuckets(ctx)
	if err != nil {
		t.Fatal(err)
	}

	for _, bucket := range buckets {
		for object := range minioClient.ListObjects(ctx, bucket.Name, minio.ListObjectsOptions{}) {
			if object.Err != nil {
				t.Fatal(object.Err)
			}

			if err := minioClient.RemoveObject(ctx, bucket.Name, object.Key, minio.RemoveObjectOptions{}); err != nil {
				t.Fatal(err)
			}
		}

		err = minioClient.RemoveBucket(ctx, bucket.Name)
		if err != nil {
			t.Fatal(err)
		}
	}
}

// NewEvidenceTestingStores returns a database with preexisting username test and case
func NewEvidenceTestingStores(t *testing.T) (Stores, User, *Case, error) {
	t.Helper()

	const (
		// caseNumber is a case number for testing
		caseNumber = 2
		// caseYear is a case year for testing
		caseYear = 2023
	)

	stores, err := GetTestStores(t)
	if err != nil {
		t.Fatalf("Error getting test stores: %v", err)
	}
	// create user for testing
	userReq := CreateUserParams{
		Username: "test",
		Password: "test",
	}

	user, err := stores.CreateUser(context.Background(), userReq)
	if err != nil {
		t.Fatalf("Error adding user: %v", err)
	}

	// create CaseTypeID and CaseCourtID
	caseTypeID, err := stores.DBStore.GetCaseTypeIDByName(context.Background(), "KM")
	if err != nil {
		t.Fatalf("Error getting CaseTypeID: %v", err)
	}

	caseCourtID, err := stores.DBStore.GetCourtIDByShortName(context.Background(), "OSPG")
	if err != nil {
		t.Fatalf("Error getting CaseCourtID: %v", err)
	}

	// Create a case for existing case test
	cs := CreateCaseParams{
		CaseTypeID:  caseTypeID,
		CaseNumber:  caseNumber,
		CaseYear:    caseYear,
		CaseCourtID: caseCourtID,
	}

	createdCase, err := stores.CreateCase(context.Background(), user.ID, cs)
	if err != nil {
		t.Fatalf("Error creating case: %v", err)
	}

	return stores, user, createdCase, nil
}
