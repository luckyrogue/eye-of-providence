// Package attribution — Phase A of the README §4 attribution pipeline.
//
// Состояние:
//   - Phase A (NOW):  worker подписан на `events` table, маппит каждое
//                     событие на attribution category по плоским правилам
//                     (source/category полей, без contents). Заполняет
//                     `attribution_events`, на котором живёт donut chart
//                     дашборда.
//   - Phase A.5:      macOS clipboard tracker в Rust агенте; clipboard
//                     events корреллируются с IDE paste-burst'ами по времени.
//   - Phase B:        Browser ext → agent → worker provenance chain
//                     (sha256 clipboard hash + AI-origin flag) → точное
//                     различение pasted_ai vs pasted_other.
//   - Phase C:        Direct Copilot/Cursor accept-API hooks (вместо
//                     burst-detection) + agent-edit detection (Claude Code
//                     / Cursor agent file writes).
//
// Что НЕ делает Phase A:
//   - Per-hunk attribution (одно событие = одна attribution-запись).
//     README §4.2 требует per-hunk; в Phase B перейдём к этому через
//     diff snapshots в IDE extension.
//   - AI-origin для paste — все paste'ы пока классифицируются как
//     `pasted_other` или (если IDE отметил category='ai') `ai_inline`.
//
// Контракт:
//   `Worker.Run(ctx)` блокирует ctx.Done() и каждые `pollInterval` (default
//   60s) обрабатывает события за окно `[lastSeen, now - lagSecs]`. Lag
//   нужен чтобы избежать гонок с ingest'ом (events могут прийти out-of-order).
//
//   Последний обработанный timestamp хранится в Postgres `worker_state`
//   таблице (см. PG-миграцию ниже). При рестарте worker восстанавливает
//   позицию — exactly-once семантика в пределах CH idempotency окна.
package attribution

import (
	"context"
	"errors"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	// PollInterval по умолчанию — 60 сек. Меньше = свежее данные на дашборде,
	// но больше нагрузка на CH. 60s — компромисс: дашборд видит атрибуцию
	// через минуту после события.
	defaultPollInterval = 60 * time.Second

	// LagSecs — отступ от now() для query'я. Защита от race condition:
	// если ingest пишет событие с ts=now() и worker читает в тот же момент,
	// событие может уйти в gap между батчами. Lag окно даёт ingest'у время
	// зафиксировать запись.
	defaultLagSecs = 15

	// MaxBatchSize — лимит на одну итерацию. Если за окно набралось больше,
	// worker сделает несколько итераций. Защищает от OOM при backfill после
	// длительного простоя.
	defaultMaxBatchSize = 5000
)

// Worker — основной runner attribution-пайплайна.
type Worker struct {
	CH      driver.Conn
	PG      *pgxpool.Pool
	Logger  *zap.Logger
	Poll    time.Duration
	LagSecs int
	Batch   int
}

// New — конструктор с дефолтами; nil-logger заменяется на no-op.
func New(ch driver.Conn, pg *pgxpool.Pool, logger *zap.Logger) *Worker {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Worker{
		CH:      ch,
		PG:      pg,
		Logger:  logger,
		Poll:    defaultPollInterval,
		LagSecs: defaultLagSecs,
		Batch:   defaultMaxBatchSize,
	}
}

// Run — основной цикл. Возвращается на ctx.Done(). Любой error на одной
// итерации логируется и не прерывает цикл — следующая попытка через `Poll`.
func (w *Worker) Run(ctx context.Context) error {
	if err := ensureWorkerStateTable(ctx, w.PG); err != nil {
		return err
	}
	w.Logger.Info("attribution worker started",
		zap.Duration("poll", w.Poll),
		zap.Int("lag_secs", w.LagSecs),
		zap.Int("batch", w.Batch),
	)

	// Первый tick — немедленно после старта (без ожидания poll-интервала),
	// чтобы worker сразу подхватил всё накопленное pending данных.
	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			w.Logger.Info("attribution worker stopping", zap.Error(ctx.Err()))
			return nil
		case <-timer.C:
			start := time.Now()
			n, err := w.tick(ctx)
			elapsed := time.Since(start)
			if err != nil {
				w.Logger.Warn("attribution tick failed",
					zap.Error(err),
					zap.Duration("elapsed", elapsed),
				)
			} else if n > 0 {
				w.Logger.Info("attribution tick",
					zap.Int("events_processed", n),
					zap.Duration("elapsed", elapsed),
				)
			}
			timer.Reset(w.Poll)
		}
	}
}

