CREATE TABLE IF NOT EXISTS providers (
    id         UUID PRIMARY KEY,
    name       VARCHAR(255) NOT NULL UNIQUE,
    type       VARCHAR(50) NOT NULL,
    config     JSONB NOT NULL DEFAULT '{}',
    status     VARCHAR(20) NOT NULL DEFAULT 'active',
    priority   INTEGER NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT providers_type_check CHECK (type IN ('stripe','paypal','adyen','razorpay')),
    CONSTRAINT providers_status_check CHECK (status IN ('active','inactive','degraded'))
);
CREATE INDEX IF NOT EXISTS idx_providers_type ON providers(type);
CREATE INDEX IF NOT EXISTS idx_providers_status ON providers(status);
