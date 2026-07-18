CREATE TABLE IF NOT EXISTS payments (
    id              UUID PRIMARY KEY,
    merchant_id     UUID NOT NULL REFERENCES merchants(id),
    provider_id     UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
    amount          BIGINT NOT NULL,
    currency        VARCHAR(3) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    external_id     VARCHAR(255) NOT NULL DEFAULT '',
    idempotency_key VARCHAR(255) NOT NULL DEFAULT '',
    description     TEXT NOT NULL DEFAULT '',
    metadata        JSONB NOT NULL DEFAULT '{}',
    error_message   TEXT NOT NULL DEFAULT '',
    attempt_count   INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT payments_status_check CHECK (status IN ('pending','processing','completed','failed','refunded','cancelled'))
);
CREATE INDEX IF NOT EXISTS idx_payments_merchant ON payments(merchant_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_idempotency ON payments(merchant_id, idempotency_key) WHERE idempotency_key != '';
CREATE INDEX IF NOT EXISTS idx_payments_created ON payments(created_at DESC);

CREATE TABLE IF NOT EXISTS routing_rules (
    id          UUID PRIMARY KEY,
    merchant_id UUID NOT NULL REFERENCES merchants(id),
    provider_id UUID NOT NULL REFERENCES providers(id),
    priority    INTEGER NOT NULL DEFAULT 100,
    weight      INTEGER NOT NULL DEFAULT 100,
    currency    VARCHAR(3) NOT NULL DEFAULT '',
    min_amount  BIGINT NOT NULL DEFAULT 0,
    max_amount  BIGINT NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT true
);
CREATE INDEX IF NOT EXISTS idx_routing_merchant ON routing_rules(merchant_id);
