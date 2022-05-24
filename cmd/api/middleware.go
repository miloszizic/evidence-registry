package api

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

func (app *Application) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			app.invalidAuthorisationHeaderFormat(w, r)
			return
		}
		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			app.invalidAuthorisationHeaderFormat(w, r)
			return
		}
		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			app.invalidAuthorisationHeaderFormat(w, r)
			return
		}
		accessToken := fields[1]
		payload, err := app.tokenMaker.VerifyToken(accessToken)
		if err != nil {
			if err.Error() == "token has expired" {
				app.tokenExpired(w, r)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), authorizationPayloadKey, payload)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func (app *Application) MiddlewarePermissionChecker(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxPayload := r.Context().Value(authorizationPayloadKey)
		payload := ctxPayload.(*Payload)
		user, err := app.stores.UserDB.GetByUsername(payload.Username)
		if err != nil {
			switch {
			case err == sql.ErrNoRows:
				app.unauthorizedUser(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
		}
		if user.Role != "admin" {
			app.unauthorizedUser(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), authorizationPayloadKey, payload)
		next.ServeHTTP(w, r.WithContext(ctx))
	})

}
