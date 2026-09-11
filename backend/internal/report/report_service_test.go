package report

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockReportRepository struct {
	createError        error
	getError           error
	listError          error
	addCommentError    error
	addLikeError       error
	removeLikeError    error
	alternativeError   error
	createdReport      *Report
	activeReport       *Report
	activeReports      []*Report
	createdComment     *ReportComment
	createdLike        *ReportLike
	createdAlternative *ReportAlternative
	removedLikeUserID  int64
}

func (m *mockReportRepository) Create(ctx context.Context, report *Report) (int64, error) {
	m.createdReport = report
	if m.createError != nil {
		return 0, m.createError
	}
	return 1, nil
}

func (m *mockReportRepository) GetActive(ctx context.Context, reportID int64) (*Report, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	return m.activeReport, nil
}

func (m *mockReportRepository) GetActiveForUser(ctx context.Context, reportID, userID int64) (*Report, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	return m.activeReport, nil
}

func (m *mockReportRepository) ListActive(ctx context.Context, limit, offset int) ([]*Report, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	return m.activeReports, nil
}

func (m *mockReportRepository) ListActiveByUser(ctx context.Context, userID int64, limit, offset int) ([]*Report, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	return m.activeReports, nil
}

func (m *mockReportRepository) AddComment(ctx context.Context, reportID int64, comment ReportComment) error {
	m.createdComment = &comment
	return m.addCommentError
}

func (m *mockReportRepository) AddLike(ctx context.Context, reportID int64, like ReportLike) error {
	m.createdLike = &like
	return m.addLikeError
}

func (m *mockReportRepository) RemoveLike(ctx context.Context, reportID, userID int64) error {
	m.removedLikeUserID = userID
	return m.removeLikeError
}

func (m *mockReportRepository) AddAlternative(ctx context.Context, reportID int64, alternative ReportAlternative) error {
	m.createdAlternative = &alternative
	return m.alternativeError
}

func testReport(t *testing.T) *Report {
	t.Helper()

	brand := &Brand{id: 1, name: "brand"}
	seller := &Seller{id: 2, name: "seller"}
	category := &Category{id: 3, name: "category"}
	report, err := NewReport("product", "description", category, brand, seller, "", 10)
	require.NoError(t, err)
	return report
}

func TestReportService_Create_Success(t *testing.T) {
	mockRepo := &mockReportRepository{}
	service := NewReportService(mockRepo)
	image, err := NewReportImage("reports/image.jpg", "https://example.com/image.jpg")
	require.NoError(t, err)

	reportID, err := service.Create(t.Context(), "product", "description", &Category{id: 3, name: "category"}, &Brand{id: 1, name: "brand"}, &Seller{id: 2, name: "seller"}, "", 10, []ReportImage{*image})

	assert.NoError(t, err)
	assert.Equal(t, int64(1), reportID)
	require.NotNil(t, mockRepo.createdReport)
	assert.Equal(t, "product", mockRepo.createdReport.GetProductName())
	assert.Len(t, mockRepo.createdReport.GetImages(), 1)
}

func TestReportService_Create_ValidationError(t *testing.T) {
	mockRepo := &mockReportRepository{}
	service := NewReportService(mockRepo)

	reportID, err := service.Create(t.Context(), "", "description", &Category{id: 3}, &Brand{id: 1}, &Seller{id: 2}, "", 10, nil)

	assert.Error(t, err)
	assert.Zero(t, reportID)
	assert.Nil(t, mockRepo.createdReport)
}

func TestReportService_Create_RepositoryError(t *testing.T) {
	mockRepo := &mockReportRepository{createError: errors.New("database error")}
	service := NewReportService(mockRepo)

	reportID, err := service.Create(t.Context(), "product", "description", &Category{id: 3}, &Brand{id: 1}, &Seller{id: 2}, "", 10, nil)

	assert.Error(t, err)
	assert.Zero(t, reportID)
}

func TestReportService_GetAndList(t *testing.T) {
	activeReport := testReport(t)
	mockRepo := &mockReportRepository{activeReport: activeReport, activeReports: []*Report{activeReport}}
	service := NewReportService(mockRepo)

	got, err := service.GetActive(t.Context(), 1)
	assert.NoError(t, err)
	assert.Same(t, activeReport, got)

	gotForUser, err := service.GetActiveForUser(t.Context(), 1, 10)
	assert.NoError(t, err)
	assert.Same(t, activeReport, gotForUser)

	listed, err := service.ListActive(t.Context(), 20, 0)
	assert.NoError(t, err)
	assert.Len(t, listed, 1)

	listedByUser, err := service.ListActiveByUser(t.Context(), 10, 20, 0)
	assert.NoError(t, err)
	assert.Len(t, listedByUser, 1)
}

func TestReportService_AddComment_Success(t *testing.T) {
	mockRepo := &mockReportRepository{activeReport: testReport(t)}
	service := NewReportService(mockRepo)

	err := service.AddComment(t.Context(), 1, 10, "comment")

	assert.NoError(t, err)
	require.NotNil(t, mockRepo.createdComment)
	assert.Equal(t, "comment", mockRepo.createdComment.GetComment())
	assert.Equal(t, int64(10), mockRepo.createdComment.GetCreatedByUserID())
}

func TestReportService_AddLike_Success(t *testing.T) {
	mockRepo := &mockReportRepository{activeReport: testReport(t)}
	service := NewReportService(mockRepo)

	err := service.AddLike(t.Context(), 1, 10)

	assert.NoError(t, err)
	require.NotNil(t, mockRepo.createdLike)
	assert.Equal(t, int64(10), mockRepo.createdLike.GetCreatedByUserID())
}

func TestReportService_RemoveLike_Success(t *testing.T) {
	report := testReport(t)
	like, err := NewReportLike(10)
	require.NoError(t, err)
	require.NoError(t, report.AddLike(*like))
	mockRepo := &mockReportRepository{activeReport: report}
	service := NewReportService(mockRepo)

	err = service.RemoveLike(t.Context(), 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, int64(10), mockRepo.removedLikeUserID)
}

func TestReportService_AddAlternative_Success(t *testing.T) {
	mockRepo := &mockReportRepository{activeReport: testReport(t)}
	service := NewReportService(mockRepo)

	err := service.AddAlternative(t.Context(), 1, "alternative", &Brand{id: 1, name: "brand"}, &Seller{id: 2, name: "seller"}, "", 10)

	assert.NoError(t, err)
	require.NotNil(t, mockRepo.createdAlternative)
	assert.Equal(t, "alternative", mockRepo.createdAlternative.GetProductName())
	assert.Equal(t, int64(10), mockRepo.createdAlternative.GetCreatedByUserID())
}

func TestReportService_RepositoryError(t *testing.T) {
	databaseError := errors.New("database error")
	mockRepo := &mockReportRepository{activeReport: testReport(t), addCommentError: databaseError}
	service := NewReportService(mockRepo)

	err := service.AddComment(t.Context(), 1, 10, "comment")

	assert.ErrorIs(t, err, databaseError)
}
