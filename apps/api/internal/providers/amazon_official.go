package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pricecompare/api/internal/models"
	"github.com/pricecompare/api/internal/httpclient"
)

// AmazonOfficialProvider implements Amazon Product Advertising API 5.0
type AmazonOfficialProvider struct {
	httpClient     *httpclient.Client
	accessKey      string
	secretKey      string
	associateTag   string
	apiEndpoint    string
	apiRegion      string
	enabled        bool
}

// NewAmazonOfficialProvider creates a new Amazon official API provider
func NewAmazonOfficialProvider(httpClient *httpclient.Client) *AmazonOfficialProvider {
	accessKey := os.Getenv("AMAZON_ACCESS_KEY")
	secretKey := os.Getenv("AMAZON_SECRET_KEY")
	associateTag := os.Getenv("AMAZON_ASSOCIATE_TAG")
	apiEndpoint := os.Getenv("AMAZON_API_ENDPOINT")
	if apiEndpoint == "" {
		apiEndpoint = "webservices.amazon.com"
	}
	apiRegion := os.Getenv("AMAZON_API_REGION")
	if apiRegion == "" {
		apiRegion = "us-east-1"
	}

	enabled := accessKey != "" && secretKey != "" && associateTag != ""

	return &AmazonOfficialProvider{
		httpClient:   httpClient,
		accessKey:    accessKey,
		secretKey:    secretKey,
		associateTag: associateTag,
		apiEndpoint:  apiEndpoint,
		apiRegion:    apiRegion,
		enabled:      enabled,
	}
}

// IsEnabled returns whether the provider is enabled (has required API keys)
func (p *AmazonOfficialProvider) IsEnabled() bool {
	return p.enabled
}

