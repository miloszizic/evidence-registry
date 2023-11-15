package api

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/miloszizic/der/service"
)

// CreateUserHandler is validating and creates a new user.
func (app *Application) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	// minimum number of characters required for a password
	const passwordMinLength = 8
	// maximum number of characters required for a password
	const passwordMaxLength = 72

	params, err := paramsParser[service.CreateUserParams](app, r)
	if err != nil {
		app.logger.Errorw("Error parsing params from request", "error", err)
		app.respondError(w, r, err)

		return
	}

	exists, err := app.stores.DBStore.UserExists(r.Context(), params.Username)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	// expectedUser validation
	params.Validator.CheckField(params.Username != "", "Username", "Username is required")
	params.Validator.CheckField(!exists, "Username", "Username is already in use")
	// email validation
	params.Validator.CheckField(params.Email != "", "Email", "Email is required")
	// password validation
	params.Validator.CheckField(params.Password != "", "Password", "Password is required")
	params.Validator.CheckField(len(params.Password) >= passwordMinLength, "Password", "Password is too short")
	params.Validator.CheckField(len(params.Password) <= passwordMaxLength, "Password", "Password is too long")
	params.Validator.CheckField(NotIn(params.Password, service.CommonPasswords...), "Password", "Password is too common")

	// return any error received while doing validation
	if params.Validator.HasErrors() {
		app.failedValidation(w, r, params.Validator)
		return
	}

	_, err = app.stores.CreateUser(r.Context(), params)
	if err != nil {
		app.respondError(w, r, err)
		return
	}

	app.respond(w, r, http.StatusCreated, envelope{"User": "successfully created"})
}

// UpdateUserPasswordHandler updates a user password in the database
func (app *Application) UpdateUserPasswordHandler(w http.ResponseWriter, r *http.Request) {
	// minimum number of characters required for a password
	const passwordMinLength = 8
	// maximum number of characters required for a password
	const passwordMaxLength = 72

	userID, err := userIDParser(r)
	if err != nil {
		app.logger.Errorw("Error parsing user ID from request", "error", err)
		app.respondError(w, r, err)

		return
	}

	params, err := paramsParser[service.UpdateUserPasswordParams](app, r)
	if err != nil {
		app.logger.Errorw("Error parsing params from request", "error", err)
		app.respondError(w, r, err)

		return
	}
	// expectedUser validation
	params.Validator.CheckField(params.Password != "", "Password", "Password is required")
	// password validation
	params.Validator.CheckField(len(params.Password) >= passwordMinLength, "Password", "Password is too short")
	params.Validator.CheckField(len(params.Password) <= passwordMaxLength, "Password", "Password is too long")
	params.Validator.CheckField(NotIn(params.Password, service.CommonPasswords...), "Password", "Password is too common")

	// return any error received while doing validation
	if params.Validator.HasErrors() {
		app.failedValidation(w, r, params.Validator)
		return
	}

	// get user id from db to update user
	user, err := app.stores.DBStore.GetUserByID(r.Context(), userID)
	if err != nil {
		app.logger.Errorw("Error getting user by ID", "error", err)
		app.respondError(w, r, err)

		return
	}

	params.ID = user.ID

	err = app.stores.UpdateUserPassword(r.Context(), params)
	if err != nil {
		app.logger.Errorw("Error updating user", "error", err)
		app.respondError(w, r, err)

		return
	}

	app.respond(w, r, http.StatusOK, envelope{"User": "successfully updated password"})
}

// AddRoleToUserHandler adds a role to a user in the database
func (app *Application) AddRoleToUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDParser(r)
	if err != nil {
		app.logger.Errorw("Error parsing user ID from request", "error", err)
		app.respondError(w, r, err)

		return
	}

	roleID, err := roleIDParser(r)
	if err != nil {
		app.logger.Errorw("Error parsing role ID from request", "error", err)
		app.respondError(w, r, err)

		return
	}

	err = app.stores.AddRoleToUser(r.Context(), userID, roleID)
	if err != nil {
		app.logger.Errorw("Error adding role to user", "error", err)
		app.respondError(w, r, err)

		return
	}

	app.respond(w, r, http.StatusOK, envelope{"User": "successfully added role to user"})
}

