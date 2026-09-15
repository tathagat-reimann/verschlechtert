package report

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"verschlechtert/backend/internal/user"

	"github.com/go-chi/chi/v5"

	"firebase.google.com/go/v4/auth"

	httpmw "verschlechtert/backend/internal/httmw"
	errorMessage "verschlechtert/backend/internal/util"
	responseUtil "verschlechtert/backend/internal/util"
)

type IReportService interface {
	Create(ctx context.Context, productName, description string, category *Category, brand *Brand, seller *Seller, productURL string, createdByUserID int64, images []ReportImage) (int64, error)
	GetActive(ctx context.Context, locale string, reportID int64) (*Report, error)
	GetActiveForUser(ctx context.Context, locale string, reportID, userID int64) (*Report, error)
	ListActive(ctx context.Context, locale string, limit, offset int) ([]*Report, error)
	ListActiveByUser(ctx context.Context, locale string, userID int64, limit, offset int) ([]*Report, error)
	AddComment(ctx context.Context, reportID, userID int64, comment string) error
	AddLike(ctx context.Context, reportID, userID int64) error
	RemoveLike(ctx context.Context, reportID, userID int64) error
	AddAlternative(ctx context.Context, reportID int64, productName string, brand *Brand, seller *Seller, productURL string, createdByUserID int64) error
}

type IAuthMiddleware interface {
	UserFromContext(ctx context.Context) (*auth.Token, bool)
}

type ReportHandler struct {
	reportService   IReportService
	userService     user.IUserService
	userFromContext func(ctx context.Context) (*auth.Token, bool)
	authClient      *auth.Client
}

func (h *ReportHandler) RegisterRoutes(r chi.Router) {
	r.Get("/reports", h.ListActive)
	r.Get("/reports/{id}", h.GetActive)

	r.Post("/api/reports/{id}/comments", h.AddComment)
	r.Post("/api/reports/{id}/alternatives", h.AddAlternative)
	r.Post("/api/reports/{id}/like", h.ToggleLike)

	r.Post("/reports", h.Create)
}

func NewReportHandler(reportService IReportService, userService user.IUserService, UserFromContext func(ctx context.Context) (*auth.Token, bool), authClient *auth.Client) *ReportHandler {
	return &ReportHandler{reportService: reportService, userService: userService, userFromContext: UserFromContext, authClient: authClient}
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
	correlationID := httpmw.CorrelationIDFromContext(r.Context())

	userID, err := h.currentUserID(r)
	if err != nil {
		responseUtil.WriteErrorJSON(w, errorMessage.Unauthorized, correlationID)
		return
	}
	limit, offset := pagination(r)
	locale := h.getLocaleFromClaims(r)
	reports, err := h.reportService.ListActive(r.Context(), locale, limit, offset)
	if err != nil {
		responseUtil.WriteErrorJSON(w, errorMessage.ErrLoadingAllReports, correlationID)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"reports": mapReports(reports, userID), "hasMore": len(reports) == limit})
}

func (h *ReportHandler) GetActive(w http.ResponseWriter, r *http.Request) {
	correlationID := httpmw.CorrelationIDFromContext(r.Context())

	reportID, ok := reportID(w, r)
	if !ok {
		return
	}
	userID, err := h.currentUserID(r)
	if err != nil {
		responseUtil.WriteErrorJSON(w, errorMessage.Unauthorized, correlationID)
		return
	}
	locale := h.getLocaleFromClaims(r)
	report, err := h.reportService.GetActive(r.Context(), locale, reportID)
	if err != nil {
		responseUtil.WriteErrorJSON(w, errorMessage.ReportNotFound, correlationID)
		return
	}
	writeJSON(w, http.StatusOK, mapReport(report, userID))
}

