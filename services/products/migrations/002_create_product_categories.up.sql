-- 002_create_product_categories.up.sql

CREATE TABLE product_categories (
    name VARCHAR(50) PRIMARY KEY,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE products
ADD CONSTRAINT products_category_fkey
FOREIGN KEY (category)
REFERENCES product_categories (name)
ON UPDATE CASCADE
ON DELETE RESTRICT;
