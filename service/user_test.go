//go:build integration

package service_test

import (
	"context"
	"testing"

	"github.com/miloszizic/der/service"
)

func TestCreateUser(t *testing.T) {
	stores, err := service.GetTestStores(t)
	if err != nil {
		t.Fatalf("Error getting test stores: %v", err)
	}

	// Create the initial user
	initialUser := service.CreateUserParams{
		Username: "user1", // Changed the username to "user1"
		Password: "password",
	}

	_, err = stores.CreateUser(context.Background(), initialUser)
	if err != nil {
		t.Fatalf("Error creating initial user: %v", err)
	}

	tests := []struct {
		name    string
		user    service.CreateUserParams
		wantErr bool
	}{
		{
			name: "successful with valid input",
			user: service.CreateUserParams{
				Username: "user2", // Changed the username to "user2"
				Password: "password",
			},
			wantErr: false,
		},
		{
			name: "failed with duplicate username",
			user: service.CreateUserParams{
				Username: "user1", // Use the username of the initial user
				Password: "password",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := stores.CreateUser(context.Background(), tt.user)
			if (err != nil) != tt.wantErr { // The condition was incorrect, it should be != not ==
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	stores, err := service.GetTestStores(t)
	if err != nil {
		t.Fatalf("Error getting test stores: %v", err)
	}

	// Create the initial user
	initialUser := service.CreateUserParams{
		Username: "user1",
		Password: "password",
	}

	createdUser, err := stores.CreateUser(context.Background(), initialUser)
	if err != nil {
		t.Fatalf("Error creating initial user: %v", err)
	}

	roleID, err := stores.DBStore.GetRoleID(context.Background(), "admin")
	if err != nil {
		t.Fatalf("Error getting role id: %v", err)
	}

	tests := []struct {
		name    string
		user    service.UpdateUserParams
		wantErr bool
	}{
		{
			name: "successful update with valid input",
			user: service.UpdateUserParams{
				ID:        createdUser.ID,
				Username:  "user1",
				FirstName: "Updated Firstname",
				LastName:  "Updated Lastname",
				Email:     "new_email@example.com",
				RoleID:    roleID, // Adjust this based on your db state
			},
			wantErr: false,
		},
		{
			name: "failed with non-exist user",
			user: service.UpdateUserParams{
				Username:  "non_exist_user",
				FirstName: "Firstname",
				LastName:  "Lastname",
				Email:     "non_exist_email@example.com",
				RoleID:    roleID,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := stores.UpdateUser(context.Background(), tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