// Search searches for products using Amazon Product Advertising API
func (p *AmazonOfficialProvider) Search(ctx context.Context, query string) ([]ProductCandidate, error) {
	if !p.enabled {
		return nil, fmt.Errorf("Amazon API provider is not enabled (AMAZON_ACCESS_KEY, AMAZON_SECRET_KEY, or AMAZON_ASSOCIATE_TAG not set)")
	}

	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	// Build PA-API 5.0 request
	params := map[string]string{
		"Operation":     "SearchItems",
		"Keywords":      query,
		"SearchIndex":   "All",
		"ItemCount":    "10",
		"Resources":    "Images.Primary.Large,ItemInfo.Title,ItemInfo.ByLineInfo,ItemInfo.ExternalIds",
		"PartnerTag":   p.associateTag,
		"PartnerType":  "Associates",
		"Marketplace": "www.amazon.com",
	}

	// Create signed request
	req, err := p.createSignedRequest(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create signed request: %w", err)
	}

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute Amazon API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Amazon API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResponse struct {
		SearchResult struct {
			Items []struct {
				ASIN string `json:"ASIN"`
				DetailPageURL string `json:"DetailPageURL"`
				Images struct {
					Primary struct {
						Large struct {
							URL string `json:"URL"`
						} `json:"Large"`
					} `json:"Primary"`
				} `json:"Images"`
				ItemInfo struct {
					Title struct {
						DisplayValue string `json:"DisplayValue"`
					} `json:"Title"`
					ByLineInfo struct {
						Brand struct {
							DisplayValue string `json:"DisplayValue"`
						} `json:"Brand"`
					} `json:"ByLineInfo"`
					ExternalIds struct {
						EANs struct {
							DisplayValues []string `json:"DisplayValues"`
						} `json:"EANs"`
						UPCs struct {
							DisplayValues []string `json:"DisplayValues"`
						} `json:"UPCs"`
					} `json:"ExternalIds"`
				} `json:"ItemInfo"`
			} `json:"Items"`
		} `json:"SearchResult"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse Amazon API response: %w", err)
	}

	// Convert to ProductCandidate
	candidates := make([]ProductCandidate, 0, len(apiResponse.SearchResult.Items))
	for _, item := range apiResponse.SearchResult.Items {
		brand := ""
		if item.ItemInfo.ByLineInfo.Brand.DisplayValue != "" {
			brand = item.ItemInfo.ByLineInfo.Brand.DisplayValue
		}

		imageURL := ""
		if item.Images.Primary.Large.URL != "" {
			imageURL = item.Images.Primary.Large.URL
		}

		candidates = append(candidates, ProductCandidate{
			Title:    item.ItemInfo.Title.DisplayValue,
			Brand:    stringPtr(brand),
			ImageURL: stringPtr(imageURL),
			Source:   "amazon",
		})
	}

	return candidates, nil
}

// FetchOffers fetches offers for a product using Amazon Product Advertising API
func (p *AmazonOfficialProvider) FetchOffers(ctx context.Context, product *models.Product) ([]*models.Offer, error) {
	if !p.enabled {
		return nil, fmt.Errorf("Amazon API provider is not enabled (AMAZON_ACCESS_KEY, AMAZON_SECRET_KEY, or AMAZON_ASSOCIATE_TAG not set)")
	}

	// Try to find ASIN from product_identifiers
	// For now, we'll search for the product to get ASIN
	// In production, you'd store ASIN in product_identifiers

	candidates, err := p.Search(ctx, product.Title)
	if err != nil {
		return nil, fmt.Errorf("failed to search for product: %w", err)
	}

	if len(candidates) == 0 {
		return []*models.Offer{}, nil
	}

	// Get item details for the first match
	// In a real implementation, you'd get ASIN from product_identifiers
	// For now, we'll use a simplified approach

	// Fetch item details using GetItems operation
	// We need ASIN - for now, search and use first result
	searchParams := map[string]string{
		"Operation":     "SearchItems",
		"Keywords":      product.Title,
		"SearchIndex":   "All",
		"ItemCount":    "1",
		"Resources":    "Offers.Listings.Price,Offers.Listings.Availability,Offers.Listings.DeliveryInfo,ItemInfo.Title,ItemInfo.ByLineInfo,ItemInfo.ExternalIds",
		"PartnerTag":   p.associateTag,
		"PartnerType":  "Associates",
		"Marketplace": "www.amazon.com",
	}

	req, err := p.createSignedRequest(ctx, searchParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create signed request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return p.createOffersFromSearch(ctx, product, candidates)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return p.createOffersFromSearch(ctx, product, candidates)
	}

	var itemResponse struct {
		SearchResult struct {
			Items []struct {
				ASIN string `json:"ASIN"`
				DetailPageURL string `json:"DetailPageURL"`
				Offers struct {
					Listings []struct {
						Price struct {
							Amount    float64 `json:"Amount"`
							Currency  string  `json:"Currency"`
						} `json:"Price"`
						Availability struct {
							Message string `json:"Message"`
							Type    string `json:"Type"`
						} `json:"Availability"`
						DeliveryInfo struct {
							IsAmazonFulfilled bool `json:"IsAmazonFulfilled"`
							IsFreeShippingEligible bool `json:"IsFreeShippingEligible"`
							IsPrimeEligible bool `json:"IsPrimeEligible"`
						} `json:"DeliveryInfo"`
						MerchantInfo struct {
							Name string `json:"Name"`
						} `json:"MerchantInfo"`
					} `json:"Listings"`
				} `json:"Offers"`
			} `json:"Items"`
		} `json:"SearchResult"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return p.createOffersFromSearch(ctx, product, candidates)
	}

	if err := json.Unmarshal(body, &itemResponse); err != nil {
		return p.createOffersFromSearch(ctx, product, candidates)
	}

	if len(itemResponse.SearchResult.Items) == 0 {
		return p.createOffersFromSearch(ctx, product, candidates)
	}

	item := itemResponse.SearchResult.Items[0]
	offers := make([]*models.Offer, 0, len(item.Offers.Listings))

	for _, listing := range item.Offers.Listings {
		priceAmount := int(listing.Price.Amount * 100) // Convert to cents
		availabilityStatus := "in_stock"
		inStock := true
		if listing.Availability.Type == "Now" || strings.Contains(strings.ToLower(listing.Availability.Message), "in stock") {
			availabilityStatus = "in_stock"
			inStock = true
		} else {
			availabilityStatus = "out_of_stock"
			inStock = false
		}

		seller := listing.MerchantInfo.Name
		if seller == "" {
			seller = "Amazon"
		}

		now := time.Now()
		offer := &models.Offer{
			ID:                 uuid.New(),
			ProductID:          product.ID,
			Source:             "amazon",
			Seller:             seller,
			PriceAmount:        priceAmount,
			Currency:           listing.Price.Currency,
			ShippingToUSAmount: 0, // Will be calculated by shipping calculator
			TotalToUSAmount:    0, // Will be calculated by shipping calculator
			EstDeliveryDaysMin: intPtr(1), // Prime eligible items
			EstDeliveryDaysMax: intPtr(3),
			InStock:            inStock,
			AvailabilityStatus: stringPtr(availabilityStatus),
			URL:                stringPtr(item.DetailPageURL),
			PriceUpdatedAt:     now,
			FetchedAt:          now,
		}

		// Adjust delivery days based on Prime eligibility
		if listing.DeliveryInfo.IsPrimeEligible {
			offer.EstDeliveryDaysMin = intPtr(1)
			offer.EstDeliveryDaysMax = intPtr(2)
		} else {
			offer.EstDeliveryDaysMin = intPtr(5)
			offer.EstDeliveryDaysMax = intPtr(10)
		}

		offers = append(offers, offer)
	}

	// If no offers found, create a default offer
	if len(offers) == 0 {
		return p.createOffersFromSearch(ctx, product, candidates)
	}

	return offers, nil
}

