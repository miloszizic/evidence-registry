package api

import (
	"github.com/miloszizic/der/internal/data"
	"net/http"
	"time"
)

func (*Application) Ping(w http.ResponseWriter, _ *http.Request) {
	_, err := w.Write([]byte("pong"))
	if err != nil {
		return
	}
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
	app.respond(w, r, http.StatusCreated, envelope{"User": "successfully created"})

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
	app.respond(w, r, http.StatusOK, envelope{"Login": response})

}

type LoginUserResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
	User                 data.User `json:"user"`
}

func (app *Application) LoginUser(request *data.UserRequest) (*LoginUserResponse, error) {
	if request.Username == "" || request.Password == "" {
		return nil, data.NewErrorf(data.ErrCodeInvalidCredentials, "username and password are required")
	}
	user, err := app.stores.User.GetByUsername(request.Username)
	if err != nil {
		return nil, data.WrapErrorf(err, data.ErrCodeUnknown, "User.GetByUsername")
	}
	match, err := user.Password.Matches(request.Password)
	if err != nil {
		return nil, data.WrapErrorf(err, data.ErrCodeUnknown, "Password.Matches")
	}
	if !match {
		return nil, data.NewErrorf(data.ErrCodeInvalidCredentials, "invalid credentials")
	}
	accessToken, accessPayload, err := app.tokenMaker.CreateToken(
		user.Username,
		app.config.AccessTokenDuration)
	if err != nil {
		return nil, data.WrapErrorf(err, data.ErrCodeUnknown, "tokenMaker.CreateToken")
	}
	rsp := LoginUserResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiresAt,
		User:                 *user,
	}
	return &rsp, nil
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
