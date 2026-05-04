package store

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
)

type ClickHouseStore struct {
	conn driver.Conn
}

// OpenClickHouse — DSN формата clickhouse://user:pass@host:port/db.
func OpenClickHouse(dsn string) (*ClickHouseStore, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}
	pass, _ := u.User.Password()
	db := strings.TrimPrefix(u.Path, "/")
	if db == "" {
		db = "default"
	}
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{u.Host},
		Auth: clickhouse.Auth{
			Database: db,
			Username: u.User.Username(),
			Password: pass,
		},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("clickhouse ping: %w", err)
	}
	return &ClickHouseStore{conn: conn}, nil
}

func (s *ClickHouseStore) Insert(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}
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

func (s *ClickHouseStore) AggregateByCategory(ctx context.Context, userID string, since time.Time) (map[string]uint64, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.conn.Query(ctx, `
		SELECT category, sum(duration_ms)
		FROM events
		WHERE user_id = ? AND ts >= ?
		GROUP BY category
	`, uid, since)
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

func (s *ClickHouseStore) Heatmap(ctx context.Context, userID string, since time.Time) ([]HeatmapCell, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.conn.Query(ctx, `
		SELECT toDayOfWeek(ts) AS dow,
		       toHour(ts) AS hour,
		       category,
		       sum(duration_ms) AS ms
		FROM events
		WHERE user_id = ? AND ts >= ?
		GROUP BY dow, hour, category
	`, uid, since)
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
		// ClickHouse toDayOfWeek: 1=Monday … 7=Sunday → конвертим в 0=Sunday … 6=Saturday
		cell.DayOfWeek = int(dow % 7)
		cell.Hour = int(hour)
		out = append(out, cell)
	}
	return out, rows.Err()
}

func (s *ClickHouseStore) Close() error {
	return s.conn.Close()
}

func safeUUID(s string) (uuid.UUID, error) {
	if s == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(s)
}
