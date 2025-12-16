package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pricecompare/api/internal/models"
)

type DemoProvider struct {
	mockProducts []ProductCandidate
	mockOffers   map[string][]*models.Offer
}

func NewDemoProvider() *DemoProvider {
	provider := &DemoProvider{
		mockOffers: make(map[string][]*models.Offer),
	}

	// Initialize mock data
	provider.mockProducts = []ProductCandidate{
		{
			Title:    "Wireless Bluetooth Headphones",
			Brand:    stringPtr("AudioTech"),
			Model:    stringPtr("ATH-500BT"),
			ImageURL: stringPtr("https://example.com/images/headphones.jpg"),
			Source:   "demo",
		},
		{
			Title:    "Smart Watch Pro",
			Brand:    stringPtr("TechTime"),
			Model:    stringPtr("TT-SW-2024"),
			ImageURL: stringPtr("https://example.com/images/watch.jpg"),
			Source:   "demo",
		},
		{
			Title:    "USB-C Charging Cable",
			Brand:    stringPtr("ChargeMax"),
			ImageURL: stringPtr("https://example.com/images/cable.jpg"),
			Source:   "demo",
		},
	}

	return provider
}

func (p *DemoProvider) Search(ctx context.Context, query string) ([]ProductCandidate, error) {
	// Simple mock search - return all products if query matches any word
	results := []ProductCandidate{}
	queryLower := toLower(query)

	for _, product := range p.mockProducts {
		if contains(toLower(product.Title), queryLower) {
			results = append(results, product)
		}
	}

	return results, nil
}

func (p *DemoProvider) FetchOffers(ctx context.Context, product *models.Product) ([]*models.Offer, error) {
	// Generate mock offers based on product
	offers := []*models.Offer{
		{
			ID:                 uuid.New(),
			ProductID:          product.ID,
			Source:             "demo",
			Seller:             "DemoSeller A",
			PriceAmount:        4999, // $49.99
			Currency:           "USD",
			ShippingToUSAmount: 999,  // $9.99
			TotalToUSAmount:    5998, // $59.98
			EstDeliveryDaysMin: intPtr(3),
			EstDeliveryDaysMax: intPtr(5),
			InStock:            true,
			URL:                stringPtr("https://example.com/seller-a/product"),
			FetchedAt:          time.Now(),
		},
		{
			ID:                 uuid.New(),
			ProductID:          product.ID,
			Source:             "demo",
			Seller:             "DemoSeller B",
			PriceAmount:        5499, // $54.99
			Currency:           "USD",
			ShippingToUSAmount: 1499, // $14.99
			TotalToUSAmount:    6998, // $69.98
			EstDeliveryDaysMin: intPtr(5),
			EstDeliveryDaysMax: intPtr(7),
			InStock:            true,
			URL:                stringPtr("https://example.com/seller-b/product"),
			FetchedAt:          time.Now(),
		},
		{
			ID:                 uuid.New(),
			ProductID:          product.ID,
			Source:             "demo",
			Seller:             "DemoSeller C",
			PriceAmount:        4799, // $47.99
			Currency:           "USD",
			ShippingToUSAmount: 1999, // $19.99
			TotalToUSAmount:    6798, // $67.98
			EstDeliveryDaysMin: intPtr(7),
			EstDeliveryDaysMax: intPtr(10),
			InStock:            false,
			URL:                stringPtr("https://example.com/seller-c/product"),
			FetchedAt:          time.Now(),
		},
	}

	return offers, nil
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func toLower(s string) string {
	lower := ""
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			lower += string(r + 32)
		} else {
			lower += string(r)
		}
	}
	return lower
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

