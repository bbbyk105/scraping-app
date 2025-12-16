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
	ID                  uuid.UUID  `json:"id"`
	ProductID           uuid.UUID  `json:"product_id"`
	Source              string     `json:"source"`
	Seller              string     `json:"seller"`
	PriceAmount         int        `json:"price_amount"`         // cents
	Currency            string     `json:"currency"`
	ShippingToUSAmount  int        `json:"shipping_to_us_amount"` // cents
	TotalToUSAmount     int        `json:"total_to_us_amount"`    // cents
	EstDeliveryDaysMin  *int       `json:"est_delivery_days_min,omitempty"`
	EstDeliveryDaysMax  *int       `json:"est_delivery_days_max,omitempty"`
	InStock             bool       `json:"in_stock"`
	URL                 *string    `json:"url,omitempty"`
	FetchedAt           time.Time  `json:"fetched_at"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

