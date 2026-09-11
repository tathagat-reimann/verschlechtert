//go:build integration
// +build integration

// can be run with `go test -tags=integration`

package report

import (
	"context"
	"fmt"
	"testing"

	"verschlechtert/backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestStore(t *testing.T) *ReportRepository {
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
	}

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("Could not ping database: %v", err)
	}

	return NewReportRepository(pool)
}

func TestReportRepository_CreateAndGetActive(t *testing.T) {
	repo := setupTestStore(t)
	require.NotNil(t, repo)

	internalUserID := int64(-999)
	brand := &Brand{id: -999, name: "unknown"}
	seller := &Seller{id: -999, name: "unknown"}
	category := &Category{id: -999, name: "unknown"}
	image, err := NewReportImage("reports/test/image.jpg", "https://example.com/image.jpg")
	require.NoError(t, err)

	productName := fmt.Sprintf("integration-product-%s", t.Name())
	report, err := NewReport(productName, "integration description", category, brand, seller, "", internalUserID)
	require.NoError(t, err)
	require.NoError(t, report.AddImage(*image))

	reportID, err := repo.Create(t.Context(), report)
	require.NoError(t, err)
	require.NotZero(t, reportID)

	comment, err := NewReportComment("integration comment", internalUserID)
	require.NoError(t, err)
	require.NoError(t, repo.AddComment(t.Context(), reportID, *comment))

	like, err := NewReportLike(internalUserID)
	require.NoError(t, err)
	require.NoError(t, repo.AddLike(t.Context(), reportID, *like))

	alternative, err := NewReportAlternative("integration alternative", brand, seller, "", internalUserID)
	require.NoError(t, err)
	require.NoError(t, repo.AddAlternative(t.Context(), reportID, *alternative))

	loadedReport, err := repo.GetActive(t.Context(), reportID)
	require.NoError(t, err)
	assert.Equal(t, reportID, loadedReport.GetID())
	assert.Equal(t, productName, loadedReport.GetProductName())
	assert.Equal(t, int64(1), int64(len(loadedReport.GetImages())))
	assert.Equal(t, int64(1), int64(len(loadedReport.GetComments())))
	assert.Equal(t, int64(1), int64(len(loadedReport.GetLikes())))
	assert.Equal(t, int64(1), int64(len(loadedReport.GetAlternatives())))

	repo.RemoveLike(t.Context(), reportID, internalUserID)

	loadedReport, err = repo.GetActiveForUser(t.Context(), reportID, internalUserID)
	require.NoError(t, err)
	assert.Equal(t, reportID, loadedReport.GetID())
	assert.Equal(t, productName, loadedReport.GetProductName())
	assert.Equal(t, int64(1), int64(len(loadedReport.GetImages())))
	assert.Equal(t, int64(1), int64(len(loadedReport.GetComments())))
	assert.Equal(t, int64(0), int64(len(loadedReport.GetLikes())))
	assert.Equal(t, int64(1), int64(len(loadedReport.GetAlternatives())))

	listedReports, err := repo.ListActive(t.Context(), 20, 0)
	require.NoError(t, err)
	assert.Contains(t, reportIDs(listedReports), reportID)

	userReports, err := repo.ListActiveByUser(t.Context(), internalUserID, 20, 0)
	require.NoError(t, err)
	assert.Contains(t, reportIDs(userReports), reportID)
}

func reportIDs(reports []*Report) []int64 {
	ids := make([]int64, 0, len(reports))
	for _, report := range reports {
		ids = append(ids, report.GetID())
	}
	return ids
}
