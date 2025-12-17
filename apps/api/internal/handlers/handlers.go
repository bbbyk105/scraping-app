package handlers

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/pricecompare/api/internal/jobs"
	"github.com/pricecompare/api/internal/models"
	"github.com/pricecompare/api/internal/providers"
	"github.com/pricecompare/api/internal/repository"
	"github.com/pricecompare/api/internal/shipping"
)

type Handlers struct {
	productRepo        *repository.ProductRepository
	offerRepo          *repository.OfferRepository
	identifierRepo     *repository.ProductIdentifierRepository
	sourceProductRepo  *repository.SourceProductRepository
	providerManager    *providers.Manager
	asynqClient        *asynq.Client
	shippingCalc       *shipping.Calculator
	logger             *zap.Logger
}

func New(
	productRepo *repository.ProductRepository,
	offerRepo *repository.OfferRepository,
	identifierRepo *repository.ProductIdentifierRepository,
	sourceProductRepo *repository.SourceProductRepository,
	providerManager *providers.Manager,
	asynqClient *asynq.Client,
	shippingCalc *shipping.Calculator,
	logger *zap.Logger,
) *Handlers {
	return &Handlers{
		productRepo:       productRepo,
		offerRepo:         offerRepo,
		identifierRepo:    identifierRepo,
		sourceProductRepo: sourceProductRepo,
		providerManager:   providerManager,
		asynqClient:       asynqClient,
		shippingCalc:      shippingCalc,
		logger:            logger,
	}
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}

func (h *Handlers) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

func (h *Handlers) Search(c *fiber.Ctx) error {
	query := c.Query("query", "")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "query parameter is required",
		})
	}

	limit := 20
	products, err := h.productRepo.Search(query, limit)
	if err != nil {
		h.logger.Error("Search failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to search products",
		})
	}

	// Get cheapest offer for each product
	type ProductWithMinPrice struct {
		*models.Product
		MinPriceCents *int `json:"min_price_cents,omitempty"`
	}

	results := make([]ProductWithMinPrice, 0, len(products))
	for _, product := range products {
		offers, err := h.offerRepo.GetByProductID(product.ID)
		if err != nil {
			h.logger.Warn("Failed to get offers", zap.Error(err))
			results = append(results, ProductWithMinPrice{Product: product})
			continue
		}

		var minPrice *int
		if len(offers) > 0 {
			min := offers[0].TotalToUSAmount
			for _, offer := range offers {
				if offer.TotalToUSAmount < min {
					min = offer.TotalToUSAmount
				}
			}
			minPrice = &min
		}

		results = append(results, ProductWithMinPrice{
			Product:       product,
			MinPriceCents: minPrice,
		})
	}

	return c.JSON(fiber.Map{
		"products": results,
	})
}

func (h *Handlers) GetProduct(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid product id",
		})
	}

	product, err := h.productRepo.GetByID(id)
	if err != nil {
		h.logger.Error("Get product failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get product",
		})
	}

	if product == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "product not found",
		})
	}

	return c.JSON(product)
}

func (h *Handlers) GetProductOffers(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid product id",
		})
	}

	offers, err := h.offerRepo.GetByProductID(id)
	if err != nil {
		h.logger.Error("Get offers failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get offers",
		})
	}

	return c.JSON(fiber.Map{
		"offers": offers,
	})
}

// CompareProductOffers returns offers for a product with sorting options.
// Supported sort keys: total, fastest, newest, in_stock
func (h *Handlers) CompareProductOffers(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid product id",
		})
	}

	sortKey := c.Query("sort", "total")
	if sortKey != "total" && sortKey != "fastest" && sortKey != "newest" && sortKey != "in_stock" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid sort key. must be one of: total, fastest, newest, in_stock",
		})
	}

	offers, err := h.offerRepo.GetByProductIDWithSort(id, sortKey)
	if err != nil {
		h.logger.Error("Get offers for compare failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get offers",
		})
	}

	return c.JSON(fiber.Map{
		"offers": offers,
	})
}

type ResolveURLRequest struct {
	URL string `json:"url"`
}

