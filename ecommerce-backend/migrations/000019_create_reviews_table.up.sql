CREATE TABLE reviews (
    id         BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_id   BIGINT NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    rating     SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    title      VARCHAR(255) NOT NULL,
    body       TEXT NOT NULL,
    status     VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- One review per user per product — not per order, so buying the same
-- product twice doesn't let someone post two reviews of it.
CREATE UNIQUE INDEX idx_reviews_product_user ON reviews (product_id, user_id);
CREATE INDEX idx_reviews_status ON reviews (status);
