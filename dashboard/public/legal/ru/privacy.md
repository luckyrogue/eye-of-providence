# Privacy Notice — Eye of Providence

**Last updated:** 2026-05-19. Поставщик услуги: индивидуальный maintainer
`main@rysdavletov.org`. Контакт по приватности, GDPR/CCPA и security — этот же e-mail.

Этот документ описывает что мы собираем, на каком правовом основании, как
храним и как пользователь может реализовать свои права (включая GDPR и CCPA
эквиваленты). Self-hosted инсталляции — данные не покидают вашу
инфраструктуру; этот документ применим к managed-инстансу
`https://eop.rysdavletov.org`.

## 1. Что мы собираем

### 1.1 Никогда не покидает машину пользователя

- **Содержимое файлов**, исходники, открытые в IDE.
- **Промпты в AI-чатах** и **ответы AI-моделей** (browser extension знает
  что пользователь на странице ChatGPT, но не читает текст диалога).
- **Сами keystrokes** — мы храним только счётчики, не последовательность.
- **Содержимое clipboard** — только sha256 + размер (см. §1.2).
- Заголовки и контент окон в **приватных/incognito** режимах и из user-defined
  blacklist.
- Содержимое **скриншотов** — мы их не делаем.

### 1.2 Что отправляется в backend

| Категория | Конкретно | Зачем |
|---|---|---|
| Идентификация | `user_id` (UUID), `device_id` (UUID), `session_id` | Привязка событий к аккаунту |
| Foreground app | bundle id (macOS) или process name (Windows) | "С чем работаешь" |
| Длительности | `duration_ms` фокуса в приложении | Active time / AFK |
| Ввод (counters) | `chars_in` (keystroke count, БЕЗ контента), `mouse_clicks` | Manual vs AI-assisted differentiation |
| Clipboard fingerprint | `sha256` хеш + размер в байтах | Атрибуция paste-событий (AI vs other) |
| AI-канал (если применимо) | provider (`openai`/`anthropic`/...), channel (`chat`/`inline`/`agent`/`cli`) | AI usage breakdown |
| Проектная атрибуция (опц.) | `project_id`, `file_lang` (только тип файла, не путь) | Per-project отчёты |
| Метаданные авторизации | email (для логина), GitHub login (если OAuth), hashed_token для API ключей | Auth |
| Reports | сгенерированный отчёт в markdown (создан AI-моделью из агрегатов выше) | History |

### 1.3 Что отправляется третьим лицам

