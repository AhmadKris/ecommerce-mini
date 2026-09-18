CREATE TABLE products (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    slug        VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    price       NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (price >= 0),
    stock       INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
    category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    image_url   VARCHAR(500) NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

-- Partial unique index so a soft-deleted product's slug can be reused.
CREATE UNIQUE INDEX idx_products_slug_active ON products (slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_category_id ON products (category_id) WHERE deleted_at IS NULL;
