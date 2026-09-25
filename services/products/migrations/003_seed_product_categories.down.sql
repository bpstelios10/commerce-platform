-- 003_seed_product_categories.down.sql

DELETE FROM product_categories
WHERE name IN ('CLOTHES', 'ACCESSORY', 'MAGNET', 'POSTCARD', 'JEWELRY');
    