# Documentation

Human-facing docs for Eye of Providence (alpha v0.1.x). AI rules: [`AGENTS.md`](../AGENTS.md),
[`backend/AGENTS.md`](../backend/AGENTS.md) (includes DDD bounded contexts).

## Product & compliance

| Document | Purpose |
| --- | --- |
| [`alpha-install.md`](alpha-install.md) | Install agent, VS Code extension, browser extension (alpha participants) |
| [`privacy.md`](privacy.md) | GDPR, data handling |
| [`attribution.md`](attribution.md) | AI vs human signals (hooks, extensions) |
| [`data-model.md`](data-model.md) | Postgres + ClickHouse schemas, MVs, Redis cache, event flow |
| [`api/openapi.yaml`](api/openapi.yaml) | Partial OpenAPI (`Event`, auth, ingest, public reads) |

## Operations

| Document | Purpose |
| --- | --- |
| [`self-hosting.md`](self-hosting.md) | Docker Compose quickstart |
| [`deploy-dokploy.md`](deploy-dokploy.md) | Dokploy production deploy |
| [`disaster-recovery.md`](disaster-recovery.md) | Backup & restore runbook |
| [`threat-model.md`](threat-model.md) | STRIDE by component |

## Integrations

| Document | Purpose |
| --- | --- |
| [`integrations-pr-comment.md`](integrations-pr-comment.md) | PR comment bot API |
| [`agents-publishing.md`](agents-publishing.md) | Tauri/extension release process |
| [`ci-hardening.md`](ci-hardening.md) | Branch protection + production env approval (manual UI checklist) |
| [`tech-debt.md`](tech-debt.md) | Active tech-debt register, 13 clusters |

## Maintainers

Internal backlog and release checklists: [`internal/`](internal/).
