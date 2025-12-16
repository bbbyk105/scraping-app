package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

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
	productRepo     *repository.ProductRepository
	offerRepo       *repository.OfferRepository
	providerManager *providers.Manager
	asynqClient     *asynq.Client
	shippingCalc    *shipping.Calculator
	logger          *zap.Logger
}

func New(
	productRepo *repository.ProductRepository,
	offerRepo *repository.OfferRepository,
	providerManager *providers.Manager,
	asynqClient *asynq.Client,
	shippingCalc *shipping.Calculator,
	logger *zap.Logger,
) *Handlers {
	return &Handlers{
		productRepo:     productRepo,
		offerRepo:       offerRepo,
		providerManager: providerManager,
		asynqClient:     asynqClient,
		shippingCalc:    shippingCalc,
		logger:          logger,
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

	if req.Source != "demo" && req.Source != "public_html" && req.Source != "all" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid source. must be 'demo', 'public_html', or 'all'",
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

