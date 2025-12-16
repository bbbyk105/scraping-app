CREATE TABLE offers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    source TEXT NOT NULL,
    seller TEXT NOT NULL,
    price_amount INTEGER NOT NULL,
    currency TEXT NOT NULL DEFAULT 'USD',
    shipping_to_us_amount INTEGER NOT NULL DEFAULT 0,
    total_to_us_amount INTEGER NOT NULL,
    est_delivery_days_min INTEGER,
    est_delivery_days_max INTEGER,
    in_stock BOOLEAN DEFAULT true,
    url TEXT,
    fetched_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_offers_unique ON offers(product_id, source, seller, COALESCE(url, ''));

CREATE INDEX idx_offers_product_id ON offers(product_id);
CREATE INDEX idx_offers_source ON offers(source);
CREATE INDEX idx_offers_fetched_at ON offers(fetched_at);
CREATE INDEX idx_offers_total_to_us_amount ON offers(total_to_us_amount);

