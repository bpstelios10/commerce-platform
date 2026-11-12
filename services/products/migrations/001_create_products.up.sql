-- 001_create_products.up.sql

CREATE TABLE products (
    product_id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL,
    description TEXT,
    price NUMERIC(10,2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1,

    CONSTRAINT product_price_positive CHECK (price > 0),
    CONSTRAINT product_version_positive CHECK (version > 0)
);
