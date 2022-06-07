package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"github.com/miloszizic/der/internal/data"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
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

func newTestServer(t *testing.T) *Application {
	//Making test application
	config, err := data.LoadProductionConfig("")
	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}
	logger := testLogger(io.Discard)
	tokenMaker, err := NewPasetoMaker(config.SymmetricKey)
	if err != nil {
		t.Errorf("failed to create tokenMaker maker: %v", err)
	}
	db, err := data.FromPostgresDB(config.Database.ConnectionInfo())
	if err != nil {
		t.Errorf("failed to create database: %v", err)
	}
	resetTestPostgresDB(db, t)
	minioCfg := config.Minio
	minioClient, err := data.FromMinio(
		minioCfg.Endpoint,
		minioCfg.AccessKey,
		minioCfg.SecretKey,
	)
	restartTestMinio(minioClient, t)
	if err != nil {
		t.Errorf("failed to create ostorage client: %v", err)
	}
	app := &Application{
		logger:     logger,
		tokenMaker: tokenMaker,
		config:     config,
		stores:     data.NewStores(db, minioClient),
	}
	return app

}

func resetTestPostgresDB(sqlDB *sql.DB, t *testing.T) {
	if _, err := sqlDB.Exec("TRUNCATE TABLE users,user_cases,evidences,cases,comments CASCADE;"); err != nil {
		t.Errorf("failed to truncate tables: %v", err)
	}
	if _, err := sqlDB.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1;"); err != nil {
		t.Errorf("failed to truncate tables: %v", err)
	}
	if _, err := sqlDB.Exec("ALTER SEQUENCE cases_id_seq RESTART WITH 1;"); err != nil {
		t.Errorf("failed to truncate tables: %v", err)
	}
	if _, err := sqlDB.Exec("ALTER SEQUENCE evidences_id_seq RESTART WITH 1;"); err != nil {
		t.Errorf("failed to truncate tables: %v", err)
	}
	if _, err := sqlDB.Exec("ALTER SEQUENCE comments_id_seq RESTART WITH 1;"); err != nil {
		t.Errorf("failed to truncate tables: %v", err)
	}
}

func restartTestMinio(client *minio.Client, t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	buckets, err := client.ListBuckets(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, bucket := range buckets {
		objects := client.ListObjects(ctx, bucket.Name, minio.ListObjectsOptions{})
		for object := range objects {
			err := client.RemoveObject(ctx, bucket.Name, object.Key, minio.RemoveObjectOptions{})
			if err != nil {
				t.Fatal(err)
			}
		}
		if err := client.RemoveBucket(context.Background(), bucket.Name); err != nil {
			t.Fatal(err)
		}
	}
}
func TestPingHandlerReturnedAPong(t *testing.T) {
	app := newTestServer(t)

	ping, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Errorf("Failed to create ping request: %s", err)
	}
	response := httptest.NewRecorder()
	app.Ping(response, ping)
	if response.Body.String() != "pong" {
		t.Errorf("Expected pong, got %s", response.Body.String())
	}

}
func TestLogin(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name: "successful with correct credentials",
			requestBody: map[string]interface{}{
				"username": "Simba",
				"password": "opsAdmin",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "unsuccessful with incorrect credentials",
			requestBody: map[string]interface{}{
				"username": "Simba",
				"password": "opsAdmin1",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "unsuccessful with missing username",
			requestBody: map[string]interface{}{
				"password": "opsAdmin",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "unsuccessful with missing password",
			requestBody: map[string]interface{}{
				"username": "Simba",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "unsuccessful with missing username and password",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestServer(t)
			user := &data.User{Username: "Simba"}
			err := user.Password.Set("opsAdmin")
			if err != nil {
				t.Errorf("failed to set password: %v", err)
			}
			err = app.stores.User.Add(user)
			if err != nil {
				t.Errorf("Failed to add user: %s", err)
			}
			body, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Errorf("Failed to marshal request body: %s", err)
			}
			response := httptest.NewRecorder()
			request, err := http.NewRequest("POST", "/Login", bytes.NewReader(body))
			if err != nil {
				t.Errorf("Error creating a new request: %v", err)
			}
			app.Login(response, request)
			if status := response.Code; status != tt.expectedStatus {
				t.Errorf("Handler returned wrong status code. Expected: %d. Got: %d.", tt.expectedStatus, status)
			}
		})
	}
}

func TestLoginFailedWithBadJSONRequest(t *testing.T) {
	app := newTestServer(t)
	response := httptest.NewRecorder()
	request, err := http.NewRequest("POST", "/Login", bytes.NewReader([]byte("badJsonRequest")))
	if err != nil {
		t.Errorf("Error creating a new request: %v", err)
	}
	app.Login(response, request)
	if status := response.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code. Expected: %d. Got: %d.", http.StatusBadRequest, status)
	}
}
func TestCreateUserHandlerCreatedANewUserSuccessfully(t *testing.T) {
	app := newTestServer(t)
	//Making test user
	createUserBody := map[string]interface{}{
		"username": "Simba",
		"password": "opsAdmin1",
	}
	requestBody, err := json.Marshal(createUserBody)
	if err != nil {
		t.Errorf("Failed to marshal request body: %s", err)
	}
	request, err := http.NewRequest("POST", "/register", bytes.NewReader(requestBody))
	if err != nil {
		t.Errorf("Error creating a new request: %v", err)
	}
	response := httptest.NewRecorder()
	app.CreateUserHandler(response, request)
	if status := response.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code. Expected: %d. Got: %d.", http.StatusCreated, status)
	}
}
func TestCreateUserHandlerFailedWithBadJSONRequest(t *testing.T) {
	app := newTestServer(t)
	requestBody, err := json.Marshal("badJsonRequest")
	if err != nil {
		t.Errorf("Failed to marshal request body: %s", err)
	}
	request, err := http.NewRequest("POST", "/register", bytes.NewReader(requestBody))
	if err != nil {
		t.Errorf("Error creating a new request: %v", err)
	}
	response := httptest.NewRecorder()
	app.CreateUserHandler(response, request)
	if status := response.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code. Expected: %d. Got: %d.", http.StatusBadRequest, status)
	}
}
