package data_test

import (
	"context"
	"database/sql"
	"evidence/internal/data"
	"github.com/minio/minio-go/v7"
	"testing"
)

//getUserService returns a user service with a test database connection.
func GetTestStores(t *testing.T) (data.Stores, error) {
	config, err := data.LoadProductionConfig(false)
	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}
	db, err := data.FromPostgresDB(config.Database.ConnectionInfo())
	if err != nil {
		t.Errorf("Error connecting to database: %v", err)
	}
	resetTestPostgresDB(db, t)
	minioCfg := config.Minio
	minioClient, err := data.FromMinio(
		minioCfg.Endpoint,
		minioCfg.AccessKey,
		minioCfg.SecretKey,
	)
	restartTestMinio(minioClient, t)

	newStores := data.NewStores(db, minioClient)

	return newStores, nil
}
func resetTestPostgresDB(sqlDB *sql.DB, t *testing.T) {
	if _, err := sqlDB.Exec("TRUNCATE TABLE users,user_cases,evidences,cases,comments CASCADE;"); err != nil {
		t.Fatal(err)
	}
	if _, err := sqlDB.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1;"); err != nil {
		t.Fatal(err)
	}
	if _, err := sqlDB.Exec("ALTER SEQUENCE cases_id_seq RESTART WITH 1;"); err != nil {
		t.Fatal(err)
	}
	if _, err := sqlDB.Exec("ALTER SEQUENCE evidences_id_seq RESTART WITH 1;"); err != nil {
		t.Fatal(err)
	}
	if _, err := sqlDB.Exec("ALTER SEQUENCE comments_id_seq RESTART WITH 1;"); err != nil {
		t.Fatal(err)
	}
	return
}
func restartTestMinio(minioClient *minio.Client, t *testing.T) {
	buckets, err := minioClient.ListBuckets(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, bucket := range buckets {
		for object := range minioClient.ListObjects(context.Background(), bucket.Name, minio.ListObjectsOptions{}) {
			if object.Err != nil {
				t.Fatal(object.Err)
			}
			if err := minioClient.RemoveObject(context.Background(), bucket.Name, object.Key, minio.RemoveObjectOptions{}); err != nil {
				t.Fatal(err)
			}

		}
		err = minioClient.RemoveBucket(context.Background(), bucket.Name)
		if err != nil {
			t.Fatal(err)
		}
	}

}

func TestMinioConnectionToTestEnv(t *testing.T) {
	config, err := data.LoadProductionConfig(false)
	if err != nil {
		t.Errorf("Failed to load config: %s", err)
	}
	minioCFG := config.Minio
	minio, err := data.FromMinio(minioCFG.Endpoint, minioCFG.AccessKey, minioCFG.SecretKey)
	alive := minio.IsOnline()
	if !alive {
		t.Errorf("expexted ostorage to be online, but it was not")
	}
}
