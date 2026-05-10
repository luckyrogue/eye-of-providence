# PR-comment integration (GitHub / GitLab)

Постит markdown-комментарий с AI-attribution breakdown в PR (GitHub) или MR
(GitLab) после merge.

## API

```
POST /v1/integrations/pr-comment
Authorization: Bearer <eop_token>            # JWT либо API token (admin scope)
Content-Type: application/json

{
  "provider":       "github" | "gitlab",
  "host":           "https://api.github.com",  // optional. Default for GH:
                                               //   api.github.com
                                               // Default for GL: gitlab.com
                                               // Self-hosted: твой URL
  "repo":           "luckyrogue/eye-of-providence",
  "pr_number":      42,                        // GitHub PR / GitLab MR IID
  "shas":           ["abc1234", "def5678"],
  "provider_token": "ghp_..."                  // PAT с repo-scope (GH) или
                                               // api-scope (GL)
}
```

Response 200:

```json
{
  "posted": true,
  "aggregate": {
    "total_commits": 5,
    "with_attribution": 4,
    "lines_added": 320,
    "lines_removed": 47,
    "ai_percent": 73.4
  },
  "comment_md": "### :eye: Eye of Providence ..."
}
```

Response 502: `{error: "provider rejected", provider_status: 403, provider_body: "..."}`
если GH/GL отвергли наш request (bad token, repo not found, etc.).

## GitHub Actions example

`.github/workflows/eop-pr-comment.yml`:

```yaml
name: EoP attribution comment
on:
  pull_request:
    types: [closed]

jobs:
  comment:
    if: github.event.pull_request.merged == true
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 0 }
      - name: Get commit SHAs
        id: shas
        run: |
          shas=$(git log --format=%H ${{ github.event.pull_request.base.sha }}..${{ github.event.pull_request.head.sha }} | jq -R . | jq -s .)
          echo "list=$shas" >> $GITHUB_OUTPUT
      - name: Post EoP comment
        run: |
          curl -fsS -X POST "$EOP_API/v1/integrations/pr-comment" \
            -H "Authorization: Bearer $EOP_TOKEN" \
            -H "Content-Type: application/json" \
            -d "$(jq -n --argjson shas '${{ steps.shas.outputs.list }}' '{
              provider: "github",
              repo: "${{ github.repository }}",
              pr_number: ${{ github.event.pull_request.number }},
              shas: $shas,
              provider_token: "${{ secrets.GITHUB_TOKEN }}"
            }')"
        env:
          EOP_API: https://eop.rysdavletov.org/api
          EOP_TOKEN: ${{ secrets.EOP_TOKEN }}
```

`secrets.GITHUB_TOKEN` уже инжектится Actions runtime — отдельный PAT не
нужен. `secrets.EOP_TOKEN` — `eop_<token>` со scope `admin` или `read`,
создаётся в settings → API tokens.

## GitLab CI example

`.gitlab-ci.yml`:

```yaml
eop-mr-comment:
  stage: post
  rules:
    - if: $CI_MERGE_REQUEST_IID && $CI_MERGE_REQUEST_EVENT_TYPE == "merge_train"
  script:
    - |
      shas=$(git log --format=%H "$CI_MERGE_REQUEST_DIFF_BASE_SHA..$CI_COMMIT_SHA" | jq -R . | jq -s -c .)
      curl -fsS -X POST "$EOP_API/v1/integrations/pr-comment" \
        -H "Authorization: Bearer $EOP_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$(jq -nc --argjson shas "$shas" '{
          provider: "gitlab",
          host: "'"$CI_SERVER_URL"'",
          repo: "'"$CI_PROJECT_PATH"'",
          pr_number: '"$CI_MERGE_REQUEST_IID"',
          shas: $shas,
          provider_token: "'"$EOP_GITLAB_TOKEN"'"
        }')"
```

`$EOP_GITLAB_TOKEN` — GitLab PAT со scope `api` (для notes write).
Self-hosted GitLab — `$CI_SERVER_URL` уже подставит правильный host.

## Comment format

Шапка:

```markdown
### :eye: Eye of Providence — coding attribution

**AI-assisted: 73%** `███████░░░`

| Metric | Value |
|---|---|
| Commits | 5 (with attribution: 4) |
| Lines added | +320 |
| Lines removed | -47 |

[View team breakdown →](https://eop.rysdavletov.org/team)
```

Если у commits нет AI-attribution (агент не установлен), показываем graceful
copy с install-link'ами вместо percentage.

## Privacy

Comment не содержит:
- Содержимое diff'а
- Имена файлов
- Author email'ы

Только counts (commits, lines added/removed) + aggregate AI percentage.

## Security

- Provider token не сохраняется на нашей стороне — request → forward →
  discard. JWT/API token логируется только prefix.
- `repo` валидируется на формат `owner/name`; `pr_number` ≥1; `shas` max 500
  per request (защита от amplification attack).
- HTTP errors от GH/GL пробрасываются как 502 с `provider_status` +
  truncated `provider_body` (≤200 chars) — для UI debugging без leak'а.

## Roadmap

- **Auto-trigger**: receive GH/GL webhook `pull_request.closed` →
  автоматически постить (без CI шага). Требует hosted webhook receiver.
- **Per-team integration storage**: сохранить provider_token (encrypted) и
  repo, чтобы curl-команда упростилась до `?team=<id>` query param.
- **GitLab discussion threads**: open thread "AI attribution" привязанный к
  diff lines с разбивкой по файлам (требует diff data).
