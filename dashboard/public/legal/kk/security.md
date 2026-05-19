# Қауіпсіздік — Eye of Providence

**Соңғы жаңарту:** 2026-05-19 · **Нұсқа:** v0.1 alpha

Бұл бет қауіпсіздікпен қалай жұмыс істейтінімізді сипаттайды: қазір
не бар, осалдық туралы қалай хабарлау керек, не күтуге болады. Ішкі
инженерлік қауіп моделі үшін
[`docs/threat-model.md`](https://github.com/luckyrogue/eye-of-providence/blob/main/docs/threat-model.md)
қараңыз; бұл бет — публичная сводка.

## Осалдық туралы хабарлау

**Көпшілікке арналған GitHub issue ашпаңыз.**
**`main@rysdavletov.org`** мекенжайына жазыңыз:

- Мәселе сипаттамасы және impact
- Қайталау қадамдары немесе PoC
- Әсер еткен нұсқа немесе commit SHA
- Follow-up үшін байланыс

Мақсат — **48 сағат ішінде растау** және **5 жұмыс күн** ішінде
жөндеу timeline'ын беру.

Scope ішінде: backend API, дашборд, агент (desktop / browser ext /
VS Code), Docker-image және CI-инфрақұрылым.

Scope сыртында: үшінші тұлғалар жүргізетін self-hosted инстанстар;
үшінші сервистер (ClickHouse Cloud, Resend, Dokploy); DoS-тестілеу
(rate-limit бар, бірақ ол exploitation surface емес).

## Қолдау көрсетілетін нұсқалар

Security-фикстер `main`-де шығады. Alpha-да LTS-тармақтарды қолдамаймыз.

| Нұсқа                | Қолдау |
| -------------------- | ------ |
| `main`               | ✅    |
| `v0.1.x-alpha.*`     | ✅ (тек latest) |
| pre-alpha rolling тэгтер | ❌ |

## Қазіргі security-posture

Бүгін жұмыс істеп тұрған нақты бақылау тетіктері:

### Backend және деректер

- **Auth:** bcrypt (cost 10), JWT HS256 `token_version` revocation-мен.
  Екінші фактор үшін WebAuthn / passkey қолдауы.
- **Rate limits:** auth-endpoints-та 10 req/min, `/v1/*`-те 120 req/min.
- **Пайдаланушы изоляциясы:** әр аналитикалық сұраныс JWT subject
  бойынша сүзіледі; cross-user қол жеткізу SQL деңгейінде мүмкін емес.
- **GDPR:** `GET /v1/me/export` барлық деректеріңізді JSON ретінде
  қайтарады; `DELETE /v1/me/data` оларды қайтымсыз өшіреді.

### Image және supply chain

- **Cosign signed:** `ghcr.io/luckyrogue/eop`-қа push'нутый әр Docker-
  image Sigstore keyless OIDC арқылы қол қойылған. Тексеру:
  ```bash
  cosign verify ghcr.io/luckyrogue/eop:<sha> \
    --certificate-identity-regexp '^https://github.com/luckyrogue/eye-of-providence/' \
    --certificate-oidc-issuer 'https://token.actions.githubusercontent.com'
  ```
- **SLSA Build L3 provenance:** image нақты commit-тен біздің CI
  тарапынан құрылғанының тексерілетін дәлелі.
- **CycloneDX SBOM:** әр image Software Bill of Materials-ті attestation
  ретінде жариялайды.

Толық recipes:
[`.github/SECURITY.md`](https://github.com/luckyrogue/eye-of-providence/blob/main/.github/SECURITY.md).

### Агент (desktop)

- **Шифрланған жергілікті буфер:** SQLite ішіндегі оқиғалар AES-256-GCM
  арқылы шифрланған. Кілт OS keyring-те сақталған (macOS Keychain,
  Windows Credential Manager, GNOME keyring).
- **Pairing-токендер** keyring-те, ешқашан plaintext-те емес.
- **Privacy инварианттары:** агент ешқашан файл мазмұнын, prompt'тарды,
  AI жауаптарын, raw keystrokes, clipboard мәтінін оқымайды. Тек
  санағыштар, хештер мен timestamps құрылғыдан шығады. Толық деректер
  картасы үшін [Privacy Notice](/privacy) §1 қараңыз.

### CI / әзірлеу

- Әр PR-да CodeQL static analysis (Go + JS/TS).
- Source және image-ке Trivy + OSV scans; PR-да dependency-review.
- gitleaks әр commit-ті кездейсоқ commit'нутый secret-терге сканерлейді.
- Step-Security `harden-runner` runner egress аудитін жасайды.

### Белгілі шектеулер (ашық мойындау)

- **Alpha-installer'лар қол қойылмаған.** Apple Developer ID және
  Windows EV cert әлі сатып алынбаған. Workaround үшін install guide
  ішіндегі [Why is this unsigned?](/docs/install#почему-installer-не-подписан)
  қараңыз. Image signing (Cosign) әсер етпеген — backend-image толық
  тексерілетін.
- **Branch protection** alpha-1 follow-up аясында
  [`docs/ci-hardening.md`](https://github.com/luckyrogue/eye-of-providence/blob/main/docs/ci-hardening.md)
  бойынша қосылуда. Қазірге дейін барлық merge CI-ден өтеді, бірақ
  required reviewers enforced емес.
- **PITR Postgres үшін** конфигурацияланбаған; RPO = 24 сағ (daily dump).
  Қатаңырақ RPO GA-ға қарай мақсат. Толық
  [`docs/disaster-recovery.md`](https://github.com/luckyrogue/eye-of-providence/blob/main/docs/disaster-recovery.md).

## Responsible disclosure

Bug bounty әзірге ұсынбаймыз. Зерттеушілерді (рұқсатпен) тиісті
CHANGELOG-та және фикс commit message-да атап өтеміз.

Көпшілікке ашудан бұрын ақылға қонымды терезе сұраймыз — әдетте
бастапқы acknowledgement-та келісетін timeline, default 90 күн.

## Бет жаңартулары

Security posture-ң маңызды өзгерістері әр релиздегі
[`CHANGELOG.md`](https://github.com/luckyrogue/eye-of-providence/blob/main/CHANGELOG.md)
**Security** қосалқы бөлімінде қадағаланады.
