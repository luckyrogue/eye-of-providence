package store

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/metrics"
)

type ClickHouseStore struct {
	conn driver.Conn
}

func OpenClickHouse(dsn string) (*ClickHouseStore, error) {
	opts, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}
	if opts.DialTimeout == 0 {
		opts.DialTimeout = 5 * time.Second
	}
	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := conn.Ping(pingCtx); err != nil {
		return nil, fmt.Errorf("clickhouse ping: %w", err)
	}
	return &ClickHouseStore{conn: conn}, nil
}

// Conn — escape hatch для тех, кому нужен low-level driver.Conn (например,
// attribution worker, использующий PrepareBatch + custom queries). Через
// EventStore-интерфейс эти операции не выразить, а тащить весь worker в
// этот package было бы overkill.
func (s *ClickHouseStore) Conn() driver.Conn {
	return s.conn
}

func (s *ClickHouseStore) Insert(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}
	defer metrics.ClickHouseWrite.ObserveSince(time.Now())
	batch, err := s.conn.PrepareBatch(ctx, `
		INSERT INTO events (
			ts, user_id, device_id, session_id, app_bundle, category, source,
			ai_provider, ai_channel, project_id, file_lang,
			duration_ms, chars_in, lines_added, lines_removed, meta
		)
	`)
	if err != nil {
		return err
	}
	for _, e := range events {
		userUUID, err := safeUUID(e.UserID)
		if err != nil {
			continue
		}
		deviceUUID, _ := safeUUID(e.DeviceID)
		sessionUUID, _ := safeUUID(e.SessionID)
		projectUUID, _ := safeUUID(e.ProjectID)
		if err := batch.Append(
			e.TS, userUUID, deviceUUID, sessionUUID,
			e.AppBundle, e.Category, e.Source,
			e.AIProvider, e.AIChannel, projectUUID, e.FileLang,
			e.DurationMS, e.CharsIn, e.LinesAdded, e.LinesRemoved, "",
		); err != nil {
			return err
		}
	}
	return batch.Send()
}

