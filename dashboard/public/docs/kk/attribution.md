# Code attribution v2

Мақсат: hunks/strokes бойынша «AI vs human» классификациясында 90%+ дәлдік. Ағымдағы
v1 алгоритмі (paste size + AI domains) ~70% береді. v2 3 сигнал
көзін қосады:

1. **VS Code/Cursor inline accept** — heuristic burst-detection
2. **Claude Code agent edits** — [Hooks API](https://docs.claude.com/en/docs/claude-code/hooks) арқылы
3. **Browser AI domains** — claude.ai, chatgpt.com, perplexity, т.б. (v1, өзгеріссіз)

## Категориялар және пайдалы атрибуция

Екі категория қабаты:

- **Raw `events.category`** (ingest): `idle`, `manual`, `ai`, `reading`, `refactor`, `other`.
- **Derived `attribution_events.category`** (worker): `typed`, `pasted_ai`, `pasted_other`,
  `ai_inline`, `ai_agent`, `refactor`, `unknown` — қараңыз
  [`002_attribution_events.up.sql`](../backend/internal/migrate/sql/clickhouse/002_attribution_events.up.sql).

Төмендегі кесте **raw** events-ті сигналдарға маппингтейді (worker Phase B алдында):

| raw `category` | `ai_provider` | `ai_channel` | Сигнал |
|---|---|---|---|
| `manual` | — | — | typing / AI белгісіз paste |
| `ai` | `copilot` | `inline` | VS Code burst немесе single paste >= threshold |
| `ai` | `cursor` | `inline` | Cursor (`vscode.env.appName === "Cursor"`) + сол heuristic |
| `ai` | `claude-code` | `agent` | PostToolUse `Edit\|Write\|MultiEdit` үстінде `eop-hook` |
| `ai` | `openai` / `anthropic` | `chat` | browser-ext: chatgpt.com / claude.ai |
| `refactor` | — | — | IDE-дағы құрылымдық өзгеріс |
| `reading` | — | — | edit жоқ focus |

## VS Code / Cursor extension

Файл: `ide-vscode/src/extension.ts`. Қос heuristic:

1. **Single big insert** (бір `contentChange` ішінде `inserted >= 80 chars`) →
   `ai_inline`. Классикалық «Copilot suggestion accept» немесе chat-тан paste.
2. **Burst** (<100ms ішінде бірнеше `contentChange`, жиыны >= 80) →
   де `ai_inline`. Copilot/Cursor inline streaming 1–3 char токенмен жазады;
   қалыпты typing — keystroke арасында >150ms.

`ai_provider` `vscode.env.appName` арқылы:
- `"Cursor"` → `cursor`
- әйтпесе → `copilot` (VS Code inline үшін default)

`app_bundle` IDE-ға байланысты: `com.todesktop.230313mzl4w4u92` (Cursor) немесе
`com.microsoft.VSCode`.

### Белгілі шектеулер

- Шынайы Copilot Accept event `proposed API`
  (`commands.onWillExecuteCommand`) қажет — stable VS Code-та әзірше жоқ.
  Burst-heuristic 85%+ жағдайды қамтиды, қалғаны paste ретінде.
- Cursor `aichat.acceptDiff` public API арқылы intercept етілмейді — сол heuristic.
- Copilot өте баяу stream етсе (>100ms char арасында), burst
  detection өткізіп typing деп атрибуциялауы мүмкін. `BURST_WINDOW_MS` tune етіңіз.

## Claude Code hook

Ең күшті сигнал: Claude Code әр
`Edit`/`Write`/`MultiEdit`-те hook триггерлейді — нөлдік ambiguity.

### Орнату

1. Hook binary құрастыру:

   ```sh
   cd backend && go build -o ~/.local/bin/eop-hook ./cmd/eop-hook
   ```

   (немесе `go install ./cmd/eop-hook` → `$GOBIN`).

2. EOP token алу (тек **local dev**; production-да `dev-token` өшірілген):

   ```sh
   # backend EOP_ENABLE_DEV_TOKEN=true (development default)
   curl -X POST http://localhost:8080/v1/auth/dev-token | jq -r .token
   ```

   `~/.zshrc` / `~/.bashrc` сақтау:

   ```sh
   export EOP_TOKEN="<token>"
   export EOP_BACKEND="http://localhost:8080"   # немесе self-host URL
   ```

   Production-да dashboard API token (`eop_<…>`) немесе device pairing.

3. Claude Code settings-те hook қосу (`~/.claude/settings.json` user-wide
   немесе репода `.claude/settings.json` project-only):

   ```json
   {
     "hooks": {
       "PostToolUse": [{
         "matcher": "Edit|Write|MultiEdit",
         "hooks": [{ "type": "command", "command": "eop-hook" }]
       }]
     }
   }
   ```

4. Claude Code session қайта іске қосу. Әр Edit/Write event жібереді:
   `category=ai`, `ai_provider=claude-code`, `ai_channel=agent`.

### Не есептеледі

- `Write` — `content` толық ұзындығы `chars_in`, `\n` → `lines_added`.
- `Edit` — `new_string` ұзындығы chars, `\n` in `new_string` → `lines_added`,
  `\n` in `old_string` → `lines_removed`.
- `MultiEdit` — барлық edits жиынтығы.
- `Read` / `Bash` / `Grep` — елемейді (attribution мағынасы жоқ).

### Privacy

Hook тек counts (chars/lines/lang) жібереді, **файл мазмұнын емес**. Ingest валидациясы —
`domain.ValidEvent()` —
[`backend/internal/ingest/domain/event.go`](../backend/internal/ingest/domain/event.go);
wire-format-та мәтін өрісі жоқ.

### Failure-mode

`EOP_TOKEN` жоқ болса — hook үнсіз шығады (exit 0). Network error —
stderr, бірақ Claude Code workflow блокталмайды. Backend қолжетімсіз кезде tool циклін
бұзбау үшін by design.

## Browser extension

Файл: `browser-extension/src/entrypoints/background.ts`. v2-де өзгеріссіз — AI домендерінде
focus бақылайды, `category=ai`, `ai_provider` mapping
`ai-domains.ts`, `ai_channel=chat` (inline емес).

## Tauri agent

Файл: `agent/src-tauri/src/core/`. Native қолданбаларда focus events
(VS Code, Cursor, Claude Desktop). AI attribution үшін parallel VS Code extension /
browser extension-ға сүйенеді. v2 өзгермейді.

## Roadmap

- Cursor **Apply**: `eop.captureCursorApply` wrapper,
  `cmd+enter` override → Cursor apply + explicit `ai` event.
- **Copilot Accept**: stable `commands.onWillExecuteCommand` немесе
  `vscode.copilot` API. Қазіргі stable-та burst-heuristic — ең жақсы.
- **JetBrains plugin**: IntelliJ отбасы үшін VS Code аналогы — кейінге қалдырылған.
