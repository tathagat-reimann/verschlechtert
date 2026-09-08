package domain

type Brand struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Seller struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Report struct {
	ID           int64               `json:"id"`
	ProductName  string              `json:"productName"`
	Category     *Category           `json:"category,omitempty"`
	Brand        *Brand              `json:"brand,omitempty"`
	Seller       *Seller             `json:"seller,omitempty"`
	ProductURL   string              `json:"productUrl"`
	Images       []ReportImage       `json:"images"`
	Comments     []ReportComment     `json:"comments"`
	Likes        []ReportLike        `json:"likes"`
	Alternatives []ReportAlternative `json:"alternatives"`
	CreatedBy    *User               `json:"createdBy"`
	CreatedAt    string              `json:"createdAt"`
}

type ReportImage struct {
	ID  int64  `json:"id"`
	URL string `json:"url"`
}

type ReportComment struct {
	ID        int64  `json:"id"`
	Comment   string `json:"content"`
	CreatedBy *User  `json:"createdBy"`
	CreatedAt string `json:"createdAt"`
}

type ReportLike struct {
	ID        int64  `json:"id"`
	CreatedBy *User  `json:"createdBy"`
	CreatedAt string `json:"createdAt"`
}

type ReportAlternative struct {
	ID          int64  `json:"id"`
	ProductName string `json:"productName"`
	Brand       string `json:"brand"`
	Seller      string `json:"seller"`
	ProductURL  string `json:"productUrl"`
	CreatedBy   *User  `json:"suggestedBy"`
	CreatedAt   string `json:"createdAt"`
}
