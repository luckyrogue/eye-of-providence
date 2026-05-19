# Безопасность — Eye of Providence

**Последнее обновление:** 2026-05-19 · **Версия:** v0.1 alpha

Эта страница описывает как мы работаем с безопасностью: что есть сейчас,
как сообщить об уязвимости, чего ожидать в ответ. Для внутренней
инженерной модели угроз см.
[`docs/threat-model.md`](https://github.com/luckyrogue/eye-of-providence/blob/main/docs/threat-model.md);
эта страница — публичная сводка.

## Сообщить об уязвимости

**Пожалуйста, не открывайте публичный GitHub issue.** Напишите на
**`main@rysdavletov.org`**:

- Описание проблемы и её impact
- Шаги воспроизведения или PoC
- Затронутая версия или commit SHA
- Ваш контакт для follow-up

Цель — **подтвердить получение в течение 48 часов** и предоставить
timeline исправления в течение **5 рабочих дней**.

В scope: backend API, дашборд, агент (desktop / browser ext /
VS Code), Docker-образ и CI-инфраструктура.

Вне scope: self-hosted инстансы, запущенные третьими лицами;
третьи сервисы (ClickHouse Cloud, Resend, Dokploy); DoS-тестирование
(у нас есть rate-limit, но это не surface для эксплуатации).

## Поддерживаемые версии

Security-фиксы выходят в `main`. В alpha мы не поддерживаем LTS-ветки.

| Версия               | Поддерживается |
| -------------------- | -------------- |
| `main`               | ✅            |
| `v0.1.x-alpha.*`     | ✅ (только latest) |
| pre-alpha rolling тэги | ❌          |

## Текущий security-posture

Конкретные контроли в работе сегодня:

### Backend и данные

- **Auth:** bcrypt (cost 10), JWT HS256 с `token_version` revocation.
  Поддержка WebAuthn / passkey для второго фактора.
- **Rate limits:** 10 req/min на auth-endpoints, 120 req/min на `/v1/*`.
- **Изоляция по пользователю:** каждый аналитический запрос фильтруется
  по JWT subject; cross-user доступ невозможен на уровне SQL.
- **GDPR:** `GET /v1/me/export` возвращает все ваши данные как JSON;
  `DELETE /v1/me/data` безвозвратно их стирает.

### Образ и supply chain

- **Cosign signed:** каждый Docker-образ, push'нутый в
  `ghcr.io/luckyrogue/eop`, подписан через Sigstore keyless OIDC.
  Verify:
  ```bash
  cosign verify ghcr.io/luckyrogue/eop:<sha> \
    --certificate-identity-regexp '^https://github.com/luckyrogue/eye-of-providence/' \
    --certificate-oidc-issuer 'https://token.actions.githubusercontent.com'
  ```
- **SLSA Build L3 provenance:** верифицируемое доказательство что образ
  собран нашим CI из конкретного commit'а. Verify через `gh attestation
  verify oci://ghcr.io/luckyrogue/eop:<sha> --owner luckyrogue
  --predicate-type 'https://slsa.dev/provenance/v1'`.
- **CycloneDX SBOM:** каждый образ публикует Software Bill of Materials
  как attestation. Fetch через `gh attestation download`.

Полные recipes:
[`.github/SECURITY.md`](https://github.com/luckyrogue/eye-of-providence/blob/main/.github/SECURITY.md).

### Агент (desktop)

- **Зашифрованный локальный буфер:** события в SQLite зашифрованы
  AES-256-GCM. Ключ хранится в OS keyring (macOS Keychain, Windows
  Credential Manager, GNOME keyring).
- **Pairing-токены** в keyring, никогда в plaintext.
- **Privacy-инварианты:** агент никогда не читает содержимое файлов,
  промпты, ответы AI, raw keystrokes, текст clipboard'а. Только счётчики,
  хеши и timestamps уходят с устройства. См.
  [Privacy Notice](/privacy) §1 для полной карты данных.

### CI / разработка

- CodeQL static analysis (Go + JS/TS) на каждом PR.
- Trivy + OSV scans по source и образу; dependency-review на PR.
- gitleaks сканирует каждый commit на случайно закоммиченные секреты.
- Step-Security `harden-runner` аудитит egress runner'ов.

### Известные ограничения (открыто признаём)

- **Alpha-installers не подписаны.** Apple Developer ID и Windows EV
  cert ещё не куплены. См.
  [Why is this unsigned?](/docs/install#почему-installer-не-подписан)
  в install guide для workaround'ов. Image signing (Cosign) не затронут
  — backend-образ полностью верифицируем.
- **Branch protection** включается по
  [`docs/ci-hardening.md`](https://github.com/luckyrogue/eye-of-providence/blob/main/docs/ci-hardening.md)
  в рамках alpha-1 follow-up. Пока все merge проходят через CI, но
  required reviewers не enforced.
- **PITR для Postgres** не настроен; RPO = 24 ч (daily dump).
  Жёстче RPO целимся к GA. Детали в
  [`docs/disaster-recovery.md`](https://github.com/luckyrogue/eye-of-providence/blob/main/docs/disaster-recovery.md).

## Responsible disclosure

Bug bounty пока не предлагаем. Зачисляем исследователей (с разрешения)
в соответствующий CHANGELOG entry и commit message фикса.

Просим разумное окно — обычно timeline, который мы согласуем в
первичном acknowledgement, default 90 дней — до публичного раскрытия.

## Обновления страницы

Существенные изменения security posture отслеживаются в
[`CHANGELOG.md`](https://github.com/luckyrogue/eye-of-providence/blob/main/CHANGELOG.md)
под подсекцией **Security** каждого релиза.
