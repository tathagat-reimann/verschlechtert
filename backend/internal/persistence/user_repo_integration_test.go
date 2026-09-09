//go:build integration
// +build integration

// can be run with `go test -tags=integration`

package persistence

import (
	"context"
	"testing"
	"verschlechtert/backend/internal/config"
	"verschlechtert/backend/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

func setupTestStore(t *testing.T) *UserRepository {
	t.Helper()

	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://app:app@localhost:5432/appdb?sslmode=disable")
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "./serviceAccountKey.json")
	t.Setenv("APP_ENV", "development")
	t.Setenv("CORS_ALLOWED_ORIGIN", "http://localhost:5173")

	cfg := config.Load()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("Could not connect to database: %v (set DATABASE_URL)", err)
		return nil
	}
	// defer pool.Close()

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("Could not ping database: %v", err)
		return nil
	}

	return NewUserRepository(pool)
}

func Test_SaveUser(t *testing.T) {
	//given

	repo := setupTestStore(t)
	assert.NotNil(t, repo)

	// when
	user, _ := domain.NewUser("f", "d", "p", "e", "")
	err := repo.Save(t.Context(), user)

	// then
	assert.NoError(t, err)
}
