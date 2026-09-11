package report

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ReportRepository struct {
	pool *pgxpool.Pool
}

func NewReportRepository(pool *pgxpool.Pool) *ReportRepository {
	return &ReportRepository{pool: pool}
}

func (repo *ReportRepository) GetActive(ctx context.Context, reportID int64) (*Report, error) {
	report, err := repo.get(ctx, `WHERE r.id = $1 AND r.active = true`, reportID)
	if err != nil {
		return nil, fmt.Errorf("getting active report: %w", err)
	}
	if err := repo.loadDetails(ctx, report); err != nil {
		return nil, err
	}
	return report, nil
}

func (repo *ReportRepository) GetActiveForUser(ctx context.Context, reportID, userID int64) (*Report, error) {
	report, err := repo.get(ctx, `
		WHERE r.id = $1
		  AND r.submitted_by_user_id = $2
		  AND r.active = true
	`, reportID, userID)
	if err != nil {
		return nil, fmt.Errorf("getting active user report: %w", err)
	}
	if err := repo.loadDetails(ctx, report); err != nil {
		return nil, err
	}
	return report, nil
}

func (repo *ReportRepository) ListActive(ctx context.Context, limit, offset int) ([]*Report, error) {
	return repo.list(ctx, `
		WHERE r.active = true
		ORDER BY r.created_at DESC, r.id DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
}

func (repo *ReportRepository) ListActiveByUser(ctx context.Context, userID int64, limit, offset int) ([]*Report, error) {
	return repo.list(ctx, `
		WHERE r.submitted_by_user_id = $1
		  AND r.active = true
		ORDER BY r.created_at DESC, r.id DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
}

func (repo *ReportRepository) get(ctx context.Context, suffix string, args ...any) (*Report, error) {
	var (
		reportID, sellerID, brandID, categoryID, createdByUserID int64
		sellerName, brandName, categoryName                      string
		productName, productURL, description                     string
		createdAt                                                time.Time
	)

	err := repo.pool.QueryRow(ctx, `
		SELECT r.id, r.seller_id, s.name, r.brand_id, b.name,
		       r.category_id, c.name_de, r.submitted_by_user_id,
		       r.product_name, COALESCE(r.product_url, ''), r.description, r.created_at
		FROM report r
		JOIN seller s ON s.id = r.seller_id
		JOIN brand b ON b.id = r.brand_id
		JOIN category c ON c.id = r.category_id
	`+suffix, args...).Scan(
		&reportID, &sellerID, &sellerName, &brandID, &brandName,
		&categoryID, &categoryName, &createdByUserID, &productName,
		&productURL, &description, &createdAt,
	)
	if err != nil {
		return nil, err
	}

	return &Report{
		id:              reportID,
		productName:     productName,
		description:     description,
		category:        &Category{id: categoryID, name: categoryName},
		brand:           &Brand{id: brandID, name: brandName},
		seller:          &Seller{id: sellerID, name: sellerName},
		productURL:      productURL,
		createdByUserID: createdByUserID,
		createdAt:       createdAt,
	}, nil
}

func (repo *ReportRepository) list(ctx context.Context, suffix string, args ...any) ([]*Report, error) {
	rows, err := repo.pool.Query(ctx, `
		SELECT r.id, r.seller_id, s.name, r.brand_id, b.name,
		       r.category_id, c.name_de, r.submitted_by_user_id,
		       r.product_name, COALESCE(r.product_url, ''), r.description, r.created_at
		FROM report r
		JOIN seller s ON s.id = r.seller_id
		JOIN brand b ON b.id = r.brand_id
		JOIN category c ON c.id = r.category_id
	`+suffix, args...)
	if err != nil {
		return nil, fmt.Errorf("listing reports: %w", err)
	}
	defer rows.Close()

	reports := make([]*Report, 0)
	for rows.Next() {
		var (
			reportID, sellerID, brandID, categoryID, createdByUserID int64
			sellerName, brandName, categoryName                      string
			productName, productURL, description                     string
			createdAt                                                time.Time
		)
		if err := rows.Scan(
			&reportID, &sellerID, &sellerName, &brandID, &brandName,
			&categoryID, &categoryName, &createdByUserID, &productName,
			&productURL, &description, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("scanning report: %w", err)
		}

		report := &Report{
			id:              reportID,
			productName:     productName,
			description:     description,
			category:        &Category{id: categoryID, name: categoryName},
			brand:           &Brand{id: brandID, name: brandName},
			seller:          &Seller{id: sellerID, name: sellerName},
			productURL:      productURL,
			createdByUserID: createdByUserID,
			createdAt:       createdAt,
		}
		if err := repo.loadLikes(ctx, report); err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading reports: %w", err)
	}
	return reports, nil
}

func (repo *ReportRepository) loadDetails(ctx context.Context, report *Report) error {
	if err := repo.loadImages(ctx, report); err != nil {
		return err
	}
	if err := repo.loadComments(ctx, report); err != nil {
		return err
	}
	if err := repo.loadLikes(ctx, report); err != nil {
		return err
	}
	return repo.loadAlternatives(ctx, report)
}

func (repo *ReportRepository) loadImages(ctx context.Context, report *Report) error {
	rows, err := repo.pool.Query(ctx, `
		SELECT id, storage_path, image_url, sort_order
		FROM report_image
		WHERE report_id = $1
		ORDER BY sort_order
	`, report.id)
	if err != nil {
		return fmt.Errorf("loading report images: %w", err)
	}
	defer rows.Close()

	report.images = make([]ReportImage, 0)
	for rows.Next() {
		var image ReportImage
		var sortOrder int16
		if err := rows.Scan(&image.id, &image.storagePath, &image.url, &sortOrder); err != nil {
			return fmt.Errorf("scanning report image: %w", err)
		}
		report.images = append(report.images, image)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("reading report images: %w", err)
	}
	return nil
}

func (repo *ReportRepository) loadComments(ctx context.Context, report *Report) error {
	rows, err := repo.pool.Query(ctx, `
		SELECT id, user_id, body, created_at
		FROM report_comment
		WHERE report_id = $1
		ORDER BY created_at ASC, id ASC
	`, report.id)
	if err != nil {
		return fmt.Errorf("loading report comments: %w", err)
	}
	defer rows.Close()

	report.comments = make([]ReportComment, 0)
	for rows.Next() {
		var comment ReportComment
		if err := rows.Scan(&comment.id, &comment.createdByUserID, &comment.comment, &comment.createdAt); err != nil {
			return fmt.Errorf("scanning report comment: %w", err)
		}
		report.comments = append(report.comments, comment)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("reading report comments: %w", err)
	}
	return nil
}

func (repo *ReportRepository) loadLikes(ctx context.Context, report *Report) error {
	rows, err := repo.pool.Query(ctx, `
		SELECT user_id, created_at
		FROM report_like
		WHERE report_id = $1
		ORDER BY created_at ASC, user_id ASC
	`, report.id)
	if err != nil {
		return fmt.Errorf("loading report likes: %w", err)
	}
	defer rows.Close()

	report.likes = make([]ReportLike, 0)
	for rows.Next() {
		var like ReportLike
		if err := rows.Scan(&like.createdByUserID, &like.createdAt); err != nil {
			return fmt.Errorf("scanning report like: %w", err)
		}
		report.likes = append(report.likes, like)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("reading report likes: %w", err)
	}
	return nil
}

func (repo *ReportRepository) loadAlternatives(ctx context.Context, report *Report) error {
	rows, err := repo.pool.Query(ctx, `
		SELECT ra.id, ra.suggested_by_user_id, ra.brand_id, b.name,
		       ra.seller_id, s.name, ra.product_name,
		       COALESCE(ra.product_url, ''), ra.created_at
		FROM report_alternative ra
		JOIN brand b ON b.id = ra.brand_id
		JOIN seller s ON s.id = ra.seller_id
		WHERE ra.report_id = $1
		ORDER BY ra.created_at ASC, ra.id ASC
	`, report.id)
	if err != nil {
		return fmt.Errorf("loading report alternatives: %w", err)
	}
	defer rows.Close()

	report.alternatives = make([]ReportAlternative, 0)
	for rows.Next() {
		var alternative ReportAlternative
		var brandID, sellerID int64
		var brandName, sellerName string
		if err := rows.Scan(
			&alternative.id, &alternative.createdByUserID, &brandID, &brandName,
			&sellerID, &sellerName, &alternative.productName,
			&alternative.productURL, &alternative.createdAt,
		); err != nil {
			return fmt.Errorf("scanning report alternative: %w", err)
		}
		alternative.brand = &Brand{id: brandID, name: brandName}
		alternative.seller = &Seller{id: sellerID, name: sellerName}
		report.alternatives = append(report.alternatives, alternative)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("reading report alternatives: %w", err)
	}
	return nil
}

func (repo *ReportRepository) Create(ctx context.Context, report *Report) (int64, error) {
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("starting report transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var reportID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO report (
			seller_id, brand_id, category_id, submitted_by_user_id,
			product_name, product_url, description
		)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7)
		RETURNING id
	`, report.seller.id, report.brand.id, report.category.id, report.createdByUserID,
		report.productName, report.productURL, report.description).Scan(&reportID)
	if err != nil {
		return 0, fmt.Errorf("creating report: %w", err)
	}

	for sortOrder, image := range report.images {
		if _, err := tx.Exec(ctx, `
			INSERT INTO report_image (report_id, storage_path, image_url, sort_order)
			VALUES ($1, $2, $3, $4)
		`, reportID, image.storagePath, image.url, sortOrder+1); err != nil {
			return 0, fmt.Errorf("creating report image: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("committing report: %w", err)
	}
	return reportID, nil
}

func (repo *ReportRepository) AddComment(ctx context.Context, reportID int64, comment ReportComment) error {
	_, err := repo.pool.Exec(ctx, `
		INSERT INTO report_comment (report_id, user_id, body)
		VALUES ($1, $2, $3)
	`, reportID, comment.createdByUserID, comment.comment)
	if err != nil {
		return fmt.Errorf("creating report comment: %w", err)
	}
	return nil
}

func (repo *ReportRepository) AddLike(ctx context.Context, reportID int64, like ReportLike) error {
	_, err := repo.pool.Exec(ctx, `
		INSERT INTO report_like (report_id, user_id)
		VALUES ($1, $2)
	`, reportID, like.createdByUserID)
	if err != nil {
		return fmt.Errorf("creating report like: %w", err)
	}
	return nil
}

func (repo *ReportRepository) RemoveLike(ctx context.Context, reportID, userID int64) error {
	_, err := repo.pool.Exec(ctx, `
		DELETE FROM report_like WHERE report_id = $1 AND user_id = $2
	`, reportID, userID)
	if err != nil {
		return fmt.Errorf("removing report like: %w", err)
	}
	return nil
}

func (repo *ReportRepository) AddAlternative(ctx context.Context, reportID int64, alternative ReportAlternative) error {
	_, err := repo.pool.Exec(ctx, `
		INSERT INTO report_alternative (
			report_id, suggested_by_user_id, brand_id, seller_id, product_name, product_url
		)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''))
	`, reportID, alternative.createdByUserID, alternative.brand.id, alternative.seller.id,
		alternative.productName, alternative.productURL)
	if err != nil {
		return fmt.Errorf("creating report alternative: %w", err)
	}
	return nil
}
