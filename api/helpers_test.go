package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/miloszizic/der/service"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// zap testing logger
func testLogger(console io.Writer) *zap.SugaredLogger {
	config := zap.NewDevelopmentEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	testWriter := zapcore.AddSync(console)
	encoder := zapcore.NewConsoleEncoder(config)
	core := zapcore.NewCore(encoder, testWriter, zapcore.InfoLevel)
	logger := zap.New(core, zap.AddCaller())
	sugarLogger := logger.Sugar()

	return sugarLogger
}

func TestNotBlank(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value string
		want  bool
	}{
		{"test", true},
		{"   ", false},
		{"", false},
		{"\n\t", false},
	}

	for _, tt := range tests {
		if got := NotBlank(tt.value); got != tt.want {
			t.Errorf("NotBlank(%q) = %v; want %v", tt.value, got, tt.want)
		}
	}
}

func TestMinRunes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value string
		n     int
		want  bool
	}{
		{"test", 3, true},
		{"test", 5, false},
		{"", 1, false},
		{"🙂", 1, true}, // emoji has 1 rune
	}

	for _, tt := range tests {
		if got := MinRunes(tt.value, tt.n); got != tt.want {
			t.Errorf("MinRunes(%q, %d) = %v; want %v", tt.value, tt.n, got, tt.want)
		}
	}
}

func TestMaxRunes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value string
		n     int
		want  bool
	}{
		{"test", 5, true},
		{"test", 3, false},
		{"", 1, true},
		{"🙂", 2, true}, // emoji has 1 rune
	}

	for _, tt := range tests {
		if got := MaxRunes(tt.value, tt.n); got != tt.want {
			t.Errorf("MaxRunes(%q, %d) = %v; want %v", tt.value, tt.n, got, tt.want)
		}
	}
}

func TestBetween(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value, min, max int
		want            bool
	}{
		{5, 1, 10, true},
		{0, 1, 10, false},
		{11, 1, 10, false},
	}

	for _, tt := range tests {
		if got := Between(tt.value, tt.min, tt.max); got != tt.want {
			t.Errorf("Between(%d, %d, %d) = %v; want %v", tt.value, tt.min, tt.max, got, tt.want)
		}
	}
}

func TestMatches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value string
		rx    *regexp.Regexp
		want  bool
	}{
		{"test@example.com", RgxEmail, true},
		{"test@", RgxEmail, false},
	}

	for _, tt := range tests {
		if got := Matches(tt.value, tt.rx); got != tt.want {
			t.Errorf("Matches(%q) = %v; want %v", tt.value, got, tt.want)
		}
	}
}

func TestIn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value    int
		safelist []int
		want     bool
	}{
		{5, []int{1, 2, 3, 4, 5}, true},
		{0, []int{1, 2, 3, 4, 5}, false},
	}

	for _, tt := range tests {
		if got := In(tt.value, tt.safelist...); got != tt.want {
			t.Errorf("In(%d) = %v; want %v", tt.value, got, tt.want)
		}
	}
}

func TestAllIn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		values   []int
		safelist []int
		want     bool
	}{
		{[]int{1, 2, 3}, []int{1, 2, 3, 4, 5}, true},
		{[]int{1, 2, 3, 6}, []int{1, 2, 3, 4, 5}, false},
	}

	for _, tt := range tests {
		if got := AllIn(tt.values, tt.safelist...); got != tt.want {
			t.Errorf("AllIn(%v) = %v; want %v", tt.values, got, tt.want)
		}
	}
}