func (s *ClickHouseStore) ListRecent(ctx context.Context, userID string, limit int) ([]Event, error) {
	defer metrics.ClickHouseRead.ObserveSince(time.Now())
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.conn.Query(ctx, `
		SELECT ts, user_id, device_id, session_id, app_bundle, category, source,
		       ai_provider, ai_channel, project_id, file_lang,
		       duration_ms, chars_in, lines_added, lines_removed
		FROM events
		WHERE user_id = ?
		ORDER BY ts DESC
		LIMIT ?
	`, uid, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Event, 0, limit)
	for rows.Next() {
		var e Event
		var userU, deviceU, sessionU, projectU uuid.UUID
		if err := rows.Scan(
			&e.TS, &userU, &deviceU, &sessionU,
			&e.AppBundle, &e.Category, &e.Source,
			&e.AIProvider, &e.AIChannel, &projectU, &e.FileLang,
			&e.DurationMS, &e.CharsIn, &e.LinesAdded, &e.LinesRemoved,
		); err != nil {
			return nil, err
		}
		e.UserID = userU.String()
		if deviceU != uuid.Nil {
			e.DeviceID = deviceU.String()
		}
		if sessionU != uuid.Nil {
			e.SessionID = sessionU.String()
		}
		if projectU != uuid.Nil {
			e.ProjectID = projectU.String()
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

const dailyAggThreshold = 30 * 24 * time.Hour

func (s *ClickHouseStore) AggregateByCategory(ctx context.Context, userID string, since time.Time) (map[string]uint64, error) {
	defer metrics.ClickHouseRead.ObserveSince(time.Now())
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	var query string
	var arg time.Time
	if time.Since(since) >= dailyAggThreshold {
		query = `SELECT category, sum(duration_ms) FROM events_daily_agg
		         WHERE user_id = ? AND date >= ? GROUP BY category`
		arg = since.Truncate(24 * time.Hour)
	} else {
		query = `SELECT category, sum(duration_ms) FROM events_hourly_agg
		         WHERE user_id = ? AND bucket_ts >= ? GROUP BY category`
		arg = since
	}
	rows, err := s.conn.Query(ctx, query, uid, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]uint64{}
	for rows.Next() {
		var category string
		var sum uint64
		if err := rows.Scan(&category, &sum); err != nil {
			return nil, err
		}
		out[category] = sum
	}
	return out, rows.Err()
}

func (s *ClickHouseStore) AggregateByCategoryBulk(ctx context.Context, userIDs []string, since time.Time) (map[string]map[string]uint64, error) {
	defer metrics.ClickHouseRead.ObserveSince(time.Now())
	out := map[string]map[string]uint64{}
	if len(userIDs) == 0 {
		return out, nil
	}
	uids := make([]uuid.UUID, 0, len(userIDs))
	for _, s := range userIDs {
		u, err := uuid.Parse(s)
		if err != nil {
			continue
		}
		uids = append(uids, u)
	}
	if len(uids) == 0 {
		return out, nil
	}

	var query string
	var arg time.Time
	if time.Since(since) >= dailyAggThreshold {
		query = `SELECT user_id, category, sum(duration_ms) FROM events_daily_agg
		         WHERE user_id IN (?) AND date >= ? GROUP BY user_id, category`
		arg = since.Truncate(24 * time.Hour)
	} else {
		query = `SELECT user_id, category, sum(duration_ms) FROM events_hourly_agg
		         WHERE user_id IN (?) AND bucket_ts >= ? GROUP BY user_id, category`
		arg = since
	}
	rows, err := s.conn.Query(ctx, query, uids, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var uid uuid.UUID
		var category string
		var sum uint64
		if err := rows.Scan(&uid, &category, &sum); err != nil {
			return nil, err
		}
		key := uid.String()
		if out[key] == nil {
			out[key] = map[string]uint64{}
		}
		out[key][category] = sum
	}
	return out, rows.Err()
}

func (s *ClickHouseStore) Heatmap(ctx context.Context, userID string, since time.Time, tz string) ([]HeatmapCell, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	if tz == "" {
		tz = "UTC"
	}
	rows, err := s.conn.Query(ctx, `
		SELECT toDayOfWeek(toTimeZone(bucket_ts, ?)) AS dow,
		       toHour(toTimeZone(bucket_ts, ?)) AS hour,
		       category,
		       sum(duration_ms) AS ms
		FROM events_hourly_agg
		WHERE user_id = ? AND bucket_ts >= ?
		GROUP BY dow, hour, category
	`, tz, tz, uid, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []HeatmapCell{}
	for rows.Next() {
		var dow, hour uint8
		var cell HeatmapCell
		if err := rows.Scan(&dow, &hour, &cell.Category, &cell.MS); err != nil {
			return nil, err
		}

		cell.DayOfWeek = int(dow % 7)
		cell.Hour = int(hour)
		out = append(out, cell)
	}
	return out, rows.Err()
}

func (s *ClickHouseStore) LanguageBreakdown(ctx context.Context, userID string, since time.Time) ([]LangCell, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.conn.Query(ctx, `
		SELECT file_lang, category, sum(chars_in), sum(duration_ms)
		FROM events_hourly_agg
		WHERE user_id = ? AND bucket_ts >= ? AND file_lang != ''
		GROUP BY file_lang, category
	`, uid, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []LangCell{}
	for rows.Next() {
		var c LangCell
		if err := rows.Scan(&c.Lang, &c.Category, &c.Chars, &c.MS); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *ClickHouseStore) ActiveUserIDs(ctx context.Context, since time.Time) ([]string, error) {
	rows, err := s.conn.Query(ctx, `
		SELECT DISTINCT user_id FROM events_hourly_agg WHERE bucket_ts >= ?
	`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []string{}
	for rows.Next() {
		var u uuid.UUID
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		out = append(out, u.String())
	}
	return out, rows.Err()
}

func (s *ClickHouseStore) DailyTrend(ctx context.Context, userID string, since time.Time, tz string) ([]TrendPoint, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	if tz == "" {
		tz = "UTC"
	}
	rows, err := s.conn.Query(ctx, `
		SELECT toDate(toTimeZone(bucket_ts, ?)) AS d, category, sum(chars_in), sum(duration_ms)
		FROM events_hourly_agg
		WHERE user_id = ? AND bucket_ts >= ?
		GROUP BY d, category
		ORDER BY d, category
	`, tz, uid, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []TrendPoint{}
	for rows.Next() {
		var day time.Time
		var p TrendPoint
		if err := rows.Scan(&day, &p.Category, &p.Chars, &p.MS); err != nil {
			return nil, err
		}
		p.Date = day.Format("2006-01-02")
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *ClickHouseStore) Close() error {
	return s.conn.Close()
}

func (s *ClickHouseStore) Ping(ctx context.Context) error {
	return s.conn.Ping(ctx)
}

func safeUUID(s string) (uuid.UUID, error) {
	if s == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(s)
}
