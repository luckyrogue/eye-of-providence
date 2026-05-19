# Changelog

## 0.0.1 — Alpha

- Manual / AI inline / refactor attribution per language with burst detection
  for streaming completions.
- Pairing-code authentication via dashboard. Tokens stored in
  VS Code SecretStorage.
- Status bar item with `idle / sending / auth-required / paused` states.
- Commands: `eop.pair`, `eop.logout`, `eop.flush`, `eop.openDashboard`,
  `eop.showLog`.
- Auto-migration from legacy `eop.token` setting into SecretStorage.
