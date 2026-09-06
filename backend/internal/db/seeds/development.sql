-- Apply with:
-- docker compose exec -T postgres psql -U app -d appdb < backend/internal/db/seeds/development.sql

INSERT INTO sellers (name, website_url, country_code)
VALUES
    ('ALDI Nord', 'https://www.aldi-nord.de', 'DE'),
    ('ALDI SÜD', 'https://www.aldi-sued.de', 'DE'),
    ('Amazon.de', 'https://www.amazon.de', 'DE'),
    ('dm-drogerie markt', 'https://www.dm.de', 'DE'),
    ('EDEKA', 'https://www.edeka.de', 'DE'),
    ('Kaufland', 'https://www.kaufland.de', 'DE'),
    ('Lidl', 'https://www.lidl.de', 'DE'),
    ('MediaMarkt', 'https://www.mediamarkt.de', 'DE'),
    ('REWE', 'https://www.rewe.de', 'DE'),
    ('ROSSMANN', 'https://www.rossmann.de', 'DE')
ON CONFLICT (lower(name)) DO NOTHING;

INSERT INTO categories (slug)
VALUES
    ('beverages'),
    ('cleaning'),
    ('clothing'),
    ('cosmetics'),
    ('electronics'),
    ('food'),
    ('household'),
    ('personal-care'),
    ('pet-care'),
    ('supplements')
ON CONFLICT (lower(slug)) DO NOTHING;

INSERT INTO category_translations (category_id, locale, name)
SELECT category.id, translation.locale, translation.name
FROM (VALUES
    ('beverages', 'de', 'Getränke'),
    ('beverages', 'en', 'Beverages'),
    ('cleaning', 'de', 'Reinigung'),
    ('cleaning', 'en', 'Cleaning'),
    ('clothing', 'de', 'Bekleidung'),
    ('clothing', 'en', 'Clothing'),
    ('cosmetics', 'de', 'Kosmetik'),
    ('cosmetics', 'en', 'Cosmetics'),
    ('electronics', 'de', 'Elektronik'),
    ('electronics', 'en', 'Electronics'),
    ('food', 'de', 'Lebensmittel'),
    ('food', 'en', 'Food'),
    ('household', 'de', 'Haushalt'),
    ('household', 'en', 'Household'),
    ('personal-care', 'de', 'Körperpflege'),
    ('personal-care', 'en', 'Personal care'),
    ('pet-care', 'de', 'Tierbedarf'),
    ('pet-care', 'en', 'Pet care'),
    ('supplements', 'de', 'Nahrungsergänzung'),
    ('supplements', 'en', 'Supplements')
) AS translation(slug, locale, name)
JOIN categories AS category ON category.slug = translation.slug
ON CONFLICT (category_id, locale) DO UPDATE SET name = EXCLUDED.name;

INSERT INTO brands (name)
VALUES
    ('Acer'),
    ('Adidas'),
    ('Adobe'),
    ('Aesop'),
    ('Alnatura'),
    ('Apple'),
    ('Ariel'),
    ('Arla'),
    ('Asus'),
    ('Avène'),
    ('Balea'),
    ('alverde'),
    ('babylove'),
    ('Barilla'),
    ('Bauknecht'),
    ('Bayer'),
    ('Beats'),
    ('Beiersdorf'),
    ('Belkin'),
    ('Ben & Jerry''s'),
    ('Beste Wahl'),
    ('Bepanthen'),
    ('Bertolli'),
    ('Bio Company'),
    ('Bioderma'),
    ('Birkenstock'),
    ('Bosch'),
    ('Braun'),
    ('Bridgestone'),
    ('Brita'),
    ('Bruno Banani'),
    ('Bübchen'),
    ('Catrice'),
    ('Cien'),
    ('Coca-Cola'),
    ('Colgate'),
    ('Converse'),
    ('Dallmayr'),
    ('Denkmit'),
    ('domol'),
    ('EDEKA Bio'),
    ('enerBiO'),
    ('Dr. Oetker'),
    ('Ecover'),
    ('Ehrmann'),
    ('Elmex'),
    ('Fackelmann'),
    ('Fanta'),
    ('Fisher-Price'),
    ('Frosch'),
    ('Gut & Günstig'),
    ('Gut Bio'),
    ('Garmin'),
    ('Garnier'),
    ('Gilette'),
    ('Gorenje'),
    ('Haribo'),
    ('Head & Shoulders'),
    ('Hipp'),
    ('Hugo Boss'),
    ('Iglo'),
    ('Ikea'),
    ('Isana'),
    ('Ja!'),
    ('Jack Wolfskin'),
    ('Jacobs'),
    ('Kärcher'),
    ('K-Classic'),
    ('K-Beauty'),
    ('Kellogg''s'),
    ('Kérastase'),
    ('Kikkoman'),
    ('Kneipp'),
    ('Knorr'),
    ('Lenovo'),
    ('Levi''s'),
    ('LG'),
    ('L''Oréal'),
    ('Lindt'),
    ('Lacura'),
    ('Lupilu'),
    ('Logitech'),
    ('Lowa'),
    ('Maggi'),
    ('Mammut'),
    ('Manner'),
    ('Mars'),
    ('Miele'),
    ('Milka'),
    ('Milbona'),
    ('Milsani'),
    ('Monster Energy'),
    ('Müller'),
    ('Naduria'),
    ('Nescafé'),
    ('Nestlé'),
    ('Nike'),
    ('Nivea'),
    ('Nintendo'),
    ('Nivea Men'),
    ('Odol-med3'),
    ('Oral-B'),
    ('Oreo'),
    ('Pampers'),
    ('Parkside'),
    ('Patagonia'),
    ('Persil'),
    ('Philips'),
    ('Pilos'),
    ('PlayStation'),
    ('Puma'),
    ('Ritter Sport'),
    ('Samsung'),
    ('Sebamed'),
    ('Siemens'),
    ('Silvercrest'),
    ('Sony'),
    ('Speick'),
    ('Spotify'),
    ('Tchibo'),
    ('Tandil'),
    ('Tempo'),
    ('Tefal'),
    ('The North Face'),
    ('Toffifee'),
    ('Tupperware'),
    ('Vanish'),
    ('Varta'),
    ('Weleda'),
    ('W5'),
    ('Whiskas'),
    ('Wrigley''s'),
    ('Zalando'),
    ('Zewa')
