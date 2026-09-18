ALTER TABLE orders
    ADD COLUMN shipping_cost NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (shipping_cost >= 0);
