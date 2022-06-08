package data_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/miloszizic/der/internal/data"
	"github.com/minio/minio-go/v7"
	"testing"
)

//getUserService returns a user service with a test database connection.
func GetTestStores(t *testing.T) (data.Stores, error) {
	config, err := data.LoadProductionConfig("")
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
	if err != nil {
		t.Errorf("Error connecting to minio: %v", err)
	}
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

func TestCreateUserCreatedNewUser(t *testing.T) {
	tests := []struct {
		name    string
		user    *data.UserRequest
		wantErr bool
	}{
		{
			name: "successful with valid input",
			user: &data.UserRequest{
				Username: "user",
				Password: "password",
			},
			wantErr: false,
		},
		{
			name: "fails with empty username",
			user: &data.UserRequest{
				Username: "",
				Password: "password",
			},
			wantErr: true,
		},
		{
			name: "fails with empty password",
			user: &data.UserRequest{
				Username: "user",
				Password: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stores, err := GetTestStores(t)
			if err != nil {
				t.Errorf("Error getting test stores: %v", err)
			}
			err = stores.CreateUser(tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("User.Add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
func TestMinioForTestingShouldBeOnline(t *testing.T) {
	config, err := data.LoadProductionConfig("")
	if err != nil {
		t.Errorf("Failed to load config: %s", err)
	}
	minioCFG := config.Minio
	client, err := data.FromMinio(minioCFG.Endpoint, minioCFG.AccessKey, minioCFG.SecretKey)
	if err != nil {
		t.Errorf("Failed to connect to client: %s", err)
	}
	alive := client.IsOnline()
	if !alive {
		t.Errorf("expexted ostorage to be online, but it was not")
	}
}
func TestCreateCaseInDBAndOBS(t *testing.T) {
	tests := []struct {
		name string
		cs   *data.Case
		want error
	}{
		{
			name: "with no name fails",
			cs: &data.Case{
				Name: "",
			},
			want: data.ErrInvalidRequest,
		},
		{
			name: "with not allowed special characters fails",
			cs: &data.Case{
				Name: "test/test",
			},
			want: data.ErrInvalidRequest,
		},
		{
			name: "with existing case fails",
			cs: &data.Case{
				Name: "case",
			},
			want: data.ErrAlreadyExists,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// get test stores
			stores, err := GetTestStores(t)
			if err != nil {
				t.Errorf("Error getting test stores: %v", err)
			}
			//create user for testing
			user := &data.User{
				Username: "test",
			}
			err = user.Password.Set("test")
			if err != nil {
				t.Errorf("Error setting password: %v", err)
			}
			// set user ID
			user.ID = 1
			err = stores.User.Add(user)
			if err != nil {
				t.Errorf("Error adding user: %v", err)
			}
			//Add a case for existing case test
			cs := &data.Case{
				Name: "case",
			}
			err = stores.CreateCase(user, cs.Name)
			if err != nil {
				t.Errorf("Error creating case: %v", err)
			}
			// testing CreateCase
			err = stores.CreateCase(user, tt.cs.Name)
			if !errors.Is(err, tt.want) {
				t.Errorf("got :%v, wanted :%v", err, tt.want)

			}
		})
	}
}
func TestCreateCaseSuccessfullyCreatedNewCaseInDBAndOBS(t *testing.T) {
	stores, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error getting test stores: %v", err)
	}
	user := &data.User{
		Username: "test",
	}
	err = user.Password.Set("test")
	if err != nil {
		t.Errorf("failed to set password: %v", err)
	}
	err = stores.User.Add(user)
	if err != nil {
		t.Errorf("failed to add user: %v", err)
	}
	want := &data.Case{
		Name: "test",
	}
	user.ID = 1
	err = stores.CreateCase(user, want.Name)
	if err != nil {
		t.Errorf("Error creating case: %v", err)
	}
	got, err := stores.GetCaseByID(1)
	if err != nil {
		t.Errorf("Error getting case: %v", err)
	}
	if !cmp.Equal(got, want, cmpopts.IgnoreFields(data.Case{}, "ID")) {
		t.Errorf(cmp.Diff(got, want))
	}
}
func TestCreateCaseFailsIfCaseAlreadyExists(t *testing.T) {
	stores, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error getting test stores: %v", err)
	}
	user := &data.User{
		Username: "test",
	}
	err = user.Password.Set("test")
	if err != nil {
		t.Errorf("failed to set password: %v", err)
	}
	err = stores.User.Add(user)
	if err != nil {
		t.Errorf("failed to add user: %v", err)
	}
	want := &data.Case{
		Name: "test",
	}
	user.ID = 1
	err = stores.CreateCase(user, want.Name)
	if err != nil {
		t.Errorf("Error creating case: %v", err)
	}
	err = stores.CreateCase(user, want.Name)
	if err == nil {
		t.Errorf("Expected error creating case, but got none")
	}
}
func TestCreateCaseFailsIfUserDoesNotExist(t *testing.T) {
	stores, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error getting test stores: %v", err)
	}
	cs := &data.Case{
		Name: "test",
	}
	user := &data.User{
		Username: "test",
	}
	err = stores.CreateCase(user, cs.Name)
	if err == nil {
		t.Errorf("Expected error creating case, but got none")
	}
}
func TestRemoveCaseRemovedACaseSuccessfullyFromDBAndOBS(t *testing.T) {
	stores, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error getting test stores: %v", err)
	}
	user := &data.User{
		Username: "test",
	}
	err = user.Password.Set("test")
	if err != nil {
		t.Errorf("failed to set password: %v", err)
	}
	err = stores.User.Add(user)
	if err != nil {
		t.Errorf("failed to add user: %v", err)
	}
	want := &data.Case{
		Name: "test",
	}
	user.ID = 1
	err = stores.CreateCase(user, want.Name)
	if err != nil {
		t.Errorf("Error creating case: %v", err)
	}
	err = stores.RemoveCase("test")
	if err != nil {
		t.Errorf("Error removing case: %v", err)
	}
	_, err = stores.GetCaseByID(1)
	if err == nil {
		t.Errorf("Expected error getting case, but got none")
	}
}
func TestRemoveCaseFailedForMissingCase(t *testing.T) {
	stores, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error getting test stores: %v", err)
	}
	err = stores.RemoveCase("test")
	if err == nil {
		t.Errorf("Expected error removing case, but got none")
	}
}
func TestRemoveCaseThatOnlyExistsInDBFails(t *testing.T) {
	stores, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error getting test stores: %v", err)
	}
	user := &data.User{
		Username: "test",
	}
	err = user.Password.Set("test")
	if err != nil {
		t.Errorf("failed to set password: %v", err)
	}
	err = stores.User.Add(user)
	if err != nil {
		t.Errorf("failed to add user: %v", err)
	}
	user.ID = 1
	cs := &data.Case{
		Name: "test",
	}
	err = stores.DBStore.AddCase(cs, user)
	if err != nil {
		t.Errorf("Error creating case: %v", err)
	}
	err = stores.RemoveCase("test")
	if err == nil {
		t.Errorf("Expected error removing case, but got none")
	}

}

