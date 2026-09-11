package report

import "context"

type IReportRepository interface {
	Create(ctx context.Context, report *Report) (int64, error)
	GetActive(ctx context.Context, reportID int64) (*Report, error)
	GetActiveForUser(ctx context.Context, reportID, userID int64) (*Report, error)
	ListActive(ctx context.Context, limit, offset int) ([]*Report, error)
	ListActiveByUser(ctx context.Context, userID int64, limit, offset int) ([]*Report, error)
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

func (s *ReportService) GetActive(ctx context.Context, reportID int64) (*Report, error) {
	return s.reportRepo.GetActive(ctx, reportID)
}

func (s *ReportService) GetActiveForUser(ctx context.Context, reportID, userID int64) (*Report, error) {
	return s.reportRepo.GetActiveForUser(ctx, reportID, userID)
}

func (s *ReportService) ListActive(ctx context.Context, limit, offset int) ([]*Report, error) {
	return s.reportRepo.ListActive(ctx, limit, offset)
}

func (s *ReportService) ListActiveByUser(ctx context.Context, userID int64, limit, offset int) ([]*Report, error) {
	return s.reportRepo.ListActiveByUser(ctx, userID, limit, offset)
}

func (s *ReportService) AddComment(ctx context.Context, reportID, userID int64, comment string) error {
	report, err := s.reportRepo.GetActive(ctx, reportID)
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

	return s.reportRepo.AddComment(ctx, reportID, *newComment)
}

func (s *ReportService) AddLike(ctx context.Context, reportID, userID int64) error {
	report, err := s.reportRepo.GetActive(ctx, reportID)
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

	return s.reportRepo.AddLike(ctx, reportID, *like)
}

func (s *ReportService) RemoveLike(ctx context.Context, reportID, userID int64) error {
	report, err := s.reportRepo.GetActive(ctx, reportID)
	if err != nil {
		return err
	}
	if err := report.RemoveLike(userID); err != nil {
		return err
	}

	return s.reportRepo.RemoveLike(ctx, reportID, userID)
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
	report, err := s.reportRepo.GetActive(ctx, reportID)
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

	return s.reportRepo.AddAlternative(ctx, reportID, *alternative)
}
