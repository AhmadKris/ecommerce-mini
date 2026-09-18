ALTER TABLE order_items ADD COLUMN product_name VARCHAR(255) NOT NULL DEFAULT '';

-- Backfill existing rows from the current product name — the best
-- approximation available for orders placed before this column existed.
-- New rows populate it correctly at checkout time from then on.
UPDATE order_items oi
SET product_name = p.name
FROM products p
WHERE oi.product_id = p.id;

ALTER TABLE order_items ALTER COLUMN product_name DROP DEFAULT;
