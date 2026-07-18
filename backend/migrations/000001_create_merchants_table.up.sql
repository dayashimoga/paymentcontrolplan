-- Create merchants table for the Merchant bounded context.
-- This is the aggregate root table for merchant management.

CREATE TABLE IF NOT EXISTS merchants (
    id          UUID PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    api_key     VARCHAR(68) NOT NULL,
    webhook_url TEXT NOT NULL DEFAULT '',
    status      VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT merchants_name_unique UNIQUE (name),
    CONSTRAINT merchants_api_key_unique UNIQUE (api_key),
    CONSTRAINT merchants_status_check CHECK (status IN ('active', 'suspended', 'inactive'))
);

-- Indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_merchants_api_key ON merchants(api_key);
CREATE INDEX IF NOT EXISTS idx_merchants_status ON merchants(status);
CREATE INDEX IF NOT EXISTS idx_merchants_created_at ON merchants(created_at DESC);
