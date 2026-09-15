package report

import (
	"context"
)

type IReportRepository interface {
	Create(ctx context.Context, report *Report) (int64, error)
	GetActive(ctx context.Context, locale string, reportID int64) (*Report, error)
	GetActiveForUser(ctx context.Context, locale string, reportID, userID int64) (*Report, error)
	ListActive(ctx context.Context, locale string, limit, offset int) ([]*Report, error)
	ListActiveByUser(ctx context.Context, locale string, userID int64, limit, offset int) ([]*Report, error)
	AddComment(ctx context.Context, reportID int64, comment ReportComment) error
	AddLike(ctx context.Context, reportID int64, like ReportLike) error
	RemoveLike(ctx context.Context, reportID, userID int64) error
	AddAlternative(ctx context.Context, reportID int64, alternative ReportAlternative) error
}

type ReportService struct {
	reportRepo IReportRepository
}

func NewReportService(reportRepo IReportRepository) *ReportService {
	return &ReportService{reportRepo: reportRepo}
}

func (s *ReportService) Create(
	ctx context.Context,
	productName, description string,
	category *Category,
	brand *Brand,
	seller *Seller,
	productURL string,
	createdByUserID int64,
	images []ReportImage,
) (int64, error) {
	report, err := NewReport(productName, description, category, brand, seller, productURL, createdByUserID)
	if err != nil {
		return 0, err
	}

	for _, image := range images {
		if err := report.AddImage(image); err != nil {
			return 0, err
		}
	}

	return s.reportRepo.Create(ctx, report)
}

func (s *ReportService) GetActive(ctx context.Context, locale string, reportID int64) (*Report, error) {
	report, err := s.reportRepo.GetActive(ctx, locale, reportID)
	if err != nil {
		return nil, err
	}
	return report, nil
}

func (s *ReportService) GetActiveForUser(ctx context.Context, locale string, reportID, userID int64) (*Report, error) {
	report, err := s.reportRepo.GetActiveForUser(ctx, locale, reportID, userID)
	if err != nil {
		return nil, err
	}
	return report, nil
}

func (s *ReportService) ListActive(ctx context.Context, locale string, limit, offset int) ([]*Report, error) {
	reports, err := s.reportRepo.ListActive(ctx, locale, limit, offset)
	if err != nil {
		return nil, err
	}
	return reports, nil
}

func (s *ReportService) ListActiveByUser(ctx context.Context, locale string, userID int64, limit, offset int) ([]*Report, error) {
	reports, err := s.reportRepo.ListActiveByUser(ctx, locale, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	return reports, nil
}

func (s *ReportService) AddComment(ctx context.Context, reportID, userID int64, comment string) error {
	report, err := s.reportRepo.GetActive(ctx, "de", reportID)
	if err != nil {
		return err
	}

	newComment, err := NewReportComment(comment, userID)
	if err != nil {
		return err
	}
	if err := report.AddComment(*newComment); err != nil {
		return err
	}

	if err := s.reportRepo.AddComment(ctx, reportID, *newComment); err != nil {
		return err
	}
	return nil
}

func (s *ReportService) AddLike(ctx context.Context, reportID, userID int64) error {
	report, err := s.reportRepo.GetActive(ctx, "de", reportID)
	if err != nil {
		return err
	}

	like, err := NewReportLike(userID)
	if err != nil {
		return err
	}
	if err := report.AddLike(*like); err != nil {
		return err
	}

	if err := s.reportRepo.AddLike(ctx, reportID, *like); err != nil {
		return err
	}
	return nil
}

func (s *ReportService) RemoveLike(ctx context.Context, reportID, userID int64) error {
	report, err := s.reportRepo.GetActive(ctx, "de", reportID)
	if err != nil {
		return err
	}
	if err := report.RemoveLike(userID); err != nil {
		return err
	}

	if err := s.reportRepo.RemoveLike(ctx, reportID, userID); err != nil {
		return err
	}
	return nil
}

func (s *ReportService) AddAlternative(
	ctx context.Context,
	reportID int64,
	productName string,
	brand *Brand,
	seller *Seller,
	productURL string,
	createdByUserID int64,
) error {
	report, err := s.reportRepo.GetActive(ctx, "de", reportID)
	if err != nil {
		return err
	}

	alternative, err := NewReportAlternative(productName, brand, seller, productURL, createdByUserID)
	if err != nil {
		return err
	}
	if err := report.AddAlternative(*alternative); err != nil {
		return err
	}

	if err := s.reportRepo.AddAlternative(ctx, reportID, *alternative); err != nil {
		return err
	}
	return nil
}
