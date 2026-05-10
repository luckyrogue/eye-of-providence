-- Webhook format — controls payload shape для Slack/Discord-style receivers,
-- которые ожидают определённый JSON.
--
-- raw   — наш канонический {event, data, sent_at} с HMAC-подписью.
-- slack — Slack incoming-webhook совместимый {text, blocks, attachments}.
--         Slack игнорирует HMAC header (нет verification на их стороне),
--         поэтому signature всё равно шлётся, но не критична для slack-format.
ALTER TABLE webhooks
  ADD COLUMN IF NOT EXISTS format TEXT NOT NULL DEFAULT 'raw';

-- CHECK constraint защищает от опечаток в API.
ALTER TABLE webhooks
  DROP CONSTRAINT IF EXISTS webhooks_format_check;
ALTER TABLE webhooks
  ADD CONSTRAINT webhooks_format_check CHECK (format IN ('raw', 'slack'));
