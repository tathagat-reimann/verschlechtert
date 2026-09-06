package authmw

import (
	"context"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
)

type contextKey string

const userContextKey contextKey = "firebaseUser"

// NewVerifier creates a Firebase Auth client used to verify ID tokens.
func NewVerifier(ctx context.Context, app *firebase.App) (*auth.Client, error) {
	return app.Auth(ctx)
}

// Middleware verifies the Firebase ID token sent in the Authorization header
// and stores the decoded token in the request context.
func Middleware(client *auth.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}
			idToken := strings.TrimPrefix(header, "Bearer ")

			token, err := client.VerifyIDToken(r.Context(), idToken)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, token)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserFromContext returns the decoded Firebase token stored by Middleware.
func UserFromContext(ctx context.Context) (*auth.Token, bool) {
	token, ok := ctx.Value(userContextKey).(*auth.Token)
	return token, ok
}
