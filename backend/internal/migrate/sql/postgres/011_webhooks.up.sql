-- Webhooks — outbound delivery на user-defined URL'ы при events
-- (commit.ingested, report.generated, ...).
--
-- secret хранится в plaintext: используется для HMAC sign'инга outgoing
-- payload'ов. User видит секрет ровно раз при создании (UI показывает
-- prefix "whk_..." после, как для PAT).
CREATE TABLE IF NOT EXISTS webhooks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url             TEXT NOT NULL,
    secret          TEXT NOT NULL, -- HMAC-SHA256 key
    events          TEXT[] NOT NULL DEFAULT '{}',
    active          BOOLEAN NOT NULL DEFAULT true,
    last_delivery_at TIMESTAMPTZ,
    last_status     INTEGER, -- HTTP status code, -1 для network errors
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_webhooks_user_active ON webhooks(user_id) WHERE active = true;
CREATE INDEX IF NOT EXISTS idx_webhooks_events ON webhooks USING gin(events);
