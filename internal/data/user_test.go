////go:build integration
package database_test

import (
	"evidence/internal/data/database"
	"testing"

	_ "github.com/lib/pq"
)

func TestCreatingUser(t *testing.T) {
	testCases := []struct {
		name     string
		user     *database.User
		password string
		wantErr  bool
	}{
		{
			name: "successfully with correct database",
			user: &database.User{
				Username: "simba",
			},
			password: "123456",
			wantErr:  false,
		},
		{
			name: "failed with incorrect database",
			user: &database.User{
				Username: "",
			},
			password: "123456",
			wantErr:  true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store, err := GetTestStores(t)
			if err != nil {
				t.Fatal(err)
			}
			err = tc.user.Password.Set(tc.password) // set password
			if err != nil {
				t.Fatal(err)
			}
			err = store.UserDB.Add(tc.user)
			if err != nil {
				if !tc.wantErr {
					t.Fatal(err)
				}
			}
		})
	}
}
func TestUserCredentials(t *testing.T) {
	testCases := []struct {
		name        string
		user        *database.User
		setPassword string
		password    string
		match       bool
	}{{
		name: "with correct password matches",
		user: &database.User{
			Username: "Simba",
		},
		setPassword: "opsAdmin",
		password:    "opsAdmin",
		match:       true,
	},
		{
			name: "with correct password matches",
			user: &database.User{
				Username: "phoebe",
			},
			setPassword: "opsAdmin",
			password:    "opsAdmin",
			match:       true,
		},
		{
			name: "with incorrect password does not match",
			user: &database.User{
				Username: "Mufasa",
			},
			setPassword: "opsAdmin",
			password:    "opsAdmin1",
			match:       false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store, err := GetTestStores(t)
			if err != nil {
				t.Fatal(err)
			}
			err = tc.user.Password.Set(tc.setPassword)
			if err != nil {
				t.Error(err)
			}
			err = store.UserDB.Add(tc.user)
			if err != nil {
				t.Error(err)
			}
			got, err := store.UserDB.GetByUsername(tc.user.Username)
			if err != nil {
				t.Errorf("retrieving specified user failed with error: %v", err)
			}
			match, err := got.Password.Matches(tc.password)
			if err != nil {
				t.Errorf("retrieving specified user failed with error: %v", err)
			}
			if match != tc.match {
				t.Errorf("got: %t, wanted: %t", match, tc.match)
			}
		})
	}
}
func TestAddingUserWithEmptyPasswordFails(t *testing.T) {
	user := &database.User{
		Username: "simba",
	}
	err := user.Password.Set("")
	if err == nil {
		t.Errorf("setting empty password should have failed but didn't")
	}
}
func TestDeletionOfUser(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		addUser  *database.User
		id       int64
		wantErr  bool
	}{
		{
			name:     "that exists successful",
			password: "123456",
			addUser: &database.User{
				Username: "Simba",
			},
			id:      1,
			wantErr: false,
		},
		{
			name:     "that also exists successful",
			password: "123456",
			addUser: &database.User{
				Username: "Pheobe",
			},
			id:      1,
			wantErr: false,
		},
		{
			name:     "that does not exist unsuccessful",
			password: "123456",
			addUser: &database.User{
				Username: "Mufasa",
			},
			id:      10,
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store, err := GetTestStores(t)
			if err != nil {
				t.Fatal(err)
			}
			err = tc.addUser.Password.Set(tc.password)
			if err != nil {
				t.Fatal(err)
			}
			err = store.UserDB.Add(tc.addUser)
			if err != nil {
				t.Error(err)
			}
			err = store.UserDB.Remove(tc.id) // delete user
			gotErr := err != nil
			if gotErr != tc.wantErr {
				t.Errorf("expected error to be %v but got %v", tc.wantErr, gotErr)
			}
		})
	}
}

func TestSearchingForNonExistingUserByIDFailed(t *testing.T) {
	store, err := GetTestStores(t)
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.UserDB.GetByID(1)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}
