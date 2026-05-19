# Құпиялылық туралы хабарлама — Eye of Providence

**Соңғы жаңарту:** 2026-05-19. Қызмет көрсетуші: жеке maintainer
`main@rysdavletov.org`. Құпиялылық, GDPR/CCPA және security бойынша байланыс — сол e-mail.

Бұл құжат не жинайтынымызды, құқықтық негізді, деректерді қалай сақтайтынымызды және
пайдаланушы өз құқықтарын қалай іске асыратынын сипаттайды (GDPR және CCPA
эквиваленттері қоса). Self-hosted орнатуларда деректер сіздің
инфрақұрылымыңыздан шықпайды; бұл құжат managed-инстансқа
`https://eop.rysdavletov.org` қолданылады.

## 1. Не жинаймыз

### 1.1 Пайдаланушы машинасынан ешқашан шықпайды

- **Файл мазмұны**, бастапқы код, IDE-да ашылған файлдар.
- **AI чатындағы prompt-тар** және **AI модель жауаптары** (browser extension
  ChatGPT бетінде екеніңізді біледі, бірақ диалог мәтінін оқымайды).
- **Өз keystroke-тары** — тек санағыштар, реттілік емес.
- **Clipboard мазмұны** — тек sha256 + өлшем (қараңыз §1.2).
- **Жеке/incognito** режимдеріндегі және user-defined blacklist-тегі
  терезе тақырыптары мен мазмұны.
- **Скриншот** мазмұны — скриншот жасамаймыз.

### 1.2 Backend-ке не жіберіледі

| Санат | Нақты | Неге |
|---|---|---|
| Сәйкестендіру | `user_id` (UUID), `device_id` (UUID), `session_id` | Оқиғаларды аккаунтқа байлау |
| Foreground app | bundle id (macOS) немесе process name (Windows) | «Немен жұмыс істейсіз» |
| Ұзақтықтар | қолданбадағы фокус `duration_ms` | Active time / AFK |
| Енгізу (санағыштар) | `chars_in` (keystroke count, МАЗМҰНСЫЗ), `mouse_clicks` | Manual vs AI-assisted differentiation |
| Clipboard fingerprint | `sha256` hash + байт өлшемі | Paste оқиғаларын атрибуциялау (AI vs other) |
| AI арна (қолданылса) | provider (`openai`/`anthropic`/...), channel (`chat`/`inline`/`agent`/`cli`) | AI usage breakdown |
| Жоба атрибуциясы (опц.) | `project_id`, `file_lang` (тек файл түрі, жол емес) | Per-project есептер |
| Auth метадеректері | email (login үшін), GitHub login (OAuth болса), API кілттер үшін hashed_token | Auth |
| Reports | markdown есеп (жоғарыдағы агрегаттардан AI модель жасайды) | History |

### 1.3 Үшінші тарапқа не жіберіледі

