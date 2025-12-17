package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/pricecompare/api/internal/models"
	"github.com/pricecompare/api/internal/providers"
	"github.com/pricecompare/api/internal/repository"
	"github.com/pricecompare/api/internal/shipping"
)

type Processor struct {
	productRepo      *repository.ProductRepository
	offerRepo        *repository.OfferRepository
	identifierRepo   *repository.ProductIdentifierRepository
	providerManager  *providers.Manager
	shippingCalc     *shipping.Calculator
	logger           *zap.Logger
}

func NewProcessor(
	productRepo *repository.ProductRepository,
	offerRepo *repository.OfferRepository,
	identifierRepo *repository.ProductIdentifierRepository,
	providerManager *providers.Manager,
	shippingCalc *shipping.Calculator,
	logger *zap.Logger,
) *Processor {
	return &Processor{
		productRepo:     productRepo,
		offerRepo:       offerRepo,
		identifierRepo:  identifierRepo,
		providerManager: providerManager,
		shippingCalc:    shippingCalc,
		logger:          logger,
	}
}

func (p *Processor) HandleFetchPrices(ctx context.Context, t *asynq.Task) error {
	var payload FetchPricesPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	p.logger.Info("Processing fetch_prices job", zap.String("source", payload.Source))

	sources := []string{}
	if payload.Source == "all" {
		sources = p.providerManager.List()
	} else {
		sources = []string{payload.Source}
	}

	for _, sourceName := range sources {
		provider, err := p.providerManager.Get(sourceName)
		if err != nil {
			p.logger.Warn("Provider not found", zap.String("source", sourceName))
			continue
		}

		if err := p.fetchFromProvider(ctx, provider, sourceName); err != nil {
			p.logger.Error("Failed to fetch from provider",
				zap.String("source", sourceName),
				zap.Error(err),
			)
		}
	}

	return nil
}

func (p *Processor) fetchFromProvider(ctx context.Context, provider providers.Provider, sourceName string) error {
	// For demo provider, we use predefined search queries
	// For public_html, we parse sample files
	// For walmart/amazon, use predefined search queries

	if sourceName == "demo" {
		queries := []string{"headphones", "watch", "cable"}
		for _, query := range queries {
			candidates, err := provider.Search(ctx, query)
			if err != nil {
				p.logger.Error("Search failed", zap.Error(err))
				continue
			}

			for _, candidate := range candidates {
				if err := p.processCandidate(ctx, candidate, provider, sourceName); err != nil {
					p.logger.Error("Failed to process candidate", zap.Error(err))
				}
			}
		}
	} else if sourceName == "public_html" {
		// Search all products from sample files
		candidates, err := provider.Search(ctx, "")
		if err != nil {
			return fmt.Errorf("failed to search: %w", err)
		}

		for _, candidate := range candidates {
			if err := p.processCandidate(ctx, candidate, provider, sourceName); err != nil {
				p.logger.Error("Failed to process candidate", zap.Error(err))
			}
		}
	} else if sourceName == "live" {
		// For live provider, use predefined search queries
		// In production, these could come from a configuration or database
		queries := []string{"headphones", "watch", "laptop"}
		for _, query := range queries {
			candidates, err := provider.Search(ctx, query)
			if err != nil {
				p.logger.Error("Search failed", zap.Error(err))
				continue
			}

			// Limit number of products per query to avoid too many requests
			maxProducts := 5
			for i, candidate := range candidates {
				if i >= maxProducts {
					break
				}
				if err := p.processCandidate(ctx, candidate, provider, sourceName); err != nil {
					p.logger.Error("Failed to process candidate", zap.Error(err))
				}
			}
		}
	} else if sourceName == "walmart" || sourceName == "amazon" {
		// For official API providers, use predefined search queries
		// In production, these could come from a configuration or database
		queries := []string{"headphones", "laptop", "smartphone", "tablet", "watch", "minecraft", "game", "toy", "book"}
		for i, query := range queries {
			// Add delay between requests to avoid rate limiting
			if i > 0 {
				// Wait 1 second between requests for rate limiting
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(1 * time.Second):
				}
			}

			candidates, err := provider.Search(ctx, query)
			if err != nil {
				p.logger.Error("Search failed", zap.Error(err), zap.String("query", query))
				// If rate limited, wait longer before next request
				if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "Too many requests") {
					p.logger.Warn("Rate limited, waiting 5 seconds", zap.String("query", query))
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(5 * time.Second):
					}
				}
				continue
			}

			// Limit number of products per query to avoid too many API requests
			maxProducts := 5 // Reduced from 10 to avoid rate limiting
			for j, candidate := range candidates {
				if j >= maxProducts {
					break
				}
				if err := p.processCandidate(ctx, candidate, provider, sourceName); err != nil {
					p.logger.Error("Failed to process candidate", zap.Error(err))
				}
			}
		}
	}

	return nil
}

