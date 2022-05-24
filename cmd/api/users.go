package api

import (
	"evidence/internal/data"
	"net/http"
	"time"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}
type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}
type loginUserResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
	User                 data.User `json:"user"`
}

func (app *Application) Ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func (app *Application) Login(w http.ResponseWriter, r *http.Request) {
	var req loginUserRequest
	err := app.readJSON(w, r, &req)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// Login
	user, err := app.stores.UserDB.GetByUsername(req.Username)
	if err != nil {
		app.invalidCredentialsResponse(w, r)
		return
	}
	match, err := user.Password.Matches(req.Password)
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	accessToken, accessPayload, err := app.tokenMaker.CreateToken(
		user.Username,
		app.config.AccessTokenDuration)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	rsp := loginUserResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiresAt,
		User:                 *user,
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"database": rsp}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// create user handler
func (app *Application) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest

	err := app.readJSON(w, r, &req)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	user := &data.User{
		Username: req.Username,
	}

	err = user.Password.Set(req.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// create user
	err = app.stores.UserDB.Add(user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