- **Google Gemini API** (`gemini-2.5-flash`) — апталық/айлық мәтіндік
  есеп жасау үшін сандық агрегаттар алады. Prompt сіздің aggregated metrics
  (сағат, %, top apps) тұрады. ЖІБЕРМЕЙМІЗ:
  app bundle толық жолы, raw events немесе §1.1-тен ешнәрсе.
  [Google AI terms](https://ai.google.dev/gemini-api/terms) бойынша free-tier
  prompt-тарды модельді жақсартуға пайдалана алады; paid tier — жоқ. Қараңыз §6.
- **Resend** (`api.resend.com`) — transactional email (verification,
  password reset). Тек email + хат мазмұнын алады.
  [Resend privacy](https://resend.com/legal/privacy).
- **GitHub** — OAuth login болса, GitHub email
  + login қайтарады. Қараңыз [GitHub OAuth scope](https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/scopes-for-oauth-apps).
- **GHCR (Docker registry)** — desktop agent Tauri updater арқылы
  жаңартуларды жүктейді; бұл трафик GitHub-қа барады.

## 2. Құқықтық негіз (GDPR Art. 6)

- **Art. 6(1)(b) Performance of a contract** — қызмет көрсету үшін
  пайдаланушы деректерін негізгі өңдеу (өз уақытыңызды бақылау).
- **Art. 6(1)(a) Consent** — marketing email (әзірше жібермейміз), сондай-ақ
  Google Gemini-ге агрегат жіберу (Settings → AI Reports ішінде explicit opt-in).
- **Art. 6(1)(f) Legitimate interest** — security logs (audit_log), rate-limit
  бұзушылықтарын мониторинг. Баланстық мүдде — қызметті қорғау.

Children: қызмет 16-дан кіші (немесе қолданылатын юрисдикцияларда 13)
тұлғаларға арналмаған. Жасты тексермейміз; кәмелетке толмаған тіркелуді
тапсақ, `DELETE /v1/me/data` арқылы аккаунтты дереу өшіреміз.

## 3. Retention (сақтау мерзімі)

| Не | Қайда | Мерзім |
|---|---|---|
| Events (raw) | ClickHouse `events` table | 18 ай TTL, partition бойынша auto-drop |
| Attribution events (derived) | ClickHouse `attribution_events` | 18 ай TTL |
| User profile | Postgres `users` | Аккаунт өшірілгенге дейін |
| Audit log | Postgres `audit_log` | 24 ай |
| Reports (AI-generated) | Postgres `reports` | Аккаунт өшірілгенге дейін |
| Agent local SQLite buffer | Пайдаланушы локальді дискі | 7 күн (сағат сайын GC) |
| Backend logs | stdout → hosting aggregator | 30 күн |

## 4. Пайдаланушы құқықтары

### 4.1 Access + portability (Art. 15, 20)

`GET /v1/me/export` (Bearer JWT қажет) барлық деректеріңізді
машина оқитын JSON қайтарады: profile, devices, projects, consent, reports, API
tokens (`hashed_token` жоқ), толық event тарихы (соңғы ~200k cap).

Dashboard: **Settings → Privacy → Export my data**.

### 4.2 Erasure / right to be forgotten (Art. 17)

`DELETE /v1/me/data` өшіреді:
- ClickHouse-тағы барлық events (`ALTER ... DELETE WHERE user_id = ?`);
- Postgres-та reports + api_tokens + consent + projects + devices + user row.

Dashboard: **Settings → Danger zone → Delete all my data**. Әрекет
кері қайтарылмайды; local SQLite buffer өшірілмейді (пайдаланушы машинасында —
**Quit & wipe local data** арқылы қолмен тазалаңыз).

### 4.3 Rectification, restriction, objection

`main@rysdavletov.org` e-mail, тақырып `[GDPR DSAR]`. Жауап мерзімі —
30 күн (Art. 12(3)).

## 5. Қауіпсіздік

Қараңыз [SECURITY.md](../.github/SECURITY.md). Қысқаша:
- Backend: bcrypt cost 10, JWT HS256 `token_version` revocation, password reset 1h TTL,
  rate-limit (auth endpoint 10/min, /v1 120/min).
- Image: signed (Cosign keyless), SLSA L3 attestation, CycloneDX SBOM —
  self-host кезінде тексеріледі.
- Agent: SQLite buffer AES-256-GCM шифрланған, кілт OS Keychain/DPAPI-да.

Incident response: 48 сағат acknowledgement, 5 жұмыс күні remediation timeline.

## 6. Sub-processors

Деректерді өңдейтін үшінші тараптардың толық тізімі:

| Sub-processor | Мақсат | Орналасу | Келісім |
|---|---|---|---|
| Dokploy (managed болса) | Backend + DB hosting | EU/US (deploy-қа байланысты) | DPA сұрау бойынша |
| Google (Gemini API) | AI есеп генерациясы | US | [Google Cloud DPA](https://cloud.google.com/terms/data-processing-addendum) |
| Resend | Transactional emails | US | [Resend DPA](https://resend.com/legal/dpa) |
| GitHub | OAuth + GHCR | US | [GitHub DPA](https://docs.github.com/en/site-policy/privacy-policies/global-privacy-practices) |

Тізім өзгерістері release notes арқылы 30 күн бұрын хабарланады.

## 7. International transfers

Backend hosting self-host таңдауына байланысты. Managed-инстанс —
Frankfurt, Germany (EU). Google Gemini (US) деректер беру —
[Standard Contractual Clauses](https://commission.europa.eu/law/law-topic/data-protection/international-dimension-data-protection/standard-contractual-clauses-scc_en) бойынша.

## 8. Self-hosted instances

Self-host (`docker-compose.full.yml`) кезінде деректер сіздің
инфрақұрылымыңыздан шықпайды, мынадан басқа:
- `EOP_GEMINI_API_KEY` болса Gemini API (бос қалдыруға болады).
- `EOP_RESEND_API_KEY` болса Resend (бос қалдыруға болады).
- `EOP_GITHUB_CLIENT_ID` болса GitHub OAuth.

Self-hosted maintainer — деректерді өңдеушінің өзі. Бұл құжат
сізді міндеттемейді; baseline ретінде пайдаланыңыз.

## 9. Өзгерістер

Бұл құжат git-те версияланады (`docs/privacy.md`). Маңызды
өзгерістер:
- release notes (`CHANGELOG.md`);
- dashboard in-app notification (managed-инстанс);
- product updates-ке келіскен пайдаланушыларға email.

## 10. Шағымдар және реттеушілер

EU: елдегі supervisory authority-ға шағым беруге құқығыңыз бар
([толық тізім](https://edpb.europa.eu/about-edpb/about-edpb/members_en)).
Maintainer ЕО-да емес, Art. 27 representative тағайындалмаған
(қызмет indie сегментінен тыс ЕО тұрғындарына жүйелі бағытталмаған;
threshold асып кетсе қайта қараймыз).

US California (CCPA): disclosure / deletion / opt-out сұраулары — сол
email.
