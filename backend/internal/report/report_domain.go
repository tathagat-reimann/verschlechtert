package report

import (
	"errors"
	"strings"
	"time"
)

type Brand struct {
	id   int64
	name string
}

func NewBrand(name string) (*Brand, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("brand name cannot be empty")
	}
	return &Brand{name: name}, nil
}

func (b *Brand) GetID() int64 { return b.id }

func (b *Brand) GetName() string { return b.name }

type Seller struct {
	id   int64
	name string
}

func NewSeller(name string) (*Seller, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("seller name cannot be empty")
	}
	return &Seller{name: name}, nil
}

func (s *Seller) GetID() int64 { return s.id }

func (s *Seller) GetName() string { return s.name }

type Category struct {
	id   int64
	name string
}

func NewCategory(name string) (*Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("category name cannot be empty")
	}
	return &Category{name: name}, nil
}

func (c *Category) GetID() int64 { return c.id }

func (c *Category) GetName() string { return c.name }

type Report struct {
	id              int64
	productName     string
	description     string
	category        *Category
	brand           *Brand
	seller          *Seller
	productURL      string
	images          []ReportImage
	comments        []ReportComment
	likes           []ReportLike
	alternatives    []ReportAlternative
	createdByUserID int64
	createdAt       time.Time
}

func NewReport(productName, description string, category *Category, brand *Brand, seller *Seller,
	productURL string, createdByUserID int64) (*Report, error) {
	productName = strings.TrimSpace(productName)
	description = strings.TrimSpace(description)
	productURL = strings.TrimSpace(productURL)
	if productName == "" {
		return nil, errors.New("productName cannot be empty")
	}
	if description == "" {
		return nil, errors.New("description cannot be empty")
	}
	if category == nil {
		return nil, errors.New("category cannot be nil")
	}
	if brand == nil {
		return nil, errors.New("brand cannot be nil")
	}
	if seller == nil {
		return nil, errors.New("seller cannot be nil")
	}
	if createdByUserID == 0 {
		return nil, errors.New("createdByUserID cannot be zero")
	}

	return &Report{
		productName:     productName,
		description:     description,
		category:        category,
		brand:           brand,
		seller:          seller,
		productURL:      productURL,
		createdByUserID: createdByUserID,
	}, nil
}

func (r *Report) GetID() int64 { return r.id }

func (r *Report) GetProductName() string { return r.productName }

func (r *Report) GetDescription() string { return r.description }

func (r *Report) GetCategory() *Category { return r.category }

func (r *Report) GetBrand() *Brand { return r.brand }

func (r *Report) GetSeller() *Seller { return r.seller }

func (r *Report) GetProductURL() string { return r.productURL }

func (r *Report) GetImages() []ReportImage {
	return append([]ReportImage(nil), r.images...)
}

func (r *Report) GetComments() []ReportComment {
	return append([]ReportComment(nil), r.comments...)
}

func (r *Report) GetLikes() []ReportLike {
	return append([]ReportLike(nil), r.likes...)
}

func (r *Report) GetAlternatives() []ReportAlternative {
	return append([]ReportAlternative(nil), r.alternatives...)
}

func (r *Report) GetCreatedByUserID() int64 { return r.createdByUserID }

func (r *Report) GetCreatedAt() time.Time { return r.createdAt }

func (r *Report) AddImage(image ReportImage) error {
	if len(r.images) >= 5 {
		return errors.New("a report can contain at most 5 images")
	}
	if strings.TrimSpace(image.GetStoragePath()) == "" {
		return errors.New("storagePath cannot be empty")
	}
	if strings.TrimSpace(image.GetURL()) == "" {
		return errors.New("url cannot be empty")
	}

	for _, existingImage := range r.images {
		if existingImage.GetURL() == image.GetURL() {
			return errors.New("image already exists in report")
		}
	}

	r.images = append(r.images, image)
	return nil
}

func (r *Report) AddComment(comment ReportComment) error {
	if strings.TrimSpace(comment.GetComment()) == "" {
		return errors.New("comment cannot be empty")
	}

	r.comments = append(r.comments, comment)
	return nil
}