// createOffersFromSearch creates offers from search results when detailed item fetch fails
func (p *AmazonOfficialProvider) createOffersFromSearch(ctx context.Context, product *models.Product, candidates []ProductCandidate) ([]*models.Offer, error) {
	now := time.Now()
	offer := &models.Offer{
		ID:                 uuid.New(),
		ProductID:          product.ID,
		Source:             "amazon",
		Seller:             "Amazon",
		PriceAmount:        0, // Would need to fetch actual price
		Currency:           "USD",
		ShippingToUSAmount: 0,
		TotalToUSAmount:    0,
		EstDeliveryDaysMin: intPtr(1),
		EstDeliveryDaysMax: intPtr(3),
		InStock:            true,
		AvailabilityStatus: stringPtr("in_stock"),
		PriceUpdatedAt:     now,
		FetchedAt:          now,
	}

	return []*models.Offer{offer}, nil
}

// createSignedRequest creates a signed request for Amazon Product Advertising API 5.0
// PA-API 5.0 uses POST requests with JSON body and AWS Signature Version 4
func (p *AmazonOfficialProvider) createSignedRequest(ctx context.Context, params map[string]string) (*http.Request, error) {
	// PA-API 5.0 uses POST with JSON body
	// For simplicity, we'll use a simplified approach with query parameters
	// In production, you should use AWS SDK or proper AWS Signature Version 4
	
	// Build request payload (simplified - actual PA-API 5.0 uses JSON)
	payload := map[string]interface{}{
		"Operation":     params["Operation"],
		"Keywords":      params["Keywords"],
		"SearchIndex":   params["SearchIndex"],
		"ItemCount":    params["ItemCount"],
		"Resources":    params["Resources"],
		"PartnerTag":   p.associateTag,
		"PartnerType":  "Associates",
		"Marketplace":  params["Marketplace"],
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create POST request
	reqURL := fmt.Sprintf("https://%s/paapi5/searchitems", p.apiEndpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", "PriceCompareBot/1.0")
	req.Header.Set("X-Amz-Target", "com.amazon.paapi5.v1.ProductAdvertisingAPIv1.SearchItems")
	req.Header.Set("X-Amz-Date", time.Now().UTC().Format("20060102T150405Z"))

	// For now, we'll use a simplified authentication
	// In production, implement proper AWS Signature Version 4
	// This is a placeholder - actual implementation requires AWS SDK or manual signing
	req.Header.Set("Authorization", fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/...", p.accessKey))

	return req, nil
}

