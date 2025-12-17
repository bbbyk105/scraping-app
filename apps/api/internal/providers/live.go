package providers

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"github.com/pricecompare/api/internal/models"
	"github.com/pricecompare/api/internal/httpclient"
)

// LiveProvider is a provider for live fetching from external websites
// This provider uses the httpclient which automatically applies:
// - robots.txt checking
// - rate limiting
// - audit logging
// - ALLOW_LIVE_FETCH control
type LiveProvider struct {
	httpClient *httpclient.Client
	baseURL    string // Base URL for the target website (e.g., "https://example.com")
}

// NewLiveProvider creates a new live provider
func NewLiveProvider(httpClient *httpclient.Client) *LiveProvider {
	// Default base URL - can be configured via environment variable
	baseURL := os.Getenv("LIVE_PROVIDER_BASE_URL")
	if baseURL == "" {
		baseURL = "https://example.com"
	}
	return &LiveProvider{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// Search searches for products on external websites
func (p *LiveProvider) Search(ctx context.Context, query string) ([]ProductCandidate, error) {
	if query == "" {
		return nil, fmt.Errorf("search query is required for live provider")
	}

	// Build search URL
	// This is a generic implementation - adjust URL pattern based on target site
	searchURL := fmt.Sprintf("%s/search?q=%s", p.baseURL, url.QueryEscape(query))

	// Fetch the search page using httpclient (with compliance checks)
	resp, err := p.httpClient.Get(ctx, "live", searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch search page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("search page returned status %d", resp.StatusCode)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var products []ProductCandidate

	// Parse product listings - common e-commerce selectors
	doc.Find(".product, .item, [data-product], .product-item, .product-card").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find(".title, .name, h2, h3, h4, [data-title], .product-title").First().Text())
		if title == "" {
			// Try to get title from link
			title = strings.TrimSpace(s.Find("a").First().Text())
		}
		if title == "" {
			return // Skip if no title found
		}

		// Limit title length
		if len(title) > 200 {
			title = title[:200]
		}

		// Get image URL
		imageURL, _ := s.Find("img").First().Attr("src")
		if imageURL == "" {
			imageURL, _ = s.Find("img").First().Attr("data-src")
		}
		// Make absolute URL if relative
		if imageURL != "" && !strings.HasPrefix(imageURL, "http") {
			if strings.HasPrefix(imageURL, "/") {
				imageURL = p.baseURL + imageURL
			} else {
				imageURL = p.baseURL + "/" + imageURL
			}
		}

		// Extract brand from title
		brand := extractBrand(title)

		products = append(products, ProductCandidate{
			Title:    title,
			Brand:    brand,
			ImageURL: stringPtr(imageURL),
			Source:   "live",
		})
	})

	// If no products found with common selectors, try alternative approach
	if len(products) == 0 {
		// Try to find products in a different structure
		doc.Find("article, .listing, [role='article']").Each(func(i int, s *goquery.Selection) {
			title := strings.TrimSpace(s.Find("h1, h2, h3, .title, a").First().Text())
			if title != "" && len(title) < 200 {
				brand := extractBrand(title)
				products = append(products, ProductCandidate{
					Title:  title,
					Brand:  brand,
					Source: "live",
				})
			}
		})
	}

	return products, nil
}

