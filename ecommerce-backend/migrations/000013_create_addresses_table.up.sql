CREATE TABLE addresses (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    recipient_name  VARCHAR(255) NOT NULL,
    phone           VARCHAR(30) NOT NULL,
    address_line    TEXT NOT NULL,
    city            VARCHAR(255) NOT NULL,
    province        VARCHAR(255) NOT NULL,
    postal_code     VARCHAR(20) NOT NULL,
    is_default      BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_addresses_user_id ON addresses (user_id);
