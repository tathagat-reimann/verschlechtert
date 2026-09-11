package report

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"verschlechtert/backend/internal/user"

	"github.com/go-chi/chi/v5"

	"firebase.google.com/go/v4/auth"
)

type IReportService interface {
	GetActive(ctx context.Context, reportID int64) (*Report, error)
	GetActiveForUser(ctx context.Context, reportID, userID int64) (*Report, error)
	ListActive(ctx context.Context, limit, offset int) ([]*Report, error)
	ListActiveByUser(ctx context.Context, userID int64, limit, offset int) ([]*Report, error)
	AddComment(ctx context.Context, reportID, userID int64, comment string) error
	AddLike(ctx context.Context, reportID, userID int64) error
	RemoveLike(ctx context.Context, reportID, userID int64) error
	AddAlternative(ctx context.Context, reportID int64, productName string, brand *Brand, seller *Seller, productURL string, createdByUserID int64) error
}

type IAuthMiddleware interface {
	UserFromContext(ctx context.Context) (*auth.Token, bool)
}

type ReportHandler struct {
	reportService IReportService
	userService   user.IUserService
	authmw        IAuthMiddleware
}

func NewReportHandler(reportService IReportService, userService user.IUserService, authmw IAuthMiddleware) *ReportHandler {
	return &ReportHandler{reportService: reportService, userService: userService, authmw: authmw}
}

type reportResponse struct {
	ID             int64                 `json:"id"`
	ProductName    string                `json:"productName"`
	Description    string                `json:"description"`
	ProductURL     string                `json:"productUrl,omitempty"`
	Brand          string                `json:"brand"`
	Category       string                `json:"category"`
	Seller         string                `json:"seller"`
	CreatedAt      string                `json:"createdAt"`
	Images         []imageResponse       `json:"images,omitempty"`
	Comments       []commentResponse     `json:"comments,omitempty"`
	Alternatives   []alternativeResponse `json:"alternatives,omitempty"`
	LikeCount      int                   `json:"likeCount"`
	LikedByMe      bool                  `json:"likedByMe"`
	IsOwner        bool                  `json:"isOwner"`
	HasAlternative bool                  `json:"hasAlternative"`
}

type imageResponse struct {
	ImageURL  string `json:"imageUrl"`
	SortOrder int16  `json:"sortOrder"`
}

type commentResponse struct {
	ID        int64  `json:"id"`
	Body      string `json:"body"`
	AuthorID  int64  `json:"authorId"`
	CreatedAt string `json:"createdAt"`
}

type alternativeResponse struct {
	ID            int64  `json:"id"`
	ProductName   string `json:"productName"`
	Brand         string `json:"brand"`
	Seller        string `json:"seller"`
	ProductURL    string `json:"productUrl,omitempty"`
	SuggestedByID int64  `json:"suggestedById"`
	CreatedAt     string `json:"createdAt"`
}

func (h *ReportHandler) ListActive(w http.ResponseWriter, r *http.Request) {
	userID, err := h.currentUserID(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	limit, offset := pagination(r)
	reports, err := h.reportService.ListActive(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, "could not load reports", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"reports": mapReports(reports, userID), "hasMore": len(reports) == limit})
}

