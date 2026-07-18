CREATE TABLE IF NOT EXISTS webhooks (
    id          UUID PRIMARY KEY,
    merchant_id UUID NOT NULL REFERENCES merchants(id),
    payment_id  UUID NOT NULL REFERENCES payments(id),
    url         TEXT NOT NULL,
    event_type  VARCHAR(100) NOT NULL,
    payload     TEXT NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts    INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 5,
    next_retry  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_webhooks_status ON webhooks(status, next_retry);
CREATE INDEX IF NOT EXISTS idx_webhooks_payment ON webhooks(payment_id);

CREATE TABLE IF NOT EXISTS audit_logs (
    id          UUID PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,
    entity_id   UUID NOT NULL,
    action      VARCHAR(50) NOT NULL,
    actor_id    UUID NOT NULL,
    changes     JSONB NOT NULL DEFAULT '{}',
    ip_address  VARCHAR(45) NOT NULL DEFAULT '',
    user_agent  TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_audit_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_actor ON audit_logs(actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_logs(created_at DESC);

CREATE TABLE IF NOT EXISTS reconciliation_records (
    id              UUID PRIMARY KEY,
    payment_id      UUID NOT NULL REFERENCES payments(id),
    provider_id     UUID NOT NULL,
    internal_amount BIGINT NOT NULL,
    external_amount BIGINT NOT NULL,
    internal_status VARCHAR(20) NOT NULL,
    external_status VARCHAR(50) NOT NULL,
    is_matched      BOOLEAN NOT NULL DEFAULT false,
    discrepancy     VARCHAR(100) NOT NULL DEFAULT '',
    reconciled_at   TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_recon_unmatched ON reconciliation_records(is_matched) WHERE is_matched = false;
CREATE INDEX IF NOT EXISTS idx_recon_payment ON reconciliation_records(payment_id);