func (h *ReportHandler) Create(w http.ResponseWriter, r *http.Request) {
	correlationID := httpmw.CorrelationIDFromContext(r.Context())

	userID, err := h.currentUserID(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var request struct {
		BrandID     int64  `json:"brandId"`
		CategoryID  int64  `json:"categoryId"`
		SellerID    int64  `json:"sellerId"`
		ProductName string `json:"productName"`
		Description string `json:"description"`
		ProductURL  string `json:"productUrl"`
		Images      []struct {
			StoragePath string `json:"storagePath"`
			ImageURL    string `json:"imageUrl"`
		} `json:"images"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		// http.Error(w, "invalid report", http.StatusBadRequest)
		responseUtil.WriteErrorJSON(w, errorMessage.ErrJsonDecoding, correlationID)
		return
	}
	if request.BrandID <= 0 || request.CategoryID <= 0 || request.SellerID <= 0 || strings.TrimSpace(request.ProductName) == "" || strings.TrimSpace(request.Description) == "" {
		// http.Error(w, "invalid report", http.StatusBadRequest)
		responseUtil.WriteErrorJSON(w, errorMessage.ErrInvalidRequest, correlationID)
		return
	}

	category := &Category{id: request.CategoryID}
	brand := &Brand{id: request.BrandID}
	seller := &Seller{id: request.SellerID}
	images := make([]ReportImage, 0, len(request.Images))
	for _, image := range request.Images {
		if strings.TrimSpace(image.StoragePath) == "" || strings.TrimSpace(image.ImageURL) == "" {
			// http.Error(w, "invalid report image", http.StatusBadRequest)
			responseUtil.WriteErrorJSON(w, errorMessage.ErrInvalidReportImage, correlationID)
			return
		}
		reportImage, err := NewReportImage(image.StoragePath, image.ImageURL)
		if err != nil {
			// http.Error(w, "invalid report image", http.StatusBadRequest)
			responseUtil.WriteErrorJSON(w, errorMessage.ErrInvalidReportImage, correlationID)
			return
		}
		images = append(images, *reportImage)
	}

	reportID, err := h.reportService.Create(r.Context(), request.ProductName, request.Description, category, brand, seller, request.ProductURL, userID, images)
	if err != nil {
		responseUtil.WriteErrorJSON(w, errorMessage.ErrCreateReport, correlationID)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"id": reportID})
}

func (h *ReportHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	correlationID := httpmw.CorrelationIDFromContext(r.Context())

	reportID, ok := reportID(w, r)
	if !ok {
		return
	}
	userID, err := h.currentUserID(r)
	if err != nil {
		// http.Error(w, "unauthorized", http.StatusUnauthorized)
		responseUtil.WriteErrorJSON(w, errorMessage.Unauthorized, correlationID)
		return
	}
	var request struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || strings.TrimSpace(request.Body) == "" {
		// http.Error(w, "invalid comment", http.StatusBadRequest)
		responseUtil.WriteErrorJSON(w, errorMessage.ErrInvalidRequest, correlationID)
		return
	}
	if err := h.reportService.AddComment(r.Context(), reportID, userID, request.Body); err != nil {
		responseUtil.WriteErrorJSON(w, errorMessage.ErrAddComment, correlationID)
		return
	}
	// w.WriteHeader(http.StatusCreated)
	writeJSON(w, http.StatusCreated, map[string]any{"body": strings.TrimSpace(request.Body), "authorId": userID})
}

func (h *ReportHandler) ToggleLike(w http.ResponseWriter, r *http.Request) {
	correlationID := httpmw.CorrelationIDFromContext(r.Context())

	reportID, ok := reportID(w, r)
	if !ok {
		return
	}
	userID, err := h.currentUserID(r)
	if err != nil {
		// http.Error(w, "unauthorized", http.StatusUnauthorized)
		responseUtil.WriteErrorJSON(w, errorMessage.Unauthorized, correlationID)
		return
	}
	locale := h.getLocaleFromClaims(r)
	report, err := h.reportService.GetActive(r.Context(), locale, reportID)
	if err != nil {
		responseUtil.WriteErrorJSON(w, errorMessage.ReportNotFound, correlationID)
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
		// http.Error(w, "could not toggle like", http.StatusInternalServerError)
		responseUtil.WriteErrorJSON(w, errorMessage.ErrToggleLike, correlationID)
		return
	}
	count := len(report.GetLikes())
	if liked {
		count--
	} else {
		count++
	}
	writeJSON(w, http.StatusOK, map[string]any{"liked": !liked, "count": count})
}

func (h *ReportHandler) AddAlternative(w http.ResponseWriter, r *http.Request) {
	correlationID := httpmw.CorrelationIDFromContext(r.Context())

	reportID, ok := reportID(w, r)
	if !ok {
		return
	}
	userID, err := h.currentUserID(r)
	if err != nil {
		// http.Error(w, "unauthorized", http.StatusUnauthorized)
		responseUtil.WriteErrorJSON(w, errorMessage.Unauthorized, correlationID)
		return
	}
	var request struct {
		BrandID     int64  `json:"brandId"`
		SellerID    int64  `json:"sellerId"`
		ProductName string `json:"productName"`
		ProductURL  string `json:"productUrl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		// http.Error(w, "invalid alternative", http.StatusBadRequest)
		responseUtil.WriteErrorJSON(w, errorMessage.ErrJsonDecoding, correlationID)
		return
	}

	if request.BrandID <= 0 || request.SellerID <= 0 || strings.TrimSpace(request.ProductName) == "" {
		// http.Error(w, "invalid alternative", http.StatusBadRequest)
		responseUtil.WriteErrorJSON(w, errorMessage.ErrInvalidAlternative, correlationID)
		return
	}

	brand := &Brand{id: request.BrandID}
	seller := &Seller{id: request.SellerID}
	if err := h.reportService.AddAlternative(r.Context(), reportID, request.ProductName, brand, seller, request.ProductURL, userID); err != nil {
		responseUtil.WriteErrorJSON(w, errorMessage.ErrAddAlternative, correlationID)
		return
	}
	// w.WriteHeader(http.StatusCreated)
	writeJSON(w, http.StatusCreated, map[string]any{"productName": strings.TrimSpace(request.ProductName), "suggestedById": userID})
}

func (h *ReportHandler) currentUserID(r *http.Request) (int64, error) {
	firebaseUser, ok := h.userFromContext(r.Context())
	if !ok {
		return 0, errUnauthorized
	}
	// Try to extract userID from Firebase custom claims first (set on auth)
	if userIDClaim, ok := firebaseUser.Claims["userID"].(float64); ok {
		return int64(userIDClaim), nil
	}
	// Fall back to Upsert if claims not set (e.g., old tokens)
	currentUser, err := h.userService.Upsert(r.Context(), h.authClient, firebaseUser.UID, reportClaimString(firebaseUser.Claims, "name"), reportClaimString(firebaseUser.Claims, "picture"), reportClaimString(firebaseUser.Claims, "email"), h.getLocaleFromClaims(r))
	if err != nil {
		return 0, err
	}
	if !currentUser.IsActive() {
		return 0, errInactiveUser
	}
	return currentUser.GetID(), nil
}

func (h *ReportHandler) getLocaleFromClaims(r *http.Request) string {
	firebaseUser, ok := h.userFromContext(r.Context())
	if !ok {
		return "de"
	}
	if localeClaim, ok := firebaseUser.Claims["locale"].(string); ok && localeClaim == "en" {
		return localeClaim
	}
	return "de"
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
	correlationID := httpmw.CorrelationIDFromContext(r.Context())

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		// http.Error(w, "invalid report id", http.StatusBadRequest)
		responseUtil.WriteErrorJSON(w, errorMessage.ErrInvalidReportID, correlationID)
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

func writeJSON(w http.ResponseWriter, status int, value any) {
	// w.Header().Set("Content-Type", "application/json")
	responseUtil.WriteSuccessJSON(w,
		status,
		value,
	)
	// _ = json.NewEncoder(w).Encode(value)
}

func reportClaimString(claims map[string]interface{}, key string) string {
	value, _ := claims[key].(string)
	return strings.TrimSpace(value)
}

var errUnauthorized = errors.New("unauthorized")
var errInactiveUser = errors.New("user is inactive")