func (h *ReportHandler) GetActive(w http.ResponseWriter, r *http.Request) {
	reportID, ok := reportID(w, r)
	if !ok {
		return
	}
	userID, err := h.currentUserID(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	report, err := h.reportService.GetActive(r.Context(), reportID)
	if err != nil {
		http.Error(w, "report not found", http.StatusNotFound)
		return
	}
	writeJSON(w, mapReport(report, userID))
}

func (h *ReportHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	reportID, ok := reportID(w, r)
	if !ok {
		return
	}
	userID, err := h.currentUserID(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var request struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || strings.TrimSpace(request.Body) == "" {
		http.Error(w, "invalid comment", http.StatusBadRequest)
		return
	}
	if err := h.reportService.AddComment(r.Context(), reportID, userID, request.Body); err != nil {
		http.Error(w, "could not add comment", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]any{"body": strings.TrimSpace(request.Body), "authorId": userID})
}

func (h *ReportHandler) ToggleLike(w http.ResponseWriter, r *http.Request) {
	reportID, ok := reportID(w, r)
	if !ok {
		return
	}
	userID, err := h.currentUserID(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	report, err := h.reportService.GetActive(r.Context(), reportID)
	if err != nil {
		http.Error(w, "report not found", http.StatusNotFound)
		return
	}
	liked := false
	for _, like := range report.GetLikes() {
		if like.GetCreatedByUserID() == userID {
			liked = true
			break
		}
	}
	if liked {
		err = h.reportService.RemoveLike(r.Context(), reportID, userID)
	} else {
		err = h.reportService.AddLike(r.Context(), reportID, userID)
	}
	if err != nil {
		http.Error(w, "could not toggle like", http.StatusInternalServerError)
		return
	}
	count := len(report.GetLikes())
	if liked {
		count--
	} else {
		count++
	}
	writeJSON(w, map[string]any{"liked": !liked, "count": count})
}

func (h *ReportHandler) AddAlternative(w http.ResponseWriter, r *http.Request) {
	reportID, ok := reportID(w, r)
	if !ok {
		return
	}
	userID, err := h.currentUserID(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var request struct {
		BrandID     int64  `json:"brandId"`
		SellerID    int64  `json:"sellerId"`
		ProductName string `json:"productName"`
		ProductURL  string `json:"productUrl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.BrandID <= 0 || request.SellerID <= 0 || strings.TrimSpace(request.ProductName) == "" {
		http.Error(w, "invalid alternative", http.StatusBadRequest)
		return
	}
	brand := &Brand{id: request.BrandID}
	seller := &Seller{id: request.SellerID}
	if err := h.reportService.AddAlternative(r.Context(), reportID, request.ProductName, brand, seller, request.ProductURL, userID); err != nil {
		http.Error(w, "could not add alternative", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]any{"productName": strings.TrimSpace(request.ProductName), "suggestedById": userID})
}

func (h *ReportHandler) currentUserID(r *http.Request) (int64, error) {
	firebaseUser, ok := h.authmw.UserFromContext(r.Context())
	if !ok {
		return 0, errUnauthorized
	}
	currentUser, err := h.userService.Upsert(r.Context(), firebaseUser.UID, reportClaimString(firebaseUser.Claims, "name"), reportClaimString(firebaseUser.Claims, "picture"), reportClaimString(firebaseUser.Claims, "email"), reportRequestLocale(r))
	if err != nil {
		return 0, err
	}
	if !currentUser.IsActive() {
		return 0, errInactiveUser
	}
	return currentUser.GetID(), nil
}

func mapReports(reports []*Report, userID int64) []reportResponse {
	result := make([]reportResponse, 0, len(reports))
	for _, report := range reports {
		result = append(result, mapReport(report, userID))
	}
	return result
}

func mapReport(report *Report, userID int64) reportResponse {
	response := reportResponse{
		ID: report.GetID(), ProductName: report.GetProductName(), Description: report.GetDescription(), ProductURL: report.GetProductURL(),
		Brand: report.GetBrand().GetName(), Category: report.GetCategory().GetName(), Seller: report.GetSeller().GetName(),
		CreatedAt: report.GetCreatedAt().Format(time.RFC3339), LikeCount: len(report.GetLikes()),
		IsOwner: report.GetCreatedByUserID() == userID,
		Images:  make([]imageResponse, 0, len(report.GetImages())), Comments: make([]commentResponse, 0, len(report.GetComments())), Alternatives: make([]alternativeResponse, 0, len(report.GetAlternatives())),
	}
	for _, like := range report.GetLikes() {
		response.LikedByMe = response.LikedByMe || like.GetCreatedByUserID() == userID
	}
	for _, image := range report.GetImages() {
		response.Images = append(response.Images, imageResponse{ImageURL: image.GetURL()})
	}
	for _, comment := range report.GetComments() {
		response.Comments = append(response.Comments, commentResponse{ID: comment.GetID(), Body: comment.GetComment(), AuthorID: comment.GetCreatedByUserID(), CreatedAt: comment.GetCreatedAt().Format(time.RFC3339)})
	}
	for _, alternative := range report.GetAlternatives() {
		response.HasAlternative = response.HasAlternative || alternative.GetCreatedByUserID() == userID
		response.Alternatives = append(response.Alternatives, alternativeResponse{ID: alternative.GetID(), ProductName: alternative.GetProductName(), Brand: alternative.GetBrand().GetName(), Seller: alternative.GetSeller().GetName(), ProductURL: alternative.GetProductURL(), SuggestedByID: alternative.GetCreatedByUserID(), CreatedAt: alternative.GetCreatedAt().Format(time.RFC3339)})
	}
	return response
}

func reportID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid report id", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

func pagination(r *http.Request) (int, int) {
	limit, offset := 24, 0
	if value, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && value > 0 && value <= 100 {
		limit = value
	}
	if value, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && value >= 0 {
		offset = value
	}
	return limit, offset
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

var reportLocalePattern = regexp.MustCompile(`^[a-z]{2}(-[A-Z]{2})?$`)

func reportRequestLocale(r *http.Request) string {
	locale := r.URL.Query().Get("locale")
	if reportLocalePattern.MatchString(locale) {
		return locale
	}
	return "de"
}

func reportClaimString(claims map[string]interface{}, key string) string {
	value, _ := claims[key].(string)
	return strings.TrimSpace(value)
}

var errUnauthorized = errors.New("unauthorized")
var errInactiveUser = errors.New("user is inactive")