// ResolveURL parses an input URL, extracts identifiers (e.g. ASIN),
// finds or creates a corresponding product, and returns it.
// For now, this supports a limited set of providers and responds politely
// when the URL cannot be handled.
func (h *Handlers) ResolveURL(c *fiber.Ctx) error {
	var req ResolveURLRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	if req.URL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "url is required",
		})
	}

	rawURL := strings.TrimSpace(req.URL)
	// 補助: スキームが無い場合は https:// を補完 (例: www.amazon.com/dp/ASIN)
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "URLの形式が正しくありません",
		})
	}

	host := strings.ToLower(parsed.Host)
	path := parsed.Path

	var (
		provider       string
		identifierType string
		identifier     string
		sourceID       string
	)

	// Very small, explicit set of supported URL patterns.
	// Example: https://www.amazon.com/dp/B08N5WRWNW
	if strings.Contains(host, "amazon.") {
		provider = "amazon"
		parts := strings.Split(path, "/")
		for i, p := range parts {
			if p == "dp" || (p == "product" && i > 0 && parts[i-1] == "gp") {
				if i+1 < len(parts) && parts[i+1] != "" {
					identifierType = "ASIN"
					identifier = parts[i+1]
					sourceID = parts[i+1]
				}
				break
			}
		}
	}

	if provider == "" || identifierType == "" || identifier == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":       "このURLは現在のバージョンでは解析対象外です",
			"description": "Amazonの商品詳細URL (https://www.amazon.com/dp/ASIN) のみ対応しています。",
		})
	}

	// Try to find an existing product via identifier
	_, existingProduct, err := h.identifierRepo.FindByTypeAndValue(identifierType, identifier)
	if err != nil {
		h.logger.Error("ResolveURL: failed to lookup identifier", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to resolve url",
		})
	}

	product := existingProduct
	if product == nil {
		// Create a minimal product placeholder. In a future iteration this can be
		// populated by a dedicated provider without violating robots/ALLOW_LIVE_FETCH.
		title := "URLから登録された商品 (" + identifierType + ": " + identifier + ")"
		product = &models.Product{
			Title: title,
		}
		if err := h.productRepo.Create(product); err != nil {
			h.logger.Error("ResolveURL: failed to create product", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to create product from url",
			})
		}

		// Save identifier mapping
		if err := h.identifierRepo.Create(&models.ProductIdentifier{
			ProductID: product.ID,
			Type:      identifierType,
			Value:     identifier,
		}); err != nil {
			h.logger.Warn("ResolveURL: failed to save identifier", zap.Error(err))
		}
	}

	// Upsert source product info
	if sourceID != "" {
		sp := &models.SourceProduct{
			ProductID: product.ID,
			Provider:  provider,
			SourceID:  sourceID,
			URL:       rawURL,
		}
		if err := h.sourceProductRepo.Upsert(sp); err != nil {
			h.logger.Warn("ResolveURL: failed to upsert source product", zap.Error(err))
		}
	}

	return c.JSON(fiber.Map{
		"product":          product,
		"identifier_type":  identifierType,
		"identifier_value": identifier,
		"provider":         provider,
	})
}

type FetchPricesRequest struct {
	Source string `json:"source"` // "demo", "public_html", or "all"
}

func (h *Handlers) FetchPrices(c *fiber.Ctx) error {
	var req FetchPricesRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Source == "" {
		req.Source = "all"
	}

	if req.Source != "demo" && req.Source != "public_html" && req.Source != "live" && req.Source != "walmart" && req.Source != "amazon" && req.Source != "all" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid source. must be 'demo', 'public_html', 'live', 'walmart', 'amazon', or 'all'",
		})
	}

	payload, err := json.Marshal(jobs.FetchPricesPayload{Source: req.Source})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create job payload",
		})
	}

	task := asynq.NewTask(jobs.TypeFetchPrices, payload)
	info, err := h.asynqClient.Enqueue(task)
	if err != nil {
		h.logger.Error("Failed to enqueue task", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to enqueue job",
		})
	}

	return c.JSON(fiber.Map{
		"job_id": info.ID,
		"status": "enqueued",
		"source": req.Source,
	})
}

func (h *Handlers) ImageSearch(c *fiber.Ctx) error {
	// Stub implementation
	return c.JSON(fiber.Map{
		"message": "Image search is not yet implemented",
		"products": []interface{}{},
	})
}