// tick — одна итерация. Читает `last_processed_at`, выгружает события,
// маппит на attribution, batch-вставляет в attribution_events, обновляет
// position в PG. Возвращает количество обработанных событий.
func (w *Worker) tick(ctx context.Context) (int, error) {
	lastSeen, err := readPosition(ctx, w.PG)
	if err != nil {
		return 0, err
	}
	now := time.Now().UTC()
	until := now.Add(-time.Duration(w.LagSecs) * time.Second)
	if !until.After(lastSeen) {
		// Окно пустое — ещё не прошёл LagSecs от последней обработки.
		return 0, nil
	}

	events, err := w.fetchEventsWindow(ctx, lastSeen, until, w.Batch)
	if err != nil {
		return 0, err
	}
	if len(events) == 0 {
		// Сдвигаем position даже если событий нет — иначе worker будет
		// бесконечно перечитывать пустое окно при затяжных периодах
		// неактивности.
		if err := writePosition(ctx, w.PG, until); err != nil {
			return 0, err
		}
		return 0, nil
	}

	attributed := make([]attribRow, 0, len(events))
	for _, e := range events {
		ar := classify(e)
		attributed = append(attributed, ar)
	}
	if err := w.insertAttribution(ctx, attributed); err != nil {
		return 0, err
	}

	// Position = ts последнего обработанного события; следующий tick
	// возьмёт с этой границы (exclusive — `> lastSeen`).
	newPos := events[len(events)-1].TS
	if newPos.Before(until) {
		newPos = until
	}
	if err := writePosition(ctx, w.PG, newPos); err != nil {
		return 0, err
	}
	return len(events), nil
}

// rawEvent — внутреннее представление строки `events` для классификатора.
// Поля, не нужные attribution, не запрашиваем (экономим bandwidth CH).
type rawEvent struct {
	TS         time.Time
	UserID     uuid.UUID
	ProjectID  uuid.UUID
	FileLang   string
	Source     string // 'os' | 'browser' | 'ide' | 'cli'
	Category   string // 'idle' | 'manual' | 'ai' | 'reading' | 'refactor' | 'other'
	AIProvider string
	AIChannel  string
	DurationMS uint32
	CharsIn    uint32
	LinesAdded uint32
}

type attribRow struct {
	TS         time.Time
	UserID     uuid.UUID
	ProjectID  uuid.UUID
	FileLang   string
	Category   string // 'typed' | 'pasted_ai' | 'pasted_other' | 'ai_inline' | 'ai_agent' | 'refactor' | 'unknown'
	AIProvider string
	Lines      uint32
	Chars      uint32
	FocusMS    uint32
}

