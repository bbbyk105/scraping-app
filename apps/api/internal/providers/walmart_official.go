package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pricecompare/api/internal/models"
	"github.com/pricecompare/api/internal/httpclient"
)

// WalmartOfficialProvider implements Walmart Data API
type WalmartOfficialProvider struct {
	httpClient *httpclient.Client
	apiKey     string
	apiBaseURL string
	apiHost    string
	enabled    bool
}

// NewWalmartOfficialProvider creates a new Walmart official API provider
func NewWalmartOfficialProvider(httpClient *httpclient.Client) *WalmartOfficialProvider {
	apiKey := os.Getenv("WALMART_API_KEY")
	apiBaseURL := os.Getenv("WALMART_API_BASE_URL")
	if apiBaseURL == "" {
		// Default to RapidAPI Walmart endpoint
		apiBaseURL = "https://walmart-data.p.rapidapi.com"
	}
	apiHost := os.Getenv("WALMART_API_HOST")
	if apiHost == "" {
		// Extract host from base URL if not provided
		if apiBaseURL != "" {
			u, err := url.Parse(apiBaseURL)
			if err == nil && u.Host != "" {
				apiHost = u.Host
			} else {
				apiHost = "walmart-data.p.rapidapi.com"
			}
		} else {
			apiHost = "walmart-data.p.rapidapi.com"
		}
	}

	enabled := apiKey != ""

	return &WalmartOfficialProvider{
		httpClient: httpClient,
		apiKey:     apiKey,
		apiBaseURL: apiBaseURL,
		apiHost:    apiHost,
		enabled:    enabled,
	}
}

// IsEnabled returns whether the provider is enabled (has API key)
func (p *WalmartOfficialProvider) IsEnabled() bool {
	return p.enabled
}

