-- Onboarding wizard state.
-- onboarding_dismissed_at: NULL = wizard ещё не пройден / не закрыт юзером.
-- Юзер либо проходит до конца (4-й шаг "первое событие" автоматом set'нет
-- эту колонку), либо вручную жмёт "skip" — тоже set'нется.

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS onboarding_dismissed_at TIMESTAMPTZ;
