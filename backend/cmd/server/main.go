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
	"verschlechtert/backend/internal/user"
)

// NewPool creates a Postgres connection pool from a DATABASE_URL connection string.
func createPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("creating pgx pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
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

	databaseURL := cfg.DatabaseURL

	pool, err := createPool(ctx, databaseURL)
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
	userHandler := user.NewUserHandler(userService, authmw.UserFromContext)

	reportRepo := report.NewReportRepository(pool)
	reportService := report.NewReportService(reportRepo)
	reportHandler := report.NewReportHandler(reportService, userService, authmw.UserFromContext)

	rt := router.NewRouter(userHandler, reportHandler, cfg.AllowedOrigins, authClient)
	handler := rt.Setup()

	// r.Group(func(r chi.Router) {
	// 	r.Use(authmw.Middleware(authClient))
	// 	getUser := func(r *http.Request) (db.User, bool, error) {
	// 		firebaseUser, ok := authmw.UserFromContext(r.Context())
	// 		if !ok {
	// 			return db.User{}, false, nil
	// 		}
	// 		user, err := db.UpsertUser(
	// 			r.Context(),
	// 			pool,
	// 			firebaseUser.UID,
	// 			claimString(firebaseUser.Claims, "name"),
	// 			claimString(firebaseUser.Claims, "picture"),
	// 			claimString(firebaseUser.Claims, "email"),
	// 		)
	// 		if !user.Active {
	// 			return db.User{}, false, errors.New("user is inactive")
	// 		}
	// 		return user, true, err
	// 	}
	// 	getCurrentUser := func(r *http.Request) (db.User, bool, error) {
	// 		user, ok, err := getUser(r)
	// 		return user, ok, err
	// 	}

	// 	r.Get("/api/reports", reportHandler.ListActive)

	// 	r.Get("/api/reports/{id}", func(w http.ResponseWriter, r *http.Request) {
	// 		user, ok, err := getCurrentUser(r)
	// 		if !ok {
	// 			http.Error(w, "unauthorized", http.StatusUnauthorized)
	// 			return
	// 		}
	// 		if err != nil {
	// 			http.Error(w, "could not load user", http.StatusInternalServerError)
	// 			return
	// 		}
	// 		reportID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	// 		if err != nil {
	// 			http.Error(w, "invalid report id", http.StatusBadRequest)
	// 			return
	// 		}
	// 		detail, err := db.GetReportDetail(r.Context(), pool, reportID, user.ID, requestLocale(r))
	// 		if err != nil {
	// 			http.Error(w, "report not found", http.StatusNotFound)
	// 			return
	// 		}
	// 		writeJSON(w, detail)
	// 	})

	// 	r.Post("/api/reports/{id}/comments", func(w http.ResponseWriter, r *http.Request) {
	// 		user, ok, err := getCurrentUser(r)
	// 		if !ok {
	// 			http.Error(w, "unauthorized", http.StatusUnauthorized)
	// 			return
	// 		}
	// 		if err != nil {
	// 			http.Error(w, "could not load user", http.StatusInternalServerError)
	// 			return
	// 		}
	// 		reportID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	// 		if err != nil {
	// 			http.Error(w, "invalid report id", http.StatusBadRequest)
	// 			return
	// 		}
	// 		var request struct {
	// 			Body string `json:"body"`
	// 		}
	// 		if err := json.NewDecoder(r.Body).Decode(&request); err != nil || strings.TrimSpace(request.Body) == "" {
	// 			http.Error(w, "invalid comment", http.StatusBadRequest)
	// 			return
	// 		}
	// 		comment, err := db.AddReportComment(r.Context(), pool, reportID, user.ID, request.Body)
	// 		if err != nil {
	// 			http.Error(w, "could not add comment", http.StatusInternalServerError)
	// 			return
	// 		}
	// 		w.WriteHeader(http.StatusCreated)
	// 		writeJSON(w, comment)
	// 	})

	// 	r.Post("/api/reports/{id}/alternatives", func(w http.ResponseWriter, r *http.Request) {
	// 		user, ok, err := getCurrentUser(r)
	// 		if !ok {
	// 			http.Error(w, "unauthorized", http.StatusUnauthorized)
	// 			return
	// 		}
	// 		if err != nil {
	// 			http.Error(w, "could not load user", http.StatusInternalServerError)
	// 			return
	// 		}
	// 		reportID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	// 		if err != nil {
	// 			http.Error(w, "invalid report id", http.StatusBadRequest)
	// 			return
	// 		}
	// 		var request struct {
	// 			BrandID     int64  `json:"brandId"`
	// 			SellerID    int64  `json:"sellerId"`
	// 			ProductName string `json:"productName"`
	// 			ProductURL  string `json:"productUrl"`
	// 		}
	// 		if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.BrandID <= 0 || request.SellerID <= 0 || strings.TrimSpace(request.ProductName) == "" {
	// 			http.Error(w, "invalid alternative", http.StatusBadRequest)
	// 			return
	// 		}
	// 		alternative, err := db.AddReportAlternative(r.Context(), pool, reportID, user.ID, request.BrandID, request.SellerID, request.ProductName, request.ProductURL)
	// 		if err != nil {
	// 			if errors.Is(err, db.ErrAlternativeAlreadySuggested) {
	// 				http.Error(w, "you already suggested an alternative for this report", http.StatusConflict)
	// 				return
	// 			}
	// 			http.Error(w, "could not add alternative", http.StatusInternalServerError)
	// 			return
	// 		}
	// 		w.WriteHeader(http.StatusCreated)
	// 		writeJSON(w, alternative)
	// 	})

	// 	r.Post("/api/reports/{id}/like", func(w http.ResponseWriter, r *http.Request) {
	// 		user, ok, err := getCurrentUser(r)
	// 		if !ok {
	// 			http.Error(w, "unauthorized", http.StatusUnauthorized)
	// 			return
	// 		}
	// 		if err != nil {
	// 			http.Error(w, "could not load user", http.StatusInternalServerError)
	// 			return
	// 		}
	// 		reportID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	// 		if err != nil {
	// 			http.Error(w, "invalid report id", http.StatusBadRequest)
	// 			return
	// 		}
	// 		liked, count, err := db.ToggleReportLike(r.Context(), pool, reportID, user.ID)
	// 		if err != nil {
	// 			http.Error(w, "could not toggle like", http.StatusInternalServerError)
	// 			return
	// 		}
	// 		writeJSON(w, map[string]any{"liked": liked, "count": count})
	// 	})

	// 	r.Get("/api/catalog/options", func(w http.ResponseWriter, r *http.Request) {
	// 		options, err := db.ListCatalogOptions(r.Context(), pool, requestLocale(r))
	// 		if err != nil {
	// 			http.Error(w, "could not load catalog options", http.StatusInternalServerError)
	// 			return
	// 		}
	// 		writeJSON(w, options)
	// 	})

	// 	r.Get("/api/me", userHandler.GetMe)

	// 	r.Patch("/api/me", func(w http.ResponseWriter, r *http.Request) {
	// 		_, ok, err := getUser(r)
	// 		if !ok {
	// 			http.Error(w, "unauthorized", http.StatusUnauthorized)
	// 			return
	// 		}
	// 		if err != nil {
	// 			http.Error(w, "could not load user", http.StatusInternalServerError)
	// 			return
	// 		}

	// 		var request struct {
	// 			Locale string `json:"locale"`
	// 		}
	// 		if err := json.NewDecoder(r.Body).Decode(&request); err != nil || !localePattern.MatchString(request.Locale) {
	// 			http.Error(w, "invalid locale", http.StatusBadRequest)
	// 			return
	// 		}
	// 		firebaseUser, _ := authmw.UserFromContext(r.Context())
	// 		updatedUser, err := db.UpdateUserLocale(r.Context(), pool, firebaseUser.UID, request.Locale)
	// 		if err != nil {
	// 			http.Error(w, "could not update user locale", http.StatusInternalServerError)
	// 			return
	// 		}
	// 		w.Header().Set("Content-Type", "application/json")
	// 		json.NewEncoder(w).Encode(updatedUser)
	// 	})

	// 	r.Get("/api/me/submissions", func(w http.ResponseWriter, r *http.Request) {
	// 		user, ok, err := getCurrentUser(r)
	// 		if !ok {
	// 			http.Error(w, "unauthorized", http.StatusUnauthorized)
	// 			return
	// 		}
	// 		if err != nil {
	// 			http.Error(w, "could not load user", http.StatusInternalServerError)
	// 			return
	// 		}
	// 		submissions, err := db.ListUserSubmissions(r.Context(), pool, user.ID, requestLocale(r))
	// 		if err != nil {
	// 			http.Error(w, "could not load submissions", http.StatusInternalServerError)
	// 			return
	// 		}
	// 		writeJSON(w, submissions)
	// 	})

	// 	r.Post("/api/submissions", func(w http.ResponseWriter, r *http.Request) {
	// 		user, ok, err := getCurrentUser(r)
	// 		if !ok {
	// 			http.Error(w, "unauthorized", http.StatusUnauthorized)
	// 			return
	// 		}
	// 		if err != nil {
	// 			http.Error(w, "could not load user", http.StatusInternalServerError)
	// 			return
	// 		}
	// 		var request struct {
	// 			ProductName string           `json:"productName"`
	// 			BrandID     int64            `json:"brandId"`
	// 			CategoryID  int64            `json:"categoryId"`
	// 			SellerID    int64            `json:"sellerId"`
	// 			ProductURL  string           `json:"productUrl"`
	// 			Description string           `json:"description"`
	// 			ObservedAt  *string          `json:"observedAt"`
	// 			Images      []db.ReportImage `json:"images"`
	// 		}
	// 		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
	// 			http.Error(w, "invalid request", http.StatusBadRequest)
	// 			return
	// 		}
	// 		slog.Debug("submission request decoded", "userId", user.ID, "request", request)
	// 		if request.ProductName == "" || request.Description == "" || len(request.Images) > 5 {
	// 			http.Error(w, "invalid submission", http.StatusBadRequest)
	// 			return
	// 		}
	// 		var observedAt *time.Time
	// 		if request.ObservedAt != nil && *request.ObservedAt != "" {
	// 			parsed, parseErr := time.Parse("2006-01-02", *request.ObservedAt)
	// 			if parseErr != nil {
	// 				http.Error(w, "invalid observed date", http.StatusBadRequest)
	// 				return
	// 			}
	// 			observedAt = &parsed
	// 		}
	// 		reportID, err := db.CreateSubmission(r.Context(), pool, user.ID, db.NewSubmission{
	// 			ProductName: request.ProductName, BrandID: request.BrandID, CategoryID: request.CategoryID,
	// 			SellerID: request.SellerID, ProductURL: request.ProductURL,
	// 			Description: request.Description, ObservedAt: observedAt, Images: request.Images,
	// 		})
	// 		if err != nil {
	// 			http.Error(w, "could not create submission", http.StatusInternalServerError)
	// 			return
	// 		}
	// 		slog.Debug("submission created", "reportId", reportID)
	// 		w.WriteHeader(http.StatusCreated)
	// 		writeJSON(w, map[string]int64{"id": reportID})
	// 	})

	// 	r.Patch("/api/me/submissions/{id}", func(w http.ResponseWriter, r *http.Request) {
	// 		user, ok, err := getCurrentUser(r)
	// 		if !ok {
	// 			http.Error(w, "unauthorized", http.StatusUnauthorized)
	// 			return
	// 		}
	// 		if err != nil {
	// 			http.Error(w, "could not load user", http.StatusInternalServerError)
	// 			return
	// 		}
	// 		reportID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	// 		if err != nil {
	// 			http.Error(w, "invalid submission id", http.StatusBadRequest)
	// 			return
	// 		}
	// 		var request struct {
	// 			ProductName string  `json:"productName"`
	// 			BrandID     int64   `json:"brandId"`
	// 			CategoryID  int64   `json:"categoryId"`
	// 			SellerID    int64   `json:"sellerId"`
	// 			Description string  `json:"description"`
	// 			ObservedAt  *string `json:"observedAt"`
	// 		}
	// 		if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.ProductName == "" || request.Description == "" {
	// 			http.Error(w, "invalid submission", http.StatusBadRequest)
	// 			return
	// 		}
	// 		var observedAt *time.Time
	// 		if request.ObservedAt != nil && *request.ObservedAt != "" {
	// 			parsed, parseErr := time.Parse("2006-01-02", *request.ObservedAt)
	// 			if parseErr != nil {
	// 				http.Error(w, "invalid observed date", http.StatusBadRequest)
	// 				return
	// 			}
	// 			observedAt = &parsed
	// 		}
	// 		if err := db.UpdateUserSubmission(r.Context(), pool, user.ID, reportID, db.UpdateSubmission{
	// 			ProductName: request.ProductName, BrandID: request.BrandID, CategoryID: request.CategoryID,
	// 			SellerID: request.SellerID, Description: request.Description,
	// 			ObservedAt: observedAt,
	// 		}); err != nil {
	// 			http.Error(w, "could not update submission", http.StatusNotFound)
	// 			return
	// 		}
	// 		w.WriteHeader(http.StatusNoContent)
	// 	})
	// })

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
