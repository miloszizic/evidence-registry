package api

import (
	"context"
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
			if err == ErrExpiredToken {
				app.tokenExpired(w, r)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
		}
		ctx := context.WithValue(r.Context(), authorizationPayloadKey, payload)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
