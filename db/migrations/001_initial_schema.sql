-- Apply with
-- docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U app -d appdb < db/migrations/001_initial_schema.sql

DROP TABLE IF EXISTS report_alternative;
DROP TABLE IF EXISTS report_like;
DROP TABLE IF EXISTS report_comment;
DROP TABLE IF EXISTS report_image;
DROP TABLE IF EXISTS report;
DROP TABLE IF EXISTS appuser;
DROP TABLE IF EXISTS category;
DROP TABLE IF EXISTS seller;
DROP TABLE IF EXISTS brand;

CREATE TABLE brand (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT brand_name_not_blank CHECK (btrim(name) <> '')
);

CREATE UNIQUE INDEX brand_name_ci_idx ON brand (lower(name));

CREATE TABLE seller (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    website_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT seller_name_not_blank CHECK (btrim(name) <> '')
);

CREATE UNIQUE INDEX seller_name_ci_idx ON seller (lower(name));

CREATE TABLE category (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name_de TEXT NOT NULL,
    name_en TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT category_name_de_not_blank CHECK (btrim(name_de) <> ''),
    CONSTRAINT category_name_en_not_blank CHECK (btrim(name_en) <> '')
);

CREATE UNIQUE INDEX category_name_de_ci_idx ON category (lower(name_de));
CREATE UNIQUE INDEX category_name_en_ci_idx ON category (lower(name_en));

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
    CONSTRAINT user_locale_valid CHECK (locale IN ('de', 'en'))
);

CREATE TABLE report (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    seller_id BIGINT NOT NULL REFERENCES seller(id) ON DELETE RESTRICT,
    brand_id BIGINT NOT NULL REFERENCES brand(id) ON DELETE RESTRICT,
    category_id BIGINT NOT NULL REFERENCES category(id) ON DELETE RESTRICT,
    submitted_by_user_id BIGINT NOT NULL REFERENCES appuser(id) ON DELETE RESTRICT,
    product_name TEXT NOT NULL,
    product_url TEXT,
    description TEXT NOT NULL,
    observed_at DATE,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT report_product_name_not_blank CHECK (btrim(product_name) <> ''),
    CONSTRAINT report_description_not_blank CHECK (btrim(description) <> '')
);

CREATE INDEX report_seller_id_idx
    ON report (seller_id);
CREATE INDEX report_brand_id_idx
    ON report (brand_id);
CREATE INDEX report_category_id_idx
    ON report (category_id);
CREATE INDEX report_submitted_by_user_id_idx
    ON report (submitted_by_user_id);
CREATE INDEX report_active_created_at_idx
    ON report (created_at DESC)
    WHERE active = true;
CREATE INDEX report_user_active_created_at_idx
    ON report (submitted_by_user_id, created_at DESC)
    WHERE active = true;

CREATE TABLE report_image (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    report_id BIGINT NOT NULL REFERENCES report(id) ON DELETE CASCADE,
    storage_path TEXT NOT NULL,
    image_url TEXT NOT NULL,
    sort_order SMALLINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT report_image_storage_path_not_blank CHECK (btrim(storage_path) <> ''),
    CONSTRAINT report_image_image_url_not_blank CHECK (btrim(image_url) <> ''),
    CONSTRAINT report_image_sort_order_valid CHECK (sort_order BETWEEN 1 AND 5),
    CONSTRAINT report_image_report_order_unique UNIQUE (report_id, sort_order)
);

CREATE INDEX report_image_report_id_idx ON report_image (report_id);

CREATE TABLE report_comment (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    report_id BIGINT NOT NULL REFERENCES report(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES appuser(id) ON DELETE RESTRICT,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT report_comment_body_not_blank CHECK (btrim(body) <> '')
);

CREATE INDEX report_comment_report_id_idx ON report_comment (report_id, created_at);

CREATE TABLE report_like (
    report_id BIGINT NOT NULL REFERENCES report(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES appuser(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (report_id, user_id)
);

CREATE TABLE report_alternative (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    report_id BIGINT NOT NULL REFERENCES report(id) ON DELETE CASCADE,
    suggested_by_user_id BIGINT NOT NULL REFERENCES appuser(id) ON DELETE RESTRICT,
    brand_id BIGINT NOT NULL REFERENCES brand(id) ON DELETE RESTRICT,
    seller_id BIGINT NOT NULL REFERENCES seller(id) ON DELETE RESTRICT,
    product_name TEXT NOT NULL,
    product_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT report_alternative_product_name_not_blank CHECK (btrim(product_name) <> ''),
    CONSTRAINT report_alternative_report_user_unique UNIQUE (report_id, suggested_by_user_id)
);

CREATE INDEX report_alternative_report_id_idx ON report_alternative (report_id, created_at);

-- Canonical "unknown" row (fixed id -999) used when a user can't find their brand/category/seller;
-- the details go in the report description instead and get consolidated later.
INSERT INTO brand (id, name) OVERRIDING SYSTEM VALUE VALUES (-999, 'unknown')
    ON CONFLICT (id) DO NOTHING;
INSERT INTO seller (id, name) OVERRIDING SYSTEM VALUE VALUES (-999, 'unknown')
    ON CONFLICT (id) DO NOTHING;
INSERT INTO category (id, name_de, name_en) OVERRIDING SYSTEM VALUE VALUES (-999, 'unknown', 'unknown')
    ON CONFLICT (id) DO NOTHING;

-- add internal user
INSERT INTO appuser (id, firebase_uid, display_name, email) OVERRIDING SYSTEM VALUE VALUES (-999, 'internal', 'Internal User', 'internal@example.com')
    ON CONFLICT (id) DO NOTHING;