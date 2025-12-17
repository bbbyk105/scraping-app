-- Extend schema for identifiers, source products, and richer offer metadata

-- product_identifiers: maps products to various identifier types (JAN/UPC/EAN/MPN/ASIN, etc.)
CREATE TABLE product_identifiers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (type, value)
);

CREATE INDEX idx_product_identifiers_product_id ON product_identifiers(product_id);

-- source_products: represents a product entity on a specific provider
CREATE TABLE source_products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    source_id TEXT NOT NULL,
    url TEXT NOT NULL,
    title TEXT,
    brand TEXT,
    image_url TEXT,
    raw_json JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (provider, source_id)
);

CREATE INDEX idx_source_products_product_id ON source_products(product_id);
CREATE INDEX idx_source_products_provider ON source_products(provider);

-- Extend offers with richer pricing / availability metadata.
ALTER TABLE offers
    ADD COLUMN fee_amount INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN tax_amount INTEGER,
    ADD COLUMN availability_status TEXT,
    ADD COLUMN estimated_delivery_date TIMESTAMP WITH TIME ZONE,
    ADD COLUMN price_updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

CREATE INDEX idx_offers_price_updated_at ON offers(price_updated_at);




