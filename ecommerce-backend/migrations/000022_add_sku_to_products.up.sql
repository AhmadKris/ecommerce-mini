ALTER TABLE products ADD COLUMN sku VARCHAR(64) NOT NULL DEFAULT '';

-- Backfill existing rows with a placeholder SKU before enforcing uniqueness —
-- admins can rename these to a real scheme later via the update endpoint.
UPDATE products SET sku = 'SKU-' || id WHERE sku = '';

ALTER TABLE products ALTER COLUMN sku DROP DEFAULT;

-- Partial unique index (not deleted_at IS NULL) so a soft-deleted product's
-- SKU can be reused, same pattern as idx_products_slug_active.
CREATE UNIQUE INDEX idx_products_sku_active ON products (sku) WHERE deleted_at IS NULL;
