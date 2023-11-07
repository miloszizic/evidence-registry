package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/miloszizic/der/db"
	"github.com/miloszizic/der/vault"
)

// CreateUserParams are used to create a new user in the database.
type CreateUserParams struct {
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	RoleID    uuid.UUID `json:"role_id"`
	Validator Validator `json:"-"`
}

// User holds the information about a user.
type User struct {
	ID        uuid.UUID      `json:"id"`
	Username  string         `json:"username"`
	Email     string         `json:"email"`
	Password  string         `json:"password"`
	RoleID    uuid.NullUUID  `json:"role_id"`
	FirstName sql.NullString `json:"first_name"`
	LastName  sql.NullString `json:"last_name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// UserWithRoles holds the information about a user with roles.
type UserWithRoles struct {
	ID        uuid.UUID      `json:"id"`
	Username  string         `json:"username"`
	Email     string         `json:"email"`
	Password  string         `json:"password"`
	RoleID    uuid.NullUUID  `json:"role_id"`
	FirstName sql.NullString `json:"first_name"`
	LastName  sql.NullString `json:"last_name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	RoleName  string         `json:"role_name"`
}

// ConvertDBAppUserToUser converts a db app user to a service app user.
func ConvertDBAppUserToUser(dbUser db.AppUser) User {
	return User{
		ID:        dbUser.ID,
		Username:  dbUser.Username,
		Email:     dbUser.Email,
		Password:  dbUser.Password,
		RoleID:    dbUser.RoleID,
		FirstName: dbUser.FirstName,
		LastName:  dbUser.LastName,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
	}
}

// ConvertDBUsersWithRolesToUserWithRoles converts a db user with roles to a service user with roles.
func ConvertDBUsersWithRolesToUserWithRoles(dbUser db.GetUsersWithRolesRow) UserWithRoles {
	return UserWithRoles{
		ID:        dbUser.ID,
		Username:  dbUser.Username,
		Email:     dbUser.Email,
		Password:  dbUser.Password,
		RoleID:    dbUser.RoleID,
		FirstName: dbUser.FirstName,
		LastName:  dbUser.LastName,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		RoleName:  dbUser.RoleName,
	}
}

// CreateUser creates a new user in the db.
func (s *Stores) CreateUser(ctx context.Context, request CreateUserParams) (User, error) {
	// check if the user exists in the db
	exist, err := s.DBStore.UserExists(ctx, request.Username)
	if err != nil {
		return User{}, fmt.Errorf("cheking user in DB store: %w , username: %q ", err, request.Username)
	}

	if exist {
		return User{}, fmt.Errorf("user %q already exists", request.Username)
	}
	// hash the password
	hashedPassword, err := vault.Hash(request.Password)
	if err != nil {
		return User{}, fmt.Errorf("hashing password: %w", err)
	}
	// handle nullable fields for the db
	FirstName := HandleNullableString(request.FirstName)
	LastName := HandleNullableString(request.LastName)
	RoleID := HandleNullableUUID(request.RoleID)

	// define CreateUserParams
	user := db.CreateUserParams{
		Username:  request.Username,
		Password:  hashedPassword,
		FirstName: FirstName,
		LastName:  LastName,
		Email:     request.Email,
		RoleID:    RoleID,
	}
	// create user in the db
	dbUser, err := s.DBStore.CreateUser(ctx, user)
	if err != nil {
		return User{}, fmt.Errorf("creating user in DB: %w , username: %q ", err, request.Username)
	}

	return ConvertDBAppUserToUser(dbUser), nil
}

// GetUserByUsername returns a user with a given username.
func (s *Stores) GetUserByUsername(ctx context.Context, username string) (User, error) {
	dbUser, err := s.DBStore.GetUserByUsername(ctx, username)
	if err != nil {
		return User{}, fmt.Errorf("getting user from DB: %w", err)
	}

	user := ConvertDBAppUserToUser(dbUser)

	return user, nil
}

// UpdateUserPasswordParams are used to update a user password in the db.
type UpdateUserPasswordParams struct {
	Password  string    `json:"password"`
	ID        uuid.UUID `json:"id"`
	Validator Validator `json:"-"`
}

