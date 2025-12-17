package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/pricecompare/api/internal/models"
)

type SourceProductRepository struct {
	db *DB
}

func NewSourceProductRepository(db *DB) *SourceProductRepository {
	return &SourceProductRepository{db: db}
}

func (r *SourceProductRepository) FindByProviderAndSourceID(provider, sourceID string) (*models.SourceProduct, error) {
	query := `
		SELECT id, product_id, provider, source_id, url, title, brand, image_url, raw_json, created_at, updated_at
		FROM source_products
		WHERE provider = $1 AND source_id = $2
		LIMIT 1
	`

	var sp models.SourceProduct
	err := r.db.QueryRow(query, provider, sourceID).Scan(
		&sp.ID,
		&sp.ProductID,
		&sp.Provider,
		&sp.SourceID,
		&sp.URL,
		&sp.Title,
		&sp.Brand,
		&sp.ImageURL,
		&sp.RawJSON,
		&sp.CreatedAt,
		&sp.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sp, nil
}

func (r *SourceProductRepository) Upsert(sp *models.SourceProduct) error {
	query := `
		INSERT INTO source_products (
			id, product_id, provider, source_id, url, title, brand, image_url, raw_json,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (provider, source_id)
		DO UPDATE SET
			product_id = EXCLUDED.product_id,
			url = EXCLUDED.url,
			title = EXCLUDED.title,
			brand = EXCLUDED.brand,
			image_url = EXCLUDED.image_url,
			raw_json = EXCLUDED.raw_json,
			updated_at = EXCLUDED.updated_at
		RETURNING id
	`

	now := time.Now()
	if sp.ID == uuid.Nil {
		sp.ID = uuid.New()
	}
	if sp.CreatedAt.IsZero() {
		sp.CreatedAt = now
	}
	sp.UpdatedAt = now

	return r.db.QueryRow(query,
		sp.ID,
		sp.ProductID,
		sp.Provider,
		sp.SourceID,
		sp.URL,
		sp.Title,
		sp.Brand,
		sp.ImageURL,
		sp.RawJSON,
		sp.CreatedAt,
		sp.UpdatedAt,
	).Scan(&sp.ID)
}


