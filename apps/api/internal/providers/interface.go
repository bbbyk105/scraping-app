package providers

import (
	"context"
	"github.com/pricecompare/api/internal/models"
)

// ProductCandidate represents a product found during search
type ProductCandidate struct {
	Title      string
	Brand      *string
	Model      *string
	ImageURL   *string
	Source     string
	Identifier *string // Optional identifier (e.g., itemId for Walmart, ASIN for Amazon)
	SourceURL  *string // Product URL from the source
}

// Provider interface for fetching product information
type Provider interface {
	// Search searches for products by query
	Search(ctx context.Context, query string) ([]ProductCandidate, error)

	// FetchOffers fetches offers for a product
	FetchOffers(ctx context.Context, product *models.Product) ([]*models.Offer, error)
}



