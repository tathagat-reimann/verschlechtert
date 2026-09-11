package report

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newReportForDomainTest(t *testing.T) *Report {
	t.Helper()

	brand := &Brand{id: 1, name: "brand"}
	seller := &Seller{id: 2, name: "seller"}
	category := &Category{id: 3, name: "category"}
	report, err := NewReport("product", "description", category, brand, seller, "", 10)
	require.NoError(t, err)
	return report
}

func TestNewReport_TrimsValuesAndInitializesEmptyChildren(t *testing.T) {
	brand := &Brand{id: 1, name: "brand"}
	seller := &Seller{id: 2, name: "seller"}
	category := &Category{id: 3, name: "category"}

	report, err := NewReport(" product ", " description ", category, brand, seller, " product-url ", 10)

	require.NoError(t, err)
	assert.Equal(t, "product", report.GetProductName())
	assert.Equal(t, "description", report.GetDescription())
	assert.Equal(t, "product-url", report.GetProductURL())
	assert.Equal(t, int64(10), report.GetCreatedByUserID())
	assert.Empty(t, report.GetImages())
	assert.Empty(t, report.GetComments())
	assert.Empty(t, report.GetLikes())
	assert.Empty(t, report.GetAlternatives())
}

func TestNewReport_ValidationErrors(t *testing.T) {
	brand := &Brand{id: 1, name: "brand"}
	seller := &Seller{id: 2, name: "seller"}
	category := &Category{id: 3, name: "category"}

	tests := []struct {
		name string
		make func() (*Report, error)
	}{
		{
			name: "missing product name",
			make: func() (*Report, error) {
				return NewReport("", "description", category, brand, seller, "", 10)
			},
		},
		{
			name: "missing description",
			make: func() (*Report, error) {
				return NewReport("product", "", category, brand, seller, "", 10)
			},
		},
		{
			name: "missing category",
			make: func() (*Report, error) {
				return NewReport("product", "description", nil, brand, seller, "", 10)
			},
		},
		{
			name: "missing brand",
			make: func() (*Report, error) {
				return NewReport("product", "description", category, nil, seller, "", 10)
			},
		},
		{
			name: "missing seller",
			make: func() (*Report, error) {
				return NewReport("product", "description", category, brand, nil, "", 10)
			},
		},
		{
			name: "missing creator",
			make: func() (*Report, error) {
				return NewReport("product", "description", category, brand, seller, "", 0)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			report, err := test.make()
			assert.Nil(t, report)
			assert.Error(t, err)
		})
	}
}

func TestReport_AddImage_LimitsAndRejectsDuplicates(t *testing.T) {
	report := newReportForDomainTest(t)

	for index := 0; index < 5; index++ {
		image, err := NewReportImage("path/"+string(rune('a'+index)), "https://example.com/"+string(rune('a'+index)))
		require.NoError(t, err)
		require.NoError(t, report.AddImage(*image))
	}

	duplicate, err := NewReportImage("path/duplicate", "https://example.com/a")
	require.NoError(t, err)
	assert.Error(t, report.AddImage(*duplicate))

	sixth, err := NewReportImage("path/sixth", "https://example.com/sixth")
	require.NoError(t, err)
	assert.Error(t, report.AddImage(*sixth))
}

func TestReport_AddCommentLikeAndAlternativeRules(t *testing.T) {
	report := newReportForDomainTest(t)

	comment, err := NewReportComment("comment", 10)
	require.NoError(t, err)
	require.NoError(t, report.AddComment(*comment))

	like, err := NewReportLike(10)
	require.NoError(t, err)
	require.NoError(t, report.AddLike(*like))
	assert.Error(t, report.AddLike(*like))
	require.NoError(t, report.RemoveLike(10))
	assert.Error(t, report.RemoveLike(10))

	brand := &Brand{id: 1, name: "brand"}
	seller := &Seller{id: 2, name: "seller"}
	alternative, err := NewReportAlternative("alternative", brand, seller, "", 10)
	require.NoError(t, err)
	require.NoError(t, report.AddAlternative(*alternative))
	assert.Error(t, report.AddAlternative(*alternative))
}

func TestReport_CollectionGettersReturnCopies(t *testing.T) {
	report := newReportForDomainTest(t)
	image, err := NewReportImage("path/image", "https://example.com/image")
	require.NoError(t, err)
	require.NoError(t, report.AddImage(*image))

	images := report.GetImages()
	images[0] = ReportImage{}

	assert.NotEmpty(t, report.GetImages()[0].GetURL())
}
