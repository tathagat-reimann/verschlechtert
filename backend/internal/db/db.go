package db

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrAlternativeAlreadySuggested is returned when a user tries to suggest a second
// alternative for the same report (only one alternative per user per report is allowed).
var ErrAlternativeAlreadySuggested = errors.New("alternative already suggested for this report")

//go:embed migrations/001_initial_schema.sql
var initialSchema []byte

type User struct {
	ID          int64  `json:"id"`
	FirebaseUID string `json:"firebaseUid"`
	DisplayName string `json:"displayName"`
	PhotoURL    string `json:"photoUrl"`
	Email       string `json:"email"`
	Locale      string `json:"locale"`
}

type Report struct {
	ID          int64      `json:"id"`
	Description string     `json:"description"`
	ObservedAt  *time.Time `json:"observedAt,omitempty"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	Product     string     `json:"product"`
	Brand       string     `json:"brand"`
	Category    string     `json:"category"`
	Seller      string     `json:"seller"`
	LikeCount   int        `json:"likeCount"`
	LikedByMe   bool       `json:"likedByMe"`
}

type Submission struct {
	ID          int64      `json:"id"`
	Description string     `json:"description"`
	ObservedAt  *time.Time `json:"observedAt,omitempty"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	Product     string     `json:"product"`
	Brand       string     `json:"brand"`
	Category    string     `json:"category"`
	Seller      string     `json:"seller"`
	BrandID     int64      `json:"brandId"`
	CategoryID  int64      `json:"categoryId"`
	SellerID    int64      `json:"sellerId"`
	LikeCount   int        `json:"likeCount"`
	LikedByMe   bool       `json:"likedByMe"`
}

type Comment struct {
	ID         int64     `json:"id"`
	Body       string    `json:"body"`
	AuthorName string    `json:"authorName"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Alternative struct {
	ID            int64     `json:"id"`
	ProductName   string    `json:"productName"`
	Brand         string    `json:"brand"`
	Seller        string    `json:"seller"`
	ProductURL    string    `json:"productUrl,omitempty"`
	AuthorName    string    `json:"authorName"`
	CreatedAt     time.Time `json:"createdAt"`
	SuggestedByMe bool      `json:"suggestedByMe"`
}

type ReportImageView struct {
	ImageURL  string `json:"imageUrl"`
	SortOrder int16  `json:"sortOrder"`
}

type ReportDetail struct {
	ID             int64             `json:"id"`
	Description    string            `json:"description"`
	ObservedAt     *time.Time        `json:"observedAt,omitempty"`
	Status         string            `json:"status"`
	CreatedAt      time.Time         `json:"createdAt"`
	Product        string            `json:"product"`
	Brand          string            `json:"brand"`
	Category       string            `json:"category"`
	Seller         string            `json:"seller"`
	ProductURL     string            `json:"productUrl,omitempty"`
	Images         []ReportImageView `json:"images"`
	Comments       []Comment         `json:"comments"`
	Alternatives   []Alternative     `json:"alternatives"`
	LikeCount      int               `json:"likeCount"`
	LikedByMe      bool              `json:"likedByMe"`
	IsOwner        bool              `json:"isOwner"`
	HasAlternative bool              `json:"hasAlternative"`
}

type UpdateSubmission struct {
	ProductName string
	BrandID     int64
	CategoryID  int64
	SellerID    int64
	Description string
	ObservedAt  *time.Time
}

type NewSubmission struct {
	ProductName string
	BrandID     int64
	CategoryID  int64
	SellerID    int64
	ProductURL  string
	Description string
	ObservedAt  *time.Time
	Images      []ReportImage
}

type ReportImage struct {
	StoragePath string `json:"storagePath"`
	ImageURL    string `json:"imageUrl"`
	SortOrder   int16  `json:"sortOrder"`
}

// UnknownCatalogID is the fixed id of the seeded "Unknown" brand/category/seller row,
// used when a user can't find a matching catalog entry.
const UnknownCatalogID int64 = -999

type CatalogOption struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug,omitempty"`
}