// UpdateUserHandler updates a user in the database
func (app *Application) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDParser(r)
	if err != nil {
		app.logger.Errorw("Error parsing user ID from request", "error", err)
		app.respondError(w, r, err)

		return
	}

	params, err := paramsParser[service.UpdateUserParams](app, r)
	if err != nil {
		app.logger.Errorw("Error parsing params from request", "error", err)
		app.respondError(w, r, err)

		return
	}
	// expectedUser validation
	params.Validator.CheckField(params.Username != "", "Username", "Username is required")
	// email validation
	params.Validator.CheckField(params.Email != "", "Email", "Email is required")

	// return any error received while doing validation
	if params.Validator.HasErrors() {
		app.failedValidation(w, r, params.Validator)
		return
	}

	// get user id from db to update user
	user, err := app.stores.DBStore.GetUserByID(r.Context(), userID)
	if err != nil {
		app.logger.Errorw("Error getting user by ID", "error", err)
		app.respondError(w, r, err)

		return
	}

	params.ID = user.ID

	_, err = app.stores.UpdateUser(r.Context(), params)
	if err != nil {
		app.logger.Errorw("Error updating user", "error", err)
		app.respondError(w, r, err)

		return
	}

	app.respond(w, r, http.StatusOK, envelope{"User": "successfully updated"})
}