// Search searches for products using Walmart API
func (p *WalmartOfficialProvider) Search(ctx context.Context, query string) ([]ProductCandidate, error) {
	if !p.enabled {
		return nil, fmt.Errorf("Walmart API provider is not enabled (WALMART_API_KEY not set)")
	}

	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	// Build API URL - Walmart Data API format
	// RapidAPI endpoint: /search (based on "Search (New)" endpoint)
	apiPath := os.Getenv("WALMART_API_PATH")
	if apiPath == "" {
		apiPath = "/search" // Default path
	}
	searchURL := fmt.Sprintf("%s%s?q=%s", p.apiBaseURL, apiPath, url.QueryEscape(query))

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers - RapidAPI format
	if p.apiKey != "" {
		req.Header.Set("X-RapidAPI-Key", p.apiKey)
		req.Header.Set("X-RapidAPI-Host", p.apiHost)
	}
	req.Header.Set("User-Agent", "PriceCompareBot/1.0")
	req.Header.Set("Accept", "application/json")

	// For API endpoints, we use direct HTTP client (robots.txt check is not needed for API)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from Walmart API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		// Log detailed error for debugging
		return nil, fmt.Errorf("Walmart API returned status %d for URL %s: %s", resp.StatusCode, searchURL, string(body))
	}

	// Parse response - RapidAPI Walmart Data API format
	var apiResponse struct {
		SearchTerms     string `json:"searchTerms"`
		AggregatedCount int    `json:"aggregatedCount"`
		SearchResult    [][]struct {
			Name         string `json:"name"`
			Image        string `json:"image"`
			Price        float64 `json:"price"`
			PriceInfo    struct {
				LinePrice string `json:"linePrice"`
				MinPrice  float64 `json:"minPrice"`
			} `json:"priceInfo"`
			ProductLink                  string `json:"productLink"`
			AvailabilityStatusDisplayValue string `json:"availabilityStatusDisplayValue"`
			IsOutOfStock                bool   `json:"isOutOfStock"`
			FulfillmentBadgeGroups      []struct {
				Text    string `json:"text"`
				SlaText string `json:"slaText"`
			} `json:"fulfillmentBadgeGroups"`
		} `json:"searchResult"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse Walmart API response: %w", err)
	}

	// Convert to ProductCandidate
	// searchResult is a 2D array, first element contains the products
	candidates := make([]ProductCandidate, 0)
	if len(apiResponse.SearchResult) > 0 {
		for _, item := range apiResponse.SearchResult[0] {
			if item.Name == "" {
				continue // Skip empty entries (ads, etc.)
			}
			// Extract itemId from Walmart URL
			// Format: https://www.walmart.com/ip/.../5461164337?...
			itemId := extractWalmartItemId(item.ProductLink)
			candidates = append(candidates, ProductCandidate{
				Title:      item.Name,
				ImageURL:   stringPtr(item.Image),
				Source:     "walmart",
				Identifier: itemId,
				SourceURL:  stringPtr(item.ProductLink),
			})
		}
	}

	return candidates, nil
}

// FetchOffers fetches offers for a product using Walmart Data API
func (p *WalmartOfficialProvider) FetchOffers(ctx context.Context, product *models.Product) ([]*models.Offer, error) {
	if !p.enabled {
		return nil, fmt.Errorf("Walmart API provider is not enabled (WALMART_API_KEY not set)")
	}

	// Search for the product to get item details
	// We'll use search results to get price information directly
	// If more detailed info is needed, we can use product-details API
	candidates, err := p.Search(ctx, product.Title)
	if err != nil {
		return nil, fmt.Errorf("failed to search for product: %w", err)
	}

	if len(candidates) == 0 {
		return []*models.Offer{}, nil
	}

	// Use search results to create offers (search API already provides price info)
	return p.createOffersFromSearch(ctx, product, candidates)
}

// createOffersFromSearch creates offers from search results when detailed item fetch fails
func (p *WalmartOfficialProvider) createOffersFromSearch(ctx context.Context, product *models.Product, candidates []ProductCandidate) ([]*models.Offer, error) {
	// Re-search to get price information from search results
	apiPath := os.Getenv("WALMART_API_PATH")
	if apiPath == "" {
		apiPath = "/search" // Default path
	}
	searchURL := fmt.Sprintf("%s%s?q=%s", p.apiBaseURL, apiPath, url.QueryEscape(product.Title))
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers - RapidAPI format
	if p.apiKey != "" {
		req.Header.Set("X-RapidAPI-Key", p.apiKey)
		req.Header.Set("X-RapidAPI-Host", p.apiHost)
	}
	req.Header.Set("User-Agent", "PriceCompareBot/1.0")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch search results: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []*models.Offer{}, nil
	}

	var searchResponse struct {
		SearchResult [][]struct {
			Name         string `json:"name"`
			Image        string `json:"image"`
			Price        float64 `json:"price"`
			PriceInfo    struct {
				LinePrice string `json:"linePrice"`
				MinPrice  float64 `json:"minPrice"`
			} `json:"priceInfo"`
			ProductLink                  string `json:"productLink"`
			AvailabilityStatusDisplayValue string `json:"availabilityStatusDisplayValue"`
			IsOutOfStock                bool   `json:"isOutOfStock"`
			FulfillmentBadgeGroups      []struct {
				Text    string `json:"text"`
				SlaText string `json:"slaText"`
			} `json:"fulfillmentBadgeGroups"`
		} `json:"searchResult"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []*models.Offer{}, nil
	}

	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return []*models.Offer{}, nil
	}

	// Find matching product by title similarity
	// searchResult is a 2D array, first element contains the products
	var matchedProduct *struct {
		Name         string `json:"name"`
		Image        string `json:"image"`
		Price        float64 `json:"price"`
		PriceInfo    struct {
			LinePrice string `json:"linePrice"`
			MinPrice  float64 `json:"minPrice"`
		} `json:"priceInfo"`
		ProductLink                  string `json:"productLink"`
		AvailabilityStatusDisplayValue string `json:"availabilityStatusDisplayValue"`
		IsOutOfStock                bool   `json:"isOutOfStock"`
		FulfillmentBadgeGroups      []struct {
			Text    string `json:"text"`
			SlaText string `json:"slaText"`
		} `json:"fulfillmentBadgeGroups"`
	}

	if len(searchResponse.SearchResult) > 0 {
		for i := range searchResponse.SearchResult[0] {
			item := &searchResponse.SearchResult[0][i]
			if item.Name == "" {
				continue // Skip empty entries
			}
			if strings.Contains(strings.ToLower(item.Name), strings.ToLower(product.Title)) ||
				strings.Contains(strings.ToLower(product.Title), strings.ToLower(item.Name)) {
				matchedProduct = item
				break
			}
		}
	}

	if matchedProduct == nil && len(searchResponse.SearchResult) > 0 && len(searchResponse.SearchResult[0]) > 0 {
		// Use first valid product if no match found
		for i := range searchResponse.SearchResult[0] {
			if searchResponse.SearchResult[0][i].Name != "" {
				matchedProduct = &searchResponse.SearchResult[0][i]
				break
			}
		}
	}

	if matchedProduct == nil {
		return []*models.Offer{}, nil
	}

	// Parse price - use minPrice if available, otherwise use price
	priceFloat := matchedProduct.Price
	if matchedProduct.PriceInfo.MinPrice > 0 {
		priceFloat = matchedProduct.PriceInfo.MinPrice
	}
	priceAmount := int(priceFloat * 100) // Convert to cents

	// Parse shipping message for delivery days from fulfillmentBadgeGroups
	shippingMessage := ""
	if len(matchedProduct.FulfillmentBadgeGroups) > 0 {
		for _, badge := range matchedProduct.FulfillmentBadgeGroups {
			if badge.SlaText != "" {
				shippingMessage = badge.Text + badge.SlaText
				break
			}
		}
	}
	estMinDays, estMaxDays := estimateDeliveryDaysFromShipping(shippingMessage)

	// Determine availability status
	availabilityStatus := "unknown"
	if matchedProduct.IsOutOfStock {
		availabilityStatus = "out_of_stock"
	} else if matchedProduct.AvailabilityStatusDisplayValue != "" {
		if strings.Contains(strings.ToLower(matchedProduct.AvailabilityStatusDisplayValue), "in stock") {
			availabilityStatus = "in_stock"
		} else if strings.Contains(strings.ToLower(matchedProduct.AvailabilityStatusDisplayValue), "out of stock") {
			availabilityStatus = "out_of_stock"
		}
	} else {
		availabilityStatus = "in_stock" // Default to in_stock if not specified
	}

	now := time.Now()
	offer := &models.Offer{
		ID:                 uuid.New(),
		ProductID:          product.ID,
		Source:             "walmart",
		Seller:             "Walmart",
		PriceAmount:        priceAmount,
		Currency:           "USD", // Walmart prices are in USD
		ShippingToUSAmount: 0,
		TotalToUSAmount:    0,
		EstDeliveryDaysMin: estMinDays,
		EstDeliveryDaysMax: estMaxDays,
		InStock:            !matchedProduct.IsOutOfStock,
		AvailabilityStatus: stringPtr(availabilityStatus),
		URL:                stringPtr(matchedProduct.ProductLink),
		PriceUpdatedAt:     now,
		FetchedAt:          now,
	}

	return []*models.Offer{offer}, nil
}

