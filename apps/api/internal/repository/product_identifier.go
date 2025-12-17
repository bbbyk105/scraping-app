package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/pricecompare/api/internal/models"
)

type ProductIdentifierRepository struct {
	db *DB
}

func NewProductIdentifierRepository(db *DB) *ProductIdentifierRepository {
	return &ProductIdentifierRepository{db: db}
}

// FindByTypeAndValue returns the identifier and associated product if it exists.
func (r *ProductIdentifierRepository) FindByTypeAndValue(idType, value string) (*models.ProductIdentifier, *models.Product, error) {
	query := `
		SELECT pi.id, pi.product_id, pi.type, pi.value, pi.created_at, pi.updated_at,
		       p.id, p.title, p.brand, p.model, p.image_url, p.created_at, p.updated_at
		FROM product_identifiers pi
		JOIN products p ON p.id = pi.product_id
		WHERE pi.type = $1 AND pi.value = $2
		LIMIT 1
	`

	var ident models.ProductIdentifier
	var product models.Product
	err := r.db.QueryRow(query, idType, value).Scan(
		&ident.ID,
		&ident.ProductID,
		&ident.Type,
		&ident.Value,
		&ident.CreatedAt,
		&ident.UpdatedAt,
		&product.ID,
		&product.Title,
		&product.Brand,
		&product.Model,
		&product.ImageURL,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	return &ident, &product, nil
}

func (r *ProductIdentifierRepository) Create(ident *models.ProductIdentifier) error {
	query := `
		INSERT INTO product_identifiers (id, product_id, type, value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	now := time.Now()
	if ident.ID == uuid.Nil {
		ident.ID = uuid.New()
	}
	ident.CreatedAt = now
	ident.UpdatedAt = now

	_, err := r.db.Exec(query,
		ident.ID,
		ident.ProductID,
		ident.Type,
		ident.Value,
		ident.CreatedAt,
		ident.UpdatedAt,
	)
	return err
}


