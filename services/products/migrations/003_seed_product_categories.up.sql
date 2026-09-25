-- 003_seed_product_categories.up.sql

INSERT INTO product_categories (name, description)
VALUES ('CLOTHES', 'Clothing products, like hoodies, t-shirts, dresses, etc.'), 
        ('ACCESSORY', 'Touristic accessories'),
        ('MAGNET', 'Touristic magnets'),
        ('POSTCARD', 'Touristic postcards'),
        ('JEWELRY', 'Jewelry items')
ON CONFLICT (name) DO NOTHING;
