CREATE TABLE promotions (
    id               BIGSERIAL PRIMARY KEY,
    code             VARCHAR(50) NOT NULL,
    type             VARCHAR(20) NOT NULL,
    value            NUMERIC(12, 2) NOT NULL CHECK (value > 0),
    minimum_purchase NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (minimum_purchase >= 0),
    usage_limit      INTEGER NOT NULL CHECK (usage_limit > 0),
    used_count       INTEGER NOT NULL DEFAULT 0 CHECK (used_count >= 0),
    starts_at        TIMESTAMPTZ NOT NULL,
    ends_at          TIMESTAMPTZ NOT NULL,
    status           VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Case-sensitive on purpose (promo codes are typically presented and typed
-- in a fixed case, e.g. all-uppercase) — the service layer normalizes to
-- uppercase before every lookup so "save10" and "SAVE10" collide correctly.
CREATE UNIQUE INDEX idx_promotions_code ON promotions (code);