func TestListCasesRetrievedAllCasesThatExistInDBAAndOBS(t *testing.T) {
	stores, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error getting test stores: %v", err)
	}
	user := &data.User{
		Username: "test",
	}
	err = user.Password.Set("test")
	if err != nil {
		t.Errorf("Error setting password: %v", err)
	}
	err = stores.User.Add(user)
	if err != nil {
		t.Errorf("Error adding user: %v", err)
	}
	user.ID = 1
	want := []data.Case{
		{
			Name: "test",
		},
		{
			Name: "test2",
		},
	}
	for _, c := range want {
		err = stores.CreateCase(user, c.Name)
		if err != nil {
			t.Errorf("Error creating case: %v", err)
		}
	}
	cases, err := stores.ListCases()
	if err != nil {
		t.Errorf("Error listing cases: %v", err)
	}
	if !cmp.Equal(cases, want, cmpopts.IgnoreFields(data.Case{}, "ID")) {
		t.Errorf(cmp.Diff(cases, want))
	}

}

func TestCreateEvidenceSuccessfullyCreatedNewEvidenceInDBAndOBS(t *testing.T) {
	stores, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error getting test stores: %v", err)
	}
	user := &data.User{
		Username: "test",
	}
	err = user.Password.Set("test")
	if err != nil {
		t.Errorf("Error setting password: %v", err)
	}

	err = stores.User.Add(user)
	if err != nil {
		t.Errorf("Error adding user: %v", err)
	}
	user.ID = 1
	// Create a case
	err = stores.CreateCase(user, "test")
	if err != nil {
		t.Errorf("Error creating evidence: %v", err)
	}
	ev := &data.Evidence{
		CaseID: 1,
		Name:   "test",
		File:   bytes.NewBufferString("test"),
	}
	cs := &data.Case{
		Name: "test",
	}
	err = stores.CreateEvidence(ev, cs)
	if err != nil {
		t.Errorf("Error creating evidence: %v", err)
	}
}
func TestCreateEvidenceRemovesEvidenceFromOBSIfAddingToDBFails(t *testing.T) {
	stores, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error getting test stores: %v", err)
	}
	user := &data.User{
		Username: "test",
	}
	err = user.Password.Set("test")
	if err != nil {
		t.Errorf("Error setting password: %v", err)
	}

	err = stores.User.Add(user)
	if err != nil {
		t.Errorf("Error adding user: %v", err)
	}
	user.ID = 1
	// Create a case
	err = stores.CreateCase(user, "test")
	if err != nil {
		t.Errorf("Error creating evidence: %v", err)
	}
	ev := &data.Evidence{
		Name: "test",
		File: bytes.NewBufferString("test"),
	}
	cs := &data.Case{
		Name: "test",
	}
	err = stores.CreateEvidence(ev, cs)
	if err == nil {
		t.Errorf("should return an error, but it did not")
	}
	_, err = stores.OBStore.GetEvidence(cs.Name, ev.Name)
	if errors.Is(err, data.ErrNotFound) {
		t.Errorf("expected err : %v, but got %v", data.ErrNotFound, err)
	}

}
func TestCreateEvidenceFailsIfEvidenceExistsInDB(t *testing.T) {
	stores, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error getting test stores: %v", err)
	}
	user := &data.User{
		Username: "test",
	}
	err = user.Password.Set("test")
	if err != nil {
		t.Errorf("Error setting password: %v", err)
	}

	err = stores.User.Add(user)
	if err != nil {
		t.Errorf("Error adding user: %v", err)
	}
	user.ID = 1
	// Create a case
	err = stores.CreateCase(user, "test")
	if err != nil {
		t.Errorf("Error creating evidence: %v", err)
	}
	ev := &data.Evidence{
		CaseID: 1,
		Name:   "test",
		File:   bytes.NewBufferString("test"),
	}
	cs := &data.Case{
		Name: "test",
	}
	_, err = stores.DBStore.CreateEvidence(ev)
	if err != nil {
		t.Errorf("Error creating evidence: %v", err)
	}
	err = stores.CreateEvidence(ev, cs)
	if err == nil {
		t.Errorf("Expected error creating evidence, but got none")
	}
	exists, err := stores.OBStore.EvidenceExists(cs.Name, ev.Name)
	if err != nil {
		t.Errorf("Error checking evidence exists: %v", err)
	}
	if exists {
		t.Errorf("Expected evidence to not exist, but it does")
	}

}
func TestCreateEvidenceFailsIfEvidenceExistsInOBS(t *testing.T) {
	stores, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error getting test stores: %v", err)
	}
	user := &data.User{
		Username: "test",
	}
	err = user.Password.Set("test")
	if err != nil {
		t.Errorf("Error setting password: %v", err)
	}

	err = stores.User.Add(user)
	if err != nil {
		t.Errorf("Error adding user: %v", err)
	}
	user.ID = 1
	// Create a case
	err = stores.CreateCase(user, "test")
	if err != nil {
		t.Errorf("Error creating evidence: %v", err)
	}
	ev := &data.Evidence{
		CaseID: 1,
		Name:   "test",
		File:   bytes.NewBufferString("test"),
	}
	cs := &data.Case{
		Name: "test",
	}
	_, err = stores.OBStore.CreateEvidence(ev, "test", ev.File)
	if err != nil {
		t.Errorf("Error creating evidence: %v", err)
	}
	err = stores.CreateEvidence(ev, cs)
	if !errors.Is(err, data.ErrAlreadyExists) {
		t.Errorf("Expected code : %v, but got %v", data.ErrAlreadyExists, err)
	}
}
func TestGetEvidenceByID(t *testing.T) {
	tests := []struct {
		name    string
		evID    int64
		csID    int64
		wantErr bool
	}{
		{
			name:    "successful retrieve evidence by correct ID",
			evID:    1,
			csID:    1,
			wantErr: false,
		},
		{
			name:    "fail to retrieve evidence by incorrect ID",
			evID:    2,
			csID:    1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stores, err := GetTestStores(t)
			if err != nil {
				t.Errorf("Error getting test stores: %v", err)
			}
			user := &data.User{
				Username: "test",
			}
			err = user.Password.Set("test")
			if err != nil {
				t.Errorf("Error setting password: %v", err)
			}

			err = stores.User.Add(user)
			if err != nil {
				t.Errorf("Error adding user: %v", err)
			}
			user.ID = 1
			// Create a case
			err = stores.CreateCase(user, "test")
			if err != nil {
				t.Errorf("Error creating evidence: %v", err)
			}
			ev := &data.Evidence{
				CaseID: 1,
				Name:   "test",
				File:   bytes.NewBufferString("test"),
			}
			cs := &data.Case{
				Name: "test",
			}
			err = stores.CreateEvidence(ev, cs)
			if err != nil {
				t.Errorf("Error creating evidence: %v", err)
			}
			_, err = stores.GetEvidenceByID(tt.evID, tt.csID)
			got := err != nil
			if got != tt.wantErr {
				t.Errorf("wated error: %v, but got %v", tt.wantErr, err)
			}
		})
	}
}
func TestDownloadEvidence(t *testing.T) {
	tests := []struct {
		name     string
		evidence *data.Evidence
		wantErr  bool
	}{
		{
			name: "successfully downloaded evidence",
			evidence: &data.Evidence{
				ID:     1,
				CaseID: 1,
				Name:   "test",
			},
			wantErr: false,
		},
		{
			name: "failed, because it doesn't exist",
			evidence: &data.Evidence{
				ID:     2,
				CaseID: 1,
				Name:   "picture", // doesn't exist
			},
			wantErr: true,
		},
		{
			name: "failed because case doesn't exist",
			evidence: &data.Evidence{
				ID:     1,
				CaseID: 2, // Case doesn't exist
				Name:   "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stores, err := GetTestStores(t)
			if err != nil {
				t.Errorf("Error getting test stores: %v", err)
			}
			user := &data.User{
				Username: "test",
			}
			err = user.Password.Set("test")
			if err != nil {
				t.Errorf("Error setting password: %v", err)
			}

			err = stores.User.Add(user)
			if err != nil {
				t.Errorf("Error adding user: %v", err)
			}
			user.ID = 1
			// Create a case
			err = stores.CreateCase(user, "test")
			if err != nil {
				t.Errorf("Error creating evidence: %v", err)
			}
			ev := &data.Evidence{
				CaseID: 1,
				Name:   "test",
				File:   bytes.NewBufferString("test"),
			}
			cs := &data.Case{
				Name: "test",
			}
			err = stores.CreateEvidence(ev, cs)
			if err != nil {
				t.Errorf("Error creating evidence: %v", err)
			}
			_, err = stores.DownloadEvidence(tt.evidence)
			got := err != nil
			if got != tt.wantErr {
				t.Errorf("wated error: %v, but got %v", tt.wantErr, err)
			}
		})
	}
}
func TestDeleteEvidence(t *testing.T) {
	tests := []struct {
		name     string
		evidence *data.Evidence
		wantErr  bool
	}{
		{
			name: "successfully deleted evidence",
			evidence: &data.Evidence{
				ID:     1,
				CaseID: 1,
				Name:   "test",
			},
			wantErr: false,
		},
		{
			name: "failed, because it doesn't exist",
			evidence: &data.Evidence{
				ID:     2,
				CaseID: 1,
				Name:   "picture", // doesn't exist
			},
			wantErr: true,
		},
		{
			name: "failed because case doesn't exist",
			evidence: &data.Evidence{
				ID:     1,
				CaseID: 2, // Case doesn't exist
				Name:   "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stores, err := GetTestStores(t)
			if err != nil {
				t.Errorf("Error getting test stores: %v", err)
			}
			user := &data.User{
				Username: "test",
			}
			err = user.Password.Set("test")
			if err != nil {
				t.Errorf("Error setting password: %v", err)
			}

			err = stores.User.Add(user)
			if err != nil {
				t.Errorf("Error adding user: %v", err)
			}
			user.ID = 1
			// Create a case
			err = stores.CreateCase(user, "test")
			if err != nil {
				t.Errorf("Error creating evidence: %v", err)
			}
			ev := &data.Evidence{
				CaseID: 1,
				Name:   "test",
				File:   bytes.NewBufferString("test"),
			}
			cs := &data.Case{
				Name: "test",
			}
			err = stores.CreateEvidence(ev, cs)
			if err != nil {
				t.Errorf("Error creating evidence: %v", err)
			}
			err = stores.DeleteEvidence(tt.evidence)
			got := err != nil
			if got != tt.wantErr {
				t.Errorf("wated error: %v, but got %v", tt.wantErr, err)
			}
		})
	}
}