// LoginUserResponse holds the parameters expected for a user login response.
type LoginUserResponse struct {
	AccessToken           string       `json:"access_token"`
	RefreshToken          string       `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  service.User `json:"user"`
}

// UserLoginHandler takes a username and password and returns a JWT token
func (app *Application) UserLoginHandler(w http.ResponseWriter, r *http.Request) {
	params, err := paramsParser[service.LoginUserParams](app, r)
	if err != nil {
		app.logger.Errorw("Error parsing params from request", "error", err)
		app.respondError(w, r, err)

		return
	}
	// expectedUser validation
	params.Validator.CheckField(params.Username != "", "Username", "Username is required")
	// password validation
	params.Validator.CheckField(params.Password != "", "Password", "Password is required")
	// return any error received while doing validation
	if params.Validator.HasErrors() {
		app.failedValidation(w, r, params.Validator)

		return
	}

	exists, err := app.stores.DBStore.UserExists(r.Context(), params.Username)
	if err != nil {
		app.logger.Errorw("Error checking if user exists", "error", err)
		app.respondError(w, r, err)

		return
	}

	if !exists {
		app.logger.Errorw("Error user does not exist", "error", err)
		app.respondError(w, r, service.ErrNotFound)

		return
	}

	user, err := app.stores.LoginUser(context.Background(), params)
	if err != nil {
		app.logger.Errorw("Error logging in user", "error", err)
		app.respondError(w, r, err)

		return
	}

	accessToken, accessPayload, err := app.tokenMaker.CreateToken(
		user.Username,
		app.config.AccessTokenDuration)
	if err != nil {
		app.logger.Errorw("Error creating access token", "error", err)
		app.serverErrorResponse(w, r, err)

		return
	}

	refreshToken, refreshPayload, err := app.tokenMaker.CreateRefreshToken(
		user.Username,
		app.config.RefreshTokenDuration)
	if err != nil {
		app.logger.Errorw("Error creating refresh token", "error", err)
		app.serverErrorResponse(w, r, err)

		return
	}

	rsp := LoginUserResponse{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiresAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiresAt,
		User:                  *user,
	}

	clientIP := r.RemoteAddr
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		clientIP = fwd
	}

	userAgent := r.UserAgent()
	err = app.stores.CreateSession(r.Context(), service.CreateSessionParams{
		UserID:           user.ID,
		RefreshPayloadID: refreshPayload.ID,
		Username:         user.Username,
		RefreshToken:     refreshToken,
		UserAgent:        userAgent,
		ClientIP:         clientIP,
		IsBlocked:        false,
		ExpiresAt:        refreshPayload.ExpiresAt,
	})

	if err != nil {
		app.logger.Errorw("Error creating session", "error", err)
		app.respondError(w, r, err)

		return
	}

	app.respond(w, r, http.StatusOK, envelope{"UserLogin": rsp})
}

// RefreshTokenHandler takes a refresh token and returns a new token and refresh token
func (app *Application) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	// verify refresh token
	params, err := paramsParser[service.RefreshUserTokenParams](app, r)
	if err != nil {
		app.logger.Errorw("Error parsing params from request", "error", err)
		app.badRequestResponse(w, r, err)

		return
	}
	// verify refresh token
	refreshPayload, err := app.tokenMaker.VerifyRefreshToken(params.RefreshToken)
	if err != nil {
		app.logger.Errorw("Error verifying refresh token", "error", err)
		app.invalidAuthorisationHeaderFormat(w, r)

		return
	}

	// get session from db
	session, err := app.stores.GetSession(r.Context(), refreshPayload.ID)
	if err != nil {
		app.logger.Errorw("Error getting session", "error", err)
		app.respondError(w, r, err)

		return
	}
	// check if session is blocked
	if session.IsBlocked {
		app.logger.Errorw("Error session is blocked", "error", err)
		app.respondError(w, r, service.ErrUnauthorized)

		return
	}
	// check if the session is expired
	if time.Now().After(session.ExpiresAt) {
		app.logger.Errorw("Error session is expired", "error", err)
		app.respondError(w, r, service.ErrUnauthorized)

		return
	}
	// check if the session refresh token matches the refresh token
	if session.RefreshToken != params.RefreshToken {
		app.logger.Errorw("Error session refresh token does not match refresh token", "error", err)
		app.respondError(w, r, service.ErrUnauthorized)

		return
	}

	// Create a new access token for the user
	accessToken, accessPayload, err := app.tokenMaker.CreateToken(
		session.Username,
		app.config.AccessTokenDuration)
	if err != nil {
		app.logger.Errorw("Error creating access token", "error", err)
		app.serverErrorResponse(w, r, err)

		return
	}

	// Create a new refresh token for the user
	newRefreshToken, newRefreshPayload, err := app.tokenMaker.CreateRefreshToken(
		session.Username,
		app.config.RefreshTokenDuration)
	if err != nil {
		app.logger.Errorw("Error creating new refresh token", "error", err)
		app.serverErrorResponse(w, r, err)

		return
	}

	// Create a new session in the database with the new refresh token and expiration date
	clientIP := r.RemoteAddr
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		clientIP = fwd
	}

	userAgent := r.UserAgent()
	err = app.stores.CreateSession(r.Context(), service.CreateSessionParams{
		UserID:           session.UserID,
		RefreshPayloadID: newRefreshPayload.ID,
		Username:         session.Username,
		RefreshToken:     newRefreshToken,
		UserAgent:        userAgent,
		ClientIP:         clientIP,
		IsBlocked:        false,
		ExpiresAt:        newRefreshPayload.ExpiresAt,
	})

	if err != nil {
		app.logger.Errorw("Error creating session", "error", err)
		app.respondError(w, r, err)

		return
	}

	// invalidate the old session (for added security)
	err = app.stores.InvalidateSession(r.Context(), session.ID)
	if err != nil {
		app.logger.Errorw("Error invalidating old session", "error", err)
		app.respondError(w, r, err)

		return
	}

	// Return the new tokens to the client
	rsp := LoginUserResponse{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiresAt,
		RefreshToken:          newRefreshToken,
		RefreshTokenExpiresAt: newRefreshPayload.ExpiresAt,
	}

	app.respond(w, r, http.StatusOK, envelope{"Token Refreshed": rsp})
}

// ListUsersHandler returns all users
func (app *Application) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := app.stores.GetUsers(r.Context())
	if err != nil {
		app.logger.Errorw("Error getting users", "error", err)
		app.respondError(w, r, err)

		return
	}

	app.respond(w, r, http.StatusOK, envelope{"Users": users})
}

// GetUsersWithRoleHandler returns all users
func (app *Application) GetUsersWithRoleHandler(w http.ResponseWriter, r *http.Request) {
	users, err := app.stores.GetUsersWithRoles(r.Context())
	if err != nil {
		app.logger.Errorw("Error getting users", "error", err)
		app.respondError(w, r, err)

		return
	}

	app.respond(w, r, http.StatusOK, envelope{"Users": users})
}

// GetUserHandler takes a id and returns a user
func (app *Application) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := userIDParser(r)
	if err != nil {
		app.logger.Errorw("Error parsing user ID from request", "error", err)
		app.respondError(w, r, err)

		return
	}

	user, err := app.stores.GetUser(r.Context(), id)
	if err != nil {
		app.respondError(w, r, err)

		return
	}

	app.respond(w, r, http.StatusOK, envelope{"User": user})
}

// DeleteUserParams holds the parameters expected for a user delete request.
type DeleteUserParams struct {
	ID uuid.UUID `json:"id"`
}

// DeleteUserHandler deletes a user from the database
func (app *Application) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	// parse params from request
	input, err := paramsParser[DeleteUserParams](app, r)
	if err != nil {
		app.logger.Errorw("Error parsing params from request", "error", err)
		app.respondError(w, r, err)

		return
	}
	// check if user exists
	_, err = app.stores.GetUser(r.Context(), input.ID)
	if err != nil {
		app.respondError(w, r, err)
		return
	}

	err = app.stores.DeleteUser(r.Context(), input.ID)
	if err != nil {
		app.logger.Errorw("Error deleting user", "error", err)
		app.respondError(w, r, err)

		return
	}

	app.respond(w, r, http.StatusOK, envelope{"User": "successfully deleted"})
}
