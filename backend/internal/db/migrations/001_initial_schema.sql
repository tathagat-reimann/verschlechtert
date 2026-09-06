CREATE TABLE brands (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT brands_name_not_blank CHECK (btrim(name) <> '')
);

CREATE UNIQUE INDEX brands_name_ci_idx ON brands (lower(name));

CREATE TABLE sellers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    website_url TEXT,
    country_code TEXT NOT NULL DEFAULT 'DE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT sellers_name_not_blank CHECK (btrim(name) <> ''),
    CONSTRAINT sellers_country_code_valid CHECK (country_code ~ '^[A-Z]{2}$')
);

CREATE UNIQUE INDEX sellers_name_ci_idx ON sellers (lower(name));

CREATE TABLE categories (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    slug TEXT NOT NULL,
    parent_id BIGINT REFERENCES categories(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT categories_slug_not_blank CHECK (btrim(slug) <> '')
);

CREATE UNIQUE INDEX categories_slug_ci_idx ON categories (lower(slug));

CREATE TABLE category_translations (
    category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    locale TEXT NOT NULL,
    name TEXT NOT NULL,
    PRIMARY KEY (category_id, locale),
    CONSTRAINT category_translations_locale_not_blank CHECK (btrim(locale) <> ''),
    CONSTRAINT category_translations_name_not_blank CHECK (btrim(name) <> '')
);

CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    firebase_uid TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL DEFAULT 'Anonymous user',
    photo_url TEXT,
    email TEXT,
    locale TEXT NOT NULL DEFAULT 'de',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT users_firebase_uid_not_blank CHECK (btrim(firebase_uid) <> ''),
    CONSTRAINT users_locale_valid CHECK (locale ~ '^[a-z]{2}(-[A-Z]{2})?$')
);

CREATE TABLE products (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    brand_id BIGINT NOT NULL REFERENCES brands(id) ON DELETE RESTRICT,
    category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT products_name_not_blank CHECK (btrim(name) <> '')
);

CREATE UNIQUE INDEX products_brand_name_ci_idx
    ON products (brand_id, lower(name));

CREATE TABLE seller_products (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    seller_id BIGINT NOT NULL REFERENCES sellers(id) ON DELETE RESTRICT,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    product_url TEXT,
    external_identifier TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT seller_products_seller_product_unique UNIQUE (seller_id, product_id)
);

CREATE INDEX seller_products_product_id_idx ON seller_products (product_id);

CREATE TABLE deterioration_reports (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    seller_product_id BIGINT NOT NULL REFERENCES seller_products(id) ON DELETE RESTRICT,
    submitted_by_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    quality_rating SMALLINT NOT NULL,
    observed_at DATE,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT deterioration_reports_title_not_blank CHECK (btrim(title) <> ''),
    CONSTRAINT deterioration_reports_description_not_blank CHECK (btrim(description) <> ''),
    CONSTRAINT deterioration_reports_quality_rating_valid CHECK (quality_rating BETWEEN 1 AND 5),
    CONSTRAINT deterioration_reports_status_valid CHECK (status IN ('pending', 'approved', 'rejected'))
);

CREATE INDEX deterioration_reports_seller_product_id_idx
    ON deterioration_reports (seller_product_id);
CREATE INDEX deterioration_reports_submitted_by_user_id_idx
    ON deterioration_reports (submitted_by_user_id);
CREATE INDEX deterioration_reports_status_created_at_idx
    ON deterioration_reports (status, created_at DESC);