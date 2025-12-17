package providers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"github.com/pricecompare/api/internal/models"
)

type PublicHTMLProvider struct {
	samplesDir string
	userAgent  string
}

func NewPublicHTMLProvider(userAgent string) *PublicHTMLProvider {
	return &PublicHTMLProvider{
		samplesDir: "/app/samples",
		userAgent:  userAgent,
	}
}

func (p *PublicHTMLProvider) Search(ctx context.Context, query string) ([]ProductCandidate, error) {
	// In MVP, we read from sample HTML files
	// List all HTML files in samples directory
	files, err := filepath.Glob(filepath.Join(p.samplesDir, "*.html"))
	if err != nil {
		return nil, fmt.Errorf("failed to list sample files: %w", err)
	}

	var candidates []ProductCandidate
	for _, file := range files {
		products, err := p.parseHTMLFile(file)
		if err != nil {
			continue // Skip files that fail to parse
		}

		// Filter by query
		queryLower := strings.ToLower(query)
		for _, product := range products {
			titleLower := strings.ToLower(product.Title)
			if strings.Contains(titleLower, queryLower) {
				candidates = append(candidates, product)
			}
		}
	}

	return candidates, nil
}

func (p *PublicHTMLProvider) FetchOffers(ctx context.Context, product *models.Product) ([]*models.Offer, error) {
	// Search for offers in all sample files
	files, err := filepath.Glob(filepath.Join(p.samplesDir, "*.html"))
	if err != nil {
		return nil, fmt.Errorf("failed to list sample files: %w", err)
	}

	var allOffers []*models.Offer
	for _, file := range files {
		offers, err := p.parseOffersFromHTML(file, product.ID)
		if err != nil {
			continue // Skip files that fail to parse
		}
		allOffers = append(allOffers, offers...)
	}

	return allOffers, nil
}

func (p *PublicHTMLProvider) parseHTMLFile(filepath string) ([]ProductCandidate, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		return nil, err
	}

	var products []ProductCandidate

	// Parse based on common e-commerce HTML structure
	// Looking for product listings with title, price, etc.
	doc.Find(".product, .item, [data-product]").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find(".title, .name, h2, h3, [data-title]").First().Text())
		if title == "" {
			title = strings.TrimSpace(s.Text())
			if len(title) > 100 {
				title = title[:100]
			}
		}

		imageURL, _ := s.Find("img").First().Attr("src")
		if imageURL == "" {
			imageURL, _ = s.Find("img").First().Attr("data-src")
		}

		if title != "" {
			brand := extractBrand(title)
			products = append(products, ProductCandidate{
				Title:    title,
				Brand:    brand,
				ImageURL: stringPtr(imageURL),
				Source:   "public_html",
			})
		}
	})

	// If no products found with common selectors, try to extract from page structure
	if len(products) == 0 {
		title := strings.TrimSpace(doc.Find("title, h1").First().Text())
		if title != "" {
			products = append(products, ProductCandidate{
				Title:  title,
				Source: "public_html",
			})
		}
	}

	return products, nil
}

func (p *PublicHTMLProvider) parseOffersFromHTML(filepath string, productID uuid.UUID) ([]*models.Offer, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		return nil, err
	}

	var offers []*models.Offer

	// Parse offers from HTML
	doc.Find(".offer, .listing, .seller-item, [data-offer]").Each(func(i int, s *goquery.Selection) {
		seller := strings.TrimSpace(s.Find(".seller, .vendor, [data-seller]").First().Text())
		if seller == "" {
			seller = "Unknown Seller"
		}

		priceText := strings.TrimSpace(s.Find(".price, [data-price], .amount").First().Text())
		priceAmount := parsePrice(priceText)

		url, _ := s.Find("a").First().Attr("href")
		if url != "" && !strings.HasPrefix(url, "http") {
			url = "https://example.com" + url
		}

		if priceAmount > 0 {
			// Estimate shipping (will be recalculated by shipping calculator)
			shipping := estimateShippingFromPrice(priceAmount)

			offers = append(offers, &models.Offer{
				ID:                 uuid.New(),
				ProductID:          productID,
				Source:             "public_html",
				Seller:             seller,
				PriceAmount:        priceAmount,
				Currency:           "USD",
				ShippingToUSAmount: shipping,
				TotalToUSAmount:    priceAmount + shipping,
				EstDeliveryDaysMin: intPtr(5),
				EstDeliveryDaysMax: intPtr(10),
				InStock:            true,
				URL:                stringPtr(url),
				FetchedAt:          time.Now(),
			})
		}
	})

	// If no offers found with common selectors, try to create one from the page
	if len(offers) == 0 {
		priceText := strings.TrimSpace(doc.Find(".price, [data-price], .cost").First().Text())
		priceAmount := parsePrice(priceText)

		if priceAmount > 0 {
			offers = append(offers, &models.Offer{
				ID:                 uuid.New(),
				ProductID:          productID,
				Source:             "public_html",
				Seller:             "Sample Site",
				PriceAmount:        priceAmount,
				Currency:           "USD",
				ShippingToUSAmount: estimateShippingFromPrice(priceAmount),
				TotalToUSAmount:    priceAmount + estimateShippingFromPrice(priceAmount),
				EstDeliveryDaysMin: intPtr(7),
				EstDeliveryDaysMax: intPtr(14),
				InStock:            true,
				URL:                stringPtr("https://example.com/product"),
				FetchedAt:          time.Now(),
			})
		}
	}

	return offers, nil
}

func parsePrice(text string) int {
	// Remove currency symbols and whitespace
	text = strings.ReplaceAll(text, "$", "")
	text = strings.ReplaceAll(text, "USD", "")
	text = strings.ReplaceAll(text, ",", "")
	text = strings.TrimSpace(text)

	// Try to parse as float
	if price, err := strconv.ParseFloat(text, 64); err == nil {
		return int(price * 100) // Convert to cents
	}

	return 0
}

func estimateShippingFromPrice(priceCents int) int {
	priceUSD := float64(priceCents) / 100.0
	if priceUSD < 20.0 {
		return 999 // $9.99
	} else if priceUSD < 50.0 {
		return 1499 // $14.99
	}
	return 1999 // $19.99
}

func extractBrand(title string) *string {
	// Simple heuristic: first word might be brand
	// Only return brand if there are multiple words (single word products are not brands)
	parts := strings.Fields(title)
	if len(parts) > 1 {
		brand := parts[0]
		if len(brand) > 2 && len(brand) < 20 {
			return &brand
		}
	}
	return nil
}