type CatalogOptions struct {
	Brands              []CatalogOption `json:"brands"`
	Categories          []CatalogOption `json:"categories"`
	Sellers             []CatalogOption `json:"sellers"`
	UnspecifiedBrand    CatalogOption   `json:"unspecifiedBrand"`
	UnspecifiedCategory CatalogOption   `json:"unspecifiedCategory"`
	UnspecifiedSeller   CatalogOption   `json:"unspecifiedSeller"`
}

// NewPool creates a Postgres connection pool from a DATABASE_URL connection string.
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("creating pgx pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}
	return pool, nil
}

// ApplyMigrations creates the database schema required by the application.
func ApplyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`); err != nil {
		return fmt.Errorf("creating schema migrations table: %w", err)
	}

	var applied bool
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM schema_migrations WHERE version = 1
		)
	`).Scan(&applied); err != nil {
		return fmt.Errorf("checking initial schema migration: %w", err)
	}
	if applied {
		return nil
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("starting initial schema migration: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, string(initialSchema)); err != nil {
		return fmt.Errorf("applying initial schema migration: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES (1)`); err != nil {
		return fmt.Errorf("recording initial schema migration: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing initial schema migration: %w", err)
	}

	return nil
}

// UpsertUser creates or refreshes the local profile for a verified Firebase user.
func UpsertUser(ctx context.Context, pool *pgxpool.Pool, firebaseUID, displayName, photoURL, email string) (User, error) {
	var user User
	err := pool.QueryRow(ctx, `
		INSERT INTO users (firebase_uid, display_name, photo_url, email)
		VALUES ($1, COALESCE(NULLIF($2, ''), 'Anonymous user'), NULLIF($3, ''), NULLIF($4, ''))
		ON CONFLICT (firebase_uid) DO UPDATE SET
			display_name = COALESCE(NULLIF(EXCLUDED.display_name, ''), users.display_name),
			photo_url = COALESCE(EXCLUDED.photo_url, users.photo_url),
			email = COALESCE(EXCLUDED.email, users.email),
			updated_at = now(),
			last_seen_at = now()
		RETURNING id, firebase_uid, display_name, COALESCE(photo_url, ''), COALESCE(email, ''), locale
	`, firebaseUID, displayName, photoURL, email).Scan(
		&user.ID,
		&user.FirebaseUID,
		&user.DisplayName,
		&user.PhotoURL,
		&user.Email,
		&user.Locale,
	)
	if err != nil {
		return User{}, fmt.Errorf("upserting user: %w", err)
	}
	return user, nil
}

// UpdateUserLocale changes the application language preference for a user.
func UpdateUserLocale(ctx context.Context, pool *pgxpool.Pool, firebaseUID, locale string) (User, error) {
	var user User
	err := pool.QueryRow(ctx, `
		UPDATE users
		SET locale = $2, updated_at = now()
		WHERE firebase_uid = $1
		RETURNING id, firebase_uid, display_name, COALESCE(photo_url, ''), COALESCE(email, ''), locale
	`, firebaseUID, locale).Scan(
		&user.ID,
		&user.FirebaseUID,
		&user.DisplayName,
		&user.PhotoURL,
		&user.Email,
		&user.Locale,
	)
	if err != nil {
		return User{}, fmt.Errorf("updating user locale: %w", err)
	}
	return user, nil
}

// ListLatestReports returns up to limit reports starting at offset, plus whether more are available.
func ListLatestReports(ctx context.Context, pool *pgxpool.Pool, search string, limit, offset int, locale string, currentUserID int64) ([]Report, bool, error) {
	rows, err := pool.Query(ctx, `
		SELECT dr.id, dr.description, dr.observed_at, dr.status, dr.created_at,
		       dr.product_name, b.name, COALESCE(ct.name, c.slug), s.name,
		       (SELECT count(*) FROM report_likes rl WHERE rl.report_id = dr.id),
		       EXISTS (SELECT 1 FROM report_likes rl WHERE rl.report_id = dr.id AND rl.user_id = $4)
		FROM deterioration_reports dr
		JOIN brands b ON b.id = dr.brand_id
		JOIN categories c ON c.id = dr.category_id
		LEFT JOIN category_translations ct ON ct.category_id = c.id AND ct.locale = $3
		JOIN sellers s ON s.id = dr.seller_id
		WHERE ($1 = '' OR dr.product_name ILIKE '%' || $1 || '%' OR b.name ILIKE '%' || $1 || '%' OR s.name ILIKE '%' || $1 || '%')
		ORDER BY dr.created_at DESC, dr.id DESC
		LIMIT $2 OFFSET $5
	`, search, limit+1, locale, currentUserID, offset)
	if err != nil {
		return nil, false, fmt.Errorf("listing reports: %w", err)
	}
	defer rows.Close()

	reports := make([]Report, 0)
	for rows.Next() {
		var report Report
		if err := rows.Scan(&report.ID, &report.Description, &report.ObservedAt, &report.Status, &report.CreatedAt, &report.Product, &report.Brand, &report.Category, &report.Seller, &report.LikeCount, &report.LikedByMe); err != nil {
			return nil, false, fmt.Errorf("scanning report: %w", err)
		}
		reports = append(reports, report)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("reading reports: %w", err)
	}
	hasMore := len(reports) > limit
	if hasMore {
		reports = reports[:limit]
	}
	slog.Debug("listed latest reports", "search", search, "offset", offset, "count", len(reports), "hasMore", hasMore)
	return reports, hasMore, nil
}

func CreateSubmission(ctx context.Context, pool *pgxpool.Pool, userID int64, submission NewSubmission) (int64, error) {
	if len(submission.Images) > 5 {
		return 0, fmt.Errorf("a report can contain at most 5 images")
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("starting submission transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var reportID int64
	if err := tx.QueryRow(ctx, `
		INSERT INTO deterioration_reports (seller_id, brand_id, category_id, submitted_by_user_id, product_name, product_url, description, observed_at)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7, $8)
		RETURNING id
	`, submission.SellerID, submission.BrandID, submission.CategoryID, userID, submission.ProductName, submission.ProductURL, submission.Description, submission.ObservedAt).Scan(&reportID); err != nil {
		return 0, fmt.Errorf("creating deterioration report: %w", err)
	}
	for index, image := range submission.Images {
		if image.SortOrder == 0 {
			image.SortOrder = int16(index + 1)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO report_images (report_id, storage_path, image_url, sort_order)
			VALUES ($1, $2, $3, $4)
		`, reportID, image.StoragePath, image.ImageURL, image.SortOrder); err != nil {
			return 0, fmt.Errorf("creating report image: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("committing submission: %w", err)
	}
	slog.Debug("created submission", "userId", userID, "reportId", reportID)
	return reportID, nil
}

func ListUserSubmissions(ctx context.Context, pool *pgxpool.Pool, userID int64, locale string) ([]Submission, error) {
	rows, err := pool.Query(ctx, `
		SELECT dr.id, dr.description, dr.observed_at, dr.status, dr.created_at,
		       dr.product_name, b.id, b.name, c.id, COALESCE(ct.name, c.slug), s.id, s.name,
		       (SELECT count(*) FROM report_likes rl WHERE rl.report_id = dr.id),
		       EXISTS (SELECT 1 FROM report_likes rl WHERE rl.report_id = dr.id AND rl.user_id = $1)
		FROM deterioration_reports dr
		JOIN brands b ON b.id = dr.brand_id
		JOIN categories c ON c.id = dr.category_id
		LEFT JOIN category_translations ct ON ct.category_id = c.id AND ct.locale = $2
		JOIN sellers s ON s.id = dr.seller_id
		WHERE dr.submitted_by_user_id = $1
		ORDER BY dr.created_at DESC
	`, userID, locale)
	if err != nil {
		return nil, fmt.Errorf("listing submissions: %w", err)
	}
	defer rows.Close()
	submissions := make([]Submission, 0)
	for rows.Next() {
		var submission Submission
		if err := rows.Scan(&submission.ID, &submission.Description, &submission.ObservedAt, &submission.Status, &submission.CreatedAt, &submission.Product, &submission.BrandID, &submission.Brand, &submission.CategoryID, &submission.Category, &submission.SellerID, &submission.Seller, &submission.LikeCount, &submission.LikedByMe); err != nil {
			return nil, fmt.Errorf("scanning submission: %w", err)
		}
		submissions = append(submissions, submission)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading submissions: %w", err)
	}
	slog.Debug("listed user submissions", "userId", userID, "count", len(submissions))
	return submissions, nil
}

// GetReportDetail loads a single report's full details, including images, comments,
// like count/state, and whether currentUserID is the report's owner.
func GetReportDetail(ctx context.Context, pool *pgxpool.Pool, reportID, currentUserID int64, locale string) (ReportDetail, error) {
	var detail ReportDetail
	var submittedByUserID int64
	err := pool.QueryRow(ctx, `
		SELECT dr.id, dr.description, dr.observed_at, dr.status, dr.created_at,
		       dr.product_name, b.name, COALESCE(ct.name, c.slug), s.name, COALESCE(dr.product_url, ''),
		       dr.submitted_by_user_id,
		       (SELECT count(*) FROM report_likes rl WHERE rl.report_id = dr.id),
		       EXISTS (SELECT 1 FROM report_likes rl WHERE rl.report_id = dr.id AND rl.user_id = $2)
		FROM deterioration_reports dr
		JOIN brands b ON b.id = dr.brand_id
		JOIN categories c ON c.id = dr.category_id
		LEFT JOIN category_translations ct ON ct.category_id = c.id AND ct.locale = $3
		JOIN sellers s ON s.id = dr.seller_id
		WHERE dr.id = $1
	`, reportID, currentUserID, locale).Scan(
		&detail.ID, &detail.Description, &detail.ObservedAt, &detail.Status, &detail.CreatedAt,
		&detail.Product, &detail.Brand, &detail.Category, &detail.Seller, &detail.ProductURL,
		&submittedByUserID, &detail.LikeCount, &detail.LikedByMe,
	)
	if err != nil {
		return ReportDetail{}, fmt.Errorf("loading report: %w", err)
	}
	detail.IsOwner = submittedByUserID == currentUserID

	detail.Images = make([]ReportImageView, 0)
	imageRows, err := pool.Query(ctx, `SELECT image_url, sort_order FROM report_images WHERE report_id = $1 ORDER BY sort_order`, reportID)
	if err != nil {
		return ReportDetail{}, fmt.Errorf("loading report images: %w", err)
	}
	for imageRows.Next() {
		var image ReportImageView
		if err := imageRows.Scan(&image.ImageURL, &image.SortOrder); err != nil {
			imageRows.Close()
			return ReportDetail{}, fmt.Errorf("scanning report image: %w", err)
		}
		detail.Images = append(detail.Images, image)
	}
	imageRows.Close()
	if err := imageRows.Err(); err != nil {
		return ReportDetail{}, fmt.Errorf("reading report images: %w", err)
	}

	detail.Comments = make([]Comment, 0)
	commentRows, err := pool.Query(ctx, `
		SELECT rc.id, rc.body, u.display_name, rc.created_at
		FROM report_comments rc
		JOIN users u ON u.id = rc.user_id
		WHERE rc.report_id = $1
		ORDER BY rc.created_at ASC
	`, reportID)
	if err != nil {
		return ReportDetail{}, fmt.Errorf("loading comments: %w", err)
	}
	for commentRows.Next() {
		var comment Comment
		if err := commentRows.Scan(&comment.ID, &comment.Body, &comment.AuthorName, &comment.CreatedAt); err != nil {
			commentRows.Close()
			return ReportDetail{}, fmt.Errorf("scanning comment: %w", err)
		}
		detail.Comments = append(detail.Comments, comment)
	}
	commentRows.Close()
	if err := commentRows.Err(); err != nil {
		return ReportDetail{}, fmt.Errorf("reading comments: %w", err)
	}

	detail.Alternatives = make([]Alternative, 0)
	altRows, err := pool.Query(ctx, `
		SELECT ra.id, ra.product_name, b.name, s.name, COALESCE(ra.product_url, ''), u.display_name, ra.created_at, ra.suggested_by_user_id
		FROM report_alternatives ra
		JOIN brands b ON b.id = ra.brand_id
		JOIN sellers s ON s.id = ra.seller_id
		JOIN users u ON u.id = ra.suggested_by_user_id
		WHERE ra.report_id = $1
		ORDER BY ra.created_at ASC
	`, reportID)
	if err != nil {
		return ReportDetail{}, fmt.Errorf("loading alternatives: %w", err)
	}
	for altRows.Next() {
		var alt Alternative
		var suggestedByUserID int64
		if err := altRows.Scan(&alt.ID, &alt.ProductName, &alt.Brand, &alt.Seller, &alt.ProductURL, &alt.AuthorName, &alt.CreatedAt, &suggestedByUserID); err != nil {
			altRows.Close()
			return ReportDetail{}, fmt.Errorf("scanning alternative: %w", err)
		}
		alt.SuggestedByMe = suggestedByUserID == currentUserID
		if alt.SuggestedByMe {
			detail.HasAlternative = true
		}
		detail.Alternatives = append(detail.Alternatives, alt)
	}
	altRows.Close()
	if err := altRows.Err(); err != nil {
		return ReportDetail{}, fmt.Errorf("reading alternatives: %w", err)
	}
	return detail, nil
}

// AddReportComment adds a comment from userID to the given report.
func AddReportComment(ctx context.Context, pool *pgxpool.Pool, reportID, userID int64, body string) (Comment, error) {
	var comment Comment
	err := pool.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO report_comments (report_id, user_id, body)
			VALUES ($1, $2, $3)
			RETURNING id, body, created_at
		)
		SELECT inserted.id, inserted.body, u.display_name, inserted.created_at
		FROM inserted
		JOIN users u ON u.id = $2
	`, reportID, userID, body).Scan(&comment.ID, &comment.Body, &comment.AuthorName, &comment.CreatedAt)
	if err != nil {
		return Comment{}, fmt.Errorf("adding comment: %w", err)
	}
	return comment, nil
}

// AddReportAlternative records userID's suggested alternative product for a report.
// Returns ErrAlternativeAlreadySuggested if userID already suggested one for this report.
func AddReportAlternative(ctx context.Context, pool *pgxpool.Pool, reportID, userID, brandID, sellerID int64, productName, productURL string) (Alternative, error) {
	var alt Alternative
	err := pool.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO report_alternatives (report_id, suggested_by_user_id, brand_id, seller_id, product_name, product_url)
			VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''))
			RETURNING id, product_name, COALESCE(product_url, '') as product_url, created_at
		)
		SELECT inserted.id, inserted.product_name, b.name, s.name, inserted.product_url, u.display_name, inserted.created_at
		FROM inserted
		JOIN brands b ON b.id = $3
		JOIN sellers s ON s.id = $4
		JOIN users u ON u.id = $2
	`, reportID, userID, brandID, sellerID, productName, productURL).Scan(
		&alt.ID, &alt.ProductName, &alt.Brand, &alt.Seller, &alt.ProductURL, &alt.AuthorName, &alt.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Alternative{}, ErrAlternativeAlreadySuggested
		}
		return Alternative{}, fmt.Errorf("adding alternative: %w", err)
	}
	alt.SuggestedByMe = true
	return alt, nil
}

// ToggleReportLike adds or removes userID's like on a report and returns the new state.
func ToggleReportLike(ctx context.Context, pool *pgxpool.Pool, reportID, userID int64) (liked bool, count int, err error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return false, 0, fmt.Errorf("starting like toggle: %w", err)
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx, `DELETE FROM report_likes WHERE report_id = $1 AND user_id = $2`, reportID, userID)
	if err != nil {
		return false, 0, fmt.Errorf("removing like: %w", err)
	}
	if result.RowsAffected() == 0 {
		if _, err := tx.Exec(ctx, `INSERT INTO report_likes (report_id, user_id) VALUES ($1, $2)`, reportID, userID); err != nil {
			return false, 0, fmt.Errorf("adding like: %w", err)
		}
		liked = true
	}
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM report_likes WHERE report_id = $1`, reportID).Scan(&count); err != nil {
		return false, 0, fmt.Errorf("counting likes: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return false, 0, fmt.Errorf("committing like toggle: %w", err)
	}
	return liked, count, nil
}

func UpdateUserSubmission(ctx context.Context, pool *pgxpool.Pool, userID, reportID int64, update UpdateSubmission) error {
	result, err := pool.Exec(ctx, `
		UPDATE deterioration_reports
		SET product_name = $3, brand_id = $4, category_id = $5, seller_id = $6, description = $7, observed_at = $8
		WHERE id = $1 AND submitted_by_user_id = $2
	`, reportID, userID, update.ProductName, update.BrandID, update.CategoryID, update.SellerID, update.Description, update.ObservedAt)
	if err != nil {
		return fmt.Errorf("updating submission: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("submission not found")
	}
	return nil
}

func ListCatalogOptions(ctx context.Context, pool *pgxpool.Pool, locale string) (CatalogOptions, error) {
	options := CatalogOptions{}
	for key, query := range map[string]string{
		"brands":  "SELECT id, name FROM brands WHERE id <> $1 ORDER BY name",
		"sellers": "SELECT id, name FROM sellers WHERE id <> $1 ORDER BY name",
	} {
		rows, err := pool.Query(ctx, query, UnknownCatalogID)
		if err != nil {
			return CatalogOptions{}, fmt.Errorf("listing %s: %w", key, err)
		}
		items := make([]CatalogOption, 0)
		for rows.Next() {
			var item CatalogOption
			if err := rows.Scan(&item.ID, &item.Name); err != nil {
				rows.Close()
				return CatalogOptions{}, fmt.Errorf("scanning %s: %w", key, err)
			}
			items = append(items, item)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return CatalogOptions{}, fmt.Errorf("reading %s: %w", key, err)
		}
		if key == "brands" {
			options.Brands = items
		} else {
			options.Sellers = items
		}
	}
	if err := pool.QueryRow(ctx, `SELECT id, name FROM brands WHERE id = $1`, UnknownCatalogID).Scan(&options.UnspecifiedBrand.ID, &options.UnspecifiedBrand.Name); err != nil {
		return CatalogOptions{}, fmt.Errorf("loading unspecified brand: %w", err)
	}
	if err := pool.QueryRow(ctx, `SELECT id, name FROM sellers WHERE id = $1`, UnknownCatalogID).Scan(&options.UnspecifiedSeller.ID, &options.UnspecifiedSeller.Name); err != nil {
		return CatalogOptions{}, fmt.Errorf("loading unspecified seller: %w", err)
	}

	rows, err := pool.Query(ctx, `
		SELECT c.id, c.slug, COALESCE(ct.name, c.slug) AS name
		FROM categories c
		LEFT JOIN category_translations ct ON ct.category_id = c.id AND ct.locale = $1
		WHERE c.id <> $2
		ORDER BY c.slug
	`, locale, UnknownCatalogID)
	if err != nil {
		return CatalogOptions{}, fmt.Errorf("listing categories: %w", err)
	}
	categories := make([]CatalogOption, 0)
	for rows.Next() {
		var item CatalogOption
		if err := rows.Scan(&item.ID, &item.Slug, &item.Name); err != nil {
			rows.Close()
			return CatalogOptions{}, fmt.Errorf("scanning category: %w", err)
		}
		categories = append(categories, item)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return CatalogOptions{}, fmt.Errorf("reading categories: %w", err)
	}
	options.Categories = categories
	if err := pool.QueryRow(ctx, `
		SELECT c.id, c.slug, COALESCE(ct.name, c.slug)
		FROM categories c
		LEFT JOIN category_translations ct ON ct.category_id = c.id AND ct.locale = $1
		WHERE c.id = $2
	`, locale, UnknownCatalogID).Scan(&options.UnspecifiedCategory.ID, &options.UnspecifiedCategory.Slug, &options.UnspecifiedCategory.Name); err != nil {
		return CatalogOptions{}, fmt.Errorf("loading unspecified category: %w", err)
	}
	return options, nil
}
