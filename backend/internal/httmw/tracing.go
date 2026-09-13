package httpmw

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

// OpenTelemetryTracing returns middleware that records request tracing information
func OpenTelemetryTracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attrs := []attribute.KeyValue{
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
			attribute.String("http.target", r.URL.Path),
			attribute.String("http.host", r.Host),
			attribute.String("http.scheme", r.URL.Scheme),
			attribute.String("http.user_agent", r.UserAgent()),
			attribute.String("http.client_ip", r.RemoteAddr),
		}
		if id := CorrelationIDFromContext(r.Context()); id != "" {
			attrs = append(attrs, attribute.String("correlation.id", id))
		}

		ctx, span := tracer.Start(r.Context(), r.Method+" "+r.URL.Path,
			trace.WithAttributes(attrs...),
		)
		defer span.End()

		// Wrap response writer to capture status code
		wrapped := &statusCodeWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r.WithContext(ctx))

		// Record response status
		span.SetAttributes(
			attribute.Int("http.status_code", wrapped.statusCode),
		)
	})
}
