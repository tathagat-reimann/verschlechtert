package router

import (
	"net/http"
	authmw "verschlechtert/backend/internal/auth"
	"verschlechtert/backend/internal/httpmw"
	"verschlechtert/backend/internal/report"
	"verschlechtert/backend/internal/user"

	"firebase.google.com/go/v4/auth"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

type Router struct {
	userHandler    *user.UserHandler
	reportHandler  *report.ReportHandler
	allowedOrigins []string
	authClient     *auth.Client
}

func NewRouter(userHandler *user.UserHandler, reportHandler *report.ReportHandler,
	allowedOrigins []string, authClient *auth.Client) *Router {
	return &Router{
		userHandler:    userHandler,
		reportHandler:  reportHandler,
		allowedOrigins: allowedOrigins,
		authClient:     authClient,
	}
}

func (rt *Router) Setup() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   rt.allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	// 1MB request bodies and 10 requests/sec per IP (burst 20) are reasonable defaults for this API.
	r.Use(httpmw.MaxBodyBytes(1 << 20))
	r.Use(httpmw.NewRateLimiter(10, 20).Middleware)

	// Public routes (no auth required)
	r.Get("/health", HealthCheck)

	// Protected API routes (require Firebase auth)
	r.Route("/api", func(r chi.Router) {
		r.Use(authmw.Middleware(rt.authClient))

		// Note: handlers now have authClient for setting Firebase custom claims
		rt.userHandler.RegisterRoutes(r)
		rt.reportHandler.RegisterRoutes(r)
	})

	return r
}
