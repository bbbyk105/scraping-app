package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/pricecompare/api/internal/models"
)

type OfferRepository struct {
	db *DB
}

func NewOfferRepository(db *DB) *OfferRepository {
	return &OfferRepository{db: db}
}

func (r *OfferRepository) Create(offer *models.Offer) error {
	query := `
		INSERT INTO offers (
			id, product_id, source, seller, price_amount, currency,
			shipping_to_us_amount, total_to_us_amount,
			est_delivery_days_min, est_delivery_days_max, in_stock, url, fetched_at,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	now := time.Now()
	offer.ID = uuid.New()
	offer.FetchedAt = now
	offer.CreatedAt = now
	offer.UpdatedAt = now

	_, err := r.db.Exec(query,
		offer.ID,
		offer.ProductID,
		offer.Source,
		offer.Seller,
		offer.PriceAmount,
		offer.Currency,
		offer.ShippingToUSAmount,
		offer.TotalToUSAmount,
		offer.EstDeliveryDaysMin,
		offer.EstDeliveryDaysMax,
		offer.InStock,
		offer.URL,
		offer.FetchedAt,
		offer.CreatedAt,
		offer.UpdatedAt,
	)
	return err
}

func (r *OfferRepository) GetByProductID(productID uuid.UUID) ([]*models.Offer, error) {
	query := `
		SELECT id, product_id, source, seller, price_amount, currency,
		       shipping_to_us_amount, total_to_us_amount,
		       est_delivery_days_min, est_delivery_days_max, in_stock, url, fetched_at,
		       created_at, updated_at
		FROM offers
		WHERE product_id = $1
		ORDER BY total_to_us_amount ASC, fetched_at DESC
	`
	rows, err := r.db.Query(query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var offers []*models.Offer
	for rows.Next() {
		var offer models.Offer
		if err := rows.Scan(
			&offer.ID,
			&offer.ProductID,
			&offer.Source,
			&offer.Seller,
			&offer.PriceAmount,
			&offer.Currency,
			&offer.ShippingToUSAmount,
			&offer.TotalToUSAmount,
			&offer.EstDeliveryDaysMin,
			&offer.EstDeliveryDaysMax,
			&offer.InStock,
			&offer.URL,
			&offer.FetchedAt,
			&offer.CreatedAt,
			&offer.UpdatedAt,
		); err != nil {
			return nil, err
		}
		offers = append(offers, &offer)
	}
	return offers, rows.Err()
}

func (r *OfferRepository) Upsert(offer *models.Offer) error {
	query := `
		INSERT INTO offers (
			id, product_id, source, seller, price_amount, currency,
			shipping_to_us_amount, total_to_us_amount,
			est_delivery_days_min, est_delivery_days_max, in_stock, url, fetched_at,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (product_id, source, seller, COALESCE(url, '')) 
		DO UPDATE SET
			price_amount = EXCLUDED.price_amount,
			shipping_to_us_amount = EXCLUDED.shipping_to_us_amount,
			total_to_us_amount = EXCLUDED.total_to_us_amount,
			est_delivery_days_min = EXCLUDED.est_delivery_days_min,
			est_delivery_days_max = EXCLUDED.est_delivery_days_max,
			in_stock = EXCLUDED.in_stock,
			fetched_at = EXCLUDED.fetched_at,
			updated_at = EXCLUDED.updated_at
		RETURNING id
	`
	now := time.Now()
	if offer.ID == uuid.Nil {
		offer.ID = uuid.New()
	}
	offer.FetchedAt = now
	offer.UpdatedAt = now
	if offer.CreatedAt.IsZero() {
		offer.CreatedAt = now
	}

	err := r.db.QueryRow(query,
		offer.ID,
		offer.ProductID,
		offer.Source,
		offer.Seller,
		offer.PriceAmount,
		offer.Currency,
		offer.ShippingToUSAmount,
		offer.TotalToUSAmount,
		offer.EstDeliveryDaysMin,
		offer.EstDeliveryDaysMax,
		offer.InStock,
		offer.URL,
		offer.FetchedAt,
		offer.CreatedAt,
		offer.UpdatedAt,
	).Scan(&offer.ID)
	return err
}

func (r *OfferRepository) DeleteByProductIDAndSource(productID uuid.UUID, source string) error {
	query := `DELETE FROM offers WHERE product_id = $1 AND source = $2`
	_, err := r.db.Exec(query, productID, source)
	return err
}

