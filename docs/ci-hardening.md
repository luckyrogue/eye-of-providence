# CI hardening runbook

This file lists the **GitHub UI actions** needed to close cluster C3 from
[`docs/tech-debt.md`](tech-debt.md). Code-side prerequisites (workflow envs,
secrets documentation) were committed separately.

Once you complete each section, return here and check it off. The goal is to
have all three sections done before promoting alpha → beta.

---

## 1. Branch protection on `main`

**Why:** today (alpha) anyone with push access can force-push or merge
without review. For payment-grade hygiene we want PR-only flow with required
checks and signed commits.

**Where:** GitHub → repo → **Settings → Branches → Branch protection rules
→ Add rule** (or edit the existing rule for `main`).

Match pattern: `main`.

Tick:

- [ ] **Require a pull request before merging**
  - Required approvals: `1`
  - Dismiss stale pull request approvals when new commits are pushed: ✅
  - Require review from Code Owners: ✅ (CODEOWNERS already configured for
    security-critical paths)
- [ ] **Require status checks to pass before merging**
  - Require branches to be up to date before merging: ✅
  - Pick (after at least one PR has run them so they appear in the list):
    - `CI / backend / backend / lint`
    - `CI / backend / backend / unit`
    - `CI / backend / backend / integration`
    - `CI / agent-rust / agent-rust / cargo check`
    - `CI / security / security / codeql (go, autobuild)`
    - `CI / security / security / codeql (javascript-typescript, none)`
    - `CI / security / security / gitleaks`
    - `CI / security / security / trivy-fs`
    - `CI / security / security / osv-scanner / osv-scan`
    - `CI / docker / docker / build + scan + sign`
    - `CI / e2e / e2e / playwright`
- [ ] **Require conversation resolution before merging**
- [ ] **Require signed commits**
- [ ] **Require linear history**
- [ ] **Do not allow bypassing the above settings** → also check
      "Include administrators" so you can't accidentally self-bypass.
- [ ] **Allow force pushes:** OFF
- [ ] **Allow deletions:** OFF

Save.

**Verification:** open a test branch, push directly to `main` from CLI →
GitHub should reject with "protected branch hook declined". A PR with red
CI should NOT show the "Merge" button as enabled.

---

## 2. Production environment approval gate

**Why:** typo in `DOKPLOY_WEBHOOK` or accidental push to `main` would deploy
to prod without human review. We want a 5-minute human approval window with
a named reviewer.

**Where:** Settings → **Environments → `production` → Configure**
(create if missing; the workflow `_deploy.yml:21` already references it).

- [ ] **Required reviewers:** add `@luckyrogue` (you). Up to 6 total.
- [ ] **Wait timer:** `5` minutes
- [ ] **Deployment branches and tags:** Selected branches → only `main`

Save.

**Verification:** push a no-op commit to `main`. The CI workflow's `deploy /
dokploy` job should appear with state "Waiting" and "Review pending review".
Approve → job runs; reject → job cancelled.

---

## 3. Secrets that must exist (re-check)

Open **Settings → Secrets and variables → Actions → Repository secrets**.
Confirm each of these exists (values not visible; just the name):

- [ ] `DOKPLOY_WEBHOOK` — URL pattern documented in `.env.example` block
      "Deploy". Required for `_deploy.yml` to fire.
- [ ] (Optional, future) `APPLE_CERTIFICATE`, `APPLE_CERTIFICATE_PASSWORD`,
      `APPLE_SIGNING_IDENTITY`, `APPLE_ID`, `APPLE_PASSWORD`, `APPLE_TEAM_ID`
      — when you buy Apple Developer cert, fill these and `release.yml` will
      pick them up automatically (signed `.dmg` instead of unsigned).
- [ ] (Optional, future) `WINDOWS_CERTIFICATE`, `WINDOWS_CERTIFICATE_PASSWORD`
      — when you buy a Windows EV cert, same story for `.msi` / `.exe`.

Repository variables (Settings → Variables):

- [ ] `HEALTH_URL` (optional) — referenced in `_deploy.yml:46` for post-deploy
      smoke. Default: dashboard public URL `/healthz`.

---

## 4. After all three sections are done

Update [`docs/tech-debt.md`](tech-debt.md) cluster C3 row:
strike-through the items and add the commit ref of this checklist
completion (or a dated note like `✅ done 2026-MM-DD by @luckyrogue`).

Then promote C3 from S0/S1 to "✅ resolved" in the cluster summary table.

---

## Future hardening (post-Sprint 1)

These are tracked in `docs/tech-debt.md` under other clusters; do NOT block
the above on them:

- harden-runner `egress-policy: block` mode with explicit allow-list for
  GHCR, npm registry, GitHub API, raw.githubusercontent.com, etc.
- conventional-commits PR check (`commitlint` action) — currently
  CONTRIBUTING.md asks contributors to follow the convention; not enforced.
- canary deploy step before main rollout (`:staging` tag, 5min soak,
  promote to `:latest`).