// estimateDeliveryDaysFromShipping estimates delivery days from shipping message
func estimateDeliveryDaysFromShipping(shippingMsg string) (*int, *int) {
	shippingLower := strings.ToLower(shippingMsg)
	
	if strings.Contains(shippingLower, "arrives today") || strings.Contains(shippingLower, "delivery today") {
		return intPtr(0), intPtr(1)
	} else if strings.Contains(shippingLower, "arrives tomorrow") || strings.Contains(shippingLower, "delivery tomorrow") {
		return intPtr(1), intPtr(2)
	} else if strings.Contains(shippingLower, "arrives in 2 days") {
		return intPtr(2), intPtr(3)
	} else if strings.Contains(shippingLower, "arrives in 3+ days") {
		return intPtr(3), intPtr(7)
	} else if strings.Contains(shippingLower, "free shipping") {
		return intPtr(3), intPtr(7)
	}
	
	// Default estimate
	return intPtr(3), intPtr(7)
}

// extractWalmartItemId extracts itemId from Walmart product URL
// Format: https://www.walmart.com/ip/.../5461164337?...
func extractWalmartItemId(urlStr string) *string {
	if urlStr == "" {
		return nil
	}
	// Parse URL to extract path
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil
	}
	// Path format: /ip/.../5461164337
	// Extract the last numeric segment before query params
	pathParts := strings.Split(strings.Trim(u.Path, "/"), "/")
	for i := len(pathParts) - 1; i >= 0; i-- {
		part := pathParts[i]
		// Check if this part is numeric (itemId)
		if len(part) > 0 {
			// Try to find numeric segment (itemId is typically 8-10 digits)
			if matched := strings.TrimSpace(part); len(matched) >= 6 {
				// Check if it's numeric
				isNumeric := true
				for _, r := range matched {
					if r < '0' || r > '9' {
						isNumeric = false
						break
					}
				}
				if isNumeric {
					return &matched
				}
			}
		}
	}
	return nil
}