func TestListEvidencesThatExistInDBAAndOBSSuccessfullyRetried(t *testing.T) {
	stores, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error getting test stores: %v", err)
	}
	user := &data.User{
		Username: "test",
	}
	err = user.Password.Set("test")
	if err != nil {
		t.Errorf("Error setting password: %v", err)
	}

	err = stores.User.Add(user)
	if err != nil {
		t.Errorf("Error adding user: %v", err)
	}
	user.ID = 1
	// Create a case
	err = stores.CreateCase(user, "test")
	if err != nil {
		t.Errorf("Error creating evidence: %v", err)
	}
	evs := []data.Evidence{
		{
			CaseID: 1,
			Name:   "test",
			File:   bytes.NewBufferString("test"),
		},
		{
			CaseID: 1,
			Name:   "test2",
			File:   bytes.NewBufferString("test2"),
		},
		{
			CaseID: 1,
			Name:   "test3",
			File:   bytes.NewBufferString("test3"),
		},
	}

	for _, ev := range evs {
		err = stores.CreateEvidence(&ev, &data.Case{
			Name: "test",
		})
		if err != nil {
			t.Errorf("Error creating evidence: %v", err)
		}
	}
	cs := &data.Case{
		ID:   1,
		Name: "test",
	}
	got, err := stores.ListEvidences(cs)
	fmt.Println(got)
	if err != nil {
		t.Errorf("Error listing all evidences from the case: %v", err)
	}
	want := 3
	if len(got) != want {
		t.Errorf("wanted %v evidences, but got %v", want, len(got))
	}
}
func TestAddEvidenceCommentAddedSuccessfully(t *testing.T) {
	stores, err := GetTestStores(t)
	if err != nil {
		t.Errorf("Error getting test stores: %v", err)
	}
	user := &data.User{
		Username: "test",
	}
	err = user.Password.Set("test")
	if err != nil {
		t.Errorf("Error setting password: %v", err)
	}

	err = stores.User.Add(user)
	if err != nil {
		t.Errorf("Error adding user: %v", err)
	}
	user.ID = 1
	// Create a case
	err = stores.CreateCase(user, "test")
	if err != nil {
		t.Errorf("Error creating evidence: %v", err)
	}
	evs := data.Evidence{
		CaseID: 1,
		Name:   "test",
		File:   bytes.NewBufferString("test"),
	}
	cs := &data.Case{
		ID:   1,
		Name: "test",
	}
	err = stores.CreateEvidence(&evs, cs)
	if err != nil {
		t.Errorf("Error creating evidence: %v", err)
	}
	comment := data.Comment{
		EvidenceID: 1,
		Text:       "text comment text",
	}
	err = stores.AddEvidenceComment(&comment)
	if err != nil {
		t.Errorf("Error adding comment: %v", err)
	}
	got, err := stores.DBStore.GetCommentsByID(1)
	if err != nil {
		t.Errorf("Error getting evidence: %v", err)
	}
	if got[0].Text != comment.Text {
		t.Errorf("wanted %v, but got %v", comment.Text, got[0].Text)
	}

}
