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

CREATE TABLE appuser (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    firebase_uid TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL DEFAULT 'Anonymous user',
    photo_url TEXT,
    email TEXT,
    locale TEXT NOT NULL DEFAULT 'de',
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT user_firebase_uid_not_blank CHECK (btrim(firebase_uid) <> ''),
    CONSTRAINT user_locale_valid CHECK (locale ~ '^[a-z]{2}(-[A-Z]{2})?$')
);

CREATE TABLE deterioration_reports (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    seller_id BIGINT NOT NULL REFERENCES sellers(id) ON DELETE RESTRICT,
    brand_id BIGINT NOT NULL REFERENCES brands(id) ON DELETE RESTRICT,
    category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    submitted_by_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    product_name TEXT NOT NULL,
    product_url TEXT,
    description TEXT NOT NULL,
    observed_at DATE,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT deterioration_reports_product_name_not_blank CHECK (btrim(product_name) <> ''),
    CONSTRAINT deterioration_reports_description_not_blank CHECK (btrim(description) <> ''),
    CONSTRAINT deterioration_reports_status_valid CHECK (status IN ('pending', 'approved', 'rejected'))
);

CREATE INDEX deterioration_reports_seller_id_idx
    ON deterioration_reports (seller_id);
CREATE INDEX deterioration_reports_brand_id_idx
    ON deterioration_reports (brand_id);
CREATE INDEX deterioration_reports_category_id_idx
    ON deterioration_reports (category_id);
CREATE INDEX deterioration_reports_submitted_by_user_id_idx
    ON deterioration_reports (submitted_by_user_id);
CREATE INDEX deterioration_reports_status_created_at_idx
    ON deterioration_reports (status, created_at DESC);

CREATE TABLE report_images (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    report_id BIGINT NOT NULL REFERENCES deterioration_reports(id) ON DELETE CASCADE,
    storage_path TEXT NOT NULL,
    image_url TEXT NOT NULL,
    sort_order SMALLINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT report_images_storage_path_not_blank CHECK (btrim(storage_path) <> ''),
    CONSTRAINT report_images_image_url_not_blank CHECK (btrim(image_url) <> ''),
    CONSTRAINT report_images_sort_order_valid CHECK (sort_order BETWEEN 1 AND 5),
    CONSTRAINT report_images_report_order_unique UNIQUE (report_id, sort_order)
);

CREATE INDEX report_images_report_id_idx ON report_images (report_id);

CREATE TABLE report_comments (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    report_id BIGINT NOT NULL REFERENCES deterioration_reports(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT report_comments_body_not_blank CHECK (btrim(body) <> '')
);

CREATE INDEX report_comments_report_id_idx ON report_comments (report_id, created_at);

CREATE TABLE report_likes (
    report_id BIGINT NOT NULL REFERENCES deterioration_reports(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (report_id, user_id)
);

CREATE TABLE report_alternatives (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    report_id BIGINT NOT NULL REFERENCES deterioration_reports(id) ON DELETE CASCADE,
    suggested_by_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    brand_id BIGINT NOT NULL REFERENCES brands(id) ON DELETE RESTRICT,
    seller_id BIGINT NOT NULL REFERENCES sellers(id) ON DELETE RESTRICT,
    product_name TEXT NOT NULL,
    product_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT report_alternatives_product_name_not_blank CHECK (btrim(product_name) <> ''),
    CONSTRAINT report_alternatives_report_user_unique UNIQUE (report_id, suggested_by_user_id)
);

CREATE INDEX report_alternatives_report_id_idx ON report_alternatives (report_id, created_at);

-- Canonical "unknown" row (fixed id -999) used when a user can't find their brand/category/seller;
-- the details go in the report description instead and get consolidated later.
INSERT INTO brands (id, name) OVERRIDING SYSTEM VALUE VALUES (-999, 'unknown')
    ON CONFLICT (id) DO NOTHING;
INSERT INTO sellers (id, name) OVERRIDING SYSTEM VALUE VALUES (-999, 'unknown')
    ON CONFLICT (id) DO NOTHING;
INSERT INTO categories (id, slug) OVERRIDING SYSTEM VALUE VALUES (-999, 'unknown')
    ON CONFLICT (id) DO NOTHING;
INSERT INTO category_translations (category_id, locale, name)
    VALUES (-999, 'de', 'unbekannt'), (-999, 'en', 'unknown')
    ON CONFLICT (category_id, locale) DO NOTHING;