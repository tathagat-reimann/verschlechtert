package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	authmw "verschlechtert/backend/internal/auth"
	"verschlechtert/backend/internal/db"
	"verschlechtert/backend/internal/httpmw"
)

func main() {
	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	pool, err := db.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer pool.Close()

	if err := db.ApplyMigrations(ctx, pool); err != nil {
		log.Fatalf("applying database migrations: %v", err)
	}

	firebaseApp, err := firebase.NewApp(ctx, nil)
	if err != nil {
		log.Fatalf("initializing firebase app: %v", err)
	}

	authClient, err := authmw.NewVerifier(ctx, firebaseApp)
	if err != nil {
		log.Fatalf("initializing firebase auth client: %v", err)
	}

	allowedOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:5173"
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{allowedOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	// 1MB request bodies and 10 requests/sec per IP (burst 20) are reasonable defaults for this API.
	r.Use(httpmw.MaxBodyBytes(1 << 20))
	r.Use(httpmw.NewRateLimiter(10, 20).Middleware)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	r.Group(func(r chi.Router) {
		r.Use(authmw.Middleware(authClient))
		getUser := func(r *http.Request) (db.User, bool, error) {
			firebaseUser, ok := authmw.UserFromContext(r.Context())
			if !ok {
				return db.User{}, false, nil
			}
			user, err := db.UpsertUser(
				r.Context(),
				pool,
				firebaseUser.UID,
				claimString(firebaseUser.Claims, "name"),
				claimString(firebaseUser.Claims, "picture"),
				claimString(firebaseUser.Claims, "email"),
			)
			return user, true, err
		}

		r.Get("/api/me", func(w http.ResponseWriter, r *http.Request) {
			user, ok, err := getUser(r)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if err != nil {
				http.Error(w, "could not load user", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(user)
		})

		r.Patch("/api/me", func(w http.ResponseWriter, r *http.Request) {
			_, ok, err := getUser(r)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if err != nil {
				http.Error(w, "could not load user", http.StatusInternalServerError)
				return
			}

			var request struct {
				Locale string `json:"locale"`
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil || !localePattern.MatchString(request.Locale) {
				http.Error(w, "invalid locale", http.StatusBadRequest)
				return
			}
			firebaseUser, _ := authmw.UserFromContext(r.Context())
			updatedUser, err := db.UpdateUserLocale(r.Context(), pool, firebaseUser.UID, request.Locale)
			if err != nil {
				http.Error(w, "could not update user locale", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(updatedUser)
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
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
	}
}

var localePattern = regexp.MustCompile(`^[a-z]{2}(-[A-Z]{2})?$`)

func claimString(claims map[string]interface{}, key string) string {
	value, ok := claims[key].(string)
	if !ok {
		return ""
	}
	return value
}
