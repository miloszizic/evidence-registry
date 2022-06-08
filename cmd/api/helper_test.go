package api

import (
	"context"
	"database/sql"
	"github.com/miloszizic/der/internal/data"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
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

// seedForHandlerTesting seeds the database with one user and one case for testing
func seedForHandlerTesting(t *testing.T, app *Application) {
	// get new test server
	user := &data.User{
		Username: "test",
	}
	err := user.Password.Set("test")
	if err != nil {
		t.Errorf("failed to set password: %v", err)
	}
	err = app.stores.User.Add(user)
	if err != nil {
		t.Fatal(err)
	}
	// create a case
	cs := &data.Case{
		Name: "test",
	}
	user.ID = 1
	err = app.stores.DBStore.AddCase(cs, user)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
	err = app.stores.OBStore.CreateCase(cs)
	if err != nil {
		t.Errorf("failed to add case: %v", err)
	}
}
