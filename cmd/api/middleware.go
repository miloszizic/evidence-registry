package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/miloszizic/der/internal/data"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

type authorization string

var (
	authorizationHeaderKey  authorization = "authorization"
	authorizationPayloadKey authorization = "authorization_payload"
	sugaredLogFormat                      = `[%s] "%s %s %s" from %s - %s %dB in %s`
)

func (app *Application) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get(string(authorizationHeaderKey))
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
		//if authorizationType != string(authorizationTypeBearer) {
		if authorizationType != "bearer" {
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
		var verr *data.Error
		ctxPayload := r.Context().Value(authorizationPayloadKey)
		payload := ctxPayload.(*Payload)
		user, err := app.stores.User.GetByUsername(payload.Username)
		if err != nil {
			switch {
			case errors.As(err, &verr) && verr.Code() == data.ErrCodeNotFound:
				app.unauthorizedUser(w, r)
				return
			default:
				app.respondError(w, r, err)
				return
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

// Logger is a middleware that logs the start and end of each request, along
// with some useful data about what was requested, what the response status was,
// and how long it took to return.
func Logger(s *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()
			defer func() {
				s.Infof(sugaredLogFormat,
					middleware.GetReqID(r.Context()), // RequestID (if set)
					r.Method,                         // Method
					r.URL.Path,                       // Path
					r.Proto,                          // Protocol
					r.RemoteAddr,                     // RemoteAddr
					statusLabel(ww.Status()),         // "200 OK"
					ww.BytesWritten(),                // Bytes Written
					time.Since(t1),                   // Elapsed
				)
			}()
			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
func statusLabel(status int) string {
	switch {
	case status >= 100 && status < 300:
		return fmt.Sprintf("%d OK", status)
	case status >= 300 && status < 400:
		return fmt.Sprintf("%d Redirect", status)
	case status >= 400 && status < 500:
		return fmt.Sprintf("%d Client Error", status)
	case status >= 500:
		return fmt.Sprintf("%d Server Error", status)
	default:
		return fmt.Sprintf("%d Unknown", status)
	}
}
