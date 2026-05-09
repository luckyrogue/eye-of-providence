-- Locale per user. Хранится при первом изменении в Settings (и идёт в email-шаблоны).
-- NULL = не задан, fallback на Mailer'е либо ru.

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS locale TEXT
    CHECK (locale IS NULL OR locale IN ('ru', 'en', 'kk', 'es'));
