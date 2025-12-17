-- Rollback for 003_extend_schema.up.sql

DROP INDEX IF EXISTS idx_offers_price_updated_at;

ALTER TABLE offers
    DROP COLUMN IF EXISTS fee_amount,
    DROP COLUMN IF EXISTS tax_amount,
    DROP COLUMN IF EXISTS availability_status,
    DROP COLUMN IF EXISTS estimated_delivery_date,
    DROP COLUMN IF EXISTS price_updated_at;

DROP INDEX IF EXISTS idx_source_products_provider;
DROP INDEX IF EXISTS idx_source_products_product_id;
DROP TABLE IF EXISTS source_products;

DROP INDEX IF EXISTS idx_product_identifiers_product_id;
DROP TABLE IF EXISTS product_identifiers;




