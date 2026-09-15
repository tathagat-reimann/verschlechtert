package httpmw

import (
	"context"
	"net/http"

	authmw "verschlechtert/backend/internal/auth"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("verschlechtert/backend")

type correlationIDContextKey struct{}

func CorrelationIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(correlationIDContextKey{}).(string); ok && id != "" {
		return id
	}
	return ""
}

func ContextWithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationIDContextKey{}, id)
}

func CorrelationIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Correlation-ID")
		if id == "" {
			id = uuid.NewString()
		}

		ctx := ContextWithCorrelationID(r.Context(), id)
		r = r.WithContext(ctx)
		w.Header().Set("X-Correlation-ID", id)

		next.ServeHTTP(w, r)
	})
}

func userIDFromContext(ctx context.Context) int64 {
	token, ok := authmw.UserFromContext(ctx)
	if !ok || token == nil {
		return 0
	}

	value, ok := token.Claims["userID"]
	if !ok {
		return 0
	}

	switch v := value.(type) {
	case float64:
		if v > 0 {
			return int64(v)
		}
	case int64:
		if v > 0 {
			return v
		}
	case int:
		if v > 0 {
			return int64(v)
		}
	}

	return 0
}

// OpenTelemetryTracing returns middleware that records request tracing information
func OpenTelemetryTracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// 1. Extract or generate correlation ID
		id := r.Header.Get("X-Correlation-ID")
		if id == "" {
			id = uuid.NewString()
		}

		// 2. Inject correlation ID into context
		ctx := ContextWithCorrelationID(r.Context(), id)

		// 3. Inject user ID into context (your existing logic)
		userID := userIDFromContext(ctx)
		if userID > 0 {
			ctx = context.WithValue(ctx, "user.id", userID)
		}

		// 4. Propagate updated context
		r = r.WithContext(ctx)

		// 5. Set correlation ID header
		w.Header().Set("X-Correlation-ID", id)

		next.ServeHTTP(w, r)
	})
}
