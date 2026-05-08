-- Manual billing — super_admin вручную активирует подписки после получения оплаты.
-- subscription_plan / subscription_until денормализованы в teams для быстрого чтения,
-- team_payments — лог всех платежей.

ALTER TABLE teams
  ADD COLUMN IF NOT EXISTS subscription_plan  TEXT NOT NULL DEFAULT 'free',
  ADD COLUMN IF NOT EXISTS subscription_until TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS subscription_note  TEXT;

CREATE TABLE IF NOT EXISTS team_payments (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  team_id       UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
  amount_cents  INTEGER,
  currency      TEXT,
  method        TEXT,             -- 'manual_transfer' | 'cash' | 'stripe' | ...
  note          TEXT,
  covers_until  TIMESTAMPTZ NOT NULL,
  paid_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  recorded_by   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_team_payments_team ON team_payments(team_id, paid_at DESC);