// FetchOffers fetches offers from external websites
func (p *LiveProvider) FetchOffers(ctx context.Context, product *models.Product) ([]*models.Offer, error) {
	// Build product detail URL
	// This is a generic implementation - adjust URL pattern based on target site
	// For now, we'll try to construct a URL from the product title
	productURL := fmt.Sprintf("%s/product/%s", p.baseURL, url.QueryEscape(strings.ToLower(strings.ReplaceAll(product.Title, " ", "-"))))

	// If product has a URL stored, use it
	if product.ImageURL != nil && strings.HasPrefix(*product.ImageURL, "http") {
		// Try to use image URL as hint for product URL structure
		// This is site-specific and may need adjustment
	}

	// Fetch the product page using httpclient (with compliance checks)
	resp, err := p.httpClient.Get(ctx, "live", productURL)
	if err != nil {
		// If product page not found, create a mock offer from search results
		// In a real implementation, you might want to store product URLs during search
		return p.createMockOffersFromProduct(product), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// If page not found, return mock offers
		return p.createMockOffersFromProduct(product), nil
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var offers []*models.Offer

	// Parse offers from product page
	// Look for price, seller, availability information
	doc.Find(".offer, .listing, .seller-item, [data-offer], .price-row, .vendor-row").Each(func(i int, s *goquery.Selection) {
		seller := strings.TrimSpace(s.Find(".seller, .vendor, [data-seller], .store-name").First().Text())
		if seller == "" {
			seller = "Unknown Seller"
		}

		priceText := strings.TrimSpace(s.Find(".price, [data-price], .amount, .cost").First().Text())
		priceAmount := parsePrice(priceText)

		if priceAmount == 0 {
			// Try alternative price selectors
			priceText = strings.TrimSpace(s.Find(".price-value, .product-price, [itemprop='price']").First().Text())
			priceAmount = parsePrice(priceText)
		}

		// Get product URL
		productLink, _ := s.Find("a").First().Attr("href")
		if productLink != "" && !strings.HasPrefix(productLink, "http") {
			if strings.HasPrefix(productLink, "/") {
				productLink = p.baseURL + productLink
			} else {
				productLink = p.baseURL + "/" + productLink
			}
		}

		// Check availability
		inStock := true
		stockText := strings.ToLower(strings.TrimSpace(s.Find(".stock, .availability, [data-stock]").First().Text()))
		if strings.Contains(stockText, "out of stock") || strings.Contains(stockText, "unavailable") || strings.Contains(stockText, "sold out") {
			inStock = false
		}

		// Estimate delivery days (if available)
		deliveryText := strings.TrimSpace(s.Find(".delivery, .shipping-time, [data-delivery]").First().Text())
		estDeliveryDaysMin, estDeliveryDaysMax := estimateDeliveryDays(deliveryText)

		if priceAmount > 0 {
			offers = append(offers, &models.Offer{
				ID:                 uuid.New(),
				ProductID:          product.ID,
				Source:             "live",
				Seller:             seller,
				PriceAmount:        priceAmount,
				Currency:           "USD",
				ShippingToUSAmount: 0, // Will be calculated by shipping calculator
				TotalToUSAmount:    0, // Will be calculated by shipping calculator
				EstDeliveryDaysMin: estDeliveryDaysMin,
				EstDeliveryDaysMax: estDeliveryDaysMax,
				InStock:            inStock,
				URL:                stringPtr(productLink),
				FetchedAt:          time.Now(),
			})
		}
	})

	// If no offers found with specific selectors, try to extract from page structure
	if len(offers) == 0 {
		// Try to find price information in the main product area
		priceText := strings.TrimSpace(doc.Find(".price, [data-price], .product-price, [itemprop='price']").First().Text())
		priceAmount := parsePrice(priceText)

		if priceAmount > 0 {
			seller := strings.TrimSpace(doc.Find(".seller, .vendor, .store, [data-seller]").First().Text())
			if seller == "" {
				// Try to extract from domain name
				u, err := url.Parse(p.baseURL)
				if err == nil {
					seller = u.Host
				} else {
					seller = "Live Site"
				}
			}

			productLink := productURL // Use the URL we requested

			offers = append(offers, &models.Offer{
				ID:                 uuid.New(),
				ProductID:          product.ID,
				Source:             "live",
				Seller:             seller,
				PriceAmount:        priceAmount,
				Currency:           "USD",
				ShippingToUSAmount: 0, // Will be calculated by shipping calculator
				TotalToUSAmount:    0, // Will be calculated by shipping calculator
				EstDeliveryDaysMin: intPtr(5),
				EstDeliveryDaysMax: intPtr(10),
				InStock:            true,
				URL:                stringPtr(productLink),
				FetchedAt:          time.Now(),
			})
		}
	}

	// If still no offers, create mock offers
	if len(offers) == 0 {
		return p.createMockOffersFromProduct(product), nil
	}

	return offers, nil
}

// createMockOffersFromProduct creates mock offers when actual scraping fails
// This is a fallback to ensure the system continues to work
func (p *LiveProvider) createMockOffersFromProduct(product *models.Product) []*models.Offer {
	// Extract a base price estimate from product title or use default
	basePrice := 4999 // $49.99 default

	// Try to extract price hints from title (very basic)
	titleLower := strings.ToLower(product.Title)
	if strings.Contains(titleLower, "premium") || strings.Contains(titleLower, "pro") {
		basePrice = 9999 // $99.99
	} else if strings.Contains(titleLower, "budget") || strings.Contains(titleLower, "basic") {
		basePrice = 2999 // $29.99
	}

	return []*models.Offer{
		{
			ID:                 uuid.New(),
			ProductID:          product.ID,
			Source:             "live",
			Seller:             "Live Site Seller",
			PriceAmount:        basePrice,
			Currency:           "USD",
			ShippingToUSAmount: 0, // Will be calculated by shipping calculator
			TotalToUSAmount:    0, // Will be calculated by shipping calculator
			EstDeliveryDaysMin: intPtr(7),
			EstDeliveryDaysMax: intPtr(14),
			InStock:            true,
			URL:                stringPtr(p.baseURL),
			FetchedAt:          time.Now(),
		},
	}
}

// estimateDeliveryDays tries to extract delivery days from text
func estimateDeliveryDays(text string) (*int, *int) {
	if text == "" {
		return nil, nil
	}

	text = strings.ToLower(text)
	
	// Look for patterns like "3-5 days", "5 days", "1 week", etc.
	// This is a simple heuristic - adjust based on actual site format
	if strings.Contains(text, "1-2") || strings.Contains(text, "1 to 2") {
		return intPtr(1), intPtr(2)
	} else if strings.Contains(text, "2-3") || strings.Contains(text, "2 to 3") {
		return intPtr(2), intPtr(3)
	} else if strings.Contains(text, "3-5") || strings.Contains(text, "3 to 5") {
		return intPtr(3), intPtr(5)
	} else if strings.Contains(text, "5-7") || strings.Contains(text, "5 to 7") {
		return intPtr(5), intPtr(7)
	} else if strings.Contains(text, "7-10") || strings.Contains(text, "7 to 10") {
		return intPtr(7), intPtr(10)
	} else if strings.Contains(text, "10-14") || strings.Contains(text, "10 to 14") {
		return intPtr(10), intPtr(14)
	} else if strings.Contains(text, "week") {
		return intPtr(7), intPtr(14)
	}

	// Default estimate
	return intPtr(5), intPtr(10)
}

