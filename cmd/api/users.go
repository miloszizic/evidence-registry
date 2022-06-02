package api

import (
	"evidence/internal/data"
	"net/http"
	"time"
)

type loginUserResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
	User                 data.User `json:"user"`
}

func (app *Application) Ping(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("pong"))
}

func (app *Application) Login(w http.ResponseWriter, r *http.Request) {
	request, err := app.userParser(w, r)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	response, err := app.LoginUser(request)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	app.respondLogin(w, r, response)

}

func (app *Application) LoginUser(request *data.UserRequest) (*loginUserResponse, error) {
	if request.Username == "" || request.Password == "" {
		return nil, data.NewErrorf(data.ErrCodeInvalidCredentials, "username and password are required")
	}
	user, err := app.stores.User.GetByUsername(request.Username)
	if err != nil {
		return nil, err
	}
	match, err := user.Password.Matches(request.Password)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, data.NewErrorf(data.ErrCodeInvalidCredentials, "invalid credentials")
	}
	accessToken, accessPayload, err := app.tokenMaker.CreateToken(
		user.Username,
		app.config.AccessTokenDuration)
	if err != nil {
		return nil, err
	}
	rsp := loginUserResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiresAt,
		User:                 *user,
	}
	return &rsp, nil
}

func (app *Application) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	request, err := app.userParser(w, r)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	err = app.stores.CreateUser(request)
	if err != nil {
		app.respondError(w, r, err)
		return
	}
	app.respondUser(w, r)

}

func (app *Application) respondUser(w http.ResponseWriter, r *http.Request) {
	err := app.writeJSON(w, http.StatusCreated, envelope{"User": "user successfully created"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
func (app *Application) respondLogin(w http.ResponseWriter, r *http.Request, response *loginUserResponse) {
	err := app.writeJSON(w, http.StatusOK, envelope{"Login": response}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) userParser(w http.ResponseWriter, r *http.Request) (*data.UserRequest, error) {
	var req data.UserRequest
	err := app.readJSON(r, &req)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return nil, err
	}
	return &req, nil
}
