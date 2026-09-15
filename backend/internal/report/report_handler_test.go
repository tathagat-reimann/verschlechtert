package report

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"verschlechtert/backend/internal/user"

	"firebase.google.com/go/v4/auth"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

type reportServiceHandlerMock struct {
	createdID       int64
	activeReport    *Report
	activeReports   []*Report
	createdUserID   int64
	comment         string
	addedLike       bool
	removedLike     bool
	alternativeName string
}

func (m *reportServiceHandlerMock) Create(_ context.Context, _, _ string, _ *Category, _ *Brand, _ *Seller, _ string, createdByUserID int64, _ []ReportImage) (int64, error) {
	m.createdUserID = createdByUserID
	return m.createdID, nil
}

func (m *reportServiceHandlerMock) GetActive(_ context.Context, _ string, _ int64) (*Report, error) {
	return m.activeReport, nil
}

func (m *reportServiceHandlerMock) GetActiveForUser(_ context.Context, _ string, _ int64, _ int64) (*Report, error) {
	return m.activeReport, nil
}

func (m *reportServiceHandlerMock) ListActive(_ context.Context, _ string, _, _ int) ([]*Report, error) {
	return m.activeReports, nil
}

func (m *reportServiceHandlerMock) ListActiveByUser(_ context.Context, _ string, _ int64, _, _ int) ([]*Report, error) {
	return m.activeReports, nil
}

func (m *reportServiceHandlerMock) AddComment(_ context.Context, _ int64, _ int64, comment string) error {
	m.comment = comment
	return nil
}

func (m *reportServiceHandlerMock) AddLike(_ context.Context, _ int64, _ int64) error {
	m.addedLike = true
	return nil
}

func (m *reportServiceHandlerMock) RemoveLike(_ context.Context, _ int64, _ int64) error {
	m.removedLike = true
	return nil
}

func (m *reportServiceHandlerMock) AddAlternative(_ context.Context, _ int64, productName string, _ *Brand, _ *Seller, _ string, _ int64) error {
	m.alternativeName = productName
	return nil
}

type reportUserServiceMock struct {
	currentUser *user.User
}

func (m *reportUserServiceMock) Upsert(_ context.Context, _ *auth.Client, _, _, _, _, _ string) (*user.User, error) {
	return m.currentUser, nil
}

func (m *reportUserServiceMock) UpdateLocale(_ context.Context, _ *auth.Client, _ *user.User, _ string) error {
	return nil
}

var authorizedUser = &auth.Token{
	UID: "test-uid",
	Claims: map[string]interface{}{
		"name":    "Test User",
		"picture": "http://example.com/photo.jpg",
		"email":   "test@example.com",
	},
}
var userOk = true

var userFromContextMock = func(ctx context.Context) (*auth.Token, bool) {
	return authorizedUser, userOk
}

func newReportHandlerForTest(service *reportServiceHandlerMock) *ReportHandler {
	currentUser, err := user.NewUser("test-uid", "Test User", "", "test@example.com", "de")
	if err != nil {
		panic(err)
	}

	return NewReportHandler(
		service,
		&reportUserServiceMock{currentUser: currentUser},
		userFromContextMock,
		nil,
	)
}

func requestWithReportID(method, target string, body io.Reader) *http.Request {
	request := httptest.NewRequest(method, target, body)
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add("id", "1")
	return request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeContext))
}

func TestReportHandler_ListActive(t *testing.T) {
	service := &reportServiceHandlerMock{activeReports: []*Report{testReport(t)}}
	handler := newReportHandlerForTest(service)

	response := httptest.NewRecorder()
	handler.ListActive(response, httptest.NewRequest(http.MethodGet, "/reports", nil))

	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"data":{"reports":[{"id":0,"productName":"product","description":"description","brand":"brand","category":"category","seller":"seller","createdAt":"0001-01-01T00:00:00Z","likeCount":0,"likedByMe":false,"isOwner":false,"hasAlternative":false}],"hasMore":false}}`, response.Body.String())
}

func TestGetActive(t *testing.T) {
	service := &reportServiceHandlerMock{activeReport: testReport(t)}
	handler := newReportHandlerForTest(service)

	response := httptest.NewRecorder()
	handler.GetActive(response, requestWithReportID(http.MethodGet, "/reports/1", nil))

	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"data":{"id":0,"productName":"product","description":"description","brand":"brand","category":"category","seller":"seller","createdAt":"0001-01-01T00:00:00Z","likeCount":0,"likedByMe":false,"isOwner":false,"hasAlternative":false}}`, response.Body.String())
}

func TestReportHandler_Create(t *testing.T) {
	service := &reportServiceHandlerMock{createdID: 42}
	handler := newReportHandlerForTest(service)

	response := httptest.NewRecorder()
	body := strings.NewReader(`{"brandId":1,"categoryId":2,"sellerId":3,"productName":"product","description":"description"}`)
	handler.Create(response, httptest.NewRequest(http.MethodPost, "/reports", body))

	assert.Equal(t, http.StatusCreated, response.Code)
	assert.JSONEq(t, `{"data":{"id":42}}`, response.Body.String())
	assert.Zero(t, service.createdUserID)
}

func TestReportHandler_AddComment(t *testing.T) {
	service := &reportServiceHandlerMock{}
	handler := newReportHandlerForTest(service)

	response := httptest.NewRecorder()
	handler.AddComment(response, requestWithReportID(http.MethodPost, "/reports/1/comments", strings.NewReader(`{"body":"comment"}`)))

	assert.Equal(t, http.StatusCreated, response.Code)
	assert.JSONEq(t, `{"data":{"body":"comment","authorId":0}}`, response.Body.String())
	assert.Equal(t, "comment", service.comment)
}

func TestReportHandler_ToggleLike_AddsLike(t *testing.T) {
	service := &reportServiceHandlerMock{activeReport: testReport(t)}
	handler := newReportHandlerForTest(service)

	response := httptest.NewRecorder()
	handler.ToggleLike(response, requestWithReportID(http.MethodPost, "/reports/1/like", nil))

	assert.Equal(t, http.StatusOK, response.Code)
	assert.True(t, service.addedLike)
	assert.JSONEq(t, `{"data":{"liked":true,"count":1}}`, response.Body.String())
}

func TestReportHandler_AddAlternative(t *testing.T) {
	service := &reportServiceHandlerMock{}
	handler := newReportHandlerForTest(service)

	response := httptest.NewRecorder()
	body := strings.NewReader(`{"brandId":1,"sellerId":2,"productName":"alternative"}`)
	handler.AddAlternative(response, requestWithReportID(http.MethodPost, "/reports/1/alternatives", body))

	assert.Equal(t, http.StatusCreated, response.Code)
	assert.JSONEq(t, `{"data":{"productName":"alternative","suggestedById":0}}`, response.Body.String())
	assert.Equal(t, "alternative", service.alternativeName)
}
