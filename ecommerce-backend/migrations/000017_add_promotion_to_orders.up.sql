ALTER TABLE orders
    ADD COLUMN discount_amount NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (discount_amount >= 0),
    ADD COLUMN promotion_id BIGINT REFERENCES promotions(id) ON DELETE SET NULL;