// NewTestEvidenceServer sets up a testing environment with a server application, a user, and a case.
// It uses testing.T to report errors in setting up the environment.
// This function is a helper function to set up the test environment for tests that require a server, a user, and a case.
func NewTestEvidenceServer(t *testing.T) (*Application, *service.User, *service.Case) {
	t.Helper()

	app := NewTestServer(t)
	stores := app.stores

	// create user for testing
	userReq := service.CreateUserParams{
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
	cs := service.CreateCaseParams{
		CaseTypeID:  caseTypeID,
		CaseNumber:  2,
		CaseYear:    2023,
		CaseCourtID: caseCourtID,
	}

	createdCase, err := stores.CreateCase(context.Background(), user.ID, cs)
	if err != nil {
		t.Fatalf("Error creating case: %v", err)
	}

	return app, &user, createdCase
}

// NewTestServer creates test server
func NewTestServer(t *testing.T) *Application {
	t.Helper()
	// making test application
	config := service.LoadDefaultConfig()

	logger := testLogger(io.Discard)

	tokenMaker, err := service.NewPasetoMaker(config.SymmetricKey)
	if err != nil {
		t.Errorf("failed to create tokenMaker maker: %v", err)
	}

	db, err := service.FromPostgresDB(config.Database.ConnectionInfo(), false, config.Env)
	if err != nil {
		t.Errorf("failed to create db: %v", err)
	}
	// restart the Postgres database
	service.ResetTestPostgresDB(t, db) // migrated from id to uuid so no reset needed

	minioCfg := config.Minio

	minioClient, err := service.FromMinio(
		minioCfg.Endpoint,
		minioCfg.AccessKey,
		minioCfg.SecretKey,
	)
	if err != nil {
		t.Errorf("failed to create ostorage client: %v", err)
	}
	// remove all the content on the minio server
	service.RestartTestMinio(context.Background(), t, minioClient)

	app := &Application{
		logger:     logger,
		tokenMaker: tokenMaker,
		config:     config,
		stores:     service.NewStores(db, minioClient),
	}

	return app
}

func TestRespondError(t *testing.T) {
	t.Parallel()

	app := &Application{}
	app.logger = zap.NewExample().Sugar()

	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "badly formed JSON",
			err:        &json.SyntaxError{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unexpected EOF",
			err:        io.ErrUnexpectedEOF,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unknown field",
			err:        errors.New("json: unknown field unknownField"),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "body too large",
			err:        errors.New("http: request body too large"),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty body",
			err:        io.EOF,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not found",
			err:        service.ErrNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "already exists",
			err:        service.ErrAlreadyExists,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "invalid request",
			err:        service.ErrInvalidRequest,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unauthorized",
			err:        service.ErrUnauthorized,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid credentials",
			err:        service.ErrInvalidCredentials,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "other error",
			err:        errors.New("some other error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		pt := tt
		t.Run(pt.name, func(t *testing.T) {
			t.Parallel()
			req, err := http.NewRequest(http.MethodPost, "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()

			app.respondError(recorder, req, pt.err)

			if recorder.Code != pt.wantStatus {
				t.Errorf("expected status code %d, got %d", pt.wantStatus, recorder.Code)
			}
		})
	}
}

// func TestHealthCheck(t *testing.T) {
//	t.Parallel()
//	app := NewTestServer(t)
//	req, err := http.NewRequest("GET", "/health", nil)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	recorder := httptest.NewRecorder()
//	app.HealthCheck(recorder, req)
//
//	if recorder.Code != http.StatusOK {
//		t.Errorf("expected status code %d, got %d", http.StatusOK, recorder.Code)
//	}
//
//	var response envelope
//	err = json.Unmarshal(recorder.Body.Bytes(), &response)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	healthStatus, ok := response["Health status"].(map[string]interface{})
//	if !ok {
//		t.Fatalf("expected health status to be of type map[string]interface{}")
//	}
//
//	if healthStatus["application"] != "online" {
//		t.Errorf("expected application to be online, got %s", healthStatus["application"])
//	}
//
//	if healthStatus["db"] != "online" {
//		t.Errorf("expected db to be online, got %s", healthStatus["db"])
//	}
//
//	if healthStatus["file_store"] != "online" {
//		t.Errorf("expected file_store to be online, got %s", healthStatus["file_store"])
//	}
//}
