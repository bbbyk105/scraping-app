package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/pricecompare/api/internal/models"
	"github.com/pricecompare/api/internal/providers"
	"github.com/pricecompare/api/internal/repository"
	"github.com/pricecompare/api/internal/shipping"
)

type Processor struct {
	productRepo     *repository.ProductRepository
	offerRepo       *repository.OfferRepository
	providerManager *providers.Manager
	shippingCalc    *shipping.Calculator
	logger          *zap.Logger
}

func NewProcessor(
	productRepo *repository.ProductRepository,
	offerRepo *repository.OfferRepository,
	providerManager *providers.Manager,
	shippingCalc *shipping.Calculator,
	logger *zap.Logger,
) *Processor {
	return &Processor{
		productRepo:     productRepo,
		offerRepo:       offerRepo,
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
	}

	return nil
}

func (p *Processor) processCandidate(
	ctx context.Context,
	candidate providers.ProductCandidate,
	provider providers.Provider,
	sourceName string,
) error {
	// Find or create product
	product, err := p.productRepo.FindByTitle(candidate.Title)
	if err != nil {
		return fmt.Errorf("failed to find product: %w", err)
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
	for _, offer := range offers {
		offer.ShippingToUSAmount = p.shippingCalc.CalculateShipping(offer.PriceAmount)
		offer.TotalToUSAmount = p.shippingCalc.CalculateTotal(offer.PriceAmount)

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

