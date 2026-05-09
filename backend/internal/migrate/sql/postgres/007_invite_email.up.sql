-- Email-driven invites.
-- Раньше invites были link-only (юзер копировал и слал руками). Теперь
-- при POST /v1/teams/:id/invites можно передать email — backend сразу
-- зашлёт письмо через Resend.
--
-- email      — на какой адрес ушло письмо (NULL = link-only invite)
-- sent_at    — когда письмо ушло (NULL = ещё не отправили или Mailer=Noop)

ALTER TABLE team_invites
  ADD COLUMN IF NOT EXISTS email   TEXT,
  ADD COLUMN IF NOT EXISTS sent_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_team_invites_email ON team_invites(email)
  WHERE email IS NOT NULL;
