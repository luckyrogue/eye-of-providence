-- Web Push subscriptions для PWA. Per-user, multi-device (один user может
-- иметь несколько subscriptions: phone + desktop + tablet).
--
-- endpoint — push-service URL (FCM / Mozilla autopush / Apple WebPush). Уникален
-- на subscription, поэтому primary disambiguation. user_id для cleanup при
-- delete data.
--
-- Auth keys — base64url-encoded p256dh + auth от browser PushManager.
-- В payload encrypt'инге используются для AES-GCM.
CREATE TABLE IF NOT EXISTS push_subscriptions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint     TEXT NOT NULL UNIQUE,
    p256dh       TEXT NOT NULL,
    auth         TEXT NOT NULL,
    user_agent   TEXT,                                  -- for UI labelling
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_push_subs_user ON push_subscriptions(user_id);