func (p *Processor) processCandidate(
	ctx context.Context,
	candidate providers.ProductCandidate,
	provider providers.Provider,
	sourceName string,
) error {
	var product *models.Product
	var err error

	// Try to find product by identifier first (for product unification)
	if candidate.Identifier != nil && *candidate.Identifier != "" {
		identifierType := getIdentifierType(sourceName)
		if identifierType != "" {
			_, existingProduct, err := p.identifierRepo.FindByTypeAndValue(identifierType, *candidate.Identifier)
			if err != nil {
				p.logger.Warn("Failed to lookup identifier", zap.Error(err))
			} else if existingProduct != nil {
				product = existingProduct
				p.logger.Info("Found existing product by identifier",
					zap.String("identifier_type", identifierType),
					zap.String("identifier_value", *candidate.Identifier),
					zap.String("product_id", product.ID.String()),
				)
			}
		}
	}

	// Fallback to title-based search if no identifier match
	if product == nil {
		product, err = p.productRepo.FindByTitle(candidate.Title)
		if err != nil {
			return fmt.Errorf("failed to find product: %w", err)
		}
	}

	if product == nil {
		product = &models.Product{
			Title:    candidate.Title,
			Brand:    candidate.Brand,
			Model:    candidate.Model,
			ImageURL: candidate.ImageURL,
		}
		if err := p.productRepo.Create(product); err != nil {
			return fmt.Errorf("failed to create product: %w", err)
		}

		// Save identifier if available
		if candidate.Identifier != nil && *candidate.Identifier != "" {
			identifierType := getIdentifierType(sourceName)
			if identifierType != "" {
				identifier := &models.ProductIdentifier{
					ProductID: product.ID,
					Type:      identifierType,
					Value:     *candidate.Identifier,
				}
				if err := p.identifierRepo.Create(identifier); err != nil {
					p.logger.Warn("Failed to save identifier", zap.Error(err))
				} else {
					p.logger.Info("Saved product identifier",
						zap.String("identifier_type", identifierType),
						zap.String("identifier_value", *candidate.Identifier),
						zap.String("product_id", product.ID.String()),
					)
				}
			}
		}
	} else {
		// Update product info if needed
		if candidate.Brand != nil {
			product.Brand = candidate.Brand
		}
		if candidate.Model != nil {
			product.Model = candidate.Model
		}
		if candidate.ImageURL != nil {
			product.ImageURL = candidate.ImageURL
		}
		if err := p.productRepo.Update(product); err != nil {
			p.logger.Warn("Failed to update product", zap.Error(err))
		}
	}

	// Delete old offers from this source
	if err := p.offerRepo.DeleteByProductIDAndSource(product.ID, sourceName); err != nil {
		p.logger.Warn("Failed to delete old offers", zap.Error(err))
	}

	// Fetch offers
	offers, err := provider.FetchOffers(ctx, product)
	if err != nil {
		return fmt.Errorf("failed to fetch offers: %w", err)
	}

	// Recalculate shipping and save offers
	now := time.Now()
	for _, offer := range offers {
		offer.ShippingToUSAmount = p.shippingCalc.CalculateShipping(offer.PriceAmount)
		offer.TotalToUSAmount = p.shippingCalc.CalculateTotal(offer.PriceAmount)
		// Update price_updated_at when price information is refreshed
		offer.PriceUpdatedAt = now

		if err := p.offerRepo.Upsert(offer); err != nil {
			p.logger.Error("Failed to upsert offer",
				zap.String("product_id", product.ID.String()),
				zap.String("seller", offer.Seller),
				zap.Error(err),
			)
		}
	}

	return nil
}

// getIdentifierType returns the identifier type for a given source
func getIdentifierType(sourceName string) string {
	switch sourceName {
	case "walmart":
		return "itemId" // Walmart itemId
	case "amazon":
		return "ASIN" // Amazon ASIN
	default:
		return "" // Unknown source
	}
}