func (w *Worker) fetchEventsWindow(ctx context.Context, after, until time.Time, limit int) ([]rawEvent, error) {
	// `>` after / `<=` until — strict вход, inclusive выход. Один и тот же
	// ts на границе будет обработан ровно один раз.
	const q = `
		SELECT ts, user_id, project_id, file_lang, source, category,
		       ai_provider, ai_channel, duration_ms, chars_in, lines_added
		FROM events
		WHERE ts > ? AND ts <= ?
		ORDER BY ts ASC
		LIMIT ?
	`
	rows, err := w.CH.Query(ctx, q, after, until, uint64(limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]rawEvent, 0, limit/2)
	for rows.Next() {
		var e rawEvent
		if err := rows.Scan(
			&e.TS, &e.UserID, &e.ProjectID, &e.FileLang, &e.Source, &e.Category,
			&e.AIProvider, &e.AIChannel, &e.DurationMS, &e.CharsIn, &e.LinesAdded,
		); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// classify — Phase A naive классификатор. Маппит plain event на
// attribution category на основе уже-проставленных клиентом полей
// (source/category/ai_provider/ai_channel).
//
// Phase B/C заменят это на per-hunk матчинг с clipboard и Copilot API.
func classify(e rawEvent) attribRow {
	var cat string
	switch {
	// IDE inline-accept (Copilot/Cursor) — burst detection в extension
	// уже выставил category='ai' + ai_channel='inline'.
	case e.Source == "ide" && e.Category == "ai" && e.AIChannel == "inline":
		cat = "ai_inline"

	// CLI agent (Claude Code hooks, aider wrapper, etc).
	case e.Source == "cli" && e.AIProvider != "":
		cat = "ai_agent"

	// IDE manual = typed code. Refactor — отдельная категория (rename,
	// extract method, и т.п. от LSP, не AI).
	case e.Source == "ide" && e.Category == "manual":
		cat = "typed"
	case e.Source == "ide" && e.Category == "refactor":
		cat = "refactor"

	// Browser AI-сайт (ChatGPT / Claude / Gemini), просто фокус — не пишем
	// код, это reading/AI-assist в обсуждении.
	case e.Source == "browser":
		// Phase A: считаем browser-в-фокусе как «unknown» в attribution.
		// Phase B свяжет browser copy + clipboard hash + IDE paste по
		// hash chain → пометит pasted_ai.
		cat = "unknown"

	// OS focus events — idle / reading / прочее. В attribution не вписываются
	// (это контекст, не "написанный код"). Пропускаем.
	default:
		cat = "unknown"
	}
	return attribRow{
		TS:         e.TS,
		UserID:     e.UserID,
		ProjectID:  e.ProjectID,
		FileLang:   e.FileLang,
		Category:   cat,
		AIProvider: e.AIProvider,
		Lines:      e.LinesAdded,
		Chars:      e.CharsIn,
		FocusMS:    e.DurationMS,
	}
}

func (w *Worker) insertAttribution(ctx context.Context, rows []attribRow) error {
	if len(rows) == 0 {
		return nil
	}
	const q = `INSERT INTO attribution_events (ts, user_id, project_id, file_lang, category, ai_provider, lines, chars, focus_ms)`
	batch, err := w.CH.PrepareBatch(ctx, q)
	if err != nil {
		return err
	}
	for _, r := range rows {
		if err := batch.Append(
			r.TS, r.UserID, r.ProjectID, r.FileLang, r.Category, r.AIProvider,
			r.Lines, r.Chars, r.FocusMS,
		); err != nil {
			return err
		}
	}
	return batch.Send()
}

// --- Position persistence ---------------------------------------------------
//
// `worker_state` — простая key/value таблица в Postgres. Хранит timestamp
// последнего обработанного события на каждый воркер по имени.
//
// Зачем PG, а не CH: ClickHouse не любит частые UPDATE; PG row update —
// тривиальный pgx call с low latency.

const workerName = "attribution_v1"

func ensureWorkerStateTable(ctx context.Context, pg *pgxpool.Pool) error {
	if pg == nil {
		return errors.New("worker_state: pg pool is nil")
	}
	_, err := pg.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS worker_state (
			name TEXT PRIMARY KEY,
			last_processed_at TIMESTAMPTZ NOT NULL DEFAULT '1970-01-01T00:00:00Z'
		)
	`)
	return err
}

func readPosition(ctx context.Context, pg *pgxpool.Pool) (time.Time, error) {
	var t time.Time
	err := pg.QueryRow(ctx,
		`SELECT last_processed_at FROM worker_state WHERE name = $1`,
		workerName,
	).Scan(&t)
	if err != nil {
		// NOT EXISTS — первый запуск; стартуем с 24h ago, чтобы захватить
		// недавнюю активность без шторма миллионов записей при backfill.
		startAt := time.Now().Add(-24 * time.Hour).UTC()
		if _, werr := pg.Exec(ctx,
			`INSERT INTO worker_state (name, last_processed_at) VALUES ($1, $2) ON CONFLICT (name) DO NOTHING`,
			workerName, startAt,
		); werr != nil {
			return time.Time{}, werr
		}
		return startAt, nil
	}
	return t, nil
}

func writePosition(ctx context.Context, pg *pgxpool.Pool, t time.Time) error {
	_, err := pg.Exec(ctx,
		`INSERT INTO worker_state (name, last_processed_at) VALUES ($1, $2)
		 ON CONFLICT (name) DO UPDATE SET last_processed_at = EXCLUDED.last_processed_at`,
		workerName, t,
	)
	return err
}
