package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/pricecompare/api/internal/models"
)

type ProductRepository struct {
	db *DB
}

func NewProductRepository(db *DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(product *models.Product) error {
	query := `
		INSERT INTO products (id, title, brand, model, image_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	now := time.Now()
	product.ID = uuid.New()
	product.CreatedAt = now
	product.UpdatedAt = now

	_, err := r.db.Exec(query,
		product.ID,
		product.Title,
		product.Brand,
		product.Model,
		product.ImageURL,
		product.CreatedAt,
		product.UpdatedAt,
	)
	return err
}

func (r *ProductRepository) GetByID(id uuid.UUID) (*models.Product, error) {
	query := `
		SELECT id, title, brand, model, image_url, created_at, updated_at
		FROM products
		WHERE id = $1
	`
	var product models.Product
	err := r.db.QueryRow(query, id).Scan(
		&product.ID,
		&product.Title,
		&product.Brand,
		&product.Model,
		&product.ImageURL,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) Search(query string, limit int) ([]*models.Product, error) {
	sqlQuery := `
		SELECT id, title, brand, model, image_url, created_at, updated_at
		FROM products
		WHERE to_tsvector('english', title) @@ plainto_tsquery('english', $1)
		   OR title ILIKE $2
		ORDER BY updated_at DESC
		LIMIT $3
	`
	searchPattern := "%" + query + "%"
	rows, err := r.db.Query(sqlQuery, query, searchPattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(
			&product.ID,
			&product.Title,
			&product.Brand,
			&product.Model,
			&product.ImageURL,
			&product.CreatedAt,
			&product.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}
	return products, rows.Err()
}

func (r *ProductRepository) FindByTitle(title string) (*models.Product, error) {
	query := `
		SELECT id, title, brand, model, image_url, created_at, updated_at
		FROM products
		WHERE title = $1
		LIMIT 1
	`
	var product models.Product
	err := r.db.QueryRow(query, title).Scan(
		&product.ID,
		&product.Title,
		&product.Brand,
		&product.Model,
		&product.ImageURL,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) Update(product *models.Product) error {
	query := `
		UPDATE products
		SET title = $2, brand = $3, model = $4, image_url = $5, updated_at = $6
		WHERE id = $1
	`
	product.UpdatedAt = time.Now()
	_, err := r.db.Exec(query,
		product.ID,
		product.Title,
		product.Brand,
		product.Model,
		product.ImageURL,
		product.UpdatedAt,
	)
	return err
}