- **Google Gemini API** (`gemini-2.5-flash`) — получает числовые агрегаты
  для генерации еженедельного/месячного отчёта в текстовом виде. Промпт
  состоит из ваших aggregated metrics (часы, %, top apps). НЕ отправляем:
  app bundle full path, raw events, или что-либо из §1.1. Согласно
  [Google AI terms](https://ai.google.dev/gemini-api/terms) free-tier
  использует prompts для улучшения моделей; платный tier — нет. См. §6.
- **Resend** (`api.resend.com`) — transactional email (verification,
  password reset). Получает только email-адрес + содержимое письма.
  [Resend privacy](https://resend.com/legal/privacy).
- **GitHub** — если используешь OAuth login, GitHub возвращает нам email
  + login. См. [GitHub OAuth scope](https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/scopes-for-oauth-apps).
- **GHCR (Docker registry)** — десктоп-агент скачивает обновления через
  Tauri updater; этот трафик уходит на GitHub.

## 2. Правовое основание (GDPR Art. 6)

- **Art. 6(1)(b) Performance of a contract** — основная обработка данных
  пользователя для предоставления услуги (трекинг своего времени).
- **Art. 6(1)(a) Consent** — отправка marketing emails (мы их пока не
  делаем), а также передача агрегатов в Google Gemini (явный opt-in в
  Settings → AI Reports).
- **Art. 6(1)(f) Legitimate interest** — security logs (audit_log), мониторинг
  rate-limit-нарушений. Балансирующий интерес — защита сервиса.

Children: услуга не предназначена для лиц младше 16 (или 13 в применимых
юрисдикциях). Мы не верифицируем возраст; обнаружив подростковую регистрацию,
немедленно удаляем аккаунт через `DELETE /v1/me/data`.

## 3. Retention (сроки хранения)

| Что | Где | Срок |
|---|---|---|
| Events (raw) | ClickHouse `events` table | 18 мес TTL, потом auto-drop по partition |
| Attribution events (derived) | ClickHouse `attribution_events` | 18 мес TTL |
| User profile | Postgres `users` | До удаления аккаунта |
| Audit log | Postgres `audit_log` | 24 мес |
| Reports (AI-generated) | Postgres `reports` | До удаления аккаунта |
| Local SQLite буфер агента | Локальный диск пользователя | 7 дней (GC раз в час) |
| Logs backend | stdout → агрегатор хостинга | 30 дней |

## 4. Права пользователя

### 4.1 Access + portability (Art. 15, 20)

`GET /v1/me/export` (требуется Bearer JWT) возвращает машино-читаемый JSON
со всеми вашими данными: профиль, devices, projects, consent, reports, API
tokens (без `hashed_token`), полная история событий (cap ~200k последних).

Дашборд: **Settings → Privacy → Export my data**.

### 4.2 Erasure / right to be forgotten (Art. 17)

`DELETE /v1/me/data` стирает:
- все события в ClickHouse (`ALTER ... DELETE WHERE user_id = ?`);
- reports + api_tokens + consent + projects + devices + user row в Postgres.

Дашборд: **Settings → Danger zone → Delete all my data**. Действие
необратимо; локальный SQLite-буфер не удаляется (это машина пользователя
— очистить вручную через **Quit & wipe local data**).

### 4.3 Rectification, restriction, objection

E-mail на `main@rysdavletov.org` с темой `[GDPR DSAR]`. Срок реакции —
30 дней (Art. 12(3)).

## 5. Безопасность

См. [SECURITY.md](../.github/SECURITY.md). Кратко:
- Backend: bcrypt cost 10, JWT HS256 с `token_version` revocation, 1h TTL
  на password reset, rate-limit (10/min auth endpoints, 120/min /v1).
- Image: signed (Cosign keyless), SLSA L3 attestation, CycloneDX SBOM —
  проверяется при self-host'е.
- Agent: SQLite-буфер шифрован AES-256-GCM, ключ в OS Keychain/DPAPI.

Incident response: 48h acknowledgement, 5 business days remediation timeline.

## 6. Sub-processors

Полный список третьих сторон, обрабатывающих данные:

| Sub-processor | Цель | Локация | Соглашение |
|---|---|---|---|
| Dokploy (если managed) | Hosting backend + DBs | EU/US (зависит от deploy) | DPA при request |
| Google (Gemini API) | AI-генерация отчётов | US | [Google Cloud DPA](https://cloud.google.com/terms/data-processing-addendum) |
| Resend | Transactional emails | US | [Resend DPA](https://resend.com/legal/dpa) |
| GitHub | OAuth + GHCR | US | [GitHub DPA](https://docs.github.com/en/site-policy/privacy-policies/global-privacy-practices) |

Изменения списка анонсируются за 30 дней через release notes.

## 7. International transfers

Backend hosted в зависимости от self-host выбора. Managed-инстанс —
Frankfurt, Germany (EU). Передача данных в Google Gemini (US) — на
[Standard Contractual Clauses](https://commission.europa.eu/law/law-topic/data-protection/international-dimension-data-protection/standard-contractual-clauses-scc_en).

## 8. Self-hosted instances

При self-host (`docker-compose.full.yml`) данные не покидают вашу
инфраструктуру за исключением:
- Gemini API, если задан `EOP_GEMINI_API_KEY` (можно оставить пустым).
- Resend, если задан `EOP_RESEND_API_KEY` (можно оставить пустым).
- GitHub OAuth, если задан `EOP_GITHUB_CLIENT_ID`.

Self-hosted maintainer — самостоятельный data controller. Этот документ
вас не обязывает; используйте его как baseline.

## 9. Изменения

Этот документ версионируется в git (`docs/privacy.md`). Существенные
изменения анонсируются:
- через release notes (`CHANGELOG.md`);
- через in-app notification dashboard'а (для managed-инстанса);
- по email тем пользователям, которые согласились на product updates.

## 10. Жалобы и регуляторы

EU: вы вправе подать жалобу в supervisory authority вашей страны
([полный список](https://edpb.europa.eu/about-edpb/about-edpb/members_en)).
Maintainer базируется не в ЕС, представительство по Art. 27 не назначено
(услуга не нацелена систематически на резидентов ЕС за пределами
indie-сегмента; при превышении thresholds пересмотрим).

US California (CCPA): запросы на disclosure / deletion / opt-out — на тот
же email.
