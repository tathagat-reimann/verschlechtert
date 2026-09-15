package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/jackc/pgx/v5/pgxpool"

	authmw "verschlechtert/backend/internal/auth"
	"verschlechtert/backend/internal/config"
	"verschlechtert/backend/internal/report"
	"verschlechtert/backend/internal/router"
	"verschlechtert/backend/internal/tracing"
	"verschlechtert/backend/internal/user"

	"github.com/exaring/otelpgx"
)

// NewPool creates a Postgres connection pool from a DATABASE_URL connection string.
func createPool(ctx context.Context, databaseURL, appEnvironment string) (*pgxpool.Pool, error) {
	// 1. Parse config instead of using pgxpool.New
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	// 2. Attach OpenTelemetry tracer
	tracerOptions := make([]otelpgx.Option, 0, 1)
	if appEnvironment != "production" {
		tracerOptions = append(tracerOptions, otelpgx.WithIncludeQueryParameters())
	}
	config.ConnConfig.Tracer = otelpgx.NewTracer(tracerOptions...)

	// 3. Create pool with config
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("creating pgx pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}
	if err := otelpgx.RecordStats(pool); err != nil {
		return nil, fmt.Errorf("unable to record database stats: %w", err)
	}
	return pool, nil
}

func main() {
	ctx := context.Background()

	cfg := config.Load()

	// Debug-level logging (API/DB call results) is only emitted outside production;
	// set APP_ENV=production to silence it.
	logLevel := slog.LevelDebug
	if cfg.AppEnvironment == "production" {
		logLevel = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))

	// Initialize OpenTelemetry Jaeger exporter
	// Uses OTEL_EXPORTER_OTLP_ENDPOINT env var (default: 127.0.0.1:4318 for HTTP)
	shutdownTracer := tracing.InitJaegerExporter(ctx)
	defer shutdownTracer(ctx)

	databaseURL := cfg.DatabaseURL

	pool, err := createPool(ctx, databaseURL, cfg.AppEnvironment)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer pool.Close()

	firebaseApp, err := firebase.NewApp(ctx, nil)
	if err != nil {
		log.Fatalf("initializing firebase app: %v", err)
	}

	authClient, err := authmw.NewVerifier(ctx, firebaseApp)
	if err != nil {
		log.Fatalf("initializing firebase auth client: %v", err)
	}

	userRepo := user.NewUserRepository(pool)
	userService := user.NewUserService(userRepo)
	userHandler := user.NewUserHandler(userService, authmw.UserFromContext, authClient)

	reportRepo := report.NewReportRepository(pool)
	reportService := report.NewReportService(reportRepo)
	reportHandler := report.NewReportHandler(reportService, userService, authmw.UserFromContext, authClient)

	rt := router.NewRouter(userHandler, reportHandler, cfg.AllowedOrigins, authClient)
	handler := rt.Setup()

	port := cfg.Port

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("listening on :%s", port)
		serverErr <- srv.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	case sig := <-stop:
		log.Printf("received %s, shutting down", sig)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("error during shutdown: %v", err)
		}

		if err := shutdownTracer(ctx); err != nil {
			log.Printf("error shutting down tracer: %v", err)
		}
	}
}

// func writeJSON(w http.ResponseWriter, value interface{}) {
// 	w.Header().Set("Content-Type", "application/json")
// 	_ = json.NewEncoder(w).Encode(value)
// }

// var localePattern = regexp.MustCompile(`^[a-z]{2}(-[A-Z]{2})?$`)

// // requestLocale returns the ?locale= query param, defaulting to German if absent or unsupported.
// func requestLocale(r *http.Request) string {
// 	if locale := r.URL.Query().Get("locale"); locale == "en" {
// 		return "en"
// 	}
// 	return "de"
// }

// func claimString(claims map[string]interface{}, key string) string {
// 	value, ok := claims[key].(string)
// 	if !ok {
// 		return ""
// 	}
// 	return value
// }