// UpdateUserPassword holds the information about a user password update for service layer.
func (s *Stores) UpdateUserPassword(ctx context.Context, request UpdateUserPasswordParams) error {
	// hash the password
	hashedPassword, err := vault.Hash(request.Password)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	params := db.UpdateUserPasswordParams{
		ID:       request.ID,
		Password: hashedPassword,
	}

	_, err = s.DBStore.UpdateUserPassword(ctx, params)
	if err != nil {
		return fmt.Errorf("updating user password in DB: %w , username: %q ", err, params.ID)
	}

	return nil
}

// AddRoleToUser adds a role to a user in the db. It checks if the user and the role exist in the db, so verification
// is not needed in the handler.
func (s *Stores) AddRoleToUser(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error {
	// check if the user exists in the db
	exist, err := s.DBStore.UserExistsByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("cheking user in DB store: %w , username: %q ", err, userID)
	}

	if !exist {
		// set service.ErrNotFound as the error
		return fmt.Errorf("user :%q, %w", userID, ErrNotFound)
	}

	// check if the role exists in the db
	exist, err = s.DBStore.RoleExistsByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("cheking role in DB store: %w , roleID: %q ", err, roleID)
	}

	if !exist {
		return fmt.Errorf("role %q does not exist", roleID)
	}

	params := db.AddRoleToUserParams{
		ID:     userID,
		RoleID: HandleNullableUUID(roleID),
	}

	_, err = s.DBStore.AddRoleToUser(ctx, params)
	if err != nil {
		return fmt.Errorf("adding role to user in DB: %w , username: %q ", err, userID)
	}

	return nil
}

// CreateSessionParams are used to create a new session in the database.
type CreateSessionParams struct {
	UserID           uuid.UUID `json:"user_id"`
	RefreshPayloadID uuid.UUID `json:"refresh_payload_id"`
	Username         string    `json:"username"`
	RefreshToken     string    `json:"refresh_token"`
	UserAgent        string    `json:"user_agent"`
	ClientIP         string    `json:"client_ip"`
	IsBlocked        bool      `json:"is_blocked"`
	ExpiresAt        time.Time `json:"expires_at"`
}

// CreateSession creates a new session in the db.
func (s *Stores) CreateSession(ctx context.Context, request CreateSessionParams) error {
	session := db.CreateSessionParams{
		UserID:           request.UserID,
		RefreshPayloadID: request.RefreshPayloadID,
		Username:         request.Username,
		RefreshToken:     request.RefreshToken,
		UserAgent:        request.UserAgent,
		ClientIp:         request.ClientIP,
		IsBlocked:        request.IsBlocked,
		ExpiresAt:        request.ExpiresAt,
	}

	user, err := s.DBStore.GetUser(ctx, request.UserID)
	if err != nil {
		return fmt.Errorf("getting user from DB: %w , username: %q ", err, request.UserID)
	}

	_, err = s.DBStore.CreateSession(ctx, session)
	if err != nil {
		return fmt.Errorf("creating session in DB: %w , username: %q ", err, user.Username)
	}

	return nil
}

// RefreshUserTokenParams are used to refresh a user token in the database.
type RefreshUserTokenParams struct {
	RefreshToken string `json:"refresh_token"`
}

type Session struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	UserID       uuid.UUID `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	ClientIP     string    `json:"client_ip"`
	IsBlocked    bool      `json:"is_blocked"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// GetSession returns a session with a given refresh token.
func (s *Stores) GetSession(ctx context.Context, payloadID uuid.UUID) (Session, error) {
	DBSession, err := s.DBStore.GetSession(ctx, payloadID)
	if err != nil {
		return Session{}, fmt.Errorf("getting session from DB: %w", err)
	}

	session := Session{
		ID:           DBSession.ID,
		UserID:       DBSession.UserID,
		Username:     DBSession.Username,
		RefreshToken: DBSession.RefreshToken,
		UserAgent:    DBSession.UserAgent,
		ClientIP:     DBSession.ClientIp,
		IsBlocked:    DBSession.IsBlocked,
		ExpiresAt:    DBSession.ExpiresAt,
		CreatedAt:    DBSession.CreatedAt,
	}

	return session, nil
}

