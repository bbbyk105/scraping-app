package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID        uuid.UUID  `json:"id"`
	Title     string     `json:"title"`
	Brand     *string    `json:"brand,omitempty"`
	Model     *string    `json:"model,omitempty"`
	ImageURL  *string    `json:"image_url,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Offer struct {
	ID                 uuid.UUID  `json:"id"`
	ProductID          uuid.UUID  `json:"product_id"`
	Source             string     `json:"source"`
	Seller             string     `json:"seller"`
	PriceAmount        int        `json:"price_amount"`          // cents
	Currency           string     `json:"currency"`
	ShippingToUSAmount int        `json:"shipping_to_us_amount"` // cents
	TotalToUSAmount    int        `json:"total_to_us_amount"`    // cents
	EstDeliveryDaysMin *int       `json:"est_delivery_days_min,omitempty"`
	EstDeliveryDaysMax *int       `json:"est_delivery_days_max,omitempty"`
	InStock            bool       `json:"in_stock"`
	URL                *string    `json:"url,omitempty"`
	FetchedAt          time.Time  `json:"fetched_at"`
	FeeAmount          int        `json:"fee_amount"`                     // cents
	TaxAmount          *int       `json:"tax_amount,omitempty"`           // cents
	AvailabilityStatus *string    `json:"availability_status,omitempty"`  // e.g. "in_stock", "out_of_stock", "preorder"
	EstimatedDelivery  *time.Time `json:"estimated_delivery_date,omitempty"`
	PriceUpdatedAt     time.Time  `json:"price_updated_at"` // when price info was last refreshed
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// ProductIdentifier represents various identifiers like JAN/UPC/EAN/MPN/ASIN, etc.
type ProductIdentifier struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	Type      string    `json:"type"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SourceProduct represents how a product appears on a specific provider (site)
type SourceProduct struct {
	ID        uuid.UUID  `json:"id"`
	ProductID uuid.UUID  `json:"product_id"`
	Provider  string     `json:"provider"`
	SourceID  string     `json:"source_id"`
	URL       string     `json:"url"`
	Title     *string    `json:"title,omitempty"`
	Brand     *string    `json:"brand,omitempty"`
	ImageURL  *string    `json:"image_url,omitempty"`
	RawJSON   []byte     `json:"raw_json,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

