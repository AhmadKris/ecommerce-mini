CREATE TABLE inventory_movements (
    id              BIGSERIAL PRIMARY KEY,
    product_id      BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    type            VARCHAR(20) NOT NULL,
    quantity        INTEGER NOT NULL CHECK (quantity >= 0),
    before_quantity INTEGER NOT NULL CHECK (before_quantity >= 0),
    after_quantity  INTEGER NOT NULL CHECK (after_quantity >= 0),
    reason          TEXT NOT NULL DEFAULT '',
    reference       VARCHAR(255) NOT NULL DEFAULT '',
    created_by      BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_inventory_movements_product_id ON inventory_movements (product_id);