ON CONFLICT (lower(name)) DO NOTHING;

WITH ordered_brands AS (
    SELECT id, name, row_number() OVER (ORDER BY id) - 1 AS brand_number
    FROM brands
), ordered_categories AS (
    SELECT id, row_number() OVER (ORDER BY id) - 1 AS category_number
    FROM categories
), inserted_products AS (
    INSERT INTO products (brand_id, category_id, name, description)
    SELECT
        brand.id,
        category.id,
        brand.name || ' sample product',
        'Development seed product for ' || brand.name
    FROM ordered_brands AS brand
    JOIN ordered_categories AS category
        ON category.category_number = brand.brand_number % (SELECT count(*) FROM categories)
    ON CONFLICT (brand_id, lower(name)) DO NOTHING
    RETURNING id, brand_id
)
SELECT count(*) FROM inserted_products;

WITH ordered_sellers AS (
    SELECT id, row_number() OVER (ORDER BY id) - 1 AS seller_number
    FROM sellers
), ordered_products AS (
    SELECT id, row_number() OVER (ORDER BY id) - 1 AS product_number
    FROM products
), seller_pairs AS (
    SELECT
        product.id AS product_id,
        seller.id AS seller_id,
        product.product_number
    FROM ordered_products AS product
    JOIN ordered_sellers AS seller
        ON seller.seller_number IN (
            product.product_number % (SELECT count(*) FROM sellers),
            (product.product_number + 1) % (SELECT count(*) FROM sellers)
        )
)
INSERT INTO seller_products (seller_id, product_id, external_identifier)
SELECT seller_id, product_id, 'development-seed-' || product_id
FROM seller_pairs
ON CONFLICT (seller_id, product_id) DO NOTHING;

WITH retailer_brands (seller_name, brand_name) AS (
    VALUES
        ('dm-drogerie markt', 'Balea'),
        ('dm-drogerie markt', 'alverde'),
        ('dm-drogerie markt', 'babylove'),
        ('dm-drogerie markt', 'Denkmit'),
        ('Lidl', 'Cien'),
        ('Lidl', 'Lupilu'),
        ('Lidl', 'Milbona'),
        ('Lidl', 'Parkside'),
        ('Lidl', 'Silvercrest'),
        ('Lidl', 'W5'),
        ('ROSSMANN', 'domol'),
        ('ROSSMANN', 'Isana'),
        ('ROSSMANN', 'Alterra'),
        ('ROSSMANN', 'enerBiO'),
        ('REWE', 'Ja!'),
        ('REWE', 'Beste Wahl'),
        ('EDEKA', 'Gut & Günstig'),
        ('EDEKA', 'EDEKA Bio'),
        ('Kaufland', 'K-Classic'),
        ('Kaufland', 'K-Beauty'),
        ('ALDI Nord', 'Tandil'),
        ('ALDI Nord', 'Gut Bio'),
        ('ALDI SÜD', 'Lacura'),
        ('ALDI SÜD', 'Milsani')
)
INSERT INTO seller_products (seller_id, product_id, external_identifier)
SELECT seller.id, product.id, 'development-seed-' || product.id
FROM retailer_brands AS mapping
JOIN sellers AS seller ON seller.name = mapping.seller_name
JOIN brands AS brand ON brand.name = mapping.brand_name
JOIN products AS product ON product.brand_id = brand.id
ON CONFLICT (seller_id, product_id) DO NOTHING;