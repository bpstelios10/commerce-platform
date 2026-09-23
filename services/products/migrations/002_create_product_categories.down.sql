-- 002_create_product_categories.down.sql

ALTER TABLE products
DROP CONSTRAINT products_category_fkey;

DROP TABLE product_categories;