// LoginUserParams stores the username and password for a user login request.
type LoginUserParams struct {
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Validator Validator `json:"-"`
}

// LoginUser checks if the user exists in the db and if the password matches and returns the user.
func (s *Stores) LoginUser(ctx context.Context, request LoginUserParams) (*User, error) {
	// retrieve the user from the db
	dbAppUser, err := s.DBStore.GetUserByUsername(ctx, request.Username)
	if err != nil {
		return nil, fmt.Errorf("getting user :%w ", err)
	}

	match, err := vault.Matches(request.Password, dbAppUser.Password)
	if err != nil {
		return nil, fmt.Errorf("chaking password: %w", err)
	}

	if !match {
		return nil, fmt.Errorf("%w : invalid credentials", ErrInvalidCredentials)
	}

	appUser := ConvertDBAppUserToUser(dbAppUser)

	return &appUser, nil
}

// InvalidateSession invalidates a session in the db.
func (s *Stores) InvalidateSession(ctx context.Context, payloadID uuid.UUID) error {
	err := s.DBStore.InvalidateSession(ctx, payloadID)
	if err != nil {
		return fmt.Errorf("invalidating session in DB: %w", err)
	}

	return nil
}

// UpdateUserParams are used to update a user in the db.
type UpdateUserParams struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	RoleID    uuid.UUID `json:"role_id"`
	Validator Validator `json:"-"`
}

// UpdateUser takes used ID to update a user in the db.
func (s *Stores) UpdateUser(ctx context.Context, request UpdateUserParams) (User, error) {
	// handle nullable fields for the db
	FirstName := HandleNullableString(request.FirstName)
	LastName := HandleNullableString(request.LastName)
	RoleID := HandleNullableUUID(request.RoleID)

	user := db.UpdateUserParams{
		ID:        request.ID,
		Username:  request.Username,
		FirstName: FirstName,
		LastName:  LastName,
		Email:     request.Email,
		RoleID:    RoleID,
	}

	dbUpdatedUser, err := s.DBStore.UpdateUser(ctx, user)
	if err != nil {
		return User{}, fmt.Errorf("updating user in DB: %w, username: %q ", err, request.ID)
	}

	updatedUser := ConvertDBAppUserToUser(dbUpdatedUser)

	return updatedUser, nil
}

// GetUsers returns all users in the db.
func (s *Stores) GetUsers(ctx context.Context) ([]User, error) {
	dbUsers, err := s.DBStore.GetUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting users from DB: %w", err)
	}

	users := make([]User, 0, len(dbUsers))

	for _, DBUser := range dbUsers {
		user := ConvertDBAppUserToUser(DBUser)
		users = append(users, user)
	}

	return users, nil
}

// GetUsersWithRoles returns all users with a given role in the db.
func (s *Stores) GetUsersWithRoles(ctx context.Context) ([]UserWithRoles, error) {
	DBUsers, err := s.DBStore.GetUsersWithRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting users from DB: %w", err)
	}

	users := make([]UserWithRoles, 0, len(DBUsers))

	for _, DBUser := range DBUsers {
		user := ConvertDBUsersWithRolesToUserWithRoles(DBUser)
		users = append(users, user)
	}

	return users, nil
}

// GetUser returns a user with a given ID.
func (s *Stores) GetUser(ctx context.Context, id uuid.UUID) (User, error) {
	dbUser, err := s.DBStore.GetUser(ctx, id)
	if err != nil {
		return User{}, fmt.Errorf("getting user from DB: %w", err)
	}

	user := ConvertDBAppUserToUser(dbUser)

	return user, nil
}

// DeleteUser deletes a user from the db.
func (s *Stores) DeleteUser(ctx context.Context, id uuid.UUID) error {
	err := s.DBStore.DeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("deleting user from DB: %w", err)
	}

	return nil
}
