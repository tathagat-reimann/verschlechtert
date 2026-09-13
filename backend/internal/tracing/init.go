package tracing

import (
	"context"
	"log"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// InitConsoleExporter initializes OpenTelemetry with console exporter
// Returns a shutdown function that should be called on application shutdown
func InitConsoleExporter(ctx context.Context) func(context.Context) error {
	exporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		log.Fatalf("failed to create console exporter: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)

	return tp.Shutdown
}

// InitJaegerExporter initializes OpenTelemetry with Jaeger OTLP exporter
// Uses OTEL_EXPORTER_OTLP_ENDPOINT environment variable (default: 127.0.0.1:4318)
// Note: Jaeger all-in-one exposes OTLP HTTP on 4318; 4317 is for gRPC and 14268 is the Jaeger collector API.
// Returns a shutdown function that should be called on application shutdown
func InitJaegerExporter(ctx context.Context) func(context.Context) error {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		slog.Warn("OTEL_EXPORTER_OTLP_ENDPOINT is not set, jaeger is not running!")
	}

	exporter, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(), // Remove in production with proper TLS
	)
	if err != nil {
		log.Fatalf("failed to create Jaeger exporter: %v", err)
	}

	// Create resource with service name (visible in Jaeger UI)
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("verschlechtert-backend"),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	return tp.Shutdown
}