func (r *Report) AddLike(like ReportLike) error {
	for _, existingLike := range r.likes {
		if existingLike.GetCreatedByUserID() == like.GetCreatedByUserID() {
			return errors.New("user has already liked report")
		}
	}

	r.likes = append(r.likes, like)
	return nil
}

func (r *Report) RemoveLike(userID int64) error {
	for index, like := range r.likes {
		if like.GetCreatedByUserID() == userID {
			r.likes = append(r.likes[:index], r.likes[index+1:]...)
			return nil
		}
	}

	return errors.New("user has not liked report")
}

func (r *Report) AddAlternative(alternative ReportAlternative) error {
	for _, existingAlternative := range r.alternatives {
		if existingAlternative.GetCreatedByUserID() == alternative.GetCreatedByUserID() {
			return errors.New("user has already suggested an alternative")
		}
	}

	r.alternatives = append(r.alternatives, alternative)
	return nil
}

type ReportImage struct {
	id          int64
	storagePath string
	url         string
}

func NewReportImage(storagePath, url string) (*ReportImage, error) {
	storagePath = strings.TrimSpace(storagePath)
	url = strings.TrimSpace(url)
	if storagePath == "" {
		return nil, errors.New("storagePath cannot be empty")
	}
	if url == "" {
		return nil, errors.New("url cannot be empty")
	}
	return &ReportImage{storagePath: storagePath, url: url}, nil
}

func (r *ReportImage) GetID() int64 { return r.id }

func (r *ReportImage) GetStoragePath() string { return r.storagePath }

func (r *ReportImage) GetURL() string { return r.url }

type ReportComment struct {
	id              int64
	comment         string
	createdByUserID int64
	createdAt       time.Time
}

func NewReportComment(comment string, createdByUserID int64) (*ReportComment, error) {
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return nil, errors.New("comment cannot be empty")
	}
	if createdByUserID == 0 {
		return nil, errors.New("createdByUserID cannot be zero")
	}
	return &ReportComment{comment: comment, createdByUserID: createdByUserID}, nil
}

func (r *ReportComment) GetID() int64 { return r.id }

func (r *ReportComment) GetComment() string { return r.comment }

func (r *ReportComment) GetCreatedByUserID() int64 { return r.createdByUserID }

func (r *ReportComment) GetCreatedAt() time.Time { return r.createdAt }

type ReportLike struct {
	createdByUserID int64
	createdAt       time.Time
}

func NewReportLike(createdByUserID int64) (*ReportLike, error) {
	if createdByUserID == 0 {
		return nil, errors.New("createdByUserID cannot be zero")
	}
	return &ReportLike{createdByUserID: createdByUserID}, nil
}

func (r *ReportLike) GetCreatedByUserID() int64 { return r.createdByUserID }

func (r *ReportLike) GetCreatedAt() time.Time { return r.createdAt }

type ReportAlternative struct {
	id              int64
	productName     string
	brand           *Brand
	seller          *Seller
	productURL      string
	createdByUserID int64
	createdAt       time.Time
}

func NewReportAlternative(productName string, brand *Brand, seller *Seller, productURL string, createdByUserID int64) (*ReportAlternative, error) {
	productName = strings.TrimSpace(productName)
	productURL = strings.TrimSpace(productURL)
	if productName == "" {
		return nil, errors.New("productName cannot be empty")
	}
	if brand == nil {
		return nil, errors.New("brand cannot be nil")
	}
	if seller == nil {
		return nil, errors.New("seller cannot be nil")
	}
	if createdByUserID == 0 {
		return nil, errors.New("createdByUserID cannot be zero")
	}
	return &ReportAlternative{
		productName:     productName,
		brand:           brand,
		seller:          seller,
		productURL:      productURL,
		createdByUserID: createdByUserID,
	}, nil
}

func (r *ReportAlternative) GetID() int64 { return r.id }

func (r *ReportAlternative) GetProductName() string { return r.productName }

func (r *ReportAlternative) GetBrand() *Brand { return r.brand }

func (r *ReportAlternative) GetSeller() *Seller { return r.seller }

func (r *ReportAlternative) GetProductURL() string { return r.productURL }

func (r *ReportAlternative) GetCreatedByUserID() int64 { return r.createdByUserID }

func (r *ReportAlternative) GetCreatedAt() time.Time { return r.createdAt }
