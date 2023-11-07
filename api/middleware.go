package api

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/miloszizic/der/service"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type (
	authorization string
	contextKey    string
)

var (
	userContextKey          contextKey    = "user"
	authorizationHeaderKey  authorization = "authorization"
	authorizationPayloadKey authorization = "authorization_payload"
)

// AuthMiddleware is a middleware that checks if the request is authorized. It checks the Authorization header
// and verifies the access token. If the token is valid, it adds the payload to the request context.
func (app *Application) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const maxAuthorizationHeaderFields = 2

		authorizationHeader := r.Header.Get(string(authorizationHeaderKey))
		fmt.Println("Authorization Header:", authorizationHeader)
		if len(authorizationHeader) == 0 {
			app.invalidAuthorisationHeaderFormat(w, r)
			return
		}
		fields := strings.Fields(authorizationHeader)
		if len(fields) < maxAuthorizationHeaderFields {
			app.invalidAuthorisationHeaderFormat(w, r)
			return
		}
		authorizationType := strings.ToLower(fields[0])
		if authorizationType != "bearer" {
			app.invalidAuthorisationHeaderFormat(w, r)
			return
		}
		accessToken := fields[1]
		payload, err := app.tokenMaker.VerifyToken(accessToken)
		if err != nil {
			if err != nil {
				fmt.Println("Token Verification Error:", err) // Debug print
			} else {
				fmt.Println("Token Payload:", payload) // Debug print
			}
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

// MiddlewarePermissionChecker is a middleware that checks for a specific permission required to access a route.
func (app *Application) MiddlewarePermissionChecker(permission string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxPayload := r.Context().Value(authorizationPayloadKey)
			payload, ok := ctxPayload.(*service.Payload)
			if !ok {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			user, err := app.stores.DBStore.GetUserByUsername(r.Context(), payload.Username)
			if err != nil {
				app.respondError(w, r, err)
				return
			}

			if user.RoleID.Valid {
				// Get permissions for the user's role
				permissions, err := app.stores.DBStore.GetPermissionsForRole(r.Context(), user.RoleID.UUID)
				if err != nil {
					app.respondError(w, r, err)
					return
				}

				// Check if the required permission is in the list of permissions
				for _, p := range permissions {
					// fmt.Println("Permission:", p)
					if p == permission {
						// fmt.Println("Permission Found:", p)
						ctx := context.WithValue(r.Context(), authorizationPayloadKey, payload)
						// fmt.Println("Context:", ctx)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}
			// fmt.Println("Permission Not Found:", permission)

			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		})
	}
}

// UserParserMiddleware is a middleware that parses the user from the request. It checks the Authorization header
// and verifies the access token. If the token is valid, it adds the user to the request context.
func (app *Application) UserParserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := app.requestUserParser(r)
		if err != nil {
			app.logger.Errorw("Error parsing user from request", "error", err)
			app.respondError(w, r, err)
			return
		}
		// Store the user in the context
		ctx := context.WithValue(r.Context(), userContextKey, user)

		// Call the next handler with the updated context
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
				elapsed := time.Since(t1)
				userInfo := "anonymous"

				if user, ok := r.Context().Value(userContextKey).(*service.User); ok {
					userInfo = user.Username
				}

				s.With(
					"request_id", middleware.GetReqID(r.Context()),
					"method", r.Method,
					"path", r.URL.Path,
					"protocol", r.Proto,
					"remote_addr", r.RemoteAddr,
					"status", statusLabel(ww.Status()),
					"bytes_written", ww.BytesWritten(),
					"elapsed", elapsed,
					"user_agent", r.UserAgent(),
					"referrer", r.Referer(),
					"user", userInfo,
				).Info("Request completed")
			}()
			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}

func statusLabel(status int) string {
	switch {
	case status >= 100 && status < 200:
		return fmt.Sprintf("%d OK", status)
	case status >= 200 && status < 300:
		return fmt.Sprintf("%d OK", status)
	case status >= 300 && status < 400:
		return fmt.Sprintf("%d Redirect", status)
	case status >= 400 && status < 500:
		return fmt.Sprintf("%d Client Error", status)
	case status >= 500 && status < 600:
		return fmt.Sprintf("%d Server Error", status)
	default:
		return fmt.Sprintf("%d Unknown", status)
	}
}

// recoverPanic is a middleware function that catches any panic that might occur during the execution of an HTTP request.
// It wraps an HTTP handler and returns a new handler that serves the same purpose but with panic recovery.
// If a panic occurs, it recovers from it, logs the error, and responds to the client with an error message.
func (app *Application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rv := recover(); rv != nil {
				app.logger.Info("Recovering from panic")
				stackTrace := debug.Stack()
				app.logger.Errorf("Stack trace: %s", stackTrace)
				var err error
				switch v := rv.(type) {
				case error:
					err = v
				default:
					err = fmt.Errorf("unexpected panic value: %v", rv)
				}
				app.serverErrorResponse(w, r, err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
