-- 001_create_orders.up.sql

CREATE TABLE orders (
    order_id UUID PRIMARY KEY,
    product_id VARCHAR(50) NOT NULL,
    quantity INT NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1,

    CONSTRAINT order_quantity_positive CHECK (quantity > 0),
    CONSTRAINT order_version_positive CHECK (version > 0)
);
